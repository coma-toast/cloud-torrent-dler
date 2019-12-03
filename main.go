package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/showrss"

	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/pidcheck"
)

// SeedrInstance is the instance
type SeedrInstance interface {
	List(path string) ([]os.FileInfo, error)
	Get(file string, destination string) error
	Add(magnet string) error
}

// AddMagnet adds a magnet link to Seedr for downloading
func AddMagnet(instance SeedrInstance, magnet string, wg *sync.WaitGroup) {
	defer wg.Done()
	err := instance.Add(magnet)
	if err != nil {
		log.Printf("Error adding magnet: %s\n", err)
	}
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

func main() {
	// var downloadList []string
	conf := getConf()
	pidPath := fmt.Sprintf("%s/cloud-torrent-downloader", conf.PidFilePath)
	pid := pidcheck.AlreadyRunning(pidPath)
	if pid {
		os.Exit(1)
	}

	// Do this for ever and ever
	for {
		// Channel so we can continuously monitor new episodes being added to showrss
		c := make(chan []string)
		var wg sync.WaitGroup

		wg.Add(1)
		go showrss.GetNewEpisodes(conf.ShowRSS, 3000, c, &wg)
		selectedSeedr := conf.GetSeedrInstance()
		for _, magnet := range <-c {
			wg.Add(1)
			go AddMagnet(selectedSeedr, magnet, &wg)
		}

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

		wg.Wait()
		// TODO: remove dev code
		// fmt.Println("sleep....")
		// time.Sleep(time.Second * 3)
		// fmt.Println("awake")
		// 5 min timer
		time.Sleep(time.Second * 300)
	}
}

//TODO: once errors are returned above, this is not needed
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
