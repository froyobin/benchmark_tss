package tools

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func LoadStringData(path string) ([]string, error) {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data := strings.Split(string(input), "\n")
	return data, nil
}

func AnalysisIPs(ip string) (string, bool) {
	var digitalOcean bool
	switch ip[len(ip)-1] {
	case 'd':
		digitalOcean = true
		ip = strings.Trim(ip, "d")
	case 'a':
		digitalOcean = false
		ip = strings.Trim(ip, "a")
	default:
		panic("invalid data center id")
	}
	return ip, digitalOcean
}

func GetInput(promote string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(promote)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Error().Err(err).Msg("fail to open stdin for input")
		return "", err
	}
	input = strings.Replace(input, "\n", "", -1)
	return input, nil
}

func GenerateComposeForAWS(storagePath string, targetIPs []string) (string, error) {
	tempDirPath := os.TempDir()
	folderPath := path.Join(tempDirPath, "awsConf")
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	templatePath := fmt.Sprintf("%s/docker-compose-aws.yml", storagePath)
	input, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", errors.New("cannot open the template file")
	}

	for _, rawIp := range targetIPs {
		ip, digitalOcean := AnalysisIPs(rawIp)
		if digitalOcean {
			continue
		}
		folderPathWithIP := path.Join(folderPath, ip)
		if _, err := os.Stat(folderPathWithIP); os.IsNotExist(err) {
			err := os.Mkdir(folderPathWithIP, os.ModePerm)
			if err != nil {
				return "", err
			}
		}

		lines := strings.Split(string(input), "\n")
		for li, line := range lines {
			if strings.Contains(line, "PUBIP") {
				lines[li] = strings.ReplaceAll(lines[li], "PUBIP", ip)
			}
		}
		output := strings.Join(lines, "\n")
		target := path.Join(folderPathWithIP, "docker-compose.sh")
		err = ioutil.WriteFile(target, []byte(output), 0644)
		if err != nil {
			return "", err
		}
	}
	return folderPath, nil
}

func GetRandomPick(n, m int) []int {
	selected := make(map[int]bool)
	for {
		if len(selected) == n {
			break
		}
		n := rand.Int() % m
		selected[n] = true

	}
	var picked []int
	for k := range selected {
		picked = append(picked, k)
	}
	return picked
}

func GetPeerIDFromPubKey(pubkey string) (peer.ID, error) {
	pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, pubkey)
	if err != nil {
		return "", fmt.Errorf("fail to parse account pub key(%s): %w", pubkey, err)
	}
	secpPubKey := pk.(secp256k1.PubKeySecp256k1)
	ppk, err := crypto.UnmarshalSecp256k1PublicKey(secpPubKey[:])
	if err != nil {
		return "", fmt.Errorf("fail to convert pubkey to the crypto pubkey used in libp2p: %w", err)
	}
	return peer.IDFromPublicKey(ppk)
}
