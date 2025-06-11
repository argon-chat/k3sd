package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/argon-chat/k3sd/cli/tui"
	clusterpkg "github.com/argon-chat/k3sd/pkg/cluster"
	clusterstorepkg "github.com/argon-chat/k3sd/pkg/clusterstore"
	"github.com/argon-chat/k3sd/pkg/db"
	"github.com/argon-chat/k3sd/pkg/utils"
)

func main() {
	utils.ParseFlags()

	ctx, err := db.OpenGormDB(db.GetDBPath())
	if err != nil {
		panic("failed to open database: " + err.Error())
	}
	// TODO: this really needs to be placed in a more appropriate location
	db.DbCtx = ctx

	if utils.VersionFlag {
		fmt.Printf("K3SD version: %s\n", utils.Version)
		os.Exit(0)
	}

	if utils.GenerateFlag {
		err := tui.RunGenerateTUI()
		if err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	clusters, err := clusterstorepkg.LoadClusters(utils.ConfigPath)
	if err != nil {
		log.Fatalf("failed to load clusters: %v", err)
	}

	logger := utils.NewLogger("cli")
	go logger.LogWorker()
	go logger.LogWorkerErr()
	go logger.LogWorkerFile()
	go logger.LogWorkerCmd()

	checkCommandExists()

	if utils.Uninstall {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Are you sure you want to uninstall the clusters? (yes/y/no/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		switch response {
		case "yes", "y":
			clusters, err = clusterpkg.UninstallCluster(clusters, logger)
			if err != nil {
				log.Fatalf("failed to uninstall clusters: %v", err)
			}
		case "no", "n":
			fmt.Println("Uninstallation canceled.")
			return
		default:
			log.Fatalf("And just what do you mean by %s?", response)
		}
	} else {
		clusters, err = clusterpkg.CreateCluster(clusters, logger, []string{})
		if err != nil {
			log.Fatalf("failed to create clusters: %v", err)
		}
	}

	if err := clusterstorepkg.SaveClusters(utils.ConfigPath, clusters); err != nil {
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
