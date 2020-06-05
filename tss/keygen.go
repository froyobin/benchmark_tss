package tss

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

type Request struct {
	Keys []string `json:"keys"`
}

type Response struct {
	PubKey      string `json:"pub_key"`
	PoolAddress string `json:"pool_address"`
	Status      int    `json:"status"`
	Blame       struct {
		FailReason string `json:"fail_reason"`
	} `json:"blame"`
}

func KeyGen(testPubKeys []string, IPs []string, ports []int) (string, error) {
	keyGenRespArr := make([][]byte, len(ports))
	var locker sync.Mutex
	var globalErr error
	keyGenReq := Request{
		Keys: testPubKeys,
	}
	request, err := json.Marshal(keyGenReq)
	if err != nil {
		return "", err
	}
	requestGroup := sync.WaitGroup{}

	for i := 0; i < len(ports); i++ {
		requestGroup.Add(1)
		go func(i int, request []byte, keygenRespAddr [][]byte, locker *sync.Mutex) {
			defer requestGroup.Done()
			url := fmt.Sprintf("http://%s:%d/keygen", IPs[i], ports[i])
			respByte, err := sendTestRequest(url, request)
			if err != nil {
				globalErr = err
				return
			}
			if err != nil {
				log.Error().Err(err).Msg("error in unmarshal the result")
				globalErr = err
				return
			}
			locker.Lock()
			keygenRespAddr[i] = respByte
			locker.Unlock()
		}(i, request, keyGenRespArr, &locker)
	}
	requestGroup.Wait()
	if globalErr != nil {
		log.Error().Err(err).Msg("error in keygen")
		return "", nil
	}
	for i := 0; i < len(ports); i++ {
		fmt.Printf("%d------%s\n", i, keyGenRespArr[i])
	}
	var ret Response
	json.Unmarshal(keyGenRespArr[0], &ret)
	return ret.PubKey, nil
}