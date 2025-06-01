package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyCertManagerAddon installs and configures the cert-manager addon on the cluster if enabled by flags.
// It applies the cert-manager manifests and waits for all deployments to be ready.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyCertManagerAddon(cluster *types.Cluster, logger *utils.Logger) {
	if !utils.Flags[utils.FlagCertManager] {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyCertManager(kubeconfig, logger)
}

// applyCertManager applies the cert-manager manifests and waits for deployments to be ready.
func applyCertManager(kubeconfigPath string, logger *utils.Logger) {
	clusterutils.ApplyComponentYAML("cert-manager", kubeconfigPath, "https://github.com/cert-manager/cert-manager/releases/download/v1.17.2/cert-manager.yaml", logger, nil)
	clusterutils.ApplyComponentYAML("cert-manager CRDs", kubeconfigPath, "https://github.com/cert-manager/cert-manager/releases/download/v1.17.2/cert-manager.crds.yaml", logger, nil)
	logger.Log("Waiting for cert-manager-webhook deployment to be ready...")
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "cert-manager", "cert-manager", logger)
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "cert-manager-cainjector", "cert-manager", logger)
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "cert-manager-webhook", "cert-manager", logger)
}
