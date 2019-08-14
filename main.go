package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

// config is the configuration struct
type config struct {
	BaseURL     string
	DlRoot      string
	PidFilePath string
	Username    string
	Passwd      string
}

// Folder is a folder that contains subfolders or files, or both
type Folder struct {
	SpaceMax  int64    `json:"space_max"`
	SpaceUsed int64    `json:"space_used"`
	Code      int      `json:"code"`
	Timestamp string   `json:"timestamp"`
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	ParentID  int      `json:"parent_id"`
	Torrents  []string `json:"torrents"`
	Folders   []Folder `json:"folders"`
	Files     []File   `json:"files"`
	Result    bool     `json:"result"`
}

// SubFolder is a folder in folder - a sort of, Folder Inception, if you will
type SubFolder struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	LastUpdate string `json:"last_update"`
}

// File is the file data object. Use the ID to get the file itself
type File struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Size           int64  `json:"size"`
	Hash           string `json:"hash"`
	LastUpdate     string `json:"last_update"`
	StreamAudio    bool   `json:"stream_audio"`
	StreamVideo    bool   `json:"stream_video"`
	VideoConverted string `json:"video_converted,omitempty"`
	ParentFolder   int
}

// new config instance
var (
	conf *config
)

// Credentials is the base64 encoded username:password
var Credentials string

// DeleteQueue is a list of folders to delete at the end of downloading
// there may be multiple files in a folder, so we have to wait until the end to delete
var DeleteQueue []int

// Write a pid file, but first make sure it doesn't exist with a running pid.
func alreadyRunning(pidFile string) bool {
	// Read in the pid file as a slice of bytes.
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		// Convert the file contents to an integer.
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			// Look for the pid in the process list.
			if process, err := os.FindProcess(pid); err == nil {
				// Send the process a signal zero kill.
				if err := process.Signal(syscall.Signal(0)); err == nil {
					fmt.Println("PID already running!")
					// We only get an error if the pid isn't running, or it's not ours.
					err := fmt.Errorf("pid already running: %d", pid)
					log.Print(err)
					return true
				}
				log.Print(err)

			} else {
				log.Print(err)
			}
		} else {
			log.Print(err)
		}
	} else {
		log.Print(err)
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
	return false
}

func getConf() *config {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()

	if err != nil {
		handleError(err)
	}

	conf := &config{}
	err = viper.Unmarshal(conf)

	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}

	return conf
}

func main() {
	conf = getConf()
	pidPath := fmt.Sprintf("%s/cloud-torrent-downloader", conf.PidFilePath)
	pid := alreadyRunning(pidPath)

	// Credentials is the base64 encoded username:password
	Credentials = b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", conf.Username, conf.Passwd)))

	// DeleteQueue is a list of folders to delete once all the downloads have completed.
	// Do not delete them right when the download completes, as there could be multiple files in a folder
	var DeleteQueue []int

	if !pid {
		// Start at the root folder of your choosing (i.e. Completed),
		// recursively searching down, populating the files list
		files, err := getFilesFromFolder(96452508)
		if err != nil {
			fmt.Println(fmt.Errorf("Error: %s", err))
			return
		}
		downloadFiles(files)
		deleteDownloaded(DeleteQueue)
	}
}

func deleteDownloaded(list []int) error {
	var err error
	for _, folder := range list {
		deleteResult, err := apiCall("DELETE", folder, "folder")
		if err != nil {
			fmt.Println(fmt.Errorf("Error: %s", err))
		}
		spew.Dump(deleteResult)
	}

	return err
}

// Simple api call method for gathering info - NOT the actual downloading of the file.
// callType is 'file' or 'folder'+
//TODO: return bytes and an error
func apiCall(method string, id int, callType string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", conf.BaseURL, callType)
	if id != 0 {
		url = fmt.Sprintf("%s/%d", url, id)
	}
	client := &http.Client{}
	request, err := http.NewRequest(method, url, nil)
	handleError(err)
	request.SetBasicAuth(conf.Username, conf.Passwd)
	response, err := client.Do(request)
	handleError(err)
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	handleError(err)
	// if response.StatusCode >= 400 {
	log.Print(fmt.Errorf("Response Code Error: %d. %s", response.StatusCode, string(data)))

	// }
	return data, err
}

// Get folder info
func getFolder(id int) (Folder, error) {
	var rootData Folder
	data, err := apiCall("GET", id, "folder")
	if err != nil {
		return rootData, err
	}
	err = json.Unmarshal(data, &rootData)
	handleError(err)
	return rootData, err

}

// Get the files in the folder and any subfolders
func getFilesFromFolder(folderID int) ([]File, error) {
	folder, err := getFolder(folderID)
	files := folder.Files
	if err != nil {
		return files, err
	}
	DeleteQueue = append(DeleteQueue, folder.ID)
	for _, folder := range folder.Folders {
		subfiles, err := getFilesFromFolder(folder.ID)
		if err != nil {
			fmt.Println(fmt.Errorf("Error: %s", err))
		}
		files = append(files, subfiles...)
	}
	return files, err
}

// Do the actual download of the files
func downloadFiles(files []File) error {
	var err error
	for _, file := range files {
		// * dev code
		isAVideo, _ := regexp.MatchString("(.*?).(txt|jpg)$", file.Name)
		// isAVideo, _ := regexp.MatchString("(.*?).(mkv|mp4|avi)$", file.Name)
		if isAVideo {
			//TODO: break out into separate functions
			fmt.Println("Downloading file: " + file.Name)
			path := fmt.Sprintf("%s/%s", conf.DlRoot, file.Name)
			fileURL := fmt.Sprintf("%s/file/%d", conf.BaseURL, file.ID)
			out, err := os.Create(path)
			handleError(err)
			defer out.Close()

			//TODO: break out into separate function - dlWithProgress(path) - probably new helper file?
			client := grab.NewClient()

			req, err := grab.NewRequest(path, fileURL)
			handleError(err)
			req.NoResume = true
			req.HTTPRequest.Header.Set("Authorization", "Basic "+Credentials)
			resp := client.Do(req)
			spew.Dump(resp.HTTPResponse.Status)

			// progress bar
			t := time.NewTicker(time.Second)
			defer t.Stop()

		Loop:
			for {
				select {
				case <-t.C:
					fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
						resp.BytesComplete(),
						resp.Size,
						100*resp.Progress())

				case <-resp.Done:
					// download is complete
					break Loop
				}
			}

			// check for errors
			if err := resp.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Download saved to ./%v \n", resp.Filename)

			fmt.Println("Download complete.")
		}
	}
	return err
}

//TODO: once errors are returned above, this is not needed
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
