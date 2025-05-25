// Package cluster provides functions for creating and configuring Kubernetes clusters and their nodes.
package cluster

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/argon-chat/k3sd/utils"
	"golang.org/x/crypto/ssh"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// logIfError logs an error using the provided logger if the error is not EOF.
//
// Parameters:
//
//	logger: Logger for output.
//	err: Error to check and log.
//	format: Format string for the log message.
//	args: Arguments for the format string.
func logIfError(logger *utils.Logger, err error, format string, args ...interface{}) {
	if err != nil && err.Error() != "EOF" {
		logger.LogErr(format, append(args, err)...)
	}
}

// buildSubstitutions creates a map from a variadic list of key-value string pairs.
//
// Parameters:
//
//	pairs: Alternating key, value pairs.
//
// Returns:
//
//	map[string]string: Map of substitutions.
func buildSubstitutions(pairs ...string) map[string]string {
	subs := make(map[string]string)
	for i := 0; i+1 < len(pairs); i += 2 {
		subs[pairs[i]] = pairs[i+1]
	}
	return subs
}

// forEachWorker applies a function to each worker in the slice.
//
// Parameters:
//
//	workers: Slice of Worker structs.
//	fn: Function to apply to each worker.
//
// Returns:
//
//	error: Error if any function application fails.
func forEachWorker(workers []Worker, fn func(*Worker) error) error {
	for i := range workers {
		if err := fn(&workers[i]); err != nil {
			return err
		}
	}
	return nil
}

// ensureNamespace ensures that a Kubernetes namespace exists, creating it if necessary.
//
// Parameters:
//
//	kubeconfigPath: Path to kubeconfig.
//	namespace: Namespace to ensure.
//	logger: Logger for output.
func ensureNamespace(kubeconfigPath, namespace string, logger *utils.Logger) {
	if namespace != "default" && namespace != "kube-system" {
		cmd := exec.Command("kubectl", "--kubeconfig", kubeconfigPath, "create", "namespace", namespace)
		_ = cmd.Run()
		logger.Log("Ensured namespace %s exists", namespace)
	}
}

// installHelmRelease installs a Helm release for a given component.
//
// Parameters:
//
//	component: Name of the component.
//	kubeconfigPath: Path to kubeconfig.
//	releaseName: Helm release name.
//	namespace: Kubernetes namespace.
//	repoName: Helm repository name.
//	repoURL: Helm repository URL.
//	chartName: Chart name.
//	chartVersion: Chart version.
//	valuesFile: Path to values file.
//	logger: Logger for output.
func installHelmRelease(component, kubeconfigPath, releaseName, namespace, repoName, repoURL, chartName, chartVersion, valuesFile string, logger *utils.Logger) {
	logger.Log("Installing %s via Helm...", component)
	if err := installHelmChart(kubeconfigPath, releaseName, namespace, repoName, repoURL, chartName, chartVersion, valuesFile, logger); err != nil {
		logger.Log("%s Helm install error: %v", component, err)
	}
}

// applyComponentYAML applies a YAML manifest for a component to the cluster.
//
// Parameters:
//
//	component: Name of the component.
//	kubeconfigPath: Path to kubeconfig.
//	manifest: Path or URL to manifest.
//	logger: Logger for output.
//	substitutions: Substitutions to apply in the manifest.
func applyComponentYAML(component, kubeconfigPath, manifest string, logger *utils.Logger, substitutions map[string]string) {
	logger.Log("Applying %s...", component)
	if err := applyYAMLManifest(kubeconfigPath, manifest, logger, substitutions); err != nil {
		logger.Log("%s error: %v", component, err)
	}
}

