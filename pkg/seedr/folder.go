package seedr

import (
	"os"
	"time"
)

// Folder is a folder that contains subfolders or files, or both
type Folder struct {
	SpaceMax   int64       `json:"space_max"`
	SpaceUsed  int64       `json:"space_used"`
	Code       int         `json:"code"`
	Timestamp  string      `json:"timestamp"`
	ID         int         `json:"id"`
	FolderName string      `json:"name"`
	ParentID   int         `json:"parent_id"`
	Torrents   []string    `json:"torrents"`
	Folders    []SubFolder `json:"folders"`
	Files      []File      `json:"files"`
	Result     bool        `json:"result"`
}

// Name returns the name of the folder
func (f Folder) Name() string {
	return f.FolderName
}

// Size is the size
func (f Folder) Size() int64 {
	return f.SpaceUsed
}

// Mode returns an empty os.FileMode to satisfy the os.FileInfo struct
func (f Folder) Mode() os.FileMode {
	var mode os.FileInfo
	return mode.Mode()
}

// ModTime is the last modified time
func (f Folder) ModTime() time.Time {
	t, _ := time.Parse("RFC 3339", f.Timestamp)
	return t
}

// IsDir is true because a folder, by definition, is a dir
func (f Folder) IsDir() bool {
	return true
}

// Sys is not used, so nil
func (f Folder) Sys() interface{} {
	return nil
}
