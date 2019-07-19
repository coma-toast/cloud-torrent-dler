package main

// Lmo2~C}8fDJ%yj,CpfUv

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

type RootFolder struct {
	SpaceMax  int64         `json:"space_max"`
	SpaceUsed int64         `json:"space_used"`
	Code      int           `json:"code"`
	Timestamp string        `json:"timestamp"`
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	ParentID  int           `json:"parent_id"`
	Torrents  []interface{} `json:"torrents"`
	Folders   []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Size       int64  `json:"size"`
		LastUpdate string `json:"last_update"`
	} `json:"folders"`
	Files  []interface{} `json:"files"`
	Result bool          `json:"result"`
}

type Folder struct {
	Space_max       int
	space_used      int
	code            int
	timestamp       string
	id              int
	name            string
	parent_id       int
	torrents        []string
	folders         []Folder
	files           []File
	video_converted string
	result          bool
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
	// response := callAndUnmarshal("GET", "folder", response)
	// root := getRootFolder()
	// spew.Dump(root)
	getRootFolder()
}

func getRootFolder() {
	var username string = "jdale215@gmail.com"
	var passwd string = "Lmo2~C}8fDJ%yj,CpfUv"
	client := &http.Client{}
	request, err := http.NewRequest("GET", "https://www.seedr.cc/rest/folder/", nil)
	if err != nil {
		spew.Dump(err)
	}
	request.SetBasicAuth(username, passwd)
	response, err := client.Do(request)
	if err != nil {
		spew.Dump(err)
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	spew.Dump(data)
	var rootData RootFolder
	var test interface{}
	// var json string
	err = json.Unmarshal(data, &rootData)
	handleError(err)
	json.Unmarshal(data, &test)
	spew.Dump(rootData)
	spew.Dump(test)
	// return jsonData
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
