package db

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/seedr"
	"github.com/coma-toast/cloud-torrent-dler/m/v2/pkg/showrss"

	"github.com/craigpastro/pgmq-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MediaType string

const (
	Show  MediaType = "show"
	Movie MediaType = "movie"
)

func (m MediaType) String() string {
	return string(m)
}

type Source string

const (
	ShowRSS Source = "showrss"
	YTS     Source = "yts"
)

type MagnetURI string

// Validate checks if the MagnetURI is a valid magnet link
func (m MagnetURI) Validate() error {
	magnetRegex := `^magnet:\?xt=urn:btih:[a-fA-F0-9]{40,}.*$`
	matched, err := regexp.MatchString(magnetRegex, string(m))
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("invalid magnet link")
	}
	return nil
}

type TorrentURI string

// Validate checks if the TorrentURI is a valid torrent link
func (t TorrentURI) Validate() error {
	torrentRegex := `^https?:\/\/.*\.torrent$`
	matched, err := regexp.MatchString(torrentRegex, string(t))
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("invalid torrent link")
	}
	return nil
}

// DownloadItem is the information needed for the download queue
type DownloadItem struct {
	gorm.Model
	ID            uint `gorm:"primaryKey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Downloaded    bool
	EpisodeID     showrss.Item
	FolderPath    string
	IsDir         bool
	Name          string
	TVShowName    string
	ParentSeedrID int
	SeedrID       seedr.File
	Source        Source
	ShowID        int
	TorrentHash   string
	MediaType     MediaType
	MagnetURI     MagnetURI
	TorrentData   TorrentURI
}

type DbClient struct {
	client *gorm.DB
	mq     *pgmq.PGMQ
}

type Database interface {
	Connect(host, database, user, password string) (*gorm.DB, error)
	Migrate() error
	DeleteDatabase() error
	AddSeedrUpload(item *DownloadItem) error
	GetSeedrUpload() (*DownloadItem, error)
	CreateDownloadItem(item *DownloadItem) error
	GetDownloadItemByName(name string) (*DownloadItem, error)
	GetDownloadItemByEpisodeID(id string) (*DownloadItem, error)
	GetDownloadItemBySeedrID(id int) (*DownloadItem, error)
	GetDownloadItemByID(id uint) (*DownloadItem, error)
	GetDownloadItemSByShowGUID(guid string) (*DownloadItem, error)
	UpdateDownloadItem(item *DownloadItem) error
	DeleteDownloadItem(id uint) error
}

func (db *DbClient) Connect(host, database, user, password string) error {
	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=" + database + " port=5432 sslmode=disable TimeZone=America/New_York"
	client, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	db.client = client

	db.mq, err = pgmq.New(context.Background(), dsn)
	if err != nil {
		return err
	}

	return err
}

func (db *DbClient) Migrate() error {
	err := db.mq.CreateQueue(context.Background(), "seedr_upload")
	if err != nil {
		return err
	}

	return db.client.AutoMigrate(&DownloadItem{}, &seedr.File{}, &showrss.Item{})
}

func (db *DbClient) AddSeedrUpload(item *DownloadItem) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = db.mq.Send(context.Background(), "seedr_upload", data)
	return err
}

func (db *DbClient) GetSeedrUpload() (*DownloadItem, error) {
	item, err := db.mq.Read(context.Background(), "seedr_upload", 30)
	if err != nil {
		return nil, err
	}
	var downloadItem DownloadItem
	err = json.Unmarshal(item.Message, &downloadItem)
	if err != nil {
		return nil, err
	}
	return &downloadItem, nil
}

func (db *DbClient) ArchiveSeedrUpload(id int64) (bool, error) {
	return db.mq.Archive(context.Background(), "seedr_upload", id)
}

func (db *DbClient) DeleteDatabase() error {
	return db.client.Migrator().DropTable(&DownloadItem{})
}

// CreateDownloadItem creates a new DownloadItem in the database
func (db *DbClient) CreateDownloadItem(item *DownloadItem) error {
	return db.client.Create(item).Error
}

// GetDownloadItem retrieves a DownloadItem by its ID
func (db *DbClient) GetDownloadItemByID(id uint) (*DownloadItem, error) {
	var item DownloadItem
	if err := db.client.Where(&DownloadItem{ID: id}).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (db *DbClient) GetDownloadItemSByShowGUID(guid string) (*DownloadItem, error) {
	var item DownloadItem
	show, err := db.GetShowRSSItemByGUID(guid)
	if err != nil {
		return nil, err
	}

	if err := db.client.Where(&DownloadItem{ShowID: show.ID}).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (db *DbClient) GetDownloadItemByEpisodeID(id string) (*DownloadItem, error) {
	var item DownloadItem
	episodeID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	var showItem = showrss.Item{
		ID: episodeID,
	}
	if err := db.client.Where(showItem).First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

func (db *DbClient) GetDownloadItemBySeedrID(id int) (*DownloadItem, error) {
	var item DownloadItem
	seedr, err := db.GetSeedrItemByID(id)
	if err != nil {
		return nil, err
	}

	if err := db.client.Where(&DownloadItem{SeedrID: *seedr}).FirstOrCreate(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (db *DbClient) GetDownloadItemByName(name string) (*DownloadItem, error) {
	var item DownloadItem
	if err := db.client.Where(&DownloadItem{Name: name}).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateDownloadItem updates an existing DownloadItem in the database
func (db *DbClient) UpdateDownloadItem(item *DownloadItem) error {
	return db.client.Save(item).Error
}

// DeleteDownloadItem deletes a DownloadItem by its ID
func (db *DbClient) DeleteDownloadItem(id uint) error {
	return db.client.Delete(&DownloadItem{}, id).Error
}

func (db *DbClient) CreateSeedrItem(item *seedr.File) error {
	return db.client.Create(item).Error
}

func (db *DbClient) GetSeedrItemByID(id int) (*seedr.File, error) {
	var item seedr.File
	if err := db.client.Where(&seedr.File{ID: id}).FirstOrCreate(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (db *DbClient) UpdateSeedrItem(item *seedr.File) error {
	return db.client.Save(item).Error
}

func (db *DbClient) DeleteSeedrItem(id int) error {
	return db.client.Delete(&seedr.File{}, id).Error
}

func (db *DbClient) CreateShowRSSItem(item *showrss.Item) error {
	return db.client.Create(item).Error
}

func (db *DbClient) GetShowRSSItemByGUID(guid string) (*showrss.Item, error) {
	var item showrss.Item
	if err := db.client.Where(&showrss.Item{GUID: guid}).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (db *DbClient) GetShowRSSItemsByShowID(id string) (*[]showrss.Item, error) {
	var items []showrss.Item
	showId, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	if err := db.client.Where(&showrss.Item{TVShowID: showId}).Find(&items).Error; err != nil {
		return nil, err
	}

	return &items, nil
}

func (db *DbClient) GetShowRSSItemByID(id int) (*showrss.Item, error) {
	var item showrss.Item
	if err := db.client.Where(&showrss.Item{ID: id}).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (db *DbClient) UpdateShowRSSItem(item *showrss.Item) error {
	return db.client.Save(item).Error
}

func (db *DbClient) DeleteShowRSSItem(id int) error {
	return db.client.Delete(&showrss.Item{}, id).Error
}
