package confighelper

import (
	"encoding/json"
	"io/ioutil"
)

func LoadJson(filename string, out interface{}) error {
	fileByte, fileErr := ioutil.ReadFile(filename)
	if fileErr != nil {
		return fileErr
	}
	fileErr = json.Unmarshal(fileByte, out)
	if fileErr != nil {
		return fileErr
	}
	return nil
}
