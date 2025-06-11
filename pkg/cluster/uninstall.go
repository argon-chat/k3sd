package cluster

import (
	"fmt"

	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/db"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
	"golang.org/x/crypto/ssh"
)

func uninstallWorker(client *ssh.Client, worker types.Worker, clusterAddress string, logger *utils.Logger) error {
	cmd := fmt.Sprintf("ssh %s@%s \"k3s-agent-uninstall.sh\"", worker.User, worker.Address)
	err := clusterutils.ExecuteCommands(client, []string{cmd}, worker.Password, logger)
	utils.LogIfError(logger, err, "Error uninstalling worker on %s: %v", clusterAddress)
	return err
}

func uninstallMaster(client *ssh.Client, clusterAddress string, logger *utils.Logger) error {
	err := clusterutils.ExecuteCommands(client, []string{"k3s-uninstall.sh"}, clusterAddress, logger)
	utils.LogIfError(logger, err, "Error uninstalling master on %s: %v", clusterAddress)
	return err
}

// UninstallCluster removes all K3s components from the provided clusters.
// It connects to each master and worker node, runs uninstall scripts, and updates cluster state.
//
// Parameters:
//
//	clusters: List of clusters to uninstall.
//	logger: Logger for output.
//
// Returns:
//
//	Updated list of clusters and error if any step fails.
func UninstallCluster(clusters []types.Cluster, logger *utils.Logger) ([]types.Cluster, error) {
	for ci, cluster := range clusters {
		err := db.DeleteClusterRecords(&cluster)
		if err != nil {
			return nil, fmt.Errorf("error deleting cluster records for %s: %v", cluster.Address, err)
		}
		client, err := clusterutils.SSHConnect(cluster.User, cluster.Password, cluster.Address)
		if err != nil {
			return nil, fmt.Errorf("error connecting to cluster %s: %v", cluster.Address, err)
		}
		defer func(client *ssh.Client) {
			err := client.Close()
			if err != nil {
				logger.LogErr("Error closing SSH connection to %s: %v\n", cluster.Address, err)
			} else {
				logger.Log("SSH connection to %s closed successfully.\n", cluster.Address)
			}
		}(client)

		for wi, worker := range cluster.Workers {
			if worker.Done {
				_ = uninstallWorker(client, worker, cluster.Address, logger)
				clusters[ci].Workers[wi].Done = false
			}
		}

		if cluster.Done {
			_ = uninstallMaster(client, cluster.Address, logger)
			clusters[ci].Done = false
		}
	}
	return clusters, nil
}
