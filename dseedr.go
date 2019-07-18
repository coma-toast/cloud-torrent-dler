package main

// Lmo2~C}8fDJ%yj,CpfUv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	space_max       int
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

// Client struct is the http Client
type Client struct {
	// TODO: figure out this (aka increase my Go knowledge)
	// BaseURL    *url.URL
	Domain     string
	UserAgent  string
	UserName   string
	Password   string
	httpClient *http.Client
}

// Service is the service
type Service interface {
	// GetCustomer(id int) (Customer, error)
}

func main() {
	response := callAndUnmarshal("GET", "folder", response)
	fmt.Printf(response)
	// response, err := http.Get("https://www.seedr.cc/rest/folder/")
	// if err != nil {
	// 	fmt.Printf("HTTP request failed with error %s\n", err)
	// } else {
	// 	data, _ := ioutil.ReadAll(response.Body)
	// 	fmt.Println(string(data))
	// }

	// jsonData := map[string]string{"username": "jdale215@gmail.com", "password": "Lmo2~C}8fDJ%yj,CpfUv", "id": "95429980"}
	// jsonValue, _ := json.Marshal(jsonData)
	// response, err = http.Post("https://www.seedr.cc/rest/folder/", "application/json", bytes.NewBuffer(jsonValue))
	// if err != nil {
	// 	fmt.Printf("HTTP request failed with error %s\n", err)
	// } else {
	// 	data, _ := ioutil.ReadAll(response.Body)
	// 	fmt.Println(string(data))
	// }
}

func (c Client) call(method string, url string) ([]byte, error) {
	fullURL := fmt.Sprintf("https://%s%s", c.Domain, url)

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return []byte{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func (c Client) callAndUnmarshal(method string, url string, target interface{}) error {
	response, err := c.call(method, url)
	if err != nil {
		return err
	}

	err = json.Unmarshal(response, target)
	if err != nil {
		return err
	}

	return nil
}

// ? From a tutorial. Not sure how useful this is...
func (c *Client) ListFolder() ([]Folder, error) {
	req, err := c.newRequest("GET", "/folder", nil)
	if err != nil {
		return nil, err
	}
	var folder []Folder
	_, err = c.do(req, &folder)
	return folder, err
}

func (c *Client) ListFiles(id int) ([]File, error) {
	path := fmt.Sprintf("/folder/%d", id)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var file []File
	_, err = c.do(req, &file)
	return file, err
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
