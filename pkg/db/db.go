package db

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MediaType string

const (
	Show  MediaType = "show"
	Movie MediaType = "movie"
)

type Source string

const (
	ShowRSS Source = "showrss"
	YTS     Source = "yts"
)

// DownloadItem is the information needed for the download queue
type DownloadItem struct {
	ID            uint `gorm:"primaryKey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Downloaded    string
	EpisodeID     int
	FolderPath    string
	IsDir         bool
	Name          string
	TVShowName    string
	ParentSeedrID int
	SeedrID       int
	Source        Source
	ShowID        int
	TorrentHash   string
	MediaType     MediaType
	MagnetURI     string
	// TorrentData   torrent.Torrent
}

type Database interface {
    Connect(host, database, user, password string) (*gorm.DB, error)
    Migrate(db *gorm.DB) error
    // * HERE - have the functions part of the interface. like func(db *gorm.DB) Get()

func Connect(host, database, user, password string) (*gorm.DB, error) {
	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=" + database + " port=5432 sslmode=disable TimeZone=America/New_York"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&DownloadItem{})
}

func DeleteDatabase(db *gorm.DB) error {
	return db.Migrator().DropTable(&DownloadItem{})
}

// CreateDownloadItem creates a new DownloadItem in the database
func CreateDownloadItem(db *gorm.DB, item *DownloadItem) error {
	return db.Create(item).Error
}

// GetDownloadItem retrieves a DownloadItem by its ID
func GetDownloadItem(db *gorm.DB, id uint) (*DownloadItem, error) {
	var item DownloadItem
	if err := db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateDownloadItem updates an existing DownloadItem in the database
func UpdateDownloadItem(db *gorm.DB, item *DownloadItem) error {
	return db.Save(item).Error
}

// DeleteDownloadItem deletes a DownloadItem by its ID
func DeleteDownloadItem(db *gorm.DB, id uint) error {
	return db.Delete(&DownloadItem{}, id).Error
}