// labelNode applies labels to a Kubernetes node.
//
// Parameters:
//
//	kubeconfigPath: Path to kubeconfig.
//	nodeName: Name of the node.
//	labels: Labels to apply.
//	logger: Logger for output.
//
// Returns:
//
//	error: Error if labeling fails.
func labelNode(kubeconfigPath, nodeName string, labels map[string]string, logger *utils.Logger) error {
	clientset, err := getKubeClient(kubeconfigPath)
	if err != nil {
		logger.Log("Failed to create k8s client for node %s: %v", nodeName, err)
		return err
	}
	labelBytes, err := json.Marshal(labels)
	if err != nil {
		logger.Log("Failed to marshal node labels for %s: %v", nodeName, err)
		return err
	}
	patch := fmt.Sprintf(`{"metadata":{"labels":%s}}`, string(labelBytes))
	_, err = clientset.CoreV1().Nodes().Patch(context.TODO(), nodeName, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		logger.Log("Failed to label node %s: %v", nodeName, err)
		return err
	} else {
		logger.Log("Labeled node %s", nodeName)
	}
	return nil
}

// getKubeClient creates a Kubernetes client from a kubeconfig file.
//
// Parameters:
//
//	kubeconfigPath: Path to kubeconfig.
//
// Returns:
//
//	*kubernetes.Clientset: Kubernetes client.
//	error: Error if client creation fails.
func getKubeClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// CreateCluster creates and configures all clusters and their nodes.
//
// Parameters:
//
//	clusters: List of clusters to create.
//	logger: Logger for output.
//	additional: Additional commands to run during setup.
//
// Returns:
//
//	[]Cluster: Updated clusters.
//	error: Error if creation fails.
func CreateCluster(clusters []Cluster, logger *utils.Logger, additional []string) ([]Cluster, error) {
	for ci, cluster := range clusters {
		client, err := sshConnect(cluster.User, cluster.Password, cluster.Address)
		if err != nil {
			return nil, err
		}
		defer func(client *ssh.Client) {
			_ = client.Close()
		}(client)

		if !cluster.Done {
			if err := setupMasterNode(&clusters[ci], client, logger, additional); err != nil {
				return nil, err
			}
		}

		if err := setupWorkerNodes(&clusters[ci], client, logger); err != nil {
			return nil, err
		}

		logFiles(logger)
	}
	return clusters, nil
}

// setupMasterNode sets up the master node for a cluster.
//
// Parameters:
//
//	cluster: Cluster to set up.
//	client: SSH client.
//	logger: Logger for output.
//	additional: Additional commands to run.
//
// Returns:
//
//	error: Error if setup fails.
func setupMasterNode(cluster *Cluster, client *ssh.Client, logger *utils.Logger, additional []string) error {
	if err := runBaseClusterSetup(cluster, client, logger, additional); err != nil {
		return err
	}
	kubeconfigPath := path.Join("./kubeconfigs", fmt.Sprintf("%s/%s.yaml", logger.Id, cluster.NodeName))
	labelMasterNode(cluster, kubeconfigPath, logger)
	applyOptionalComponents(cluster, kubeconfigPath, logger)
	return nil
}

// runBaseClusterSetup runs the base setup commands for a cluster's master node.
//
// Parameters:
//
//	cluster: Cluster to set up.
//	client: SSH client.
//	logger: Logger for output.
//	additional: Additional commands to run.
//
// Returns:
//
//	error: Error if setup fails.
func runBaseClusterSetup(cluster *Cluster, client *ssh.Client, logger *utils.Logger, additional []string) error {
	baseCmds := append(baseClusterCommands(*cluster), additional...)
	logger.Log("Connecting to cluster: %s", cluster.Address)
	if err := ExecuteCommands(client, baseCmds, logger); err != nil {
		return fmt.Errorf("exec master: %v", err)
	}
	cluster.Done = true
	saveKubeConfig(client, *cluster, cluster.NodeName, logger)
	return nil
}

// labelMasterNode applies labels to the master node of a cluster.
//
// Parameters:
//
//	cluster: Cluster whose master node to label.
//	kubeconfigPath: Path to kubeconfig.
//	logger: Logger for output.
func labelMasterNode(cluster *Cluster, kubeconfigPath string, logger *utils.Logger) {
	_ = labelNode(kubeconfigPath, cluster.NodeName, cluster.Labels, logger)
}

