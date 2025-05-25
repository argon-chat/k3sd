// Package cluster provides utilities for managing kubeconfig files for Kubernetes clusters.
package cluster

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/argon-chat/k3sd/utils"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// getFirstKey returns the first key in a map, or an empty string if the map is empty.
//
// Parameters:
//
//	m: Map to inspect.
//
// Returns:
//
//	string: First key.
func getFirstKey(m map[string]interface{}) string {
	for k := range m {
		return k
	}
	return ""
}

// patchKubeConfigKeys renames keys in a kubeconfig to match the node name and sets the current context.
//
// Parameters:
//
//	config: Kubeconfig to patch.
//	nodeName: Node name to use.
func patchKubeConfigKeys(config *clientcmdapi.Config, nodeName string) {
	patchKey := func(m map[string]interface{}, set func(string), del func(string)) {
		oldKey := getFirstKey(m)
		if oldKey != "" && oldKey != nodeName {
			set(oldKey)
			del(oldKey)
		}
	}
	patchKey(toMapInterface(config.Clusters),
		func(old string) { config.Clusters[nodeName] = config.Clusters[old] },
		func(old string) { delete(config.Clusters, old) },
	)
	patchKey(toMapInterface(config.AuthInfos),
		func(old string) { config.AuthInfos[nodeName] = config.AuthInfos[old] },
		func(old string) { delete(config.AuthInfos, old) },
	)
	patchKey(toMapInterface(config.Contexts),
		func(old string) { config.Contexts[nodeName] = config.Contexts[old] },
		func(old string) { delete(config.Contexts, old) },
	)
	if ctx, ok := config.Contexts[nodeName]; ok {
		ctx.Cluster = nodeName
		ctx.AuthInfo = nodeName
	}
	config.CurrentContext = nodeName
}

// toMapInterface converts a map of any type to a map of interface{}.
//
// Parameters:
//
//	m: Input map.
//
// Returns:
//
//	map[string]interface{}: Converted map.
func toMapInterface[T any](m map[string]T) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// saveKubeConfig fetches, patches, and saves the kubeconfig for a cluster node.
//
// Parameters:
//
//	client: SSH client.
//	cluster: Cluster information.
//	nodeName: Node name.
//	logger: Logger for output.
func saveKubeConfig(client *ssh.Client, cluster Cluster, nodeName string, logger *utils.Logger) {
	kubeConfig, err := readRemoteKubeConfig(client, cluster.Address, logger)
	if err != nil {
		return
	}
	config, err := parseAndPatchKubeConfig(kubeConfig, cluster.Address, nodeName, logger)
	if err != nil {
		return
	}
	writeKubeConfigToFile(config, logger.Id, nodeName, logger)
}

// readRemoteKubeConfig reads the kubeconfig file from a remote node via SSH.
//
// Parameters:
//
//	client: SSH client.
//	address: Node address.
//	logger: Logger for output.
//
// Returns:
//
//	string: Kubeconfig content.
//	error: Error if reading fails.
func readRemoteKubeConfig(client *ssh.Client, address string, logger *utils.Logger) (string, error) {
	kubeConfig, err := ExecuteRemoteScript(client, "cat /etc/rancher/k3s/k3s.yaml", logger)
	if err != nil {
		logger.Log("Failed to read kubeconfig from %s: %v\n", address, err)
		return "", err
	}
	return kubeConfig, nil
}

// parseAndPatchKubeConfig parses and patches a kubeconfig string for a specific node.
//
// Parameters:
//
//	kubeConfig: Kubeconfig content.
//	address: Node address.
//	nodeName: Node name.
//	logger: Logger for output.
//
// Returns:
//
//	*clientcmdapi.Config: Parsed and patched config.
//	error: Error if parsing fails.
func parseAndPatchKubeConfig(kubeConfig, address, nodeName string, logger *utils.Logger) (*clientcmdapi.Config, error) {
	kubeConfig = strings.Replace(kubeConfig, "127.0.0.1", address, -1)
	config, err := clientcmd.Load([]byte(kubeConfig))
	if err != nil {
		logger.Log("Failed to parse kubeconfig: %v", err)
		return nil, err
	}
	patchKubeConfigKeys(config, nodeName)
	return config, nil
}

// writeKubeConfigToFile writes a kubeconfig to a file.
//
// Parameters:
//
//	config: Kubeconfig to write.
//	loggerId: Logger/session ID.
//	nodeName: Node name.
//	logger: Logger for output.
func writeKubeConfigToFile(config *clientcmdapi.Config, loggerId, nodeName string, logger *utils.Logger) {
	newKubeConfig, err := clientcmd.Write(*config)
	if err != nil {
		logger.Log("Failed to marshal kubeconfig: %v", err)
		return
	}
	kubeConfigPath := path.Join("./kubeconfigs", fmt.Sprintf("%s/%s.yaml", loggerId, nodeName))
	if err := createFileWithErr(kubeConfigPath, string(newKubeConfig)); err != nil {
		logger.Log("Failed to write kubeconfig to file: %v", err)
	}
}

// createFileWithErr creates a file and writes content to it, creating directories as needed.
//
// Parameters:
//
//	filePath: Path to the file.
//	content: Content to write.
//
// Returns:
//
//	error: Error if writing fails.
func createFileWithErr(filePath, content string) error {
	if err := os.MkdirAll(path.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing kubeconfig to file: %v", err)
	}
	return nil
}

// logFiles logs the contents of all kubeconfig files for the current logger/session.
//
// Parameters:
//
//	logger: Logger for output.
func logFiles(logger *utils.Logger) {
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
