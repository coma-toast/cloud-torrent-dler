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
	spew.Dump("files", files)
	// downloadFiles(files)
}

func apiCall(method string, id int, callType string) []byte {
	var username = "jdale215@gmail.com"
	var passwd = "Lmo2~C}8fDJ%yj,CpfUv"
	var baseURL = "https://www.seedr.cc/rest/"
	url := fmt.Sprintf("%s/%s", baseURL, callType)
	if id != 0 {
		url = fmt.Sprintf("%s/%d", url, id)
	}
	client := &http.Client{}
	request, err := http.NewRequest(method, url, nil)
	handleError(err)
	request.SetBasicAuth(username, passwd)
	response, err := client.Do(request)
	handleError(err)
	data, _ := ioutil.ReadAll(response.Body)
	spew.Dump(data)
	defer response.Body.Close()
	return data
}

func getFolder(id int) Folder {
	var rootData Folder
	data := apiCall("GET", id, "folder")
	err := json.Unmarshal(data, &rootData)
	handleError(err)
	return rootData
}

func getFilesFromFolder(folderID int) []File {
	folder := getFolder(folderID)
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
