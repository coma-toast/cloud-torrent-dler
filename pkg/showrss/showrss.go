package showrss

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
)

// GetShows gets all episodes
func GetShows(url string) (Shows, error) {
	var result Channel
	var err error
	if xmlBytes, err := getXML(url); err != nil {
		log.Printf("Failed to get XML: %v", err)
	} else {
		xml.Unmarshal(xmlBytes, &result)
	}
	return result.Shows, err
}

// GetNewEpisodes is a loop to look for new shows added to the RSS feed to then add to the download queue
func GetNewEpisodes(url string, interval int, c chan []string, wg *sync.WaitGroup) {
	defer wg.Done()
	initState, _ := GetShows(url)
	initEpisodes := []int{}
	magnetsToAdd := []string{}
	for _, item := range initState.Item {
		initEpisodes = append(initEpisodes, item.TVEpisodeID)
	}
	sort.Ints(initEpisodes)
	fmt.Println("Loop start")
	currentState, _ := GetShows(url)
	for _, currentListItem := range currentState.Item {
		id := sort.SearchInts(initEpisodes, currentListItem.TVEpisodeID)
		// TODO: remove dev code
		if id > 0 {
			// if id >= 0 {
			// fmt.Println("Already exists, skipping...")
			continue
		}
		initEpisodes = append(initEpisodes, currentListItem.TVEpisodeID)
		magnetsToAdd = append(magnetsToAdd, currentListItem.Link)
	}
	c <- magnetsToAdd
	// spew.Dump(initState)
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
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
