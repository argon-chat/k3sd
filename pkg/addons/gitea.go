package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyGiteaAddon installs and configures the Gitea addon on the cluster if enabled.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyGiteaAddon(cluster *types.Cluster, logger *utils.Logger) {
	addon, ok := cluster.Addons["gitea"]
	if !ok || !addon.Enabled {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyGitea(cluster, kubeconfig, logger, &addon)
}

func applyGitea(clusterObj *types.Cluster, kubeconfigPath string, logger *utils.Logger, addon *types.AddonConfig) {
	substitutions := addon.Subs
	if substitutions == nil {
		substitutions = make(map[string]string)
	}
	if substitutions["${POSTGRES_USER}"] == "" {
		substitutions["${POSTGRES_USER}"] = "gitea"
	}
	if substitutions["${POSTGRES_PASSWORD}"] == "" {
		substitutions["${POSTGRES_PASSWORD}"] = "changeme"
	}
	if substitutions["${POSTGRES_DB}"] == "" {
		substitutions["${POSTGRES_DB}"] = "giteadb"
	}
	manifestPath := addon.Path
	if manifestPath == "" {
		manifestPath = clusterutils.ResolveYamlPath("gitea.yaml")
	}
	clusterutils.ApplyComponentYAML("gitea", kubeconfigPath, manifestPath, logger, substitutions)
	if ingressAddon, ok := clusterObj.Addons["gitea-ingress"]; ok && ingressAddon.Enabled {
		applyGiteaIngress(clusterObj, kubeconfigPath, logger, &ingressAddon)
	}
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "gitea", "default", logger)
}

func applyGiteaIngress(clusterObj *types.Cluster, kubeconfigPath string, logger *utils.Logger, addon *types.AddonConfig) {
	substitutions := addon.Subs
	if substitutions == nil {
		substitutions = make(map[string]string)
	}
	if substitutions["${DOMAIN}"] == "" {
		substitutions["${DOMAIN}"] = clusterObj.Domain
	}
	manifestPath := addon.Path
	if manifestPath == "" {
		manifestPath = clusterutils.ResolveYamlPath("gitea.ingress.yaml")
	}
	clusterutils.ApplyComponentYAML("gitea-ingress", kubeconfigPath, manifestPath, logger, substitutions)
}
