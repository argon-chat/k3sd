package clusterutils

import (
	"fmt"
	"os/exec"
	"strings"

	utils "github.com/argon-chat/k3sd/pkg/utils"
)

func isHelmRepoAlreadyExists(output string) bool {
	return strings.Contains(output, "already exists")
}

// InstallHelmChart installs or upgrades a Helm chart on the cluster using the provided parameters.
// It adds the repo, updates it, and runs the upgrade/install command.
//
// Parameters:
//
//	kubeconfigPath: Path to the kubeconfig file.
//	releaseName: Name of the Helm release.
//	namespace: Kubernetes namespace for the release.
//	repoName: Name of the Helm repo.
//	repoURL: URL of the Helm repo.
//	chartName: Name of the chart in the repo.
//	chartVersion: Version of the chart to install.
//	valuesFile: Path to a values.yaml file (optional).
//	logger: Logger for output.
//
// Returns:
//
//	Error if any Helm operation fails.
func InstallHelmChart(kubeconfigPath, releaseName, namespace, repoName, repoURL, chartName, chartVersion, valuesFile string, logger *utils.Logger) error {
	if err := helmRepoAdd(repoName, repoURL, logger); err != nil {
		return err
	}
	if err := helmRepoUpdate(logger); err != nil {
		return err
	}
	args := buildHelmArgs(kubeconfigPath, releaseName, namespace, repoName, chartName, chartVersion, valuesFile)
	return helmUpgradeInstall(args, logger)
}

func helmRepoAdd(repoName, repoURL string, logger *utils.Logger) error {
	cmd := exec.Command("helm", "repo", "add", repoName, repoURL)
	out, err := cmd.CombinedOutput()
	if err != nil && !isHelmRepoAlreadyExists(string(out)) {
		logger.LogErr("Helm repo add failed: %v\nOutput: %s", err, string(out))
		return fmt.Errorf("helm repo add failed: %w", err)
	}
	logger.Log("Helm repo add output: %s", string(out))
	return nil
}

func helmRepoUpdate(logger *utils.Logger) error {
	cmd := exec.Command("helm", "repo", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.LogErr("Helm repo update failed: %v\nOutput: %s", err, string(out))
		return fmt.Errorf("helm repo update failed: %w", err)
	}
	logger.Log("Helm repo update output: %s", string(out))
	return nil
}

func buildHelmArgs(kubeconfigPath, releaseName, namespace, repoName, chartName, chartVersion, valuesFile string) []string {
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	baseArgs := []string{"--kubeconfig", kubeconfigPath, "--namespace", namespace, "--version", chartVersion, "--create-namespace", "--wait", "--timeout", "600s"}
	if utils.HelmAtomic {
		baseArgs = append(baseArgs, "--atomic")
	}
	if valuesFile != "" {
		baseArgs = append(baseArgs, "-f", valuesFile)
	}
	return append([]string{"upgrade", "--install", releaseName, chartRef}, baseArgs...)
}

func helmUpgradeInstall(args []string, logger *utils.Logger) error {
	cmd := exec.Command("helm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.LogErr("Helm upgrade/install failed: %v\nOutput: %s", err, string(out))
		return fmt.Errorf("helm upgrade/install failed: %w", err)
	}
	logger.Log("Helm upgrade/install output: %s", string(out))
	return nil
}

// UninstallHelmRelease uninstalls a Helm release from the given namespace.
//
// Parameters:
//
//	kubeconfigPath: Path to the kubeconfig file.
//	releaseName: Name of the Helm release to uninstall.
//	namespace: Kubernetes namespace of the release.
//	logger: Logger for output.
//
// Returns:
//
//	Error if the Helm uninstall command fails.
func UninstallHelmRelease(kubeconfigPath, releaseName, namespace string, logger *utils.Logger) error {
	cmd := exec.Command("helm", "uninstall", releaseName, "--namespace", namespace, "--kubeconfig", kubeconfigPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.LogErr("Helm uninstall failed: %v\nOutput: %s", err, string(out))
		return err
	}
	logger.Log("Helm uninstall output: %s", string(out))
	return nil
}
