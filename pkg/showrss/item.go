package showrss

// Item is each individual item - each episode
type Item struct {
	ID           int    `gorm:"primaryKey"`
	GUID         string `xml:"guid"`
	Title        string `xml:"title"`
	Link         string `xml:"link"`
	PubDate      string `xml:"pubDate"`
	Description  string `xml:"description"`
	TVShowID     int    `xml:"show_id"`
	TVExternalID int    `xml:"external_id"`
	TVShowName   string `xml:"show_name"`
	TVEpisodeID  int    `xml:"episode_id"`
	TVRawTitle   string `xml:"raw_title"`
	TVInfoHash   string `xml:"info_hash"`
	EnclosureURL string `xml:"enclosure url"`
}

// ItemTitle returns the Title
func (i Item) ItemTitle() string {
	return i.Title
}

// ItemLink returns the Link
func (i Item) ItemLink() string {
	return i.Link
}

// ItemPubDate returns the PubDate
func (i Item) ItemPubDate() string {
	return i.PubDate
}

// ItemDescription returns the Description
func (i Item) ItemDescription() string {
	return i.Description
}

// ItemTVShowID returns the TVShowID
func (i Item) ItemTVShowID() int {
	return i.TVShowID
}

// ItemTVExternalID returns the TVExternalID
func (i Item) ItemTVExternalID() int {
	return i.TVExternalID
}

// ItemTVShowName returns the TVShowName
func (i Item) ItemTVShowName() string {
	return i.TVShowName
}

// ItemTVEpisodeID returns the TVEpisodeID
func (i Item) ItemTVEpisodeID() int {
	return i.TVEpisodeID
}

// ItemTVRawTitle returns the TVRawTitle
func (i Item) ItemTVRawTitle() string {
	return i.TVRawTitle
}

// ItemTVInfoHash returns the TVInfoHash
func (i Item) ItemTVInfoHash() string {
	return i.TVInfoHash
}

// ItemEnclosureURL returns the EnclosureURL
func (i Item) ItemEnclosureURL() string {
	return i.EnclosureURL
}
