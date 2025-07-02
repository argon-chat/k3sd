package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

func init() {
	RegisterAddonConfigBuilder("cluster-issuer", AddonConfigBuilderFunc(func(domain string, subs map[string]string) map[string]interface{} {
		if subs == nil {
			subs = map[string]string{}
		}
		subs["${DOMAIN}"] = domain
		return map[string]interface{}{
			"enabled": true,
			"subs":    subs,
		}
	}))
}

// ApplyClusterIssuerAddon installs and configures the ClusterIssuer addon on the cluster if enabled.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyClusterIssuerAddon(cluster *types.Cluster, logger *utils.Logger) {
	addon, ok := cluster.Addons["cluster-issuer"]
	if !ok || !addon.Enabled {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyClusterIssuer(kubeconfig, logger, &addon)
}

func applyClusterIssuer(kubeconfigPath string, logger *utils.Logger, addon *types.AddonConfig) {
	substitutions := addon.Subs
	manifestPath := addon.Path
	if manifestPath == "" {
		manifestPath = clusterutils.ResolveYamlPath("clusterissuer.yaml")
	}
	clusterutils.ApplyComponentYAML("clusterissuer", kubeconfigPath, manifestPath, logger, substitutions)
}

// DeleteClusterIssuerAddon uninstalls the ClusterIssuer addon from the cluster.
//
// Parameters:
//
//	cluster: The cluster to uninstall the addon from.
//	logger: Logger for output.
func DeleteClusterIssuerAddon(cluster *types.Cluster, logger *utils.Logger) {
	addon, ok := cluster.Addons["cluster-issuer"]
	if !ok {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	manifestPath := addon.Path
	if manifestPath == "" {
		manifestPath = clusterutils.ResolveYamlPath("clusterissuer.yaml")
	}
	clusterutils.DeleteComponentYAML("clusterissuer", kubeconfig, manifestPath, logger, addon.Subs)
}
