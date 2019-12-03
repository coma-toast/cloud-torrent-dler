package seedr

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
)

// Service is the service
type Service interface {
	GetFolder(id int) (Folder, error)
	GetFile(id int) (File, error)
	DownloadFileByID(id int, destination string) error
	AddMagnet(magnet string) error
}

// GetFolder gets a Folder from Seedr
func (c Client) GetFolder(id int) (Folder, error) {
	var folder Folder
	url := "/folder"
	if id != 0 {
		url = fmt.Sprintf("%s/%d", url, id)
	}
	_, err := c.call("GET", url, nil, &folder)
	if err != nil {
		return Folder{}, err
	}
	return folder, nil

}

// GetFile gets a File from Seedr
func (c Client) GetFile(id int) (File, error) {
	var file File
	url := fmt.Sprintf("/%d", id)
	_, err := c.call("GET", url, nil, &file)
	if err != nil {
		return File{}, err
	}
	return file, nil
}

// DownloadFileByID downloads a file with the specified ID
func (c Client) DownloadFileByID(id int, destination string) error {
	var err error
	url := fmt.Sprintf("/file/%d", id)
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = c.call("GET", url, nil, output)
	if err != nil {
		return err
	}

	return err
}

//AddMagnet adds a magnet link to Seedr to be downloaded
func (c Client) AddMagnet(magnet string) error {
	var err error
	url := fmt.Sprintf("/torrent/magnet")
	// payload := []string{"magnet", magnet}
	// result, err := c.call("POST", url, payload, nil)
	result, err := c.call("POST", url, magnet, nil)
	spew.Dump("Result: ", result)
	return err
}

// 	out, err := os.Create(path)
// 	handleError(err)
// 	defer out.Close()

// 	// progress bar
// 	t := time.NewTicker(time.Second)
// 	defer t.Stop()

// Loop:
// 	for {
// 		select {
// 		case <-t.C:
// 			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
// 				resp.BytesComplete(),
// 				resp.Size,
// 				100*resp.Progress())

// 		case <-resp.Done:
// 			// download is complete
// 			break Loop
// 		}
// 	}

// 	// check for errors
// 	if err := resp.Err(); err != nil {
// 		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
// 		os.Exit(1)
// 	}

// 	fmt.Printf("Download saved to ./%v \n", resp.Filename)

// 	fmt.Println("Download complete.")
