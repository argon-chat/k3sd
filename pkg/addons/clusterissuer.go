package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyClusterIssuerAddon applies the ClusterIssuer YAML to the cluster if enabled by flags.
// It substitutes the domain and applies the manifest.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyClusterIssuerAddon(cluster *types.Cluster, logger *utils.Logger) {
	if !utils.Flags[utils.FlagClusterIssuer] {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyClusterIssuer(cluster, kubeconfig, logger)
}

// applyClusterIssuer applies the ClusterIssuer manifest with domain substitutions.
func applyClusterIssuer(cluster *types.Cluster, kubeconfigPath string, logger *utils.Logger) {
	substitutions := clusterutils.BuildSubstitutions("${DOMAIN}", cluster.Domain, "DOMAIN", cluster.Domain)
	clusterutils.ApplyComponentYAML("clusterissuer", kubeconfigPath, clusterutils.ResolveYamlPath("clusterissuer.yaml"), logger, substitutions)
}
