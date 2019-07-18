package main

// Lmo2~C}8fDJ%yj,CpfUv

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

type RootFolder struct {
	space_max  int
	space_used int
	code       int
	timestamp  string
	id         int
	name       string
	parent_id  int
	torrents   []string
	folders    []Folder
	files      []File
	result     bool
}

type Folder struct {
	"space max"       int
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
	id           int
	name         string
	size         int
	hash         string
	last_update  string
	stream_audio bool
	stream_video bool
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
	root := getRootFolder()
	spew.Dump(root)
}

func getRootFolder() RootFolder {
	var username string = "jdale215@gmail.com"
	var passwd string = "Lmo2~C}8fDJ%yj,CpfUv"
	jsonData := RootFolder{}
	client := &http.Client{}
	request, err := http.NewRequest("GET", "https://www.seedr.cc/rest/folder/", nil)
	if err != nil {
		spew.Dump(err)
	}
	request.SetBasicAuth(username, passwd)
	response, err := client.Do(request)
	data, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(data, &jsonData)
	return jsonData
}
