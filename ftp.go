package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/secsy/goftp"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/helper"
)

// SeedrFTP is the struct for FTP
type SeedrFTP struct {
	Username      string
	Password      string
	client        *goftp.Client
	folderMapping map[int]string
}

// List gets a list of files or folders in a given path
func (s *SeedrFTP) List(path string) ([]DownloadItem, error) {
	if s.client == nil {
		ftpConfig := goftp.Config{
			User:               s.Username,
			Password:           s.Password,
			ConnectionsPerHost: 10,
			Timeout:            10 * time.Second,
			Logger:             os.Stderr,
		}
		var err error
		s.client, err = goftp.DialConfig(ftpConfig, "ftp.seedr.cc")
		if err != nil {
			return []DownloadItem{}, err
		}

	}

	folderList, err := s.client.ReadDir(path)
	if err != nil {
		return []DownloadItem{}, err
	}

	returnData := []DownloadItem{}

	for _, file := range folderList {
		name := helper.SanitizeText(file.Name())
		var cacheData DownloadItem
		if cache.IsSet(name) {
			cacheData = cache.Get(name)
		}
		appendData := DownloadItem{
			EpisodeID:  cacheData.EpisodeID,
			ShowID:     cacheData.ShowID,
			SeedrID:    cacheData.SeedrID,
			Name:       file.Name(),
			FolderPath: file.Name(),
		}
		returnData = append(returnData, appendData)
	}

	return returnData, nil
}

// Get downloads the files in the provided array of files
func (s *SeedrFTP) Get(file DownloadItem, destination string) error {
	var err error
	if s.client == nil {
		ftpConfig := goftp.Config{
			User:               s.Username,
			Password:           s.Password,
			ConnectionsPerHost: 10,
			Timeout:            10 * time.Second,
			Logger:             os.Stderr,
		}
		s.client, err = goftp.DialConfig(ftpConfig, "ftp.seedr.cc")
	}
	// TODO: remove dev code
	// isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi)$", file)
	isAVideo, _ := regexp.MatchString("(.*?).(jpg|iso)$", file.Name)
	if isAVideo {
		filePathArray := strings.Split(file.Name, "/")
		folder := filePathArray[len(filePathArray)-2]
		if folder == filePathArray[0] {
			folder = "Files"
		}
		filename := filePathArray[len(filePathArray)-1]
		log.Println("Downloading file " + filename)
		err := os.Mkdir(destination+"/"+folder, 0777)
		destFile, err := os.Create(destination + "/" + folder + "/" + filename)
		if err != nil {
			log.Println("Error creating destination file: ", err)
			return err
		}
		s.client.Retrieve(file.Name, destFile)
	}
	return err
}

// Add doesn't add - FTP add unsupported
func (s *SeedrFTP) Add(magnet string) error {
	err := fmt.Errorf("No add function for FTP.")
	return err
}

// DeleteFolder doesn't work yet.
func (s *SeedrFTP) DeleteFolder(id int) error {
	return nil
}

// DeleteFile doesn't work yet.
func (s *SeedrFTP) DeleteFile(id int) error {
	return nil
}

// FindID is unsupported
func (s *SeedrFTP) FindID(filename string) (int, error) {
	err := fmt.Errorf("Finding Seedr ID does not work on FTP")
	return 0, err
}

// TODO: folderMapping function for FTP
// GetPath gets the file path
func (s *SeedrFTP) GetPath(queryID int) (string, error) {
	var err error

	for id, pathName := range s.folderMapping {
		if id == queryID {
			return pathName, err
		}
	}

	err = fmt.Errorf("ID not found: %d", queryID)

	return "", err
}
