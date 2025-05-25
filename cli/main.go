// The main package for the k3sd CLI tool. Handles cluster creation, uninstallation, and logging.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/argon-chat/k3sd/cluster"
	"github.com/argon-chat/k3sd/utils"
)

// main is the entry point for the k3sd CLI tool.
// It parses command-line flags, loads cluster configuration, and either creates or uninstalls clusters.
// It also sets up logging and saves the updated cluster state.
func main() {
	utils.ParseFlags()

	if utils.VersionFlag {
		fmt.Printf("K3SD version: %s\n", utils.Version)
		os.Exit(0)
	}

	clusters, err := cluster.LoadClusters(utils.ConfigPath)
	if err != nil {
		log.Fatalf("failed to load clusters: %v", err)
	}

	logger := utils.NewLogger("cli")
	go logger.LogWorker()
	go logger.LogWorkerErr()
	go logger.LogWorkerFile()
	go logger.LogWorkerCmd()

	// Uncomment to check for required commands before proceeding.
	// checkCommandExists()

	if utils.Uninstall {
		// Prompt the user for confirmation before uninstalling clusters.
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Are you sure you want to uninstall the clusters? (yes/no): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "yes" {
			clusters, err = cluster.UninstallCluster(clusters, logger)
			if err != nil {
				log.Fatalf("failed to uninstall clusters: %v", err)
			}
		} else {
			fmt.Println("Uninstallation canceled.")
			return
		}
	} else {
		// Create or update clusters as specified in the configuration.
		clusters, err = cluster.CreateCluster(clusters, logger, []string{})
		if err != nil {
			log.Fatalf("failed to create clusters: %v", err)
		}
	}

	// Save the updated cluster state to the configuration file.
	if err := cluster.SaveClusters(utils.ConfigPath, clusters); err != nil {
		log.Fatalf("failed to save clusters: %v", err)
	}
}

// checkCommandExists verifies that all required external commands are available in the system's PATH.
// If any command is missing, the program will terminate with a fatal error.
func checkCommandExists() {
	commands := []string{
		"linkerd",
		"kubectl",
		"step",
		"ssh",
	}

	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err != nil {
			log.Fatalf("Command %s not found. Please install it.", cmd)
		}
	}
}
