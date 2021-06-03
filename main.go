package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/helper"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/pidcheck"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/seedr"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/showrss"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/yts"
)

// SeedrInstance is the instance
type SeedrInstance interface {
	Add(magnet string) (seedr.Result, error)
	AddTorrent(torrent string) (seedr.Result, error)
	DeleteFile(id int) error
	DeleteFolder(id int) error
	FindID(filename string) (int, error)
	Get(item DownloadItem, destination string) error
	GetPath(ID int) (string, error)
	List(path string) ([]DownloadItem, error)
}

// Magnet is for magnet links and their ID
type Magnet struct {
	ID         int
	link       string
	name       string
	tVShowName string
}

// ApiMagnet is the struct for data for working with Magnets via api
type ApiMagnet struct {
	Link string `json:"link"`
}

// DownloadItem is the information needed for the download queue
type DownloadItem struct {
	EpisodeID     int
	FolderPath    string
	IsDir         bool
	Name          string
	TVShowName    string
	ParentSeedrID int
	SeedrID       int
	ShowID        int
}

// One cache to rule them all
var cache = &Cache{}
var dryRun = false

func main() {
	configPath := flag.String("conf", ".", "config path")
	flag.Parse()
	conf = getConf(*configPath)

	// Logging setup
	log.SetFormatter(&log.JSONFormatter{})
	if conf.DevMode {
		log.SetLevel(log.DebugLevel)
	}
	date := time.Now().Format("01-02-2006")
	err := os.MkdirAll(conf.CachePath+"/log", 0777)
	if err != nil {
		log.WithField("error", err).Warn("Unable to create log directory")
	}
	fileLocation := fmt.Sprintf("%s/log/ctd-%s.log", conf.CachePath, date)
	file, err := os.OpenFile(fileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err == nil {
		log.SetOutput(io.MultiWriter(file, os.Stdout))
	} else {
		log.WithField("error", err).Warn("Failed to log to file, using default stderr")
	}
	log.Info("Starting up...")

	// Cache setup
	err = cache.Initialize(conf.CachePath)
	if err != nil {
		dryRun = true
		log.WithField("error", err).Warn("Error initializing cache")
	}

	selectedSeedr := conf.GetSeedrInstance()

	pidPath := fmt.Sprintf("%s/cloud-torrent-downloader", conf.PidFilePath)
	pid := pidcheck.AlreadyRunning(pidPath)
	if pid {
		log.Fatal("App already running. Exiting.")
	}

	var episodeLoopTime = time.Second * time.Duration(conf.CheckEpisodesTimer)
	if conf.DevMode {
		log.Debug("#####    Dev mode enabled.    ##### ")
		episodeLoopTime = time.Second * 5
	}

	go func() {
		log.Debug("Starting first run of checkNewEpisode")
		checkNewEpisodes(selectedSeedr)
		// ticker to control how often the loop runs
		// for range time.NewTicker(time.Second * 10).C { // * dev code
		for range time.NewTicker(episodeLoopTime).C {
			log.Debug("Timer loop running checkNewEpisodes")
			checkNewEpisodes(selectedSeedr)
		}
	}()

	magnetApi := &MagnetApi{selectedSeedr: selectedSeedr}
	go magnetApi.RunMagnetApi()

	// TODO: worker pools for downloading - they take a long time and setting a limit would be good

	var downloadLoopTime = time.Second * time.Duration(conf.CheckFilesToDownloadTimer)
	if conf.DevMode {
		downloadLoopTime = time.Second * 10
	}
	// downloadWorker()
	// Channel so we can continuously monitor new episodes being added to showrss

	go func() {
		for range time.NewTicker(downloadLoopTime).C {
			deleteQueue := make(map[string]int)
			unsortedItems, err := findAllToDownload(selectedSeedr, "", conf.UseFTP)
			if err != nil {
				log.WithField("error", err).Warn("Error finding all downloads")
			}
		unsortedLoop:
			for _, unsortedItem := range unsortedItems {
				dataLog(unsortedItem, "Unsorted loop")
				okToDeleteFolder := false
				isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi|m4v)$", unsortedItem.Name)
				if isAVideo {
					name := helper.SanitizeText(string(unsortedItem.Name[0 : len(unsortedItem.Name)-4]))
					itemCacheData := cache.Get(name)
					_ = itemCacheData
					setCacheSeedrInfo(selectedSeedr, conf.CompletedFolders[0], &unsortedItem)
					if unsortedItem.ShowID != 0 {
						path := fmt.Sprintf("%s/%s", conf.DlRoot, conf.CompletedFolders[0])
						if unsortedItem.TVShowName != "" {
							path = fmt.Sprintf("%s/%s", path, unsortedItem.TVShowName)
						}
						log.WithFields(log.Fields{
							"show":        unsortedItem.TVShowName,
							"destination": path,
						}).Info("Show found and autodownloading")
						_, err = os.Stat(path + unsortedItem.Name)
						if err != nil {
							if os.IsNotExist(err) {
								err = selectedSeedr.Get(unsortedItem, path)
								if err != nil {
									log.WithField("error", err).Warn("Error getting the file ", unsortedItem.Name)
									okToDeleteFolder = false
									delete(deleteQueue, unsortedItem.FolderPath)
									break unsortedLoop
								}
								okToDeleteFolder = true
							}
						}
						if conf.DeleteAfterDownload {
							infoLog(unsortedItem, "Deleting item")
							err = selectedSeedr.DeleteFile(unsortedItem.SeedrID)
							if err != nil {
								log.WithField("error", err).Warn("Error deleting file ", unsortedItem.Name)

							}
						}
					}
				}
				if okToDeleteFolder {
					deleteQueue[unsortedItem.FolderPath] = unsortedItem.ParentSeedrID
				}
			}
		outerLoop:
			for _, downloadFolder := range conf.CompletedFolders {
				log.WithField("folder", downloadFolder).Debug("outerLoop")
				okToDeleteFolder := true
				list, err := findAllToDownload(selectedSeedr, downloadFolder, conf.UseFTP)
				if err != nil {
					log.WithField("error", err).Warn("Error finding downloads")
				}

				for _, item := range list {
					isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi|m4v)$", item.Name)
					if conf.DevMode {
						// * dev code so you don't download huge files during testing
						isAVideo, _ = regexp.MatchString("(.*?).(txt|jpg)$", item.Name)
					}
					if isAVideo {
						dataLog(item, "Setting cache for item")
						setCacheSeedrInfo(selectedSeedr, downloadFolder, &item)
						localPath := fmt.Sprintf("%s/%s", conf.DlRoot, item.FolderPath)
						thisShouldBeDownloaded := shouldThisBeDownloaded(localPath + item.Name)
						if thisShouldBeDownloaded {
							dataLog(item, "Item will be downloaded")
							err = selectedSeedr.Get(item, localPath)
							if err != nil {
								log.WithField("error", err).Warn("Error getting file ", item.Name)
								okToDeleteFolder = false
								delete(deleteQueue, item.FolderPath)
								break outerLoop
							}

						}
					}
					if conf.DeleteAfterDownload {
						infoLog(item, "Download complete, deleting item")
						fmt.Println("Deleting item: " + item.Name)
						err = selectedSeedr.DeleteFile(item.SeedrID)
						if err != nil {
							log.WithField("error", err).Warn("Error deleting file", item.Name)
						}
					}
					if okToDeleteFolder {
						deleteQueue[item.FolderPath] = item.ParentSeedrID
					}
				}
			}
			deleteTheQueue(selectedSeedr, deleteQueue)
		}
	}()
	dontExit := make(chan bool)
	for {
		time.Sleep(1000000)
	}
	// Waiting for a channel that never comes...
	<-dontExit
}

