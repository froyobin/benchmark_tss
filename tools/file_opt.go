package tools

import (
	"io/ioutil"
	"strings"
)

func LoadStringData(path string) ([]string, error) {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data := strings.Split(string(input), "\n")
	return data, nil
}
