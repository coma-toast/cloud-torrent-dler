package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/kennygrant/sanitize"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/pidcheck"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/showrss"
)

// SeedrInstance is the instance
type SeedrInstance interface {
	Add(magnet string) error
	Get(file string, destination string) error
	GetPath(ID int) (string, error)
	FindID(filename string) (int, error)
	List(path string) ([]os.FileInfo, error)
}

// Magnet is for magnet links and their ID
type Magnet struct {
	ID   int
	link string
	name string
}

// DownloadItem is the information needed for the download queue
type DownloadItem struct {
	EpisodeID  int
	ShowID     int
	SeedrID    int
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
	for range time.NewTicker(time.Second * 10).C {
		fmt.Println("Tick...")
		for _, downloadFolder := range conf.CompletedFolder {
			list, err := findAllToDownload(selectedSeedr, downloadFolder, conf.UseFTP)
			if err != nil {
				panic(err)
			}

			for _, file := range list {
				isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi|m4v)$", file.Name)
				if isAVideo {
					// file.Name = sanitizeText(file.Name)
					setCacheSeedrInfo(selectedSeedr, file.Name)
					// spew.Dump("FILE", file)
					// folderPath := fmt.Sprintf("%s/%s/", conf.DlRoot, downloadFolder)
					// fmt.Println("folderPath: " + folderPath)
					// _, err = os.Stat(folderPath)
					// if err != nil {
					// 	if os.IsNotExist(err) {
					// 		err = selectedSeedr.Get(file.Name, folderPath)
					// 		if err != nil {
					// 			fmt.Println(err)
					// 		}

					// 		// defer addToDeleteQueue(file)

					// 	}
					// }
				}
			}
		}
	}

	// Waiting for a channel that never comes...
	<-dontExit
}

func setCacheSeedrInfo(selectedSeedr SeedrInstance, filename string) error {
	folderName := string(filename[0 : len(filename)-4])
	if !cache.IsSet(folderName) {
		folderName = sanitizeText(folderName)
		folderItem := cache.Get(folderName)
		id, err := selectedSeedr.FindID(filename)
		if err != nil {
			return err
		}

		folderItem.SeedrID = id
		folderItem.Name = filename
		err = cache.Set(folderName, folderItem)
		if err != nil {
			return err
		}
	}

	return nil
}

func sanitizeText(input string) string {
	extension := input[len(input)-4:]
	output := sanitize.BaseName(input[0 : len(input)-4])
	output = strings.ReplaceAll(output, "-", " ")
	output = output + extension

	return output
}

// func addToDeleteQueue(file DownloadItem) {
// 	var deleteItem int

// 	deleteItem.SeerID = file.SeedrID
// 	deleteItem.Name = file.Name
// 	deleteItem.FolderPath = file.Name

// 	DeleteQueue = append(DeleteQueue, deleteItem)
// 	fmt.Printf("Deleting %s\n", file.Name)
// }

func checkNewEpisodes(selectedSeedr SeedrInstance) {
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

// getShowIDFromEpisodeID does exactly that
func getShowIDFromEpisodeID(episode int, allShows showrss.Shows) (int, error) {
	for _, item := range allShows.Item {
		if item.TVEpisodeID == episode {
			return item.TVShowID, nil
		}
	}
	err := fmt.Errorf("unable to find ShowID for EpisodeID: %d", episode)

	return 0, err
}

// AddMagnet adds a magnet link to Seedr for downloading
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
			fmt.Println(part)
		}
	}
	spew.Dump(itemData)
	// os.Exit(7)

	err := cache.Set(data.name, itemData)
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
		currentItem.SeedrID = 0

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
