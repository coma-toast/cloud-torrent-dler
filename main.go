package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/db"
	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/pidcheck"
	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/seedr"
	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/showrss"
	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/yts"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// SeedrInstance is the instance
type SeedrInstance interface {
	Add(magnet string) (seedr.Result, error)
	AddTorrent(torrent string) (seedr.Result, error)
	DeleteFile(id int) error
	DeleteFolder(id int) error
	FindID(filename string) (int, error)
	Get(item db.DownloadItem, destination string) error
	GetPath(ID int) (string, error)
	List(path string) ([]db.DownloadItem, error)
}

// Magnet is for magnet links and their ID
type Magnet struct {
	ID          int
	link        string
	name        string
	tVShowName  string
	torrentHash string
}

// ApiMagnet is the struct for data for working with Magnets via api
type ApiMagnet struct {
	Link string `json:"link"`
}

// One cache to rule them all
// var cache = &Cache{}
var dryRun = false
var videoFileRegex = regexp.MustCompile("(.*?).(mkv|mp4|avi|m4v)$")
var database = db.DbClient{}

func main() {
	configPath := flag.String("conf", ".", "config path")
	migrate := flag.Bool("migrate", false, "migrate the database")
	populateDb := flag.Bool("populate", false, "populate the database")
	flag.Parse()
	conf = getConf(*configPath)

	// Logging setup
	log.SetFormatter(&log.JSONFormatter{})
	if conf.DevMode {
		log.SetLevel(log.DebugLevel)
	}
	date := time.Now().Format("01-02-2006")
	err := os.MkdirAll(filepath.Join(conf.CachePath, "log"), 0777)
	if err != nil {
		log.WithField("error", err).Warn("Unable to create log directory")
	}
	fileLocation := filepath.Join(conf.CachePath, "log", fmt.Sprintf("ctd-%s.log", date))
	file, err := os.OpenFile(fileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err == nil {
		log.SetOutput(io.MultiWriter(file, os.Stdout))
	} else {
		log.WithField("error", err).Warn("Failed to log to file, using default stderr")
	}
	log.Info("Starting up...")

	if *migrate {
		log.Info("Migrating database...")
		err := database.Connect(conf.DBHost, conf.DBDatabase, conf.DBUser, conf.DBPassword)
		if err != nil {
			log.WithField("error", err).Warn("Error connecting to database")
		}
		// err = db.DeleteDatabase(database)
		err = database.Migrate()
		if err != nil {
			log.WithField("error", err).Warn("Error migrating database")
		}
		os.Exit(0)
	}

	if *populateDb {
		log.Info("Populating database...")
		err := database.Connect(conf.DBHost, conf.DBDatabase, conf.DBUser, conf.DBPassword)
		if err != nil {
			log.WithField("error", err).Warn("Error connecting to database")
		}
		populateDatabaseTVShows()
		os.Exit(0)
	}

	selectedSeedr := conf.GetSeedrInstance()

	pidPath := filepath.Join(conf.PidFilePath, "cloud-torrent-downloader")
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

	go func() {
		for range time.NewTicker(downloadLoopTime).C {
			// deleteQueue := make(map[string]int)
			// unsortedItems, err := findAllToDownload(selectedSeedr, "", conf.UseFTP)
			// if err != nil {
			// 	log.WithField("error", err).Warn("Error finding all downloads")
			// }
			database.GetSeedrUpload()
		unsortedLoop:
			// * commenting this all out.
			// * first, check the queue for items to add to seedr.
			// * second, check the database for any items that are not downloaded
			itemsToUpload, err := database.GetSeedrUpload()
			if err != nil {
				log.WithField("error", err).Warn("Error getting items to upload")
			}
			for _, item := range itemsToUpload {
				log.WithFields(log.Fields{
					"show": item.TVShowName,
				}).Info("Uploading item to seedr")
				err = selectedSeedr.AddTorrent(item.TorrentHash)
			// for _, unsortedItem := range unsortedItems {
			// 	log.WithFields(log.Fields{
			// 		"show": unsortedItem.TVShowName,
			// 		"name": unsortedItem.Name,
			// 	}).Info("processsing unsorted item")
			// 	okToDeleteFolder := false
			// 	isAVideo := videoFileRegex.MatchString(unsortedItem.Name)

			// 	if isAVideo {
			// 		// name := helper.SanitizeText(string(unsortedItem.Name[0 : len(unsortedItem.Name)-4]))
			// 		item, err := database.GetSeedrItemByID(unsortedItem.SeedrID.ID)
			// 		if err != nil {
			// 			log.WithField("error", err).Warn("Error getting item from database")
			// 		}
			// 		downloadItem, err := database.GetDownloadItemBySeedrID(unsortedItem.SeedrID.ID)
			// 		if err != nil {
			// 			log.WithField("error", err).Warn("Error getting download item from database")
			// 		}
			// 		downloadItem.SeedrID = *item
			// 		database.UpdateDownloadItem(downloadItem)

			// 		if unsortedItem.ShowGUID.ID != 0 {
			// 			path := filepath.Join(conf.DlRoot, conf.CompletedFolders[0])
			// 			if unsortedItem.TVShowName != "" {
			// 				path = filepath.Join(path, unsortedItem.TVShowName)
			// 			}
			// 			log.WithFields(log.Fields{
			// 				"show":        unsortedItem.TVShowName,
			// 				"destination": path,
			// 			}).Info("Show found and autodownloading")
			// 			_, err = os.Stat(path + unsortedItem.Name)
			// 			if err != nil {
			// 				if os.IsNotExist(err) {
			// 					err = selectedSeedr.Get(unsortedItem, path)
			// 					if err != nil {
			// 						log.WithField("error", err).Warn("Error getting the file ", unsortedItem.Name)
			// 						okToDeleteFolder = false
			// 						delete(deleteQueue, unsortedItem.FolderPath)
			// 						break unsortedLoop
			// 					}
			// 					okToDeleteFolder = true
			// 				}
			// 			}
			// 			if conf.DeleteAfterDownload {
			// 				infoLog(unsortedItem, "Deleting item")
			// 				err = selectedSeedr.DeleteFile(unsortedItem.SeedrID.ID)
			// 				if err != nil {
			// 					log.WithField("error", err).Warn("Error deleting file ", unsortedItem.Name)
			// 				}
			// 				database.DeleteSeedrItem(unsortedItem.SeedrID.ID)
			// 			}
			// 		}

			// 		var path string
			// 		if downloadItem.MediaType.String() == "movie" {
			// 			path = filepath.Join(conf.DlRoot, conf.CompletedFolders[1])
			// 		} else if downloadItem.MediaType.String() == "show" {
			// 			path = filepath.Join(conf.DlRoot, conf.CompletedFolders[0])
			// 		}
			// 		log.WithFields(log.Fields{
			// 			"name":        unsortedItem.Name,
			// 			"destination": path,
			// 		}).Info("autodownload item found, downloading")

			// 		_, err = os.Stat(path + unsortedItem.Name)
			// 		if err != nil {
			// 			if os.IsNotExist(err) {
			// 				err = selectedSeedr.Get(unsortedItem, path)
			// 				if err != nil {
			// 					log.WithField("error", err).Warn("Error getting the file ", unsortedItem.Name)
			// 					okToDeleteFolder = false
			// 					delete(deleteQueue, unsortedItem.FolderPath)
			// 					break unsortedLoop
			// 				}
			// 				okToDeleteFolder = true
			// 			}
			// 		}

			// 		if conf.DeleteAfterDownload {
			// 			infoLog(unsortedItem, "Deleting item")
			// 			err = selectedSeedr.DeleteFile(unsortedItem.SeedrID.ID)
			// 			if err != nil {
			// 				log.WithField("error", err).Warn("Error deleting file ", unsortedItem.Name)

			// 			}
			// 		}

			// 	}
			// 	if okToDeleteFolder {
			// 		deleteQueue[unsortedItem.FolderPath] = unsortedItem.ParentSeedrID
			// 	}
			// }
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
						localPath := filepath.Join(conf.DlRoot, item.FolderPath)
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
						err = selectedSeedr.DeleteFile(item.SeedrID.ID)
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
func dataLog(item db.DownloadItem, message string) {
	contextLogger := log.WithFields(log.Fields{
		"episodeID":     item.EpisodeID,
		"folderPath":    item.FolderPath,
		"isDir":         item.IsDir,
		"name":          item.Name,
		"tVShowName":    item.TVShowName,
		"parentSeedrID": item.ParentSeedrID,
		"seedrID":       item.SeedrID,
		"showID":        item.ShowGUID.ID,
	})

	contextLogger.Debug(message)
}

func infoLog(item db.DownloadItem, message string) {
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

// func validateCacheSeedrInfo(selectedSeedr SeedrInstance, downloadFolder string, item *db.DownloadItem) error {
// 	var err error
// 	filename := item.Name
// 	folderName := string(filename[0 : len(filename)-4])
// 	cacheItem := cache.Get(folderName)

// }

// func setCacheSeedrInfo(selectedSeedr SeedrInstance, downloadFolder string, item *db.DownloadItem) error {
// 	var err error
// 	filename := item.Name
// 	folderName := helper.SanitizeText(string(filename[0 : len(filename)-4]))
// 	// folderPath := downloadFolder + "/" + folderName
// 	cacheItem := cache.Get(folderName)

// 	item.ShowGUID = cacheItem.ShowGUID
// 	item.EpisodeID = cacheItem.EpisodeID
// 	// item.FolderPath = helper.SanitizePath(folderPath)
// 	item.FolderPath = downloadFolder
// 	item.TorrentHash = cacheItem.TorrentHash
// 	item.SeedrID, err = selectedSeedr.FindID(filename)
// 	if err != nil {
// 		return err
// 	}

// 	item.TVShowName = cacheItem.TVShowName

// 	err = cache.Set(folderName, *item)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func populateDatabaseTVShows() {
	allShows, err := showrss.GetAllEpisodeItems(conf.ShowRSS)
	if err != nil {
		log.WithField("error", err).Warn("Error getting shows")
		return
	}
	for i := 0; i < len(allShows); i++ {
		showData := allShows[i]
		dbData, err := database.GetDownloadItemByShowGUID(showData.GUID)
		if err != nil {
			log.WithField("error", err).Warn("Error getting show from database")
			continue
		}
		if dbData != nil {
			dbData.ShowGUID = showData
			database.UpdateDownloadItem(dbData)
			continue
		}
		newItem := db.DownloadItem{
			ShowGUID: showData,
		}
		err = database.CreateDownloadItem(&newItem)
	}
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
			log.WithField("error", err).Warn("Error adding magnet")
			return err
		}
	}

	itemData := db.DownloadItem{
		EpisodeID:   data.ID,
		ShowGUID:    showrss.Item{ID: showID},
		SeedrID:     seedr.File{ID: 0},
		Name:        data.name,
		FolderPath:  "",
		TVShowName:  data.tVShowName,
		TorrentHash: data.torrentHash,
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
		log.WithField("error", err).Warn("Error setting cache")
		return err
	}

	err = cache.SetAutoDownload(data.torrentHash, data.autoDownload)
	if err != nil {
		log.WithField("error", err).Warn("Error setting autodownload")
		return err
	}

	return nil
}

func findAllToDownload(instance SeedrInstance, path string, ftp bool) ([]db.DownloadItem, error) {
	// TODO: the 2 api calls need test cases, test the structs
	files, err := instance.List(path)

	if err != nil {
		// instance.List("")
		return []db.DownloadItem{}, err
	}
	downloads := []db.DownloadItem{}

	for _, file := range files {
		nextPath := file.FolderPath
		file.FolderPath = path

		if !file.IsDir {
			db.Database.GetDownloadItemBySeedrID(file.ID)
			downloads = append(downloads, file)
		} else {
			newDownloads, err := findAllToDownload(instance, nextPath, ftp)

			if err != nil {
				return []db.DownloadItem{}, err
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

