package seedr

import (
	"fmt"
	"io/ioutil"

	"github.com/davecgh/go-spew/spew"
)

// Service is the service
type Service interface {
	GetFolder(id int) (Folder, error)
	GetFile(id int) (File, error)
	DownloadFileByID(id int, destination string) error
	AddMagnet(magnet string) error
}

// GetFolder gets a Folder from Seedr
func (c Client) GetFolder(id int) (Folder, error) {
	var folder Folder
	url := "/folder"
	if id != 0 {
		url = fmt.Sprintf("%s/%d", url, id)
	}
	_, err := c.call("GET", url, nil, &folder)
	if err != nil {
		return Folder{}, err
	}
	return folder, nil

}

// GetFile gets a File from Seedr
func (c Client) GetFile(id int) (File, error) {
	var file File
	url := fmt.Sprintf("/%d", id)
	_, err := c.call("GET", url, nil, &file)
	if err != nil {
		return File{}, err
	}
	return file, nil
}

// DownloadFileByID downloads a file with the specified ID
func (c Client) DownloadFileByID(id int, destination string) error {
	var err error
	url := fmt.Sprintf("/file/%d", id)
	err = c.downloadFile(url, destination)
	if err != nil {
		return err
	}

	return err
}

func (c Client) downloadFile(url string, destination string) error {
	var err error
	responseBytes, err := c.call("GET", url, nil, nil)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(destination, responseBytes, 0644)
}

//AddMagnet adds a magnet link to Seedr to be downloaded
func (c Client) AddMagnet(magnet string) error {
	var err error
	url := fmt.Sprintf("/torrent/magnet")
	result, err := c.call("POST", url, magnet, nil)
	spew.Dump("Result: ", result)
	return err
}
