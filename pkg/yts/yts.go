package yts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var URL = "https://yts.mx/api/v2/list_movies.json?quality=2160p"

func GetMovies(url string) ([]Movie, error) {
	var result YTSResult
	var err error

	jsonBytes, err := getJSON(url, "", 1)
	if err != nil {
		return []Movie{}, err
	}

	if err = json.Unmarshal(jsonBytes, &result); err != nil {
		return []Movie{}, err
	}

	return result.Data.Movies, nil
}

func SearchMovies(search string, page int) (YTSData, error) {
	var result YTSResult
	jsonBytes, err := getJSON(URL, search, page)
	if err != nil {
		return YTSData{}, err
	}

	if err = json.Unmarshal(jsonBytes, &result); err != nil {
		return YTSData{}, err
	}

	return result.Data, nil

}

func getJSON(url string, query string, page int) ([]byte, error) {
	if query != "" {
		url = fmt.Sprintf("%s&query_term=%s", url, query)
	}
	url = fmt.Sprintf("%s&page=%d", url, page)
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}