// * Logging functions
func dataLog(item DownloadItem, message string) {
	contextLogger := log.WithFields(log.Fields{
		"episodeID":     item.EpisodeID,
		"folderPath":    item.FolderPath,
		"isDir":         item.IsDir,
		"name":          item.Name,
		"tVShowName":    item.TVShowName,
		"parentSeedrID": item.ParentSeedrID,
		"seedrID":       item.SeedrID,
		"showID":        item.ShowID,
	})

	contextLogger.Debug(message)
}

func infoLog(item DownloadItem, message string) {
	contextLogger := log.WithFields(log.Fields{
		"folderPath": item.FolderPath,
		"name":       item.Name,
	})

	contextLogger.Info(message)
}

func deleteTheQueue(selectedSeedr SeedrInstance, deleteQueue map[string]int) {
	if conf.DeleteAfterDownload {
		for name, id := range deleteQueue {
			fmt.Println("Deleting folder: " + name)
			err := selectedSeedr.DeleteFolder(id)
			if err != nil {
				log.WithField("error", err).Warn("Error deleting the queue")
			}
		}
	}
}

// func validateCacheSeedrInfo(selectedSeedr SeedrInstance, downloadFolder string, item *DownloadItem) error {
// 	var err error
// 	filename := item.Name
// 	folderName := string(filename[0 : len(filename)-4])
// 	cacheItem := cache.Get(folderName)

// }

func setCacheSeedrInfo(selectedSeedr SeedrInstance, downloadFolder string, item *DownloadItem) error {
	var err error
	filename := item.Name
	folderName := helper.SanitizeText(string(filename[0 : len(filename)-4]))
	// folderPath := downloadFolder + "/" + folderName
	cacheItem := cache.Get(folderName)

	item.ShowID = cacheItem.ShowID
	item.EpisodeID = cacheItem.EpisodeID
	// item.FolderPath = helper.SanitizePath(folderPath)
	item.FolderPath = downloadFolder
	item.SeedrID, err = selectedSeedr.FindID(filename)
	if err != nil {
		return err
	}

	item.TVShowName = cacheItem.TVShowName

	err = cache.Set(folderName, *item)
	if err != nil {
		return err
	}

	return nil
}

