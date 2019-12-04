package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/showrss"
)

// SeedrInstance is the instance
type SeedrInstance interface {
	List(path string) ([]os.FileInfo, error)
	Get(file string, destination string) error
	Add(magnet string) error
}

type Magnet struct {
	ID   int
	link string
}

// One cache to rule them all
var cache = &Cache{}

func main() {

	conf := getConf()
	err := cache.Initialize(conf.CachePath)
	if err != nil {
		log.Println(err)
	}
	// TODO: re-write this mess I found on the internet
	// pidPath := fmt.Sprintf("%s/cloud-torrent-downloader", conf.PidFilePath)
	// pid := pidcheck.AlreadyRunning(pidPath)
	// if pid {
	// 	os.Exit(1)
	// }

	selectedSeedr := conf.GetSeedrInstance()
	_ = selectedSeedr

	// Channel so we can continuously monitor new episodes being added to showrss
	magnetChannel := make(chan Magnet)
	dontExit := make(chan bool)

	// ticker to control how often the loop runs
	tick := time.NewTicker(time.Second * 5)

	go func() {
		for {
			magnet := <-magnetChannel
			err := AddMagnet(selectedSeedr, magnet)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	// Do this for ever and ever
	go func() {
		for range tick.C {
			newMagnets, err := getNewEpisodes(conf.ShowRSS)
			if err != nil {
				log.Println(err)
			}
			for _, link := range newMagnets {
				magnetChannel <- link
			}
		}
	}()

	// TODO: worker pools for downloading - they take a long time and setting a limit would be good
	// list, err := findAllToDownload(selectedSeedr, conf.CompletedFolder, conf.UseFTP)
	// if err != nil {
	// 	panic(err)
	// }

	// for _, file := range list {
	// 	// spew.Dump("FILE", file)
	// 	err = selectedSeedr.Get(file, conf.DlRoot)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	// Waiting for a channel that never comes...
	<-dontExit
}

// AddMagnet adds a magnet link to Seedr for downloading
func AddMagnet(instance SeedrInstance, data Magnet) error {
	if cache.IsSet(data.ID) {
		return nil
	}
	fmt.Println("adding ", data.ID)
	// err := instance.Add(data.link)
	// if err != nil {
	// 	return err
	// }

	err := cache.Set(data.ID, data.link)
	if err != nil {
		return err
	}

	return nil
}

func findAllToDownload(instance SeedrInstance, path string, ftp bool) ([]string, error) {
	files, err := instance.List(path)

	if err != nil {
		return []string{}, err
	}
	downloads := []string{}

	for _, file := range files {
		var fullPath string
		if ftp {
			fullPath = path + "/" + file.Name()
		} else {
			fullPath = file.Name()
		}

		if !file.IsDir() {
			downloads = append(downloads, fullPath)
		} else {
			newDownloads, err := findAllToDownload(instance, fullPath, ftp)
			if err != nil {
				return []string{}, err
			}
			downloads = append(downloads, newDownloads...)
		}
	}

	return downloads, err
}

// var DeleteQueue []int

// TODO: move to a key/value map - id:magnet
// initEpisodes := []int{}
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
