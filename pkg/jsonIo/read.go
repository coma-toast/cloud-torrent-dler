package jsonIo

import (
	"encoding/json"
	"io/ioutil"
)

// ReadFile reads the cache file from disk
func ReadFile(path string, target interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		return err
	}

	return nil
}
