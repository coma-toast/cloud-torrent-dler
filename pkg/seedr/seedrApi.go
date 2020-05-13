package seedr

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
)

// Service is the service
type Service interface {
	GetFolder(id int) (Folder, error)
	GetFile(id int) (File, error)
	DeleteFile(id int) error
	DeleteFolder(id int) error
	DownloadFileByID(id int, destination string) error
	AddMagnet(magnet string) error
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
	err = c.downloadFile(url, destination)
	if err != nil {
		return err
	}

	return err
}

func (c Client) downloadFile(url string, destination string) error {
	var err error
	f, err := os.Create(destination)
	defer f.Close()
	if err != nil {
		spew.Dump("downloadFile() error: ", err)
		return err
	}
	response, err := c.stream(http.MethodGet, url, nil)
	defer response.Body.Close()
	if err != nil {
		spew.Dump("downloadFile() error: ", err)
		return err
	}

	bytes, err := io.Copy(f, response.Body)
	if err != nil {
		spew.Dump("downloadFile() error: ", err)
		return err
	}
	spew.Dump("bytes written: :", bytes)

	return err
}

//AddMagnet adds a magnet link to Seedr to be downloaded
func (c Client) AddMagnet(magnet string) error {
	var err error
	url := fmt.Sprintf("/torrent/magnet")
	result, err := c.call(http.MethodPost, url, magnet, nil)
	spew.Dump("Result: ", result)
	return err
}
