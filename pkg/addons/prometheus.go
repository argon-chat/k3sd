package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyPrometheusAddon installs and configures the Prometheus stack on the cluster if enabled.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyPrometheusAddon(cluster *types.Cluster, logger *utils.Logger) {
	addon, ok := cluster.Addons["prometheus"]
	if !ok || !addon.Enabled {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyPrometheus(kubeconfig, logger, &addon)
}

func applyPrometheus(kubeconfigPath string, logger *utils.Logger, addon *types.AddonConfig) {
	clusterutils.EnsureNamespace(kubeconfigPath, "monitoring", logger)
	valuesFile := addon.Path
	if valuesFile == "" {
		valuesFile = clusterutils.ResolveYamlPath("prom-stack-values.yaml")
	}
	if err := clusterutils.InstallHelmChart(
		kubeconfigPath,
		"kube-prom-stack",
		"monitoring",
		"prometheus-community",
		"https://prometheus-community.github.io/helm-charts",
		"kube-prometheus-stack",
		"35.5.1",
		valuesFile,
		logger,
	); err != nil {
		logger.LogErr("failed to install Prometheus Helm chart: %v", err)
	}
}
