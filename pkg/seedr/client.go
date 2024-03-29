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

	log "github.com/sirupsen/logrus"
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

// Result is the API Result
type Result struct {
	Result          bool   `json:"result"`
	Code            int    `json:"code"`
	User_torrent_id int    `json:"user_torrent_id"`
	Title           string `json:"title"`
	Torrent_hash    string `json:"torrent_hash"`
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
	if err != nil {
		return &http.Response{}, err
	}

	request.Header.Set("Authorization", "Basic "+c.credentials)
	request.Header.Add("content-type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW")
	resp, err := c.client.Do(request)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return &http.Response{}, errors.New("HTTP error occurred. Status code: " + strconv.Itoa(resp.StatusCode))
	}

	log.WithField("status code", resp.StatusCode).Info("HTTP request sent")

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
			if strings.HasPrefix(fmt.Sprintf("%v", payload), "magnet:?") {
				postData = fmt.Sprintf("------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"magnet\"\r\n\r\n%s\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--", payload)
			} else {
				postData = fmt.Sprintf("------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"torrent_url\"\r\n\r\n%s\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--", payload)
			}
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
		return responseBody, err
	}

	return responseBody, nil
}
