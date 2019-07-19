package main

// Lmo2~C}8fDJ%yj,CpfUv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type Folder struct {
	SpaceMax  int64 `json:"space_max"`
	SpaceUsed int64 `json:"space_used"`
	// Code      int   `json:"code"`
	// Timestamp string        `json:"timestamp"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID int    `json:"parent_id"`
	// Torrents  [] `json:"torrents"`
	Folders []Folder `json:"folders"`
	Files   []File   `json:"files"`
	Result  bool     `json:"result"`
}

type SubFolder struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	LastUpdate string `json:"last_update"`
}

type File struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Size           int64  `json:"size"`
	Hash           string `json:"hash"`
	LastUpdate     string `json:"last_update"`
	StreamAudio    bool   `json:"stream_audio"`
	StreamVideo    bool   `json:"stream_video"`
	VideoConverted string `json:"video_converted,omitempty"`
}

// Client struct is the http Client struct
type Client struct {
	Domain    string
	UserAgent string
	UserName  string
	Password  string
}

// Service is the service
type Service interface {
	// GetCustomer(id int) (Customer, error)
}

func main() {
	files := getFilesFromFolder(0)
	// spew.Dump(files)
	downloadFiles(files)
}

func apiCall(method string, id int, url string, isFolder bool) []byte {
	var baseUrl = "https://www.seedr.cc/rest/"
	if isFolder {
		url = fmt.Sprintf("%s/%s/%d", url, "folder", id)
	} else {
		url = fmt.Sprintf("%s/%d", url, id)
	}
	var username = "jdale215@gmail.com"
	var passwd = "Lmo2~C}8fDJ%yj,CpfUv"
	client := &http.Client{}
	request, err := http.NewRequest(method, url, nil)
	// TODO: pick up from here, making api call a function
	handleError(err)
	request.SetBasicAuth(username, passwd)
	response, err := client.Do(request)
	handleError(err)
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	return data
}

func getFolder(id int) Folder {
	var rootData Folder
	if id != 0 {
		// url := fmt.Sprintf("%s%d", url, id)
	}

	err = json.Unmarshal(data, &rootData)
	handleError(err)
	return rootData
}

func getFilesFromFolder(folderId int) []File {
	folder := getFolder(folderId)
	files := folder.Files
	for _, folder := range folder.Folders {
		subfiles := getFilesFromFolder(folder.ID)
		files = append(files, subfiles...)
	}
	return files
}

func downloadFiles(files []File) {
	for _, file := range files {
		out, err := os.Create(file.Name)
		handleError(err)
		defer out.Close()

		spew.Dump(file)
	}
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
