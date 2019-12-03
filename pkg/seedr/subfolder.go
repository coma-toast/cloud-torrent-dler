package seedr

import (
	"os"
	"time"
)

// SubFolder is a folder in folder - a sort of, Folder Inception, if you will
type SubFolder struct {
	ID            int    `json:"id"`
	SubFolderName string `json:"name"`
	SubFolderSize int64  `json:"size"`
	LastUpdate    string `json:"last_update"`
}

// Name returns the name of the subfolder
func (f SubFolder) Name() string {
	return f.SubFolderName
}

// Size is the size
func (f SubFolder) Size() int64 {
	return f.SubFolderSize
}

// Mode returns an empty os.FileMode to satisfy the os.FileInfo struct
func (f SubFolder) Mode() os.FileMode {
	var mode os.FileInfo
	return mode.Mode()
}

// ModTime is the last modified time
func (f SubFolder) ModTime() time.Time {
	t, _ := time.Parse("RFC 3339", f.LastUpdate)
	return t
}

// IsDir is true because a folder, by definition, is a dir
func (f SubFolder) IsDir() bool {
	return true
}

// Sys is not used, so nil
func (f SubFolder) Sys() interface{} {
	return nil
}
