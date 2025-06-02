package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyTraefikAddon applies the Traefik values YAML to the cluster if enabled.
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
