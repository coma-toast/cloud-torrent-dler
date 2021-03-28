package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/helper"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/pidcheck"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/showrss"
)

// SeedrInstance is the instance
type SeedrInstance interface {
	Add(magnet string) error
	DeleteFile(id int) error
	DeleteFolder(id int) error
	FindID(filename string) (int, error)
	Get(item DownloadItem, destination string) error
	GetPath(ID int) (string, error)
	List(path string) ([]DownloadItem, error)
}

// Magnet is for magnet links and their ID
type Magnet struct {
	ID   int
	link string
	name string
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
	log.Println("Starting up...")
	configPath := flag.String("conf", ".", "config path")
	flag.Parse()
	conf = getConf(*configPath)
	err := cache.Initialize(conf.CachePath)
	if err != nil {
		dryRun = true
		fmt.Println(err)
	}

	selectedSeedr := conf.GetSeedrInstance()

	pidPath := fmt.Sprintf("%s/cloud-torrent-downloader", conf.PidFilePath)
	pid := pidcheck.AlreadyRunning(pidPath)
	if pid {
		log.Println("App already running. Exiting.")
		os.Exit(1)
	}

	// Channel so we can continuously monitor new episodes being added to showrss
	dontExit := make(chan bool)
	var episodeLoopTime = time.Second * time.Duration(conf.CheckEpisodesTimer)
	if conf.DevMode {
		log.Println("Dev mode enabled.")
		episodeLoopTime = time.Second * 5
	}

	go func() {
		checkNewEpisodes(selectedSeedr)
		// ticker to control how often the loop runs
		// for range time.NewTicker(time.Second * 10).C { // * dev code
		for range time.NewTicker(episodeLoopTime).C {
			checkNewEpisodes(selectedSeedr)
		}
	}()

	magnetApi := &MagnetApi{selectedSeedr: selectedSeedr}
	magnetApi.RunMagnetApi()

	// TODO: worker pools for downloading - they take a long time and setting a limit would be good

	var downloadLoopTime = time.Second * time.Duration(conf.CheckFilesToDownloadTimer)
	if conf.DevMode {
		downloadLoopTime = time.Second * 10
	}
	// downloadWorker()
	go func() {
		for range time.NewTicker(downloadLoopTime).C {
			deleteQueue := make(map[string]int)
			unsortedItems, err := findAllToDownload(selectedSeedr, "", conf.UseFTP)
			if err != nil {
				fmt.Println(err)
			}
		unsortedLoop:
			for _, unsortedItem := range unsortedItems {
				okToDeleteFolder := false
				isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi|m4v)$", unsortedItem.Name)
				if isAVideo {
					setCacheSeedrInfo(selectedSeedr, conf.CompletedFolders[0], &unsortedItem)
					if unsortedItem.ShowID != 0 {
						log.Println("Show found and autodownloading. ", unsortedItem.TVShowName)
						path := fmt.Sprintf("%s/%s/%s%s", conf.DlRoot, conf.CompletedFolders[0], unsortedItem.TVShowName, unsortedItem.FolderPath)
						_, err = os.Stat(path + unsortedItem.Name)
						if err != nil {
							if os.IsNotExist(err) {
								err = selectedSeedr.Get(unsortedItem, path)
								if err != nil {
									fmt.Println(err)
									okToDeleteFolder = false
									delete(deleteQueue, unsortedItem.FolderPath)
									break unsortedLoop
								}
								okToDeleteFolder = true
							}
						}
						if conf.DeleteAfterDownload {
							fmt.Println("Deleting item: " + unsortedItem.Name)
							err = selectedSeedr.DeleteFile(unsortedItem.SeedrID)
							if err != nil {
								fmt.Println(err)

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
				okToDeleteFolder := true
				list, err := findAllToDownload(selectedSeedr, downloadFolder, conf.UseFTP)
				if err != nil {
					fmt.Println(err)
				}

				for _, item := range list {
					isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi|m4v)$", item.Name)
					if conf.DevMode {
						// * dev code so you don't download huge files during testing
						isAVideo, _ = regexp.MatchString("(.*?).(txt|jpg)$", item.Name)
					}
					if isAVideo {
						setCacheSeedrInfo(selectedSeedr, downloadFolder, &item)
						localPath := fmt.Sprintf("%s/%s", conf.DlRoot, item.FolderPath)
						thisShouldBeDownloaded := shouldThisBeDownloaded(localPath + item.Name)
						if thisShouldBeDownloaded {
							err = selectedSeedr.Get(item, localPath)
							if err != nil {
								fmt.Println(err)
								okToDeleteFolder = false
								delete(deleteQueue, item.FolderPath)
								break outerLoop
							}

						}
					}
					if conf.DeleteAfterDownload {
						fmt.Println("Deleting item: " + item.Name)
						err = selectedSeedr.DeleteFile(item.SeedrID)
						if err != nil {
							fmt.Println(err)

						}
					}
					if okToDeleteFolder {
						deleteQueue[item.FolderPath] = item.ParentSeedrID
					}
				}
			}
			deleteTheQueue(selectedSeedr, deleteQueue)
		}

		// Waiting for a channel that never comes...
		<-dontExit
	}()
}

func deleteTheQueue(selectedSeedr SeedrInstance, deleteQueue map[string]int) {
	if conf.DeleteAfterDownload {
		for name, id := range deleteQueue {
			fmt.Println("Deleting folder: " + name)
			var err error
			err = selectedSeedr.DeleteFolder(id)
			if err != nil {
				fmt.Println(err)
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
	folderPath := downloadFolder + "/" + folderName
	cacheItem := cache.Get(folderName)

	item.ShowID = cacheItem.ShowID
	item.EpisodeID = cacheItem.EpisodeID
	item.FolderPath = helper.SanitizePath(folderPath)
	item.SeedrID, err = selectedSeedr.FindID(filename)
	if err != nil {
		return err
	}

	err = cache.Set(folderName, *item)
	if err != nil {
		return err
	}

	return nil
}

func checkNewEpisodes(selectedSeedr SeedrInstance) {
	log.Println("Checking ShowRSS for new episodes")
	initializeMagnetList, err := getNewEpisodes(conf.ShowRSS)
	if err != nil {
		fmt.Println(err)
		return
	}

	allShows, err := showrss.GetShows(conf.ShowRSS)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, magnet := range initializeMagnetList {
		showID, err := getShowIDFromEpisodeID(magnet.ID, allShows)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = AddMagnet(selectedSeedr, magnet, showID)
		if err != nil {
			fmt.Println(err)
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
		err := instance.Add(data.link)
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
			ID:   item.ItemTVEpisodeID(),
			link: item.ItemLink(),
			name: item.ItemTitle(),
		})
	}

	return returnData, nil
}

type MagnetApi struct {
	selectedSeedr SeedrInstance
}

// RunMagnetApi is the api for adding magnet urls
func (magnetApi *MagnetApi) RunMagnetApi() {
	r := mux.NewRouter()
	r.HandleFunc("/gui", magnetApi.AllHandler)
	r.HandleFunc("/api/ping", magnetApi.PingHandler)
	r.HandleFunc("/api/magnet", magnetApi.AddMagnetHandler).Methods("POST")
	log.Printf("Magnet API starting up...\nSend a magnet link as a GET request to port %s to add", conf.Port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), r))
}

// AddMagnetHandler handles api calls adding magnets
func (magnetApi *MagnetApi) AddMagnetHandler(w http.ResponseWriter, r *http.Request) {
	var data ApiMagnet
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		log.Println("Error decoding JSON: ", err)
	} else {
		spew.Dump(data)

		magnetApi.AddRawMagnet(data.Link)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Settin'\n"))
}

// PingHandler is just a quick test to ensure api calls are working.
func (magnetApi *MagnetApi) AllHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	data := string(bytes)
	spew.Dump(data)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

// PingHandler is just a quick test to ensure api calls are working.
func (magnetApi *MagnetApi) PingHandler(w http.ResponseWriter, r *http.Request) {
	spew.Dump("here in ping")

	w.Write([]byte("Pong\n"))
}

// AddRawMagnet adds a magnet link to Seedr for downloading
func (magnetApi *MagnetApi) AddRawMagnet(magnetLink string) error {
	if !dryRun {
		fmt.Printf("Adding magnet : %s\n", magnetLink)
		err := magnetApi.selectedSeedr.Add(magnetLink)
		if err != nil {
			return err
		}
	}

	return nil
}
