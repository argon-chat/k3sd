package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyTraefikAddon applies the Traefik values YAML to the cluster if enabled by flags.
// It waits for the Traefik deployment to be ready after applying the manifest.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyTraefikAddon(cluster *types.Cluster, logger *utils.Logger) {
	if !utils.Flags[utils.FlagTraefikValues] {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyTraefikValues(kubeconfig, logger)
}

// applyTraefikValues applies the Traefik values manifest and waits for the deployment to be ready.
func applyTraefikValues(kubeconfigPath string, logger *utils.Logger) {
	clusterutils.ApplyComponentYAML("traefik-values", kubeconfigPath, clusterutils.ResolveYamlPath("traefik-values.yaml"), logger, nil)
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "traefik", "kube-system", logger)
}
