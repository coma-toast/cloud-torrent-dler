package pidcheck

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// AlreadyRunning determines if there is a process already running
func AlreadyRunning(pidFile string) bool {
	// Read in the pid file as a slice of bytes.
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			err = writePid(pidFile)
			if err != nil {
				log.Println(err)

				return true
			}

			return false
		}

		return true
	}

	if len(pidData) == 0 {
		err = writePid(pidFile)
		if err != nil {
			log.Println(err)

			return true
		}

		return false
	}

	// Convert the file contents to an integer.
	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		fmt.Println(err)
		return true
	}

	// Look for the pid in the process list.
	process, err := os.FindProcess(pid)
	log.Info(err)
	if err != nil {
		if err.Error() == "OpenProcess: The parameter is incorrect." {
			return false
		}
		fmt.Println(err)
		return true
	}
	// Send the process a signal zero kill.
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		if !strings.Contains(err.Error(), "process already finished") {
			fmt.Printf("pid already running: %d", pid)
			return true
		}
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	err = writePid(pidFile)
	if err != nil {
		fmt.Println("Failed to write pid file: ", err)
	}

	return false
}

func writePid(pidFile string) error {
	err := ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)

	return err
}
