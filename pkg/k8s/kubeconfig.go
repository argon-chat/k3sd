package k8s

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
	"golang.org/x/crypto/ssh"
)

// SaveKubeConfig retrieves the kubeconfig from a remote cluster node and saves it locally.
//
// Parameters:
//
//	client: SSH client connected to the node.
//	cluster: Cluster information.
//	nodeName: Name of the node.
//	logger: Logger for output.
//
// This function fetches the kubeconfig file from the specified cluster node using SSH,
// patches the address if needed, writes it to a local file, and optionally renames the
// kubeconfig context if the cluster specifies a custom context name.
func SaveKubeConfig(client *ssh.Client, cluster types.Cluster, nodeName string, logger *utils.Logger) {
	kubeConfig, err := readRemoteKubeConfig(client, cluster.Address, logger)
	if err != nil {
		logger.Log("Failed to read kubeconfig from %s: %v", cluster.Address, err)
		return
	}
	kubeConfig = patchKubeConfigAddress(kubeConfig, cluster.Address)
	kubeConfigPath := buildKubeConfigPath(logger.Id, nodeName)
	logIfFileWriteErr(kubeConfigPath, kubeConfig, logger)

	if cluster.Context != "" {
		oldContext := getCurrentContextFromKubeconfig(kubeConfig)
		clusterutils.RenameKubeconfigContext(kubeConfigPath, oldContext, cluster.Context, logger)
	}
}

func patchKubeConfigAddress(kubeConfig, address string) string {
	return strings.ReplaceAll(kubeConfig, "127.0.0.1", address)
}

func getCurrentContextFromKubeconfig(kubeConfig string) string {
	for _, line := range strings.Split(kubeConfig, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "current-context:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "current-context:"))
		}
	}
	return ""
}

func buildKubeConfigPath(loggerId, nodeName string) string {
	return path.Join("./kubeconfigs", fmt.Sprintf("%s/%s.yaml", loggerId, nodeName))
}

func logIfFileWriteErr(filePath, content string, logger *utils.Logger) {
	if err := createFileWithErr(filePath, content); err != nil {
		logger.Log("Failed to write kubeconfig to file: %v", err)
	}
}

func readRemoteKubeConfig(client *ssh.Client, address string, logger *utils.Logger) (string, error) {
	kubeConfig, err := clusterutils.ExecuteRemoteScript(client, "cat /etc/rancher/k3s/k3s.yaml", logger)
	if err != nil {
		logger.Log("Failed to read kubeconfig from %s: %v\n", address, err)
		return "", err
	}
	return kubeConfig, nil
}

func createFileWithErr(filePath, content string) error {
	if err := os.MkdirAll(path.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing kubeconfig to file: %v", err)
	}
	return nil
}

// LogFiles logs the contents of all kubeconfig files for the current logger session.
//
// Parameters:
//
//	logger: Logger for output.
//
// This function reads all files in the kubeconfigs directory for the current logger session
// and logs their contents using the provided logger.
func LogFiles(logger *utils.Logger) {
	dir := path.Join("./kubeconfigs", logger.Id)
	files, err := os.ReadDir(dir)
	if err != nil {
		logger.Log("read dir: %v", err)
		return
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fp := path.Join(dir, f.Name())
		data, err := os.ReadFile(fp)
		if err != nil {
			logger.Log("read file: %v", err)
			continue
		}
		logger.LogFile(fp, string(data))
	}
}

// LabelWorkerNode applies labels to the specified worker node using kubectl.
//
// Parameters:
//
//	cluster: The cluster containing the worker node.
//	worker: The worker node to label.
//	logger: Logger for output.
//
// Returns:
//
//	error: Error if labeling fails.
//
// This function builds the kubeconfig path for the given cluster and worker, then calls
// the clusterutils.LabelNode function to apply the labels using kubectl.
func LabelWorkerNode(cluster *types.Cluster, worker *types.Worker, logger *utils.Logger) error {
	kubeconfigPath := path.Join("./kubeconfigs", fmt.Sprintf("%s/%s.yaml", logger.Id, cluster.NodeName))
	return clusterutils.LabelNode(kubeconfigPath, worker.NodeName, worker.GetLabels(), logger)
}
