package main

// ENSURE RUNING THE DEPLOYMENT FROM THE PROJECT ROOT DIRECTORY!!!!!!!!

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	flag.IntVar(num, "n", 25, "how many nodes we deploy")
	flag.IntVar(option, "opt", 2, "how many nodes we deploy")
	flag.BoolVar(initConfig, "init", false, "recreate the test nodes keypairs")
	flag.StringVar(hostsTablePath, "h", "hosts.txt", "path of the host table")
	flag.StringVar(remoteFilePath, "r", "/home/ubuntu/go-tss/benchmark_docker/Data/data_local/",
		"the path of the tss configure file on the remote machine")
	flag.Parse()
}

func doJob(i int, jobs chan string, hosts []string, dones chan<- struct{}) error {
	defer func() {
		dones <- struct{}{}
	}()
	for ip := range jobs {
		out, err := remote.RunCommand(ip, pemLocation, "ufw disable", true)
		if err != nil {
			log.Error().Err(err).Msg("error in running remote command")
			return err
		}
		log.Info().Msg(out)
		// test existing of file
		out, err = remote.RunCommand(ip, pemLocation, "ls /home/ubuntu/go-tss/benchmark_docker/Data/data_local/", true)
		if err != nil {
			log.Error().Err(err).Msg("error in running remote command")
			return err
		}
		if strings.Contains(out, " No such file or directory") {
			log.Info().Msg("we create the directory")
			out, err = remote.RunCommand(ip, pemLocation, "mkdir -p /home/ubuntu/go-tss/benchmark_docker/Data/data_local/", true)
			if err != nil {
				log.Error().Err(err).Msg("error in running remote command")
				return err
			}
			log.Info().Msg(out)
		}
		// clone the tss code
		out, err = remote.RunCommand(ip, pemLocation, "git clone https://gitlab.com/thorchain/tss/go-tss.git /home/ubuntu/go-tss/go-tss", true)
		if err != nil {
			log.Error().Err(err).Msg("error in running remote command")
			return err
		}
		log.Info().Msg(out)
	}
	return nil
}

func addJob(jobs chan<- string, tasks []string) {
	for _, el := range tasks {
		// filter out the empty ip address
		if len(el) == 0 {
			continue
		}
		jobs <- el
	}
	close(jobs)
}

func deploy(initConfigure bool, hostsTablePath, remoteFilePath string, num int) error {
	var err error
	if initConfigure {
		fmt.Printf(">>>>>>%v\n", num)
		_, err = tools.CreateNewConfigure(0, num, storagePath)
		if err != nil {
			log.Error().Err(err).Msg("fail to create the nodes configure file")
			return err
		}
	}

	hostIPs, err := tools.LoadStringData(hostsTablePath)
	// as the last element of hostIP is empty, so we avoid the last entry
	hostIPs = hostIPs[:len(hostIPs)-1]
	if err != nil {
		log.Error().Err(err).Msg("fail to read the host file")
		return err
	}
	bootstrapIP := hostIPs[0]

	// everytime we run, we update the bootstrap node IP address
	err = tools.UpdateBootstrapNode(bootstrapIP, num, storagePath)
	if err != nil {
		log.Error().Err(err).Msg("fail to update the bootstrapIP")
		return err
	}
	// we set we have 5 threads do the command
	worker := 5
	working := worker
	dones := make(chan struct{}, worker)
	jobs := make(chan string, worker)
	done := false
	go addJob(jobs, hostIPs)
	for i := 0; i < worker; i++ {
		go doJob(i, jobs, hostIPs, dones)
	}

	for {
		<-dones
		working -= 1
		if working <= 0 {
			done = true
		}
		if done == true {
			break
		}
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
	// if  initConfig is set true, we will regeneraet the keysoters for num of nodes
	getParameters(&remoteFilePath, &hostsTablePath, &initConfigure, &num, &option)
	switch option {
	case 1:
		err := deploy(initConfigure, hostsTablePath, remoteFilePath, num)
		if err != nil {
			log.Error().Err(err).Msg("fail to deploy the nodes")
		}
		return
	case 2:
		// keygen test
		inputKeys, inputIPs, ports, err := prepare(pubKeyPath, hostsTablePath)
		if err != nil {
			return
		}
		timeBefore := time.Now()
		for i := 0; i < 10; i++ {
			poolPubKey, err := tss.KeyGen(inputKeys, inputIPs, ports)
			if err != nil {
				fmt.Printf("We quit As saw errors!!!")
				return
			}
			fmt.Println(poolPubKey)
		}
		fmt.Printf("time we spend is %v\n", time.Now().Sub(timeBefore)/10)
	case 3:
		// keysign test
		inputKeys, inputIPs, ports, err := prepare(pubKeyPath, hostsTablePath)
		if err != nil {
			return
		}
		timeBefore := time.Now()
		poolKey := "thorpub1addwnpepqgxtq2zxnhjp6ezspfxyn7nqqjkwauetr6cuhj70n7pj4rqx4jgpcjmvuum"
		for i := 0; i < 10; i++ {
			err := tss.KeySign("hello"+string(i), poolKey, inputIPs, ports, inputKeys)
			if err != nil {
				fmt.Printf("######we quit as saw the error!!!")
				return
			}
		}
		fmt.Printf("time we spend is %v\n", time.Now().Sub(timeBefore)/10)
	default:
		fmt.Println("not supported!!!")
		return
	}
}
