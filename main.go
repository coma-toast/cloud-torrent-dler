package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
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

// DeleteQueue is a list of folders to delete.
// Using a queue since there could be multiple files in a single folder
// so we want to wait until the loop is done before deleting the folder.
var DeleteQueue = map[string]int{}

func main() {
	conf = getConf()
	err := cache.Initialize(conf.CachePath)
	if err != nil {
		dryRun = true
		fmt.Println(err)
	}

	DeleteQueue = make(map[string]int)
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
		for range time.NewTicker(time.Minute * 1).C {
			checkNewEpisodes(selectedSeedr)
		}
	}()

	// TODO: worker pools for downloading - they take a long time and setting a limit would be good

	// downloadWorker()
	for range time.NewTicker(time.Second * 5).C {
		// for range time.NewTicker(time.Minute * 1).C {
		for _, downloadFolder := range conf.CompletedFolder {
			list, err := findAllToDownload(selectedSeedr, downloadFolder, conf.UseFTP)
			if err != nil {
				panic(err)
			}

			for _, item := range list {
				// isAVideo, _ := regexp.MatchString("(.*?).(txt|jpg)$", item.Name)
				isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi|m4v)$", item.Name)
				if isAVideo {
					setCacheSeedrInfo(selectedSeedr, downloadFolder, &item)
					localPath := fmt.Sprintf("%s/%s/", conf.DlRoot, item.FolderPath)
					_, err = os.Stat(localPath + item.Name)
					if err != nil {
						if os.IsNotExist(err) {
							err = selectedSeedr.Get(item, conf.DlRoot)
							if err != nil {
								fmt.Println(err)
							}
							// fmt.Printf("Pretend downloading %s\n", item.Name)
						}
					}
				}
				fmt.Println("Pretend delete item after downloading: " + item.Name)
				// err = selectedSeedr.DeleteFile(item.SeedrID)
				// if err != nil {
				// fmt.Println(err)
				// }

				folderID, err := selectedSeedr.FindID(item.FolderPath)
				if err != nil {
					fmt.Println(err)
					continue
				}

				DeleteQueue[item.FolderPath] = folderID
			}
			// if err != nil {
			// 	fmt.Println(err)
			// }t(item, conf.DlRoot)
			// if err != nil {
			// 	fmt.Println(err)
			// }
		}
		deleteTheQueue(selectedSeedr, DeleteQueue)
	}

	// Waiting for a channel that never comes...
	<-dontExit
}

func deleteTheQueue(selectedSeedr SeedrInstance, DeleteQueue map[string]int) {
	for name, id := range DeleteQueue {
		fmt.Println("This would delete item: " + name)
		_ = id
		var err error
		// err := selectedSeedr.DeleteFolder(id)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func setCacheSeedrInfo(selectedSeedr SeedrInstance, downloadFolder string, item *DownloadItem) error {
	var err error
	filename := item.Name
	folderName := string(filename[0 : len(filename)-4])
	// if !cache.IsSet(folderName) {
	folderName = sanitizeText(folderName)
	cacheItem := cache.Get(folderName)

	item.ShowID = cacheItem.ShowID
	item.EpisodeID = cacheItem.EpisodeID
	item.Name = filename
	item.FolderPath = fmt.Sprintf("%s/%s", downloadFolder, folderName)
	item.SeedrID, err = selectedSeedr.FindID(filename)
	if err != nil {
		return err
	}

	err = cache.Set(folderName, *item)
	if err != nil {
		return err
	}
	// }

	return nil
}

func sanitizeText(input string) string {
	var extension string
	extension = input[len(input)-4:]
	output := sanitize.BaseName(input)
	output = strings.ReplaceAll(output, "-", " ")
	hasExtension, _ := regexp.MatchString("(.*?).(mkv|mp4|avi|m4v|txt|jpg)$", input)
	if hasExtension {
		output = output + extension
	}

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
		}
	}

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
