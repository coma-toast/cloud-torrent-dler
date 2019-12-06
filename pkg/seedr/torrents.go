package seedr

// Torrents is the currently downloading torrents struct
type Torrents struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Folder     string `json:"folder"`
	Size       int    `json:"size"`
	Hash       string `json:"hash"`
	Progress   string `json:"progress"`
	LastUpdate string `json:"last_update"`
}
