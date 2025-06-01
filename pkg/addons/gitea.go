package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyGiteaAddon installs and configures the Gitea addon on the cluster if enabled by flags.
// It applies the Gitea manifest, substitutes database credentials, and optionally applies ingress.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyGiteaAddon(cluster *types.Cluster, logger *utils.Logger) {
	if !utils.Flags[utils.FlagGitea] {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyGitea(cluster, kubeconfig, logger)
}

// applyGitea applies the Gitea manifest and waits for the deployment to be ready.
func applyGitea(clusterObj *types.Cluster, kubeconfigPath string, logger *utils.Logger) {
	substitutions := clusterutils.BuildSubstitutions(
		"${POSTGRES_USER}", clusterObj.Gitea.Pg.Username,
		"${POSTGRES_PASSWORD}", clusterObj.Gitea.Pg.Password,
		"${POSTGRES_DB}", clusterObj.Gitea.Pg.DbName,
	)
	clusterutils.ApplyComponentYAML("gitea", kubeconfigPath, clusterutils.ResolveYamlPath("gitea.yaml"), logger, substitutions)
	if utils.Flags[utils.FlagGiteaIngress] {
		applyGiteaIngress(clusterObj, kubeconfigPath, logger)
	}
	clusterutils.WaitForDeploymentReady(kubeconfigPath, "gitea", "default", logger)
}

// applyGiteaIngress applies the Gitea ingress manifest with domain substitutions.
func applyGiteaIngress(clusterObj *types.Cluster, kubeconfigPath string, logger *utils.Logger) {
	substitutions := clusterutils.BuildSubstitutions("${DOMAIN}", clusterObj.Domain, "DOMAIN", clusterObj.Domain)
	clusterutils.ApplyComponentYAML("gitea-ingress", kubeconfigPath, clusterutils.ResolveYamlPath("gitea.ingress.yaml"), logger, substitutions)
}