func checkNewEpisodes(selectedSeedr SeedrInstance) {
	log.Info("Checking ShowRSS for new episodes")
	initializeMagnetList, err := getNewEpisodes(conf.ShowRSS)
	if err != nil {
		log.WithField("error", err).Warn("Error getting new episodes")
		return
	}

	allShows, err := showrss.GetShows(conf.ShowRSS)
	if err != nil {
		log.WithField("error", err).Warn("Error getting shows")
		return
	}

	for _, magnet := range initializeMagnetList {
		showID, err := getShowIDFromEpisodeID(magnet.ID, allShows)
		if err != nil {
			log.WithField("error", err).Warn("Error getting ShowID from EpisodeID")
			continue
		}

		err = AddMagnet(selectedSeedr, magnet, showID)
		if err != nil {
			log.WithField("error", err).Warn("Error adding Magnet")
			continue
		}
	}
	dryRun = false
}

func getShowIDFromEpisodeID(episode int, allShows showrss.Shows) (int, error) {
	for _, item := range allShows.Item {
		if item.TVEpisodeID == episode {
			return item.TVShowID, nil
		}
	}
	err := fmt.Errorf("unable to find ShowID for EpisodeID: %d", episode)

	return 0, err
}

func shouldThisBeDownloaded(filepath string) bool {
	currentFile, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
	} else {
		if currentFile.Size() == 0 {
			return true
		}
	}

	return false
}

// AddMagnet adds a magnet link with ShowRSS ID to Seedr for downloading
func AddMagnet(instance SeedrInstance, data Magnet, showID int) error {
	if cache.IsSet(data.name) {
		return nil
	}

	if !dryRun {
		fmt.Printf("Adding magnet for episode: %s\n", data.name)
		_, err := instance.Add(data.link)
		if err != nil {
			return err
		}
	}

	itemData := DownloadItem{
		EpisodeID:  data.ID,
		ShowID:     showID,
		SeedrID:    0,
		Name:       data.name,
		FolderPath: "",
		TVShowName: data.tVShowName,
	}

	magnetParts := strings.Split(data.link, "&")
	for _, part := range magnetParts {
		if strings.Contains(part, "dn=") {
			part = strings.TrimPrefix(part, "dn=")
			part = strings.ReplaceAll(part, "+", " ")
			itemData.Name = part
		}
	}

	err := cache.Set(data.name, itemData)
	if err != nil {
		return err
	}

	return nil
}

func findAllToDownload(instance SeedrInstance, path string, ftp bool) ([]DownloadItem, error) {
	// TODO: the 2 api calls need test cases, test the structs
	files, err := instance.List(path)

	if err != nil {
		return []DownloadItem{}, err
	}
	downloads := []DownloadItem{}

	for _, file := range files {
		nextPath := file.FolderPath
		file.FolderPath = path

		if !file.IsDir {
			downloads = append(downloads, file)
		} else {
			newDownloads, err := findAllToDownload(instance, nextPath, ftp)

			if err != nil {
				return []DownloadItem{}, err
			}
			downloads = append(downloads, newDownloads...)
		}
	}

	return downloads, err
}

// getNewEpisodes is a loop to look for new shows added to the RSS feed to then add to the download queue
func getNewEpisodes(url string) ([]Magnet, error) {
	returnData := []Magnet{}
	episodes, err := showrss.GetAllEpisodeItems(url)
	if err != nil {
		return nil, err
	}
	for _, item := range episodes {
		returnData = append(returnData, Magnet{
			ID:         item.ItemTVEpisodeID(),
			link:       item.ItemLink(),
			name:       item.ItemTitle(),
			tVShowName: item.ItemTVShowName(),
		})
	}

	return returnData, nil
}

type MagnetApi struct {
	selectedSeedr SeedrInstance
}

type MainPageData struct {
	Movies []yts.Movie
	Shows  []showrss.Item
}

// RunMagnetApi is the api for adding magnet urls
func (magnetApi *MagnetApi) RunMagnetApi() {
	r := mux.NewRouter()
	r.HandleFunc("/gui", magnetApi.GuiHandler)
	r.HandleFunc("/api/ping", magnetApi.PingHandler)
	r.HandleFunc("/api/magnet", magnetApi.AddMagnetHandler).Methods("POST")
	r.HandleFunc("/api/torrent", magnetApi.AddTorrentHandler).Methods("POST")
	log.Info(fmt.Sprintf("Magnet API running. Send JSON {link: url} as a POST request to x.x.x.x:%s/api/magnet to add directly to Seedr!", conf.Port))

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// r.Use(APILoggingMiddleware)
	log.Error(http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), r))
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
	}
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
	bytes, err := ioutil.ReadAll(r.Body)
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

	data := MainPageData{
		Movies: movieData,
		Shows:  showData,
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

func APILoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"source": r.Header.Get("X-FORWARDED-FOR"),
			"url":    r.Header.Get("URL"),
		})

		next.ServeHTTP(w, r)
	})
}
