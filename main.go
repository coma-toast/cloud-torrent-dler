package main

import (
	"fmt"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/pidcheck"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/showrss"
)

// SeedrInstance is the instance
type SeedrInstance interface {
	List(path string) ([]os.FileInfo, error)
	Get(file string, destination string) error
	Add(magnet string) error
}

// Magnet is for magnet links and their ID
type Magnet struct {
	ID   int
	link string
}

// DownloadItem is the information needed for the download queue
type DownloadItem struct {
	ID         int
	Name       string
	FolderPath string
}

// One cache to rule them all
var cache = &Cache{}
var dryRun = false
var DeleteQueue = []DownloadItem{}

func main() {
	conf = getConf()
	err := cache.Initialize(conf.CachePath)
	if err != nil {
		dryRun = true
		fmt.Println(err)
	}

	selectedSeedr := conf.GetSeedrInstance()
	// _ = selectedSeedr

	pidPath := fmt.Sprintf("%s/cloud-torrent-downloader", conf.PidFilePath)
	pid := pidcheck.AlreadyRunning(pidPath)
	if pid {
		os.Exit(1)
	}

	// Channel so we can continuously monitor new episodes being added to showrss
	dontExit := make(chan bool)

	go func() {
		checkNewEpisodes(selectedSeedr)
		// ticker to control how often the loop runs
		for range time.NewTicker(time.Minute * 5).C {
			checkNewEpisodes(selectedSeedr)
		}
	}()

	// TODO: completed folder should be an array of folders to be monitored with
	// their own download destinations - for example, you can do a kids/not kids
	// download separately.

	// TODO: worker pools for downloading - they take a long time and setting a limit would be good

	// downloadWorker()
	for range time.NewTicker(time.Second * 5).C {
		for _, downloadFolder := range conf.CompletedFolder {
			list, err := findAllToDownload(selectedSeedr, downloadFolder, conf.UseFTP)
			if err != nil {
				panic(err)
			}
			// TODO: file exist checking;
			// TODO: delete queue;
			// TODO: delete;
			// os.Exit(4)

			for _, file := range list {
				// spew.Dump("FILE", file)
				filePath := fmt.Sprintf("%s/%s", conf.DlRoot, file.Name)
				spew.Dump(filePath)
				_, err := os.Stat(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						err = selectedSeedr.Get(file.Name, conf.DlRoot)
						if err != nil {
							fmt.Println(err)
						}

						defer addToDeleteQueue(file)

					}
				}
				fmt.Printf("file %s exists, skipping\n", file.Name)
			}
		}
	}

	// Waiting for a channel that never comes...
	<-dontExit
}

func addToDeleteQueue(file DownloadItem) {
	var deleteItem DownloadItem

	deleteItem.ID = file.ID
	deleteItem.Name = file.Name
	deleteItem.FolderPath = file.Name

	DeleteQueue = append(DeleteQueue, deleteItem)
	fmt.Printf("Deleting %s\n", file.Name)
}

func checkNewEpisodes(selectedSeedr SeedrInstance) {
	initializeMagnetList, err := getNewEpisodes(conf.ShowRSS)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, magnet := range initializeMagnetList {
		err := AddMagnet(selectedSeedr, magnet)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	dryRun = false
}

// AddMagnet adds a magnet link to Seedr for downloading
func AddMagnet(instance SeedrInstance, data Magnet) error {
	if cache.IsSet(data.ID) {
		return nil
	}
	if !dryRun {
		fmt.Printf("Adding magnet for episode: %d\n", data.ID)
		err := instance.Add(data.link)
		if err != nil {
			return err
		}
	}

	err := cache.Set(data.ID, data.link)
	if err != nil {
		return err
	}

	return nil
}

func findAllToDownload(instance SeedrInstance, path string, ftp bool) ([]DownloadItem, error) {
	files, err := instance.List(path)

	if err != nil {
		return []DownloadItem{}, err
	}
	downloads := []DownloadItem{}

	for _, file := range files {
		var currentItem DownloadItem
		currentItem.Name = file.Name()
		currentItem.ID = 0
		if ftp {
			currentItem.FolderPath = path + "/" + file.Name()
		} else {
			currentItem.FolderPath = file.Name()
		}

		if !file.IsDir() {
			downloads = append(downloads, currentItem)
		} else {
			newDownloads, err := findAllToDownload(instance, currentItem.FolderPath, ftp)
			if err != nil {
				return []DownloadItem{}, err
			}
			downloads = append(downloads, newDownloads...)
		}
	}

	return downloads, err
}

// var DeleteQueue []int

// getNewEpisodes is a loop to look for new shows added to the RSS feed to then add to the download queue
func getNewEpisodes(url string) ([]Magnet, error) {
	returnData := []Magnet{}
	episodes, err := showrss.GetAllEpisodeLinks(url)
	if err != nil {
		return nil, err
	}
	for episodeID, magnetLink := range episodes {
		returnData = append(returnData, Magnet{
			ID:   episodeID,
			link: magnetLink,
		})
	}

	return returnData, nil
}
