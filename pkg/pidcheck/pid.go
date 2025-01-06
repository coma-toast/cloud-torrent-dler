package pidcheck

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// AlreadyRunning determines if there is a process already running
func AlreadyRunning(pidFile string) bool {
	// Check if the pid file exists.
	_, err := os.Stat(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			err = writePid(pidFile)
			if err != nil {
				log.Println(err)
			}
			return false
		} else {
			log.Println(err)
			return true
		}
	}
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
	log.Info("pid process and error: ", process, err)
	if err != nil {
		if err.Error() == "OpenProcess: The parameter is incorrect." {
			return false
		}
		fmt.Println(err)
		return true
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
	err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
	if err != nil {
		log.Printf("Failed to write PID file %s: %v", pidFile, err)
		return err
	}
	return nil
}
