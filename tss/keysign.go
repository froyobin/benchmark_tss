package tss

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

// Request request to sign a message
type KeySignReq struct {
	PoolPubKey    string   `json:"pool_pub_key"` // pub key of the pool that we would like to send this message from
	Message       string   `json:"message"`      // base64 encoded messages to be signed
	SignerPubKeys []string `json:"signer_pub_keys"`
}

type Node struct {
	Pubkey         string `json:"pubkey"`
	BlameData      []byte `json:"data"`
	BlameSignature []byte `json:"signature,omitempty"`
}

type Blame struct {
	FailReason string `json:"fail_reason"`
	IsUnicast  bool   `json:"is_broadcast"`
	BlameNodes []Node `json:"blame_peers,omitempty"`
}

type Status byte

const (
	NA Status = iota
	Success
	Fail
)

type KeySignResponse struct {
	R      string `json:"r"`
	S      string `json:"s"`
	Status Status `json:"status"`
	Blame  Blame  `json:"blame"`
}

func KeySign(inputMsg, poolPubKey string, IPs []string, ports []int, signersPubKey []string) error {
	var locker sync.Mutex
	keySignRespArr := make([][]byte, len(ports))
	var globalErr error
	msg := base64.StdEncoding.EncodeToString([]byte(inputMsg))

	keySignReq := KeySignReq{
		PoolPubKey:    poolPubKey,
		Message:       msg,
		SignerPubKeys: signersPubKey,
	}
	request, _ := json.Marshal(keySignReq)
	requestGroup := sync.WaitGroup{}
	for i := 0; i < len(ports); i++ {
		requestGroup.Add(1)
		go func(idx int, request []byte, keySignRespArr [][]byte, locker *sync.Mutex) {
			defer requestGroup.Done()
			url := fmt.Sprintf("http://%s:%d/keysign", IPs[idx], ports[idx])

			respByte, err := sendTestRequest(url, request)
			if err != nil {
				log.Error().Err(err).Msg("fail to send request")
				globalErr = err
				panic("error in keysign")
				return
			}
			fmt.Printf("---%d::%s\n", idx, string(respByte))
			var response KeySignResponse
			err = json.Unmarshal(respByte, &response)
			if err != nil {
				panic("fail to get the valid signature")
			}
			if response.Status == Fail {
				panic("error in get signature")
			}
			locker.Lock()
			keySignRespArr[idx] = respByte
			locker.Unlock()
		}(i, request, keySignRespArr, &locker)
	}
	requestGroup.Wait()
	if globalErr != nil {
		log.Error().Err(globalErr).Msg("fail to run keysign")
		return globalErr
	}
	fmt.Printf("%v", string(keySignRespArr[0]))
	return nil
}
