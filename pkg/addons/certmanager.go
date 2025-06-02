package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyCertManagerAddon installs and configures the cert-manager addon on the cluster if enabled.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyCertManagerAddon(cluster *types.Cluster, logger *utils.Logger) {
	addon, ok := cluster.Addons["cert-manager"]
	if !ok || !addon.Enabled {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyCertManager(kubeconfig, logger, &addon)
}

func applyCertManager(kubeconfigPath string, logger *utils.Logger, addon *types.AddonConfig) {
	manifestPath := addon.Path
	if manifestPath == "" {
		manifestPath = "https://github.com/cert-manager/cert-manager/releases/download/v1.17.2/cert-manager.yaml"
	}
	clusterutils.ApplyComponentYAML("cert-manager", kubeconfigPath, manifestPath, logger, addon.Subs)
	crdsPath := "https://github.com/cert-manager/cert-manager/releases/download/v1.17.2/cert-manager.crds.yaml"
	clusterutils.ApplyComponentYAML("cert-manager CRDs", kubeconfigPath, crdsPath, logger, nil)
	logger.Log("Waiting for cert-manager-webhook deployment to be ready...")
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "cert-manager", "cert-manager", logger)
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "cert-manager-cainjector", "cert-manager", logger)
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "cert-manager-webhook", "cert-manager", logger)
}
