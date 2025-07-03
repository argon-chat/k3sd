package cluster

import (
	"fmt"
	"strings"

	"github.com/argon-chat/k3sd/pkg/addons"
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/db"
	k8s "github.com/argon-chat/k3sd/pkg/k8s"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"

	"golang.org/x/crypto/ssh"
)

// CreateCluster provisions and configures all clusters in the provided list.
// It connects to each master node, sets up the cluster, applies addons, and joins workers.
//
// Parameters:
//
//	clusters: List of clusters to create.
//	logger: Logger for output.
//	additional: Additional shell commands to run on the master node.
//
// Returns:
//
//	Updated list of clusters
func CreateCluster(clusters []types.Cluster, logger *utils.Logger, additional []string) []types.Cluster {
	for ci, cluster := range clusters {
		client, err := clusterutils.SSHConnect(cluster.User, cluster.Password, cluster.Address)
		if err != nil {
			logger.LogErr("error connecting to cluster %s: %v", cluster.Address, err)
		}

		defer closeSSHClient(client)

		if err := handleMasterNode(&clusters[ci], client, logger, additional); err != nil {
			logger.LogErr("error handling master node %s: %v", cluster.Address, err)
		}
		if err := setupWorkerNodes(&clusters[ci], client, logger); err != nil {
			logger.LogErr("error setting up worker nodes: %v", err)
		}
		linkerdMC, okMC := cluster.Addons["linkerd-mc"]
		if okMC && linkerdMC.Enabled {
			addons.LinkChannel = append(addons.LinkChannel, &clusters[ci])
		}
		k8s.LogFiles(logger)
	}

	for _, cluster := range clusters {
		version, err := db.InsertCluster(&cluster)
		if err != nil {
			logger.LogErr("error inserting cluster %s: %v", cluster.Address, err)
		}
		applyOptionalComponents(&cluster, version, logger)
	}

	for _, cluster := range addons.LinkChannel {
		addons.LinkClusters(cluster, &clusters, logger)
	}
	return clusters
}

func closeSSHClient(client *ssh.Client) {
	_ = client.Close()
}

func handleMasterNode(cluster *types.Cluster, client *ssh.Client, logger *utils.Logger, additional []string) error {
	return setupMasterNode(cluster, client, logger, additional)
}

func setupMasterNode(cluster *types.Cluster, client *ssh.Client, logger *utils.Logger, additional []string) error {
	if err := runBaseClusterSetup(cluster, client, logger, additional); err != nil {
		return err
	}
	kubeconfigPath := buildKubeconfigPath(logger.Id, cluster.NodeName)
	labelMasterNode(cluster, kubeconfigPath, logger)
	return nil
}

func buildKubeconfigPath(loggerId, nodeName string) string {
	return "./kubeconfigs/" + loggerId + "/" + nodeName + ".yaml"
}

func runBaseClusterSetup(cluster *types.Cluster, client *ssh.Client, logger *utils.Logger, additional []string) error {
	if cluster.Done {
		return nil
	}
	baseCmds := append(baseClusterCommands(*cluster), additional...)
	logger.Log("Connecting to cluster: %s", cluster.Address)
	if err := clusterutils.ExecuteCommands(client, baseCmds, cluster.Password, logger); err != nil {
		return fmt.Errorf("exec master: %v", err)
	}
	markClusterDone(cluster)
	k8s.SaveKubeConfig(client, *cluster, cluster.NodeName, logger)
	return nil
}

func markClusterDone(cluster *types.Cluster) {
	cluster.Done = true
}

func labelMasterNode(cluster *types.Cluster, kubeconfigPath string, logger *utils.Logger) {
	_ = clusterutils.LabelNode(kubeconfigPath, cluster.NodeName, cluster.GetLabels(), logger)
}

func applyOptionalComponents(cluster *types.Cluster, version int, logger *utils.Logger) {
	oldVersion, err := db.GetClusterVersion(cluster, version)
	if err != nil {
		logger.LogErr("error getting old cluster version for %s: %v", cluster.Address, err)
		return
	}
	for name, migration := range addons.AddonRegistry {
		addon, ok := cluster.Addons[name]

		if oldVersion == nil {
			if ok && addon.Enabled {
				migration.Up(cluster, logger)
			} else {
				migration.Down(cluster, logger)
			}
		} else {
			oldAddon, oldOk := oldVersion.Addons[name]
			if oldOk && !ok {
				migration.Down(cluster, logger)
				continue
			}
			if oldOk && ok && oldAddon.Enabled == addon.Enabled {
				continue
			}
			if ok {
				if addon.Enabled {
					migration.Up(cluster, logger)
				} else {
					migration.Down(cluster, logger)
				}
			}
		}
	}

	//hasEnabled := false
	//hasDisabled := false
	//for _, custom := range cluster.CustomAddons {
	//	hasEnabled = hasEnabled || (custom.Enabled && (custom.Manifest != nil || custom.Helm != nil))
	//	hasDisabled = hasDisabled || (!custom.Enabled && (custom.Manifest != nil || custom.Helm != nil))
	//}
	//if hasEnabled {
	addons.ApplyCustomAddons(cluster, logger)
	//}
	//if hasDisabled {
	addons.DeleteCustomAddons(cluster, logger)
	//}
}

