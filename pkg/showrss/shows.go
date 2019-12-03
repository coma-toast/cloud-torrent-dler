package showrss

// Channel is another layer above the Shows
type Channel struct {
	Shows Shows `xml:"channel"`
}

// Shows is the main XML response
type Shows struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	TTL         int    `xml:"ttl"`
	Description string `xml:"description"`
	Item        []Item `xml:"item"`
}

// ShowsObjectTitle is the Title
func (s Shows) ShowsObjectTitle() string {
	return s.Title
}

// ShowsObjectLink is the Link
func (s Shows) ShowsObjectLink() string {
	return s.Link
}

// ShowsObjectTTL is the TTL
func (s Shows) ShowsObjectTTL() int {
	return s.TTL
}

// ShowsObjectDescription is the Description
func (s Shows) ShowsObjectDescription() string {
	return s.Description
}

// ShowsObjectItem is the Item
func (s Shows) ShowsObjectItem() []Item {
	return s.Item
}
