package showrss

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetShows gets all episodes
func GetShows(url string) (Shows, error) {
	var result Channel
	var err error

	xmlBytes, err := getXML(url)
	if err != nil {
		return Shows{}, err
	}

	if err = xml.Unmarshal(xmlBytes, &result); err != nil {
		return Shows{}, err
	}

	return result.Shows, nil
}

// GetAllEpisodeLinks is a loop to look for new shows added to the RSS feed to then add to the download queue
func GetAllEpisodeLinks(url string) (map[int]string, error) {
	initState, err := GetShows(url)
	if err != nil {
		return nil, err
	}
	returnData := make(map[int]string)
	for _, item := range initState.Item {
		returnData[item.TVEpisodeID] = item.Link
	}
	return returnData, nil
}

func getXML(url string) ([]byte, error) {
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
