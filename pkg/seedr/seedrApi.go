package seedr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

// Service is the service
type Service interface {
	GetFolder(id int) (Folder, error)
	GetFile(id int) (File, error)
	DeleteFile(id int) error
	DeleteFolder(id int) error
	DownloadFileByID(id int, destination string) error
	AddMagnet(magnet string) (Result, error)
	AddTorrent(magnet string) (Result, error)
}

// DeleteFolder deletes a folder from Seedr
func (c Client) DeleteFolder(id int) error {
	url := fmt.Sprintf("/folder/%d", id)
	_, err := c.call(http.MethodDelete, url, nil, nil)
	return err
}

// DeleteFile deletes a file from Seedr
func (c Client) DeleteFile(id int) error {
	url := fmt.Sprintf("/file/%d", id)
	_, err := c.call(http.MethodDelete, url, nil, nil)
	return err
}

// GetFolder gets a Folder from Seedr
func (c Client) GetFolder(id int) (Folder, error) {
	var folder Folder
	url := "/folder"
	if id != 0 {
		url = fmt.Sprintf("%s/%d", url, id)
	}
	_, err := c.call(http.MethodGet, url, nil, &folder)
	if err != nil {
		return Folder{}, err
	}
	return folder, nil

}

// GetFile gets a File from Seedr
func (c Client) GetFile(id int) (File, error) {
	var file File
	url := fmt.Sprintf("/%d", id)
	_, err := c.call(http.MethodGet, url, nil, &file)
	if err != nil {
		return File{}, err
	}
	return file, nil
}

// DownloadFileByID downloads a file with the specified ID
func (c Client) DownloadFileByID(id int, destination string) error {
	var err error
	url := fmt.Sprintf("/file/%d", id)
	spew.Dump(id, url, destination)
	err = c.downloadFile(url, destination)
	if err != nil {
		return err
	}

	return err
}

func (c Client) downloadFile(url string, destination string) error {
	var err error
	f, err := os.Create(destination)
	if err != nil {
		spew.Dump("Error in downloadFile() creating destination file: ", err)
		return err
	}
	defer f.Close()
	response, err := c.stream(http.MethodGet, url, nil)
	if err != nil {
		log.WithField("error", err).Debug("Error getting stream response")
		return err
	}
	defer response.Body.Close()
	if err != nil {
		spew.Dump("Error in downloadFile() getting stream: ", err)
		return err
	}

	bytes, err := io.Copy(f, response.Body)
	if err != nil {
		spew.Dump("Error in downloadFile() copying to destination: ", err)
		return err
	}
	spew.Dump("bytes written: :", bytes)

	return err
}

//AddMagnet adds a magnet link to Seedr to be downloaded
func (c Client) AddMagnet(magnet string) (Result, error) {
	var err error
	var resultData Result
	url := fmt.Sprintf("/torrent/magnet")
	result, err := c.call(http.MethodPost, url, magnet, nil)
	if err != nil {
		log.WithError(err)
	}

	log.WithField("data", result).Debug("POST request with magnet data sent successfully")
	err = json.Unmarshal(result, &resultData)
	log.WithField("data", resultData).Debug("resultData")

	return resultData, err
}

//AddTorrent adds a torrent URL to Seedr to be downloaded
func (c Client) AddTorrent(torrentUrl string) (Result, error) {
	var err error
	var resultData Result
	url := fmt.Sprintf("/torrent/url")
	result, err := c.call(http.MethodPost, url, torrentUrl, nil)
	if err != nil {
		log.WithError(err)
	}

	log.WithField("data", result).Debug("POST request with torrent url sent successfully")
	err = json.Unmarshal(result, &resultData)
	log.WithField("data", resultData).Debug("resultData")

	return resultData, err
}