// applyOptionalComponents applies optional components to the cluster based on flags.
//
// Parameters:
//
//	cluster: Cluster to modify.
//	kubeconfigPath: Path to kubeconfig.
//	logger: Logger for output.
func applyOptionalComponents(cluster *Cluster, kubeconfigPath string, logger *utils.Logger) {
	if utils.Flags["cert-manager"] {
		applyCertManager(kubeconfigPath, logger)
	}
	if utils.Flags["traefik-values"] {
		applyTraefikValues(kubeconfigPath, logger)
	}
	if utils.Flags["clusterissuer"] {
		applyClusterIssuer(cluster, kubeconfigPath, logger)
	}
	if utils.Flags["gitea"] {
		applyGitea(cluster, kubeconfigPath, logger)
	}
	if utils.Flags["prometheus"] {
		applyPrometheus(kubeconfigPath, logger)
	}
	if utils.Flags["linkerd"] {
		runLinkerdInstall(*cluster, logger, false)
	}
	if utils.Flags["linkerd-mc"] {
		runLinkerdInstall(*cluster, logger, true)
	}
}

// applyCertManager applies the cert-manager YAMLs to the cluster.
//
// Parameters:
//
//	kubeconfigPath: Path to kubeconfig.
//	logger: Logger for output.
func applyCertManager(kubeconfigPath string, logger *utils.Logger) {
	applyComponentYAML("cert-manager", kubeconfigPath, "https://github.com/cert-manager/cert-manager/releases/download/v1.17.2/cert-manager.yaml", logger, nil)
	applyComponentYAML("cert-manager CRDs", kubeconfigPath, "https://github.com/cert-manager/cert-manager/releases/download/v1.17.2/cert-manager.crds.yaml", logger, nil)
	logger.Log("Waiting for cert-manager deployment to be ready...")
	time.Sleep(30 * time.Second)
}

// applyTraefikValues applies the Traefik values YAML to the cluster.
//
// Parameters:
//
//	kubeconfigPath: Path to kubeconfig.
//	logger: Logger for output.
func applyTraefikValues(kubeconfigPath string, logger *utils.Logger) {
	applyComponentYAML("traefik-values", kubeconfigPath, resolveYamlPath("traefik-values.yaml"), logger, nil)
}

// applyClusterIssuer applies the ClusterIssuer YAML to the cluster, with domain substitutions.
//
// Parameters:
//
//	cluster: Cluster to modify.
//	kubeconfigPath: Path to kubeconfig.
//	logger: Logger for output.
func applyClusterIssuer(cluster *Cluster, kubeconfigPath string, logger *utils.Logger) {
	substitutions := buildSubstitutions("${DOMAIN}", cluster.Domain, "DOMAIN", cluster.Domain)
	applyComponentYAML("clusterissuer", kubeconfigPath, resolveYamlPath("clusterissuer.yaml"), logger, substitutions)
}

// applyGitea applies the Gitea YAML to the cluster, with Postgres substitutions.
//
// Parameters:
//
//	cluster: Cluster to modify.
//	kubeconfigPath: Path to kubeconfig.
//	logger: Logger for output.
func applyGitea(cluster *Cluster, kubeconfigPath string, logger *utils.Logger) {
	substitutions := buildSubstitutions(
		"${POSTGRES_USER}", cluster.Gitea.Pg.Username,
		"${POSTGRES_PASSWORD}", cluster.Gitea.Pg.Password,
		"${POSTGRES_DB}", cluster.Gitea.Pg.DbName,
	)
	applyComponentYAML("gitea", kubeconfigPath, resolveYamlPath("gitea.yaml"), logger, substitutions)
	if utils.Flags["gitea-ingress"] {
		applyGiteaIngress(cluster, kubeconfigPath, logger)
	}
}

// applyGiteaIngress applies the Gitea Ingress YAML to the cluster, with domain substitutions.
//
// Parameters:
//
//	cluster: Cluster to modify.
//	kubeconfigPath: Path to kubeconfig.
//	logger: Logger for output.
func applyGiteaIngress(cluster *Cluster, kubeconfigPath string, logger *utils.Logger) {
	substitutions := buildSubstitutions("${DOMAIN}", cluster.Domain, "DOMAIN", cluster.Domain)
	applyComponentYAML("gitea-ingress", kubeconfigPath, resolveYamlPath("gitea.ingress.yaml"), logger, substitutions)
}

