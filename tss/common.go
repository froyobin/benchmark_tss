package tss

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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
		client := http.Client{Timeout: 100 * time.Second}
		resp, err = client.Post(url, "application/json", bytes.NewBuffer(request))
		if err != nil {
			log.Error().Err(err).Msgf("fail to send post")
			//	return nil, err
		}
	}
	if resp == nil {
		return nil, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("msg body--%s", body)
		return body, err
	}
	return body, nil
}
