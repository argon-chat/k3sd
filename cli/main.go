package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"log"

	"github.com/argon-chat/k3sd/cli/tui"
	clusterpkg "github.com/argon-chat/k3sd/pkg/cluster"
	clusterstorepkg "github.com/argon-chat/k3sd/pkg/clusterstore"
	"github.com/argon-chat/k3sd/pkg/db"
	"github.com/argon-chat/k3sd/pkg/utils"
)

func main() {
	utils.ParseFlags()

	err := downloadAndExtractYamls(utils.Version)
	if err != nil {
		log.Printf("yamls download failed: %v", err)
	}

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

func downloadAndExtractYamls(version string) error {
	baseURL := "https://github.com/argon-chat/k3sd/releases/download/"
	var url string
	if version != "dev" && version != "" {
		url = baseURL + "v" + version + "/yamls.tar.gz"
	} else {
		resp, err := http.Get("https://api.github.com/repos/argon-chat/k3sd/releases/latest")
		if err == nil {
			defer resp.Body.Close()
			type ghRelease struct {
				TagName string `json:"tag_name"`
			}
			var rel ghRelease
			if err := json.NewDecoder(resp.Body).Decode(&rel); err == nil && rel.TagName != "" {
				url = baseURL + rel.TagName + "/yamls.tar.gz"
			}
		}
		if url == "" {
			url = baseURL + utils.Version + "/yamls.tar.gz"
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download yamls: %s", resp.Status)
	}

	home, _ := os.UserHomeDir()
	dest := filepath.Join(home, ".k3sd", "yamls")
	os.MkdirAll(dest, 0700)
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		outName := strings.TrimPrefix(hdr.Name, "yamls/")
		outPath := filepath.Join(dest, outName)
		os.MkdirAll(filepath.Dir(outPath), 0700)
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
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
