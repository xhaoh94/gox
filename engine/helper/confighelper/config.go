package confighelper

import (
	"encoding/json"
	"os"
)

func LoadJson(filename string, out interface{}) error {
	fileByte, fileErr := os.ReadFile(filename)
	if fileErr != nil {
		return fileErr
	}
	fileErr = json.Unmarshal(fileByte, out)
	if fileErr != nil {
		return fileErr
	}
	return nil
}
