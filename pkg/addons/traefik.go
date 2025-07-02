package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyTraefikAddon installs and configures the Traefik addon on the cluster if enabled.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyTraefikAddon(cluster *types.Cluster, logger *utils.Logger) {
	addon, ok := cluster.Addons["traefik"]
	if !ok || !addon.Enabled {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyTraefikValues(kubeconfig, logger, &addon)
}

func applyTraefikValues(kubeconfigPath string, logger *utils.Logger, addon *types.AddonConfig) {
	manifestPath := addon.Path
	if manifestPath == "" {
		manifestPath = clusterutils.ResolveYamlPath("traefik-values.yaml")
	}
	clusterutils.ApplyComponentYAML("traefik-values", kubeconfigPath, manifestPath, logger, addon.Subs)
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "traefik", "kube-system", logger)
}

// DeleteTraefikAddon uninstalls the Traefik addon from the cluster.
//
// Parameters:
//
//	cluster: The cluster to uninstall the addon from.
//	logger: Logger for output.
func DeleteTraefikAddon(cluster *types.Cluster, logger *utils.Logger) {
	addon, ok := cluster.Addons["traefik"]
	if !ok {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	manifestPath := addon.Path
	if manifestPath == "" {
		manifestPath = clusterutils.ResolveYamlPath("traefik-values.yaml")
	}
	clusterutils.DeleteComponentYAML("traefik-values", kubeconfig, manifestPath, logger, addon.Subs)
}
