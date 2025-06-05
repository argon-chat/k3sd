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

// ApplyClusterIssuerAddon applies the ClusterIssuer YAML to the cluster if enabled.
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
