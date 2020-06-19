package remote

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"

	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/benchmark_tss/tools"
)

func RunCommand(ip, pemLocation, ins string, digitalOcean bool) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("bash")
	cmdWriter, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Start()
	var cmdSent, username string

	if digitalOcean {
		username = "root"
		cmdSent = fmt.Sprintf("ssh %s@%s -f %s", username, ip, ins)
	} else {
		username = "ubuntu"
		cmdSent = fmt.Sprintf("ssh -i %s  %s@%s -f sudo %s", pemLocation, username, ip, ins)
		// cmdSent = fmt.Sprintf("-i %s %s@%s -f %s", pemLocation, username, ip, ins)
	}
	fmt.Println(cmdSent)
	cmdWriter.Write([]byte(cmdSent + "\n"))
	cmdWriter.Write([]byte("exit" + "\n"))
	cmd.Run()
	err = cmd.Wait()

	return out.String(), err
}

func doCommand(ip, remoteFile, pemLocation, filePath string, digitalOcean bool) error {
	var out bytes.Buffer
	cmd := exec.Command("bash")
	cmdWriter, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stdout = &out
	cmd.Start()
	var cmdSent, username string

	if digitalOcean {
		username = "root"
		cmdSent = fmt.Sprintf("scp  %s %s@%s:%s", filePath, username, ip, remoteFile)
	} else {
		username = "ubuntu"
		cmdSent = fmt.Sprintf("scp -i %s %s %s@%s:%s", pemLocation, filePath, username, ip, remoteFile)
	}
	fmt.Println(cmdSent)

	cmdWriter.Write([]byte(cmdSent + "\n"))
	cmdWriter.Write([]byte("exit" + "\n"))

	err = cmd.Wait()
	return err
}

func addJob(jobs chan<- int, tasksNum int) {
	for i := 0; i < tasksNum; i++ {
		jobs <- i
	}
	close(jobs)
}

func sendTssRunScripts(i int, jobs chan int, ips []string, localFilePath, remoteFilePath, pemLocation, awsConfigurePath string, digitalOcean bool, dones chan<- struct{}) {
	var filePath string
	var dockerPath string
	dockerPath = fmt.Sprintf("%s/docker-compose.yml", localFilePath)
	remoteConfigure := path.Join(remoteFilePath, "run.sh")
	remoteDonfigure := path.Join(remoteFilePath, "docker-compose.yml")
	defer func() {
		dones <- struct{}{}
	}()
	for index := range jobs {

		ip := ips[index]
		if len(ip) == 0 {
			continue
		}
		ip, digitalOcean = tools.AnalysisIPs(ip)
		filePath = fmt.Sprintf("%s/%d/deployed_run.sh", localFilePath, index)
		err := doCommand(ip, remoteConfigure, pemLocation, filePath, digitalOcean)
		if err != nil {
			log.Error().Err(err).Msgf("!!!fail to send to node %s", ip)
		}
		//if !digitalOcean {
		//	awsFolder := path.Join(awsConfigurePath, ip)
		//	dockerPath = path.Join(awsFolder, "docker-compose.sh")
		//}
		err = doCommand(ip, remoteDonfigure, pemLocation, dockerPath, digitalOcean)
		if err != nil {
			log.Error().Err(err).Msgf("!!!fail to send to node %s", ip)
		}
	}
}

func SendRemote(ips []string, localFilePath, remoteFilePath, pemLocation, awsConfigurePath string, isDigitalOcean bool) error {
	var runErr error
	worker := 5
	working := worker
	dones := make(chan struct{}, worker)
	jobs := make(chan int, worker)
	done := false

	go addJob(jobs, len(ips))

	for i := 0; i < worker; i++ {
		go sendTssRunScripts(i, jobs, ips, localFilePath, remoteFilePath, pemLocation, awsConfigurePath, isDigitalOcean, dones)
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

	return runErr
}
