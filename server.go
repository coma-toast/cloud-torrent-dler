package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"

	"net/http"
	"sort"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/db"
	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/seedr"
	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/showrss"
	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/yts"
	"github.com/gorilla/mux"
)

type MagnetApi struct {
	selectedSeedr SeedrInstance
	mutex         sync.RWMutex
}

type MainPageData struct {
	Movies   []yts.Movie
	Shows    []showrss.Item
	ShowList []showrss.Item
}

type ShowPageData struct {
	Show  string
	Items []showrss.Item
}

// RunMagnetApi is the api for adding magnet urls
func (magnetApi *MagnetApi) RunMagnetApi() {
	r := mux.NewRouter()
	r.HandleFunc("/gui", magnetApi.GuiHandler)
	r.HandleFunc("/api/ping", magnetApi.PingHandler)
	r.HandleFunc("/api/item", magnetApi.AddItemMagnet).Methods(http.MethodPost)
	r.HandleFunc("/api/magnet", magnetApi.AddMagnetHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/torrent", magnetApi.AddTorrentHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/show/{showID}", magnetApi.ShowHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/data/show", magnetApi.DataShowHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/search", magnetApi.SearchHandler).Methods(http.MethodGet)
	// r.HandleFunc("/api/data/show/{showID}", magnetApi.DataShowByIDHandler).Methods("GET")
	// r.HandleFunc("/api/data/show/{showID}", magnetApi.DataShowByIDHandler).Methods("POST")
	log.Info(fmt.Sprintf("Magnet API running. Send JSON {link: url} as a POST request to x.x.x.x:%s/api/magnet to add directly to Seedr!", conf.Port))

	// * this needs to be fixed when i get back to the UI
	// Serve static files
	staticDir := "./static"
	fs := http.FileServer(http.Dir(staticDir))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// r.Use(APILoggingMiddleware)
	log.Error(http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), r))
}

func (magnetApi *MagnetApi) SearchHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	params := r.URL.Query()
	search := params.Get("search")
	page, err := strconv.Atoi(params.Get("page"))
	if err != nil {
		log.Error(err)
	}

	result, err := yts.SearchMovies(search, page)
	if err != nil {
		log.Error(err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Error(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}

func (magnetApi *MagnetApi) AddItemMagnet(w http.ResponseWriter, r *http.Request) {
	var data db.DownloadItem

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Error decoding JSON")
	} else {
		log.WithField("item", data.Name).Info("Adding Magnet/Torrent to MQ")
		err = database.AddSeedrUpload(&data)
		if err != nil {
			log.WithError(err)
		}
	}
}

// AddMagnetHandler handles api calls adding magnets
func (magnetApi *MagnetApi) AddMagnetHandler(w http.ResponseWriter, r *http.Request) {
	var data ApiMagnet
	var result seedr.Result

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Error decoding JSON")
	} else {
		log.WithField("link", data.Link).Info("Adding Magnet/Torrent to Seedr")
		result, err = magnetApi.AddRawMagnet(data.Link)
		if err != nil {
			log.WithError(err)
		}

	}
	resultData, err := json.Marshal(result)
	if err != nil {
		log.WithError(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultData)
}

// AddTorrentHandler handles api calls adding magnets
func (magnetApi *MagnetApi) AddTorrentHandler(w http.ResponseWriter, r *http.Request) {
	var data ApiMagnet
	var result seedr.Result

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Error decoding JSON")
	} else {
		log.WithField("link", data.Link).Info("Adding Magnet/Torrent to Seedr")
		result, err = magnetApi.AddRawTorrent(data.Link)
		if err != nil {
			log.WithError(err)
		}
		cache.SetAutoDownload(result.Torrent_hash, data.AutoDownload)

	}
	resultData, err := json.Marshal(result)
	if err != nil {
		log.WithError(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultData)
}

// ShowHandler handles api calls to pull a show from ShowRSS
func (magnetApi *MagnetApi) ShowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	showID := vars["showID"]
	var result showrss.Shows
	var err error

	log.WithField("link", showID).Info("Getting all ShowRSS data")
	result, err = magnetApi.GetFullShow(showID)
	if err != nil {
		log.WithError(err)
	}

	templateShow := template.Must(template.ParseFiles(conf.CachePath + "/templates/show.html"))

	data := ShowPageData{
		Show:  result.Title,
		Items: result.Item,
	}

	templateShow.Execute(w, data)

	// w.WriteHeader(http.StatusOK)
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(resultData)
}

