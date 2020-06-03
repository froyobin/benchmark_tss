package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	sdkkey "github.com/binance-chain/go-sdk/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"gitlab.com/thorchain/bepswap/thornode/cmd"
)

func setupBech32Prefix() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
	config.Seal()
}

func getP2PIDFromPrivKey(priKeyString string) (string, error) {
	priHexBytes, err := base64.StdEncoding.DecodeString(priKeyString)
	if err != nil {
		return "", fmt.Errorf("fail to decode private key: %w", err)
	}
	rawBytes, err := hex.DecodeString(string(priHexBytes))
	if err != nil {
		return "", fmt.Errorf("fail to hex decode private key: %w", err)
	}

	var keyBytesArray [32]byte
	copy(keyBytesArray[:], rawBytes[:32])
	priKey := secp256k1.PrivKeySecp256k1(keyBytesArray)

	p2pPriKey, err := crypto.UnmarshalSecp256k1PrivateKey(priKey[:])
	if err != nil {
		return "", err
	}
	id, err := peer.IDFromPrivateKey(p2pPriKey)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func UpdateBootstrapNode(ip string, num int, path string) error {
	// we always update from the first node, skip the bootstrap node
	for i := 1; i < num; i++ {
		target := fmt.Sprintf("%s/%d", path, i)
		in := target + "/run.sh"
		out := target + "/deployed_run.sh"
		input, err := ioutil.ReadFile(in)
		if err != nil {
			return err
		}

		lines := strings.Split(string(input), "\n")
		for li, line := range lines {
			if strings.Contains(line, "IPADDR") {
				lines[li] = strings.ReplaceAll(lines[li], "IPADDR", ip)
				break
			}
		}
		output := strings.Join(lines, "\n")
		err = ioutil.WriteFile(out, []byte(output), 0644)
		if err != nil {
			return errors.New("fail to write the file")
		}

	}

	return nil
}

func CreateNewConfigure(start, num int, storagePath string) ([]string, error) {
	setupBech32Prefix()
	var pubKeysBuf bytes.Buffer
	var p2pid string
	for i := start; i < start+num; i++ {
		manager, err := sdkkey.NewKeyManager()
		if err != nil {
			return nil, err
		}
		bech32Key, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, manager.GetPrivKey().PubKey())
		if err != nil {
			return nil, err
		}
		_, err = pubKeysBuf.WriteString(bech32Key + "\n")
		if err != nil {
			return nil, err
		}
		privKeyExported, err := manager.ExportAsPrivateKey()
		if err != nil {
			return nil, err
		}
		privKey := base64.StdEncoding.EncodeToString([]byte(privKeyExported))
		// we save the p2pid of the first node and set it fore the rest as the bootstrap node
		if i == start {
			p2pid, err = getP2PIDFromPrivKey(privKey)
			if err != nil {
				return nil, err
			}
		}

		templatePath := fmt.Sprintf("%s/run_template.sh", storagePath)
		input, err := ioutil.ReadFile(templatePath)
		if err != nil {
			return nil, errors.New("cannot open the template file")
		}
		lines := strings.Split(string(input), "\n")
		for li, line := range lines {
			if strings.Contains(line, "PRIVKEY") {
				lines[li] = strings.ReplaceAll(lines[li], "PRIVKEY", privKey)
				// if we are the fist node, we do not set the bootstrap,otherwise, we set it
				if i == start {
					lines[2] = strings.ReplaceAll(lines[2], "TIME", "1")
					s := strings.Split(lines[li], "-peer")
					lines[li] = s[0]
				} else {
					lines[2] = strings.ReplaceAll(lines[2], "TIME", "100")
					lines[li] = strings.ReplaceAll(lines[li], "BOOTSTRAP", p2pid)
				}
				break
			}
		}
		output := strings.Join(lines, "\n")
		target := fmt.Sprintf("%s/%d/run.sh", storagePath, i)
		err = ioutil.WriteFile(target, []byte(output), 0644)
		if err != nil {
			return nil, errors.New("fail to write the file")
		}
	}
	// we write the public key to the file
	target := fmt.Sprintf("%s/pubkeys.txt", storagePath)
	err := ioutil.WriteFile(target, pubKeysBuf.Bytes(), 0644)
	if err != nil {
		return nil, errors.New("fail to write the file")
	}
	pubKeys := strings.Split(pubKeysBuf.String(), "\n")
	return pubKeys, nil
}
