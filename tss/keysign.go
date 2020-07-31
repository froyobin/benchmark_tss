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
				return
			}
			var response KeySignResponse
			err = json.Unmarshal(respByte, &response)
			if err != nil {
				fmt.Printf("unmarshal error")
			}
			locker.Lock()
			keySignRespArr[idx] = respByte
			locker.Unlock()
		}(i, request, keySignRespArr, &locker)
	}
	requestGroup.Wait()
	if globalErr != nil {
		log.Error().Err(globalErr).Msg("fail to run keysign")
	}

	votes := make(map[string]int)
	for i, itemstr := range keySignRespArr {
		var item Response
		err := json.Unmarshal(itemstr, &item)
		if err != nil {
			log.Error().Err(err).Msgf("fail to unmarshal the keygen result")
			return err
		}

		fmt.Printf("\nresult::>>%d---status:%v-unicast(%v)->%v\n", i, item.Status, item.Blame.IsUnicast, item.Blame)
		_ = i
		for _, el := range item.Blame.BlameNodes {
			_, ok := votes[el.Pubkey]
			if !ok {
				votes[el.Pubkey] = 1
				continue
			}
			votes[el.Pubkey] += 1
		}
		//if len(poolPubKey) == 0 {
		//	poolPubKey = item.PubKey
		//} else {
		//	c.Assert(poolPubKey, Equals, item.PubKey)
		//}
	}
	fmt.Printf("------------------------------\n")
	for k, v := range votes {
		fmt.Printf("node %s :-->%d\n", k, v)
	}
	fmt.Printf("------------------------------\n")

	fmt.Printf("%v", string(keySignRespArr[0]))
	return nil
}
