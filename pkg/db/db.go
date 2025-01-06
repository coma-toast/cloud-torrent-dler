package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DownloadItem is the information needed for the download queue
type DownloadItem struct {
	EpisodeID     int
	FolderPath    string
	IsDir         bool
	Name          string
	TVShowName    string
	ParentSeedrID int
	SeedrID       int
	ShowID        int
	TorrentHash   string
}

func Connect() (*gorm.DB, error) {
	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&DownloadItem{})
}
