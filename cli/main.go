package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	. "github.com/argon-chat/k3sd/pkg/cluster"
	clusterstore "github.com/argon-chat/k3sd/pkg/clusterstore"
	. "github.com/argon-chat/k3sd/pkg/utils"
)

func main() {
	ParseFlags()

	if VersionFlag {
		fmt.Printf("K3SD version: %s\n", Version)
		os.Exit(0)
	}

	clusters, err := clusterstore.LoadClusters(ConfigPath)
	if err != nil {
		log.Fatalf("failed to load clusters: %v", err)
	}

	logger := NewLogger("cli")
	go logger.LogWorker()
	go logger.LogWorkerErr()
	go logger.LogWorkerFile()
	go logger.LogWorkerCmd()

	checkCommandExists()

	if Uninstall {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Are you sure you want to uninstall the clusters? (yes/y/no/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "yes" || response == "y" {
			clusters, err = UninstallCluster(clusters, logger)
			if err != nil {
				log.Fatalf("failed to uninstall clusters: %v", err)
			}
		} else if response == "no" || response == "n" {
			fmt.Println("Uninstallation canceled.")
			return
		} else {
			log.Fatalf("And just what do you mean by %s?", response)
		}
	} else {
		clusters, err = CreateCluster(clusters, logger, []string{})
		if err != nil {
			log.Fatalf("failed to create clusters: %v", err)
		}
	}

	if err := clusterstore.SaveClusters(ConfigPath, clusters); err != nil {
		log.Fatalf("failed to save clusters: %v", err)
	}
}

func checkCommandExists() {
	commands := []string{
		"linkerd",
		"step",
		"ssh",
	}

	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err != nil {
			log.Fatalf("Command %s not found. Please install it.", cmd)
		}
	}
}
