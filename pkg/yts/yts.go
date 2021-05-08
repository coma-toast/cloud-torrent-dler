package yts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetMovies(url string) ([]Movie, error) {
	var result YTSResult
	var err error

	jsonBytes, err := getJSON(url)
	if err != nil {
		return []Movie{}, err
	}

	if err = json.Unmarshal(jsonBytes, &result); err != nil {
		return []Movie{}, err
	}

	return result.Data.Movies, nil
}

func getJSON(url string) ([]byte, error) {
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
