package main

type mockSeedrInstance struct {
}

// Add is test
func (m mockSeedrInstance) Add(magnet string) (seedr.Result, error) {
	return seedr.Result{}, nil

}

// DeleteFile is test
func (m mockSeedrInstance) DeleteFile(id int) error {
	return nil

}

// DeleteFolder is test
func (m mockSeedrInstance) DeleteFolder(id int) error {
	return nil

}

// FindID is test
func (m mockSeedrInstance) FindID(filename string) (int, error) {
	return 0, nil

}

// Get is test
func (m mockSeedrInstance) Get(item DownloadItem, destination string) error {
	return nil
}

// GetPath is test
func (m mockSeedrInstance) GetPath(ID int) (string, error) {
	return "", nil
}

// List is test
func (m mockSeedrInstance) List(path string) ([]DownloadItem, error) {
	return []DownloadItem{}, nil
}
