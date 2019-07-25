package main

// Lmo2~C}8fDJ%yj,CpfUv

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/davecgh/go-spew/spew"
)

// Folder is a folder that contains subfolders or files, or both
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

var BaseURL = "https://www.seedr.cc/rest"
var DlRoot = "./tmp"
var Username = "jdale215@gmail.com"
var Passwd = "Lmo2~C}8fDJ%yj,CpfUv"
var Credentials = b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", Username, Passwd)))

func main() {
	files := getFilesFromFolder(0)
	// spew.Dump(files)
	downloadFiles(files)
}

func apiCall(method string, id int, callType string) []byte {
	url := fmt.Sprintf("%s/%s", BaseURL, callType)
	if id != 0 {
		url = fmt.Sprintf("%s/%d", url, id)
	}
	client := &http.Client{}
	request, err := http.NewRequest(method, url, nil)
	handleError(err)
	request.SetBasicAuth(Username, Passwd)
	response, err := client.Do(request)
	handleError(err)
	data, err := ioutil.ReadAll(response.Body)
	handleError(err)
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
		isAVideo, _ := regexp.MatchString("(.*?).(jpg|txt)$", file.Name)
		if isAVideo {
			fmt.Println("Downloading file: " + file.Name)
			path := fmt.Sprintf("%s/%s", DlRoot, file.Name)
			fileURL := fmt.Sprintf("%s/file/%d", BaseURL, file.ID)
			out, err := os.Create(path)
			handleError(err)
			defer out.Close()

			client := grab.NewClient()
			// client.HTTPClient.Transport.DisableCompression = true

			req, err := grab.NewRequest(path, fileURL)
			handleError(err)
			// ...
			req.NoResume = true
			req.HTTPRequest.Header.Set("Authorization", "Basic "+Credentials)
			resp := client.Do(req)
			spew.Dump(resp)

			// progress bar
			t := time.NewTicker(time.Second)
			defer t.Stop()

			for {
				select {
				case <-t.C:
					fmt.Printf("%.02f%% complete\n", resp.Progress())

				case <-resp.Done:
					if err := resp.Err(); err != nil {
						handleError(err)
					}

				}
			}
			fmt.Println("Download complete.")
		} else {
		}
	}
	return
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