// applyPrometheus installs the Prometheus stack via Helm in the monitoring namespace.
//
// Parameters:
//
//	kubeconfigPath: Path to kubeconfig.
//	logger: Logger for output.
func applyPrometheus(kubeconfigPath string, logger *utils.Logger) {
	ensureNamespace(kubeconfigPath, "monitoring", logger)
	installHelmRelease(
		"Prometheus stack",
		kubeconfigPath,
		"kube-prom-stack",
		"monitoring",
		"prometheus-community",
		"https://prometheus-community.github.io/helm-charts",
		"kube-prometheus-stack",
		"35.5.1",
		resolveYamlPath("prom-stack-values.yaml"),
		logger,
	)
}

// resolveYamlPath returns the full path to a YAML file for installing additional components.
// If --yamls-path is set, it is used as the prefix. Otherwise, it checks ./yamls then ~/.k3sd/yamls.
func resolveYamlPath(filename string) string {
	if utils.YamlsPath != "" {
		return path.Join(utils.YamlsPath, filename)
	}
	if _, err := os.Stat("yamls/" + filename); err == nil {
		return "yamls/" + filename
	}
	home, err := os.UserHomeDir()
	if err == nil {
		candidate := path.Join(home, ".k3sd", "yamls", filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// fallback to just the filename (may be absolute or relative)
	return filename
}

// setupWorkerNodes sets up all worker nodes for a cluster.
//
// Parameters:
//
//	cluster: Cluster to modify.
//	client: SSH client.
//	logger: Logger for output.
//
// Returns:
//
//	error: Error if setup fails.
func setupWorkerNodes(cluster *Cluster, client *ssh.Client, logger *utils.Logger) error {
	return forEachWorker(cluster.Workers, func(worker *Worker) error {
		if worker.Done {
			return nil
		}
		return joinAndLabelWorker(cluster, worker, client, logger)
	})
}

// joinAndLabelWorker joins a worker node to the cluster and applies labels.
//
// Parameters:
//
//	cluster: Cluster to modify.
//	worker: Worker node to join and label.
//	client: SSH client.
//	logger: Logger for output.
//
// Returns:
//
//	error: Error if joining or labeling fails.
func joinAndLabelWorker(cluster *Cluster, worker *Worker, client *ssh.Client, logger *utils.Logger) error {
	worker.Done = true
	token, err := ExecuteRemoteScript(client, "echo $(k3s token create)", logger)
	logIfError(logger, err, "token error for %s: %v", cluster.Address)
	if err != nil {
		return nil
	}
	if err := joinWorker(cluster, worker, client, logger, token); err != nil {
		return err
	}
	return labelWorkerNode(cluster, worker, logger)
}

// joinWorker joins a worker node to the cluster using the provided token.
//
// Parameters:
//
//	cluster: Cluster to modify.
//	worker: Worker node to join.
//	client: SSH client.
//	logger: Logger for output.
//	token: Cluster join token.
//
// Returns:
//
//	error: Error if joining fails.
func joinWorker(cluster *Cluster, worker *Worker, client *ssh.Client, logger *utils.Logger, token string) error {
	if cluster.PrivateNet {
		joinCmds := []string{
			fmt.Sprintf("ssh %s@%s \"sudo apt update && sudo apt install -y curl\"", worker.User, worker.Address),
			fmt.Sprintf("ssh %s@%s \"curl -sfL https://get.k3s.io | K3S_URL=https://%s:6443 K3S_TOKEN='%s' INSTALL_K3S_EXEC='--node-name %s' sh -\"", worker.User, worker.Address, cluster.Address, strings.TrimSpace(token), worker.NodeName),
		}
		if err := ExecuteCommands(client, joinCmds, logger); err != nil {
			return fmt.Errorf("worker join %s: %v", worker.Address, err)
		}
	} else {
		workerClient, err := sshConnect(worker.User, worker.Password, worker.Address)
		if err != nil {
			logger.Log("Failed to connect to worker %s directly: %v", worker.Address, err)
			return nil
		}
		defer workerClient.Close()
		joinCmds := []string{
			"sudo apt update && sudo apt install -y curl",
			fmt.Sprintf("curl -sfL https://get.k3s.io | K3S_URL=https://%s:6443 K3S_TOKEN='%s' INSTALL_K3S_EXEC='--node-name %s' sh -", cluster.Address, strings.TrimSpace(token), worker.NodeName),
		}
		if err := ExecuteCommands(workerClient, joinCmds, logger); err != nil {
			return fmt.Errorf("worker join %s: %v", worker.Address, err)
		}
	}
	return nil
}

// DRY: Use generic node labeling function
// labelWorkerNode applies labels to a worker node in the cluster.
//
// Parameters:
//
//	cluster: Cluster to modify.
//	worker: Worker node to label.
//	logger: Logger for output.
//
// Returns:
//
//	error: Error if labeling fails.
func labelWorkerNode(cluster *Cluster, worker *Worker, logger *utils.Logger) error {
	kubeconfigPath := path.Join("./kubeconfigs", fmt.Sprintf("%s/%s.yaml", logger.Id, cluster.NodeName))
	return labelNode(kubeconfigPath, worker.NodeName, worker.Labels, logger)
}

// pipeAndLog runs a command and logs its stdout and stderr output.
//
// Parameters:
//
//	cmd: Command to run.
//	logger: Logger for output.
func pipeAndLog(cmd *exec.Cmd, logger *utils.Logger) {
	outPipe, _ := cmd.StdoutPipe()
	errPipe, _ := cmd.StderrPipe()
	_ = cmd.Start()
	go streamOutput(outPipe, false, logger)
	go streamOutput(errPipe, true, logger)
	_ = cmd.Wait()
	logger.Log("Command executed successfully")
}

// pipeAndApply runs a command, collects its YAML output, and applies it to the cluster.
//
// Parameters:
//
//	cmd: Command to run.
//	kubeconfig: Path to kubeconfig.
//	logger: Logger for output.
func pipeAndApply(cmd *exec.Cmd, kubeconfig string, logger *utils.Logger) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	_ = cmd.Start()

	var yaml strings.Builder
	collectYAML(stdout, &yaml)
	go streamOutput(stderr, true, logger)
	_ = cmd.Wait()

	applyYAMLToCluster(yaml.String(), kubeconfig, logger)
}

// collectYAML reads lines from a reader and appends them to a YAML string builder.
//
// Parameters:
//
//	r: Reader to read from.
//	yaml: String builder to append YAML to.
func collectYAML(r io.Reader, yaml *strings.Builder) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		yaml.WriteString(scanner.Text() + "\n")
	}
}

// applyYAMLToCluster applies a YAML string to a Kubernetes cluster using kubectl.
//
// Parameters:
//
//	yaml: YAML content to apply.
//	kubeconfig: Path to kubeconfig.
//	logger: Logger for output.
func applyYAMLToCluster(yaml string, kubeconfig string, logger *utils.Logger) {
	apply := exec.Command("kubectl", "--kubeconfig", kubeconfig, "apply", "-f", "-")
	apply.Stdin = strings.NewReader(yaml)
	out, err := apply.CombinedOutput()
	if err != nil {
		log.Fatalf("apply failed: %v\n%s", err, string(out))
	}
	logger.Log("Apply output:\n%s", string(out))
}

// baseClusterCommands returns the base setup commands for a cluster's master node.
//
// Parameters:
//
//	cluster: Cluster to set up.
//
// Returns:
//
//	[]string: List of setup commands.
func baseClusterCommands(cluster Cluster) []string {
	return []string{
		"sudo apt-get update -y",
		"sudo apt-get install curl wget zip unzip -y",
		fmt.Sprintf("curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--disable traefik --node-name %s' K3S_KUBECONFIG_MODE=\"644\" sh -", cluster.NodeName),
		"sleep 10",
	}
}
