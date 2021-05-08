package yts

type YTSResult struct {
	Status         string  `json:"status"`
	Status_message string  `json:"status_message"`
	Data           YTSData `json:"data"`
}

type YTSData struct {
	Movie_count int     `json:"movie_count"`
	Limit       int     `json:"limit"`
	Page_number int     `json:"page_number"`
	Movies      []Movie `json:"movies"`
}

type Movie struct {
	Id                        int       `json:"id"`
	Url                       string    `json:"url"`
	Imdb_code                 string    `json:"imdb_code"`
	Title                     string    `json:"title"`
	Title_english             string    `json:"title_english"`
	Title_long                string    `json:"title_long"`
	Slug                      string    `json:"slug"`
	Year                      int       `json:"year"`
	Rating                    float32   `json:"rating"`
	Runtime                   int       `json:"runtime"`
	Genres                    []string  `json:"genres"`
	Summary                   string    `json:"summary"`
	Description_full          string    `json:"description_full"`
	Synopsis                  string    `json:"synopsis"`
	Yt_trailer_code           string    `json:"yt_trailer_code"`
	Language                  string    `json:"language"`
	Mpa_rating                string    `json:"mpa_rating"`
	Background_image          string    `json:"background_image"`
	Background_image_original string    `json:"background_image_original"`
	Small_cover_image         string    `json:"small_cover_image"`
	Medium_cover_image        string    `json:"medium_cover_image"`
	Large_cover_image         string    `json:"large_cover_image"`
	State                     string    `json:"state"`
	Torrents                  []Torrent `json:"torrents"`
	Date_uploaded             string    `json:"date_uploaded"`
	Date_uploaded_unix        int       `json:"date_uploaded_unix"`
}

type Torrent struct {
	Url                string `json:"url"`
	Hash               string `json:"hash"`
	Quality            string `json:"quality"`
	Torrent_type       string `json:"type"`
	Seeds              int    `json:"seeds"`
	Peers              int    `json:"peers"`
	Size               string `json:"size"`
	Size_bytes         int    `json:"size_bytes"`
	Date_uploaded      string `json:"date_uploaded"`
	Date_uploaded_unix int    `json:"date_uploaded_unix"`
}
