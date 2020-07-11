package seedr

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// Client is the Seedr Client
type Client struct {
	client      *http.Client
	Username    string
	Password    string
	credentials string
}

// Error is the error
type Error struct {
	Result       bool   `json:"result"`
	ErrorText    string `json:"error"`
	CallResponse *http.Response
}

var baseURL = "https://www.seedr.cc/rest"

func (e Error) Error() string {
	message := fmt.Sprintf("Seedr API Error: %s \n Status Code: %d", e.ErrorText, e.CallResponse.StatusCode)
	return message
}

func (c *Client) stream(method string, url string, payload interface{}) (*http.Response, error) {
	if c.credentials == "" {
		// credentials is the base64 encoded username:password
		c.credentials = b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.Username, c.Password)))
	}
	url = baseURL + url
	if c.client == nil {
		// ! An error that keeps happening in production I believe is here. Or at least this function
		// Tick...
		// Getting vampirina s02e22 hdtv x264 w4f from cache
		// Deleting item: RARBG.txt
		// Deleting item: RARBG_DO_NOT_MIRROR.exe
		// Getting vampirina s02e22 hdtv x264 w4f from cache
		// Downloading item: vampirina.s02e22.hdtv.x264-w4f.mkv to /media/USB-Drive/Shows/Kids/Vampirina S02E22 HDTV x264 W4F
		// panic: runtime error: invalid memory address or nil pointer dereference
		// [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x70fcd8]

		// goroutine 1 [running]:
		// gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/seedr.Client.downloadFile(0xc0001642d0, 0xc0000bb5e0, 0x12, 0xc0000bb640, 0x14, 0xc000336440, 0x34, 0xc0003941b0, 0xf, 0xc000086180, ...)
		//         /home/jason/git/cloud-torrent-dler/pkg/seedr/seedrApi.go:83 +0x208
		// gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/seedr.Client.DownloadFileByID(0x0, 0xc0000bb5e0, 0x12, 0xc0000bb640, 0x14, 0x0, 0x0, 0x29d2ac50, 0xc000086180, 0x5d, ...)
		//         /home/jason/git/cloud-torrent-dler/pkg/seedr/seedrApi.go:66 +0xeb
		// main.(*SeedrAPI).Get(0xc0001641e0, 0x0, 0xc0002e0c00, 0x29, 0x0, 0xc0002e0de0, 0x22, 0x81c110e, 0x29d2ac50, 0x0, ...)
		//         /home/jason/git/cloud-torrent-dler/api.go:109 +0x38e
		// main.main()
		//         /home/jason/git/cloud-torrent-dler/main.go:139 +0xb98
		c.client = &http.Client{}
	}
	// TODO: throw error if >= than 400
	var err error
	var postData string

	if payload != nil {
		if method == http.MethodPost {
			postData = fmt.Sprintf("------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"magnet\"\r\n\r\n%s\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--", payload)
		}
	}
	request, err := http.NewRequest(method, url, strings.NewReader(postData))
	request.Header.Set("Authorization", "Basic "+c.credentials)
	request.Header.Add("content-type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW")
	resp, err := c.client.Do(request)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return &http.Response{}, errors.New("HTTP error occurred. Status code: " + strconv.Itoa(resp.StatusCode))
	}
	spew.Dump(resp.StatusCode)

	return resp, err
}

func (c *Client) call(method string, url string, payload interface{}, target interface{}) ([]byte, error) {
	if c.credentials == "" {
		// credentials is the base64 encoded username:password
		c.credentials = b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.Username, c.Password)))
	}
	url = baseURL + url
	if c.client == nil {
		c.client = &http.Client{}
	}
	// TODO: throw error if >= than 400
	var err error
	var postData string

	if payload != nil {
		if method == http.MethodPost {
			postData = fmt.Sprintf("------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"magnet\"\r\n\r\n%s\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--", payload)
		}
	}
	request, err := http.NewRequest(method, url, strings.NewReader(postData))
	request.Header.Set("Authorization", "Basic "+c.credentials)
	request.Header.Add("content-type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW")
	if err != nil {
		return []byte{}, err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	//TODO: this can all be one error function, take responseBody and do all the error checks
	errorTarget := Error{}

	if target != nil {
		err = json.Unmarshal(responseBody, &target)
		if err != nil {
			return responseBody, err
		}
	}

	if errorTarget.ErrorText != "" {
		errorTarget.CallResponse = resp
		return responseBody, errorTarget
	}
	// TODO: ^ to here
	if resp.StatusCode >= 400 {
		err := fmt.Errorf("Seedr HTTP Error: %d", resp.StatusCode)
		spew.Dump(resp.Status)
		return responseBody, err
	}

	return responseBody, nil
}
