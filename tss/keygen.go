package tss

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

// Request request to do keygen
type Request struct {
	Keys        []string `json:"keys"`
	BlockHeight int64    `json:"block_height"`
	Version     string   `json:"tss_version"`
}

type Response struct {
	PubKey      string `json:"pub_key"`
	PoolAddress string `json:"pool_address"`
	Status      Status `json:"status"`
	Blame       Blame  `json:"blame"`
}

func KeyGen(testPubKeys []string, IPs []string, ports []int, blockHeight int64) (string, error) {
	keyGenRespArr := make([][]byte, len(IPs))
	var locker sync.Mutex
	var globalErr error
	keyGenReq := Request{
		Keys:        testPubKeys,
		BlockHeight: blockHeight,
		Version:     "0.16.0",
	}
	request, err := json.Marshal(keyGenReq)
	if err != nil {
		return "", err
	}
	requestGroup := sync.WaitGroup{}

	for i := 0; i < len(IPs); i++ {
		requestGroup.Add(1)
		go func(idx int, request []byte, keygenRespAddr [][]byte, locker *sync.Mutex) {
			defer requestGroup.Done()
			url := fmt.Sprintf("http://%s:%d/keygen", IPs[idx], ports[idx])
			respByte, err := sendTestRequest(url, request)
			if err != nil {
				globalErr = err
				fmt.Printf("%d--err:%v--%s\n", i, err, string(respByte))
				return
			}
			if err != nil {
				log.Error().Err(err).Msg("error in unmarshal the result")
				fmt.Printf("%d----%s", i, string(respByte))
				globalErr = err
				return
			}
			locker.Lock()
			keygenRespAddr[idx] = respByte
			locker.Unlock()
		}(i, request, keyGenRespArr, &locker)
	}
	requestGroup.Wait()
	if globalErr != nil {
		log.Error().Err(err).Msg("error in keygen")
		return "", nil
	}
	//for i := 0; i < len(IPs); i++ {
	//	fmt.Printf("%d------%s\n", i, keyGenRespArr[i])
	//}
	var ret Response
	err = json.Unmarshal(keyGenRespArr[0], &ret)
	if err != nil {
		log.Error().Err(err).Msgf("fail to unmarshal the keygen result")
		// return "", err
	}

	votes := make(map[string]int)
	for i, itemstr := range keyGenRespArr {
		var item Response
		err = json.Unmarshal(itemstr, &item)
		if err != nil {
			log.Error().Err(err).Msgf("fail to unmarshal the keygen result")
			// return "", err
			continue
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

	return ret.PubKey, nil
}
