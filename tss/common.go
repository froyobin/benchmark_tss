package tss

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"
)

func sendTestRequest(url string, request []byte) ([]byte, error) {
	var resp *http.Response
	var err error
	fmt.Printf("%s\n", url)
	if len(request) == 0 {
		resp, err = http.Get(url)
		if err != nil {
			log.Error().Err(err)
			return nil, err
		}
	} else {
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(request))
		if err != nil {
			log.Error().Err(err)
			return nil, err

		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}
	return body, nil
}
