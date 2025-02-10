package seedr

import (
	"os"
	"time"

	"gorm.io/gorm"
)

// File is the file data object. Use the ID to get the file itself
type File struct {
	gorm.Model
	_id            uint   `gorm:"primaryKey"`
	ID             int    `json:"id"`
	FileName       string `json:"name"`
	FileSize       int64  `json:"size"`
	Hash           string `json:"hash"`
	LastUpdate     string `json:"last_update"`
	StreamAudio    bool   `json:"stream_audio"`
	StreamVideo    bool   `json:"stream_video"`
	VideoConverted string `json:"video_converted,omitempty"`
	ParentFolder   int
}

// FileID gets the ID of the file
func (f File) FileID() int {
	return f.ID
}

// Name returns the name of the file
func (f File) Name() string {
	return f.FileName
}

// Size is the size
func (f File) Size() int64 {
	return f.FileSize
}

// Mode returns an empty os.FileMode to satisfy the os.FileInfo struct
func (f File) Mode() os.FileMode {
	var mode os.FileInfo
	return mode.Mode()
}

// ModTime is the last modified time
func (f File) ModTime() time.Time {
	t, _ := time.Parse("RFC 3339", f.LastUpdate)
	return t
}

// IsDir is true because a folder, by definition, is a dir
func (f File) IsDir() bool {
	return false
}

// Sys is not used, so nil
func (f File) Sys() interface{} {
	return nil
}
