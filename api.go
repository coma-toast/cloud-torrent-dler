package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/seedr"
)

// SeedrAPI is the struct for API
type SeedrAPI struct {
	Username      string
	Password      string
	client        *seedr.Client
	folderMapping map[int]string
}

// List gets a list of files or folders
func (s *SeedrAPI) List(path string) ([]os.FileInfo, error) {
	folderList := []os.FileInfo{}
	if s.client == nil {
		s.client = &seedr.Client{
			Username: s.Username,
			Password: s.Password,
		}
	}

	if s.folderMapping == nil {
		s.folderMapping = make(map[int]string)
		err := s.populateFolderMapping(0, "")
		// err := s.populateFolderMapping(23577)
		if err != nil {
			return []os.FileInfo{}, err
		}
	}

	folderID, err := s.getFolderIDFromPath(path)
	if err != nil {
		err = s.populateFolderMapping(0, "")
		if err != nil {
			return []os.FileInfo{}, err
		}
		folderID, err = s.getFolderIDFromPath(path)
		if err != nil {
			return []os.FileInfo{}, err
		}
	}

	files, err := s.client.GetFolder(folderID)
	if err != nil {
		return []os.FileInfo{}, err
	}

	for _, folder := range files.Folders {
		folderList = append(folderList, folder)
	}
	for _, file := range files.Files {
		folderList = append(folderList, file)
		s.folderMapping[file.ID] = file.FileName
	}

	return folderList, nil
}

// Get downloads the file name
func (s *SeedrAPI) Get(file string, destination string) error {
	var err error
	var downloadID = 0
	var folderLength = 0

	// TODO: figure out why the folderMapping is incorrect if you don't run this again
	err = s.populateFolderMapping(0, "")
	if err != nil {
		return err
	}
	spew.Dump(s.folderMapping)

	// * dev code
	// isAVideo, _ := regexp.MatchString("(.*?).(txt|jpg)$", file)
	isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi|m4v)$", file)
	if isAVideo {
		fmt.Printf("Downloading file: %s\n", file)

		for id, name := range s.folderMapping {
			if strings.Contains(name, file) {
				fmt.Println("found", file, id, name, folderLength)
				if len(name) > folderLength {
					downloadID = id
					folderLength = len(name)
					fmt.Printf("ID for %s is %d\n", file, downloadID)
				}
			}
		}
		if downloadID != 0 {
			fmt.Printf("DownloadFileByID(%d), file: %s\n", downloadID, file)
			if err != nil {
				return err
			}
			seedrPath, err := s.GetPath(downloadID)
			fmt.Println("seedrPath " + seedrPath)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("%s/%s/%s", destination, seedrPath, file)

			fmt.Println(path)
			os.Exit(2)
			// TODO: make subfolders. file.ParentFolder (int) -> getFolderFromID + path
			err = s.client.DownloadFileByID(downloadID, path)
			if err != nil {
				return err
			}
		}
	}

	return err
}

// Add adds a magnet link
func (s *SeedrAPI) Add(magnet string) error {
	if s.client == nil {
		s.client = &seedr.Client{
			Username: s.Username,
			Password: s.Password,
		}
	}
	err := s.client.AddMagnet(magnet)
	if err != nil {
		return err
	}

	return nil
}

// TODO: make sure we download the longest file name - folders may be named the same, or there may be duplicates

// GetPath returns the full path of the file in Seedr, for path replication locally
func (s *SeedrAPI) GetPath(queryID int) (string, error) {
	fmt.Println("queryID ", queryID)
	var err error
	if s.folderMapping == nil {
		s.folderMapping = make(map[int]string)
		err := s.populateFolderMapping(0, "")
		if err != nil {
			return "not found", err
		}
	}

	pathName := s.folderMapping[queryID]
	if len(pathName) == 0 {
		err = fmt.Errorf("folder Path lookup error - queryID not found: %d", queryID)
	}

	return pathName, err
}

func (s *SeedrAPI) getFolderIDFromPath(path string) (int, error) {
	var err error

	for id, pathName := range s.folderMapping {
		if pathName == path {
			return id, err
		}
	}

	err = fmt.Errorf("Path not found: %s", path)

	return 0, err
}

func (s *SeedrAPI) populateFolderMapping(ID int, path string) error {
	folder, err := s.client.GetFolder(ID)

	if err != nil {
		return err
	}

	if _, ok := s.folderMapping[folder.ID]; !ok {
		if ID == 0 {
			s.folderMapping[folder.ID] = folder.FolderName
		} else {
			s.folderMapping[folder.ID] = path + folder.FolderName
		}
	}

	if len(folder.Folders) > 0 {
		for _, subfolder := range folder.Folders {
			err := s.populateFolderMapping(subfolder.ID, "")
			if err != nil {
				return err
			}
		}
	}

	if len(folder.Files) > 0 {
		for _, file := range folder.Files {
			s.folderMapping[file.ID] = folder.FolderName + "/" + file.FileName
		}
	}

	return nil
}
