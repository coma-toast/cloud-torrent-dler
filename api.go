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
func (s *SeedrAPI) List(path string) ([]DownloadItem, error) {
	folderList := []DownloadItem{}
	if s.client == nil {
		s.client = &seedr.Client{
			Username: s.Username,
			Password: s.Password,
		}
	}

	if s.folderMapping == nil {
		s.folderMapping = make(map[int]string)
		err := s.populateFolderMapping(0, "")
		if err != nil {
			return []DownloadItem{}, err
		}
	}

	folderID, err := s.getFolderIDFromPath(path)
	if err != nil {
		err = s.populateFolderMapping(0, "")
		if err != nil {
			return []DownloadItem{}, err
		}
		folderID, err = s.getFolderIDFromPath(path)
		if err != nil {
			return []DownloadItem{}, err
		}
	}

	files, err := s.client.GetFolder(folderID)
	if err != nil {
		return []DownloadItem{}, err
	}

	for _, folder := range files.Folders {
		name := folder.Name()
		var cacheData DownloadItem
		if cache.IsSet(name) {
			cacheData = cache.Get(name)
		}
		appendData := DownloadItem{
			EpisodeID:     cacheData.EpisodeID,
			FolderPath:    folder.SubFolderName,
			IsDir:         folder.IsDir(),
			Name:          folder.Name(),
			ParentSeedrID: folderID,
			SeedrID:       cacheData.SeedrID,
			ShowID:        cacheData.ShowID,
		}
		folderList = append(folderList, appendData)
	}
	for _, file := range files.Files {
		name := file.Name()
		id, err := s.getFileIDFromPath(name)
		// fmt.Println(id, name)
		if err != nil {
			fmt.Println(err)
		}
		var cacheData DownloadItem
		if cache.IsSet(name) {
			cacheData = cache.Get(name)
		}

		appendData := DownloadItem{
			EpisodeID:     cacheData.EpisodeID,
			FolderPath:    file.FileName,
			IsDir:         file.IsDir(),
			Name:          file.Name(),
			ParentSeedrID: folderID,
			SeedrID:       id,
			ShowID:        cacheData.ShowID,
		}
		folderList = append(folderList, appendData)
		s.folderMapping[file.ID] = file.FileName
	}

	return folderList, nil
}

// Get downloads the file name
func (s *SeedrAPI) Get(item DownloadItem, destination string) error {
	var err error
	fmt.Printf("Downloading item: %s to %s\n", item.Name, destination+"/"+item.FolderPath)
	// Local parent folder
	path := fmt.Sprintf("%s/%s", destination, item.FolderPath)
	// Make the local parent folder
	os.MkdirAll(path, 0777)
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

// DeleteFile deletes a file from Seedr
func (s *SeedrAPI) DeleteFile(id int) error {
	return s.client.DeleteFile(id)
}

// DeleteFolder deletes a folder from Seedr
func (s *SeedrAPI) DeleteFolder(id int) error {
	return s.client.DeleteFolder(id)
}

func (s *SeedrAPI) getFolderIDFromPath(path string) (int, error) {
	var err error

	if s.folderMapping == nil {
		s.folderMapping = make(map[int]string)
		err := s.populateFolderMapping(0, "")
		if err != nil {
			return 0, err
		}
	}

	for id, pathName := range s.folderMapping {
		if pathName == path {
			return id, err
		}
	}

	err = fmt.Errorf("Path not found: %s", path)

	return 0, err
}

func (s *SeedrAPI) getFileIDFromPath(path string) (int, error) {
	var err error

	if s.folderMapping == nil {
		s.folderMapping = make(map[int]string)
		err := s.populateFolderMapping(0, "")
		if err != nil {
			return 0, err
		}
	}

	for id, pathName := range s.folderMapping {
		if strings.Contains(pathName, path) {
			return id, err
		}
	}

	err = fmt.Errorf("Path not found: %s", path)

	return 0, err
}

func (s *SeedrAPI) populateFolderMapping(ID int, path string) error {
	folder, err := s.client.GetFolder(ID)
	// folder.FolderName = helper.SanitizeText(folder.FolderName)

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
			// file.FileName = helper.SanitizeText(file.FileName)
			s.folderMapping[file.ID] = folder.FolderName + "/" + file.FileName
		}
	}

	return nil
}
