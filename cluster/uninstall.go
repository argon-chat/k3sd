// Package cluster provides functions for uninstalling clusters and their nodes.
package cluster

import (
	"fmt"

	"github.com/argon-chat/k3sd/utils"
	"golang.org/x/crypto/ssh"
)

// uninstallWorker uninstalls a worker node from a cluster via SSH.
//
// Parameters:
//
//	client: SSH client connection to the cluster.
//	worker: Worker node information.
//	clusterAddress: Address of the cluster.
//	logger: Logger for output.
//
// Returns:
//
//	error: Error if uninstallation fails.
func uninstallWorker(client *ssh.Client, worker Worker, clusterAddress string, logger *utils.Logger) error {
	cmd := fmt.Sprintf("ssh %s@%s \"k3s-agent-uninstall.sh\"", worker.User, worker.Address)
	err := ExecuteCommands(client, []string{cmd}, logger)
	logIfError(logger, err, "Error uninstalling worker on %s: %v", clusterAddress)
	return err
}

// uninstallMaster uninstalls the master node from a cluster via SSH.
//
// Parameters:
//
//	client: SSH client connection to the cluster.
//	clusterAddress: Address of the cluster.
//	logger: Logger for output.
//
// Returns:
//
//	error: Error if uninstallation fails.
func uninstallMaster(client *ssh.Client, clusterAddress string, logger *utils.Logger) error {
	err := ExecuteCommands(client, []string{"k3s-uninstall.sh"}, logger)
	logIfError(logger, err, "Error uninstalling master on %s: %v", clusterAddress)
	return err
}

// UninstallCluster uninstalls all clusters and their worker nodes.
//
// Parameters:
//
//	clusters: List of clusters to uninstall.
//	logger: Logger for output.
//
// Returns:
//
//	[]Cluster: Updated clusters with nodes marked as not done.
//	error: Error if uninstallation fails.
func UninstallCluster(clusters []Cluster, logger *utils.Logger) ([]Cluster, error) {
	for ci, cluster := range clusters {
		client, err := sshConnect(cluster.User, cluster.Password, cluster.Address)
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