func (magnetApi *MagnetApi) DataShowHandler(w http.ResponseWriter, r *http.Request) {
	var result map[string]db.DownloadItem
	var err error

	result = cache.GetAll()

	resultData, err := json.Marshal(result)
	if err != nil {
		log.WithError(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultData)
}

// Load the web front end
func (magnetApi *MagnetApi) GuiHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bytes, err := io.ReadAll(r.Body)
	// * dev code
	_ = bytes
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	templateMain := template.Must(template.ParseFiles(conf.CachePath + "/templates/main.html"))
	movieData, err := yts.GetMovies("https://yts.mx/api/v2/list_movies.json?quality=2160p")
	if err != nil {
		log.WithField("error", err).Warn("Error getting Movie Data from YTS")
	}

	showData, err := showrss.GetAllEpisodeItems(conf.ShowRSS)
	if err != nil {
		log.WithField("error", err).Warn("Error getting Show Data from ShowRSS")
	}

	showList := []showrss.Item{}
	// Get list of "all" shows (only the latest X number of shows)
	showsListAllData, err := showrss.GetShows("https://showrss.info/other/all.rss")
	if err != nil {
		log.WithField("error", err).Warn("Error getting Show List from ShowRSS")
	}

	// Get a list of all subscribed shows
	showsListSubscribedData, err := showrss.GetShows(conf.ShowRSS)
	if err != nil {
		log.WithField("error", err).Warn("Error getting Show List from ShowRSS")
	}

	for _, item := range showsListAllData.Item {
		add := true
		for _, addedItem := range showList {
			if addedItem.TVShowID == item.TVShowID {
				add = false
			}
		}
		if add {
			showList = append(showList, item)
		}
	}

	for _, item := range showsListSubscribedData.Item {
		add := true
		for _, addedItem := range showList {
			if addedItem.TVShowID == item.TVShowID {
				add = false
			}
		}
		if add {
			showList = append(showList, item)
		}
	}

	sort.SliceStable(showList, func(i, j int) bool {
		return showList[i].TVShowName < showList[j].TVShowName
	})
	// test, err := leetx.Lookup("ace", 15*time.Second)
	// spew.Dump(test)
	data := MainPageData{
		Movies:   movieData,
		Shows:    showData,
		ShowList: showList,
	}

	templateMain.Execute(w, data)

	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte(data))
}

// PingHandler is just a quick test to ensure api calls are working.
func (magnetApi *MagnetApi) PingHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Request sent to /api/ping")

	w.Write([]byte("Pong\n"))
}

// AddRawMagnet adds a magnet link to Seedr for downloading
func (magnetApi *MagnetApi) AddRawMagnet(magnetLink string) (seedr.Result, error) {
	magnetApi.mutex.Lock()
	defer magnetApi.mutex.Unlock()
	var result seedr.Result
	var err error
	if !dryRun {
		fmt.Printf("Adding magnet : %s\n", magnetLink)
		result, err = magnetApi.selectedSeedr.Add(magnetLink)
		if err != nil {
			return seedr.Result{}, err
		}
	}

	return result, nil
}

// AddRawTorrent adds a torrent link to Seedr for downloading
func (magnetApi *MagnetApi) AddRawTorrent(torrentUrl string) (seedr.Result, error) {
	magnetApi.mutex.Lock()
	defer magnetApi.mutex.Unlock()
	var result seedr.Result
	var err error
	if !dryRun {
		fmt.Printf("Adding torrent : %s\n", torrentUrl)
		result, err = magnetApi.selectedSeedr.AddTorrent(torrentUrl)
		if err != nil {
			return seedr.Result{}, err
		}
	}

	return result, nil
}

func (magnetApi *MagnetApi) GetFullShow(showID string) (showrss.Shows, error) {
	url := fmt.Sprintf("https://showrss.info/show/%s.rss", showID)
	data, err := showrss.GetShows(url)
	if err != nil {
		return showrss.Shows{}, err
	}

	return data, nil
}

func APILoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"source": r.Header.Get("X-FORWARDED-FOR"),
			"url":    r.Header.Get("URL"),
		})

		next.ServeHTTP(w, r)
	})
}