func setupWorkerNodes(cluster *types.Cluster, client *ssh.Client, logger *utils.Logger) error {
	return clusterutils.ForEachWorker(cluster.Workers, func(worker *types.Worker) error {
		return joinAndLabelWorker(cluster, worker, client, logger)
	})
}

func joinAndLabelWorker(cluster *types.Cluster, worker *types.Worker, client *ssh.Client, logger *utils.Logger) error {
	markWorkerDone(worker)
	token, err := getK3sToken(client, cluster, logger)
	if err != nil {
		return nil
	}
	if err := joinWorker(cluster, worker, client, logger, token); err != nil {
		return err
	}
	return k8s.LabelWorkerNode(cluster, worker, logger)
}

func markWorkerDone(worker *types.Worker) {
	worker.Done = true
}

func getK3sToken(client *ssh.Client, cluster *types.Cluster, logger *utils.Logger) (string, error) {
	token, err := clusterutils.ExecuteRemoteScript(client, "echo $(k3s token create)", logger)
	utils.LogIfError(logger, err, "token error for %s: %v", cluster.Address)
	return token, err
}

func joinWorker(cluster *types.Cluster, worker *types.Worker, client *ssh.Client, logger *utils.Logger, token string) error {
	if worker.Done {
		return nil
	}
	if cluster.PrivateNet {
		return joinWorkerPrivateNet(cluster, worker, client, logger, token)
	}
	return joinWorkerPublicNet(cluster, worker, logger, token)
}

func joinWorkerPrivateNet(cluster *types.Cluster, worker *types.Worker, client *ssh.Client, logger *utils.Logger, token string) error {
	joinCmds := []string{
		fmt.Sprintf("ssh %s@%s \"sudo apt update && sudo apt install -y curl\"", worker.User, worker.Address),
		fmt.Sprintf("ssh %s@%s \"curl -sfL https://get.k3s.io | K3S_URL=https://%s:6443 K3S_TOKEN='%s' INSTALL_K3S_EXEC='--node-name %s' sh -\"", worker.User, worker.Address, cluster.Address, strings.TrimSpace(token), worker.NodeName),
	}
	if err := clusterutils.ExecuteCommands(client, joinCmds, cluster.Password, logger); err != nil {
		return fmt.Errorf("worker join %s: %v", worker.Address, err)
	}
	return nil
}

func joinWorkerPublicNet(cluster *types.Cluster, worker *types.Worker, logger *utils.Logger, token string) error {
	workerClient, err := clusterutils.SSHConnect(worker.User, worker.Password, worker.Address)
	if err != nil {
		logger.Log("Failed to connect to worker %s directly: %v", worker.Address, err)
		return nil
	}
	defer func() {
		if err := workerClient.Close(); err != nil {
			logger.LogErr("failed to close worker SSH client: %v", err)
		}
	}()
	joinCmds := []string{
		"sudo apt update && sudo apt install -y curl",
		fmt.Sprintf("curl -sfL https://get.k3s.io | K3S_URL=https://%s:6443 K3S_TOKEN='%s' INSTALL_K3S_EXEC='--node-name %s' sh -", cluster.Address, strings.TrimSpace(token), worker.NodeName),
	}
	if err := clusterutils.ExecuteCommands(workerClient, joinCmds, worker.Password, logger); err != nil {
		return fmt.Errorf("worker join %s: %v", worker.Address, err)
	}
	return nil
}

func baseClusterCommands(cluster types.Cluster) []string {
	return []string{
		"sudo apt-get update -y",
		"sudo apt-get install curl wget zip unzip -y",
		fmt.Sprintf("sudo sh -c \"curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--disable traefik --node-name %s' K3S_KUBECONFIG_MODE=\"644\" sh -\"", cluster.NodeName),
		"sleep 10",
	}
}
