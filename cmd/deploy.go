package main

// ENSURE RUNING THE DEPLOYMENT FROM THE PROJECT ROOT DIRECTORY!!!!!!!!

import (
	"flag"
	"fmt"
	"os"
	"sort"
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
	flag.IntVar(num, "n", 40, "how many nodes we deploy")
	flag.IntVar(option, "opt", 2, "how many nodes we deploy")
	flag.BoolVar(initConfig, "init", false, "recreate the test nodes keypairs")
	flag.StringVar(hostsTablePath, "h", "hosts.txt", "path of the host table")
	flag.StringVar(remoteFilePath, "r", "/home/ubuntu/go-tss/benchmark_docker/Data/data_local/",
		"the path of the tss configure file on the remote machine")
	flag.Parse()
}

func doPrepareJob(jobs chan string, dones chan<- struct{}) {
	var digitalOcean bool
	defer func() {
		dones <- struct{}{}
	}()
	for ip := range jobs {
		ip, digitalOcean = tools.AnalysisIPs(ip)
		out, err := remote.RunCommand(ip, pemLocation, "ufw disable", digitalOcean)
		if err != nil {
			log.Error().Err(err).Msg("error in running disable firewall command")
			return
		}
		log.Info().Msg(out)
		// test existing of file
		out, err = remote.RunCommand(ip, pemLocation, "ls /home/ubuntu/go-tss/benchmark_docker/Data/data_local/", digitalOcean)
		if err != nil {
			log.Error().Err(err).Msg("error in running remote command")
			return
		}
		if strings.Contains(out, " No such file or directory") {
			log.Info().Msg("we create the directory")
			out, err = remote.RunCommand(ip, pemLocation, "mkdir -p /home/ubuntu/go-tss/benchmark_docker/Data/data_local/", digitalOcean)
			if err != nil {
				log.Error().Err(err).Msg("error in running remote command")
				return
			}
			log.Info().Msg(out)
		}
		// clone the tss code
		out, err = remote.RunCommand(ip, pemLocation, "git clone https://gitlab.com/thorchain/tss/go-tss.git /home/ubuntu/go-tss/go-tss", digitalOcean)
		if err != nil {
			log.Error().Err(err).Msg("error in running remote command")
			return
		}
		log.Info().Msg(out)

		if !digitalOcean {
			// clone the tss code
			out, err = remote.RunCommand(ip, pemLocation, "chown -R ubuntu.ubuntu /home/ubuntu/go-tss", digitalOcean)
			if err != nil {
				log.Error().Err(err).Msg("error in running remote command")
				return
			}
			log.Info().Msg(out)

		}

	}
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

func doDeployment(hostIPs []string, remoteFilePath, pemPath, awsConfigurePath string) error {
	var isDigitalOcean bool
	if len(pemPath) == 0 {
		isDigitalOcean = true
	}
	// we set we have 5 threads do the command
	worker := 5
	working := worker
	dones := make(chan struct{}, worker)
	jobs := make(chan string, worker)
	done := false
	go addJob(jobs, hostIPs)
	for i := 0; i < worker; i++ {
		go doPrepareJob(jobs, dones)
	}

	for {
		<-dones
		working -= 1
		if working <= 0 {
			done = true
		}
		if done {
			break
		}
	}

	// we send the configuration file and the docker compose file
	err := remote.SendRemote(hostIPs, storagePath, remoteFilePath, pemPath, awsConfigurePath, isDigitalOcean)
	if err != nil {
		log.Error().Msg("fail to update the configuration file to remote nodes")
		return err
	}
	return nil
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
	if err != nil {
		log.Error().Err(err).Msg("fail to read the host file")
		return err
	}
	bootstrapIP, _ := tools.AnalysisIPs(hostIPs[0])
	// everytime we run, we update the bootstrap node IP address
	err = tools.UpdateBootstrapNode(bootstrapIP, num, storagePath)
	if err != nil {
		log.Error().Err(err).Msg("fail to update the bootstrapIP")
		return err
	}
	err = tools.UpdateExternalIP(hostIPs, storagePath)
	if err != nil {
		log.Error().Err(err).Msg("fail to update the external IP")
		return err
	}

	awsConfigurePath, err := tools.GenerateComposeForAWS(storagePath, hostIPs)
	if err != nil {
		log.Error().Err(err).Msgf("fail to generate the aws docker compose file")
		return err
	}
	err = doDeployment(hostIPs, remoteFilePath, pemLocation, awsConfigurePath)
	if err != nil {
		return err
	}
	os.RemoveAll(awsConfigurePath)
	return nil
}

func prepare(pubKeyPath, hostsTablePath string) ([]string, []string, []int, error) {
	input, err := tools.GetInput("please input the nunber of the nodes")
	if err != nil {
		return nil, nil, nil, err
	}
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
	var inputKeys, inputIPs []string
	var ports []int
	var selected []int
	//for i := 0; i < nodeNum; i++ {
	//	selected = append(selected, i)
	//}
	selected = tools.GetRandomPick(nodeNum, len(hostIPs))
	sort.Ints(selected)
	fmt.Printf("---we selected------%v\n", selected)
	for i := 0; i < nodeNum; i++ {
		inputKeys = append(inputKeys, pubKeys[selected[i]])
		inputIPs = append(inputIPs, hostIPs[selected[i]])
		ports = append(ports, 8080)
	}
	return inputKeys, inputIPs, ports, nil
}

func runKeySign(poolKey string, inputKeys, ips []string, ports []int, i, loops int, done chan bool) {
	err := tss.KeySign("hello"+string(i), poolKey, ips, ports, inputKeys)
	if err != nil {
		fmt.Printf("######we quit as saw the error!!!")
		return
	}
	done <- true
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
		ips := make([]string, len(inputIPs))
		for i, el := range inputIPs {
			ips[i], _ = tools.AnalysisIPs(el)
		}

		input, err := tools.GetInput("please input the rounds you want to benchmark")
		if err != nil {
			return
		}
		loops, err := strconv.Atoi(input)
		if err != nil {
			log.Error().Err(err).Msg("invalid input")
			return
		}

		timeBefore := time.Now()
		for i := 0; i < loops; i++ {
			fmt.Printf("-----we run %d/%d tests\n", i, loops)
			poolPubKey, err := tss.KeyGen(inputKeys, ips, ports)
			if err != nil {
				fmt.Printf("We quit As saw errors!!!")
				return
			}
			fmt.Println(poolPubKey)
		}
		fmt.Printf("time we spend ms is %v\n", time.Since(timeBefore).Milliseconds()/int64(loops))
		fmt.Printf("time we spend is %v\n", time.Since(timeBefore)/time.Duration(loops))
	case 3:
		// keysign test
		inputKeys, inputIPs, ports, err := prepare(pubKeyPath, hostsTablePath)
		ips := make([]string, len(inputIPs))
		for i, el := range inputIPs {
			ips[i], _ = tools.AnalysisIPs(el)
		}
		if err != nil {
			return
		}
		timeBefore := time.Now()

		input, err := tools.GetInput("please input the rounds you want to benchmark")
		if err != nil {
			return
		}
		loops, err := strconv.Atoi(input)
		if err != nil {
			log.Error().Err(err).Msg("invalid input")
			return
		}
		done := make(chan bool)
		poolKey := "thorpub1addwnpepqg0vs2njqwhh90vw7sp4qvfuc8m2q8a4ddgcza0mc3k33x6ahzg9xenwkwj"
		for i := 0; i < loops; i++ {
			fmt.Printf("----------------%d\n", i)
			go runKeySign(poolKey, inputKeys, ips, ports, i, loops, done)
			select {
			case <-done:
				continue
			case <-time.After(time.Second * 60):
				panic("error timeout")
			}
		}

		fmt.Printf("time we spend is %v\n", time.Since(timeBefore)/time.Duration(loops))

	case 4:
		inputKeys, inputIPs, ports, err := prepare(pubKeyPath, hostsTablePath)
		if err != nil {
			return
		}
		ips := make([]string, len(inputIPs))
		for i, el := range inputIPs {
			ips[i], _ = tools.AnalysisIPs(el)
		}

		input, err := tools.GetInput("please input the rounds you want to benchmark")
		if err != nil {
			return
		}
		loops, err := strconv.Atoi(input)
		if err != nil {
			log.Error().Err(err).Msg("invalid input")
			return
		}
		var poolPubKey string
		timeBeforeKeyGen := time.Now()
		for i := 0; i < loops; i++ {
			fmt.Printf("-----we run %d/%d tests\n", i, loops)
			poolPubKey, err = tss.KeyGen(inputKeys, ips, ports)
			if err != nil {
				panic("###we quit as we saw error in keyGen")
			}
			// fmt.Println(poolPubKey)
		}
		timeAfterKeyGen := time.Now()
		fmt.Print("now we do the keysign test")

		timeBeforeKeySign := time.Now()
		for i := 0; i < loops; i++ {
			err := tss.KeySign("hello"+string(i), poolPubKey, ips, ports, inputKeys)
			if err != nil {
				panic("###we quit as we saw error in keysign")
			}
		}
		timeAfterKeySign := time.Now()

		fmt.Printf("\ntime we spend for keygen(ms) is %v\n", timeAfterKeyGen.Sub(timeBeforeKeyGen).Milliseconds()/int64(loops))
		fmt.Printf("time we spend for keysign(ms) is %v\n", timeAfterKeySign.Sub(timeBeforeKeySign).Milliseconds()/int64(loops))

	case 5:
		// 16Uiu2HAm8c9uDs34BYfJqb6gaBP2iCj5TayapNptE7zEmUF7bn3e

		inputKeys, _, _, _ := prepare(pubKeyPath, hostsTablePath)
		tools.SetupBech32Prefix()
		for i, el := range inputKeys {
			peer, err := tools.GetPeerIDFromPubKey(el)
			if err != nil {
				fmt.Printf("-----%d-------->%v\n", i, err)
			}

			fmt.Printf("----%d---%v\n", i, peer.String())
		}

	default:
		fmt.Println("not supported!!!")
		return
	}
}
