package showrss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	// log "github.com/sirupsen/logrus"
)

type ShowRSSClient struct {
	userId      string
	userFeedUrl string
	showUrl     string
}

func (showRss *ShowRSSClient) Init(userId string) {
	showRss.userId = userId
	showRss.showUrl = "https://showrss.info/show/"
	showRss.userFeedUrl = fmt.Sprintf("https://showrss.info/user/%s.rss?magnets=true&namespaces=true&name=null&quality=null&re=null", showRss.userId)
}

func (showRss *ShowRSSClient) GetAllAvailableShows() ([]Shows, error) {
	showId := 1
	shows := []Shows{}
	success := true
	for success {
		show, err := showRss.GetShowByID(strconv.Itoa(showId))
		if err != nil {
			success = false
			break
		}
		shows = append(shows, show)
		showId++
	}
	return shows, nil
}

// GetShows gets all episodes
func (showRss *ShowRSSClient) GetLatestSubscribedShows() (Shows, error) {
	var result Channel
	var err error

	xmlBytes, err := getXML(showRss.userFeedUrl)
	if err != nil {
		return Shows{}, err
	}

	if err = xml.Unmarshal(xmlBytes, &result); err != nil {
		return Shows{}, err
	}

	return result.Shows, nil
}

func (showRss *ShowRSSClient) GetShowByID(showId string) (Shows, error) {
	shows := Shows{}
	url := fmt.Sprintf("%s%d.rss", showRss.showUrl, showId)
	xmlBytes, err := getXML(url)
	if err != nil {
		return Shows{}, err
	}

	if err = xml.Unmarshal(xmlBytes, &shows); err != nil {
		return Shows{}, err
	}

	return shows, nil
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}
