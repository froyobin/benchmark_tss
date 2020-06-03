package main

// ENSURE RUNING THE DEPLOYMENT FROM THE PROJECT ROOT DIRECTORY!!!!!!!!

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/benchmark_tss/remote"
	"gitlab.com/thorchain/benchmark_tss/tools"
	"gitlab.com/thorchain/benchmark_tss/tss"
)

const (
	storagePath = "./storage"
	pemLocation = "~/Documents/thorchain_bin.pem"
)

func getParameters(remoteFilePath, hostsTablePath *string, initConfig *bool, num, option *int) {
	flag.IntVar(num, "n", 10, "how many nodes we deploy")
	flag.IntVar(option, "opt", 2, "how many nodes we deploy")
	flag.BoolVar(initConfig, "init", false, "recreate the test nodes keypairs")
	flag.StringVar(hostsTablePath, "h", "hosts.txt", "path of the host table")
	flag.StringVar(remoteFilePath, "r", "/home/ubuntu/go-tss/benchmark_docker/Data/data_local/",
		"the path of the tss configure file on the remote machine")
	flag.Parse()
}

func deploy(initConfigure bool, hostsTablePath, remoteFilePath string) error {
	var err error
	if initConfigure {
		_, err = tools.CreateNewConfigure(0, 10, storagePath)
		if err != nil {
			log.Error().Err(err).Msg("fail to create the nodes configure file")
			return err
		}
	}

	hostIPs, err := tools.LoadStringData(hostsTablePath)
	if err != nil {
		log.Error().Err(err).Msg("fail to read the host file")
		return err
	}
	bootstrapIP := hostIPs[0]

	// everytime we run, we update the bootstrap node IP address
	err = tools.UpdateBootstrapNode(bootstrapIP, 10, storagePath)
	if err != nil {
		log.Error().Err(err).Msg("fail to update the bootstrapIP")
		return err
	}

	for _, ip := range hostIPs {
		out, err := remote.RunCommand(ip, pemLocation, "ufw disable", true)
		if err != nil {
			log.Error().Err(err).Msg("error in running remote command")
			return err
		}
		log.Info().Msg(out)
	}

	// we send the configuration file and the docker compose file
	err = remote.SendRemote(hostIPs, storagePath, remoteFilePath, pemLocation, true)
	if err != nil {
		log.Error().Msg("fail to update the configuration file to remote nodes")
		return err
	}
	return nil
}

func prepare(pubKeyPath, hostsTablePath string) ([]string, []string, []int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("please input the nunber of the nodes")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Error().Err(err).Msg("fail to open stdin for input")
		return nil, nil, nil, err
	}
	input = strings.Replace(input, "\n", "", -1)
	nodeNum, err := strconv.Atoi(input)
	if err != nil {
		log.Error().Err(err).Msg("invalid input")
		return nil, nil, nil, err
	}
	pubKeys, err := tools.LoadStringData(pubKeyPath)
	if err != nil {
		log.Error().Err(err).Msg("fail to load the nodes public keys")
		return nil, nil, nil, err
	}
	hostIPs, err := tools.LoadStringData(hostsTablePath)
	if err != nil {
		log.Error().Err(err).Msg("fail to load the nodes public keys")
		return nil, nil, nil, err
	}
	inputKeys := pubKeys[:nodeNum]
	inputIPs := hostIPs[:nodeNum]
	ports := make([]int, nodeNum)
	for i := 0; i < nodeNum; i++ {
		ports[i] = 8080
	}
	return inputKeys, inputIPs, ports, nil
}

func main() {
	var remoteFilePath, hostsTablePath string
	var initConfigure bool
	var num, option int
	pubKeyPath := "storage/pubkeys.txt"

	getParameters(&remoteFilePath, &hostsTablePath, &initConfigure, &num, &option)
	switch option {
	case 1:
		err := deploy(initConfigure, hostsTablePath, remoteFilePath)
		if err != nil {
			log.Error().Err(err).Msg("fail to deploy the nodes")
		}
		return
	case 2:
		inputKeys, inputIPs, ports, err := prepare(pubKeyPath, hostsTablePath)
		if err != nil {
			return
		}
		poolPubKey, err := tss.KeyGen(inputKeys, inputIPs, ports)
		if err != nil {
			return
		}
		fmt.Println(poolPubKey)
	case 3:
		inputKeys, inputIPs, ports, err := prepare(pubKeyPath, hostsTablePath)
		if err != nil {
			return
		}
		poolKey := "thorpub1addwnpepqdh6jcqqwx3kpgqn5krhks2ejpm5kp38j4s720f2hkc6ygv9akfk2xrlf4x"
		tss.KeySign("hello", poolKey, inputIPs, ports, inputKeys)
	default:
		fmt.Println("not supported!!!")
		return
	}
}
