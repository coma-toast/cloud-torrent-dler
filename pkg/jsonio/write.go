package jsonio

import (
	"encoding/json"
	"os"
)

// WriteFile writes the cache file to disk
func WriteFile(path string, payload interface{}) error {
	data, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
