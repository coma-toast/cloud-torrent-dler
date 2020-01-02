package main

import (
	"fmt"
	"os"
	"strings"

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
func (s *SeedrAPI) Get(item DownloadItem, destination string) error {
	var err error

	fmt.Printf("Downloading file: %s\n", item.Name)
	if err != nil {
		return err
	}

	// Local parent folder
	path := fmt.Sprintf("%s/%s", destination, item.FolderPath)
	// Make the local parent folder
	os.MkdirAll(path, 0644)
	// Full local path
	path = fmt.Sprintf("%s/%s", path, item.Name)

	err = s.client.DownloadFileByID(item.SeedrID, path)
	if err != nil {
		return err
	}

	return err
}

// FindID gets the SeedrID for a file
func (s *SeedrAPI) FindID(filename string) (int, error) {
	s.populateFolderMapping(0, "")
	for id, name := range s.folderMapping {
		if strings.Contains(name, filename) {
			return id, nil
		}
	}

	return 0, fmt.Errorf("filename not found")
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
