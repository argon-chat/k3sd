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

// DeletePrometheusAddon uninstalls the Prometheus stack from the cluster by uninstalling the Helm release.
//
// Parameters:
//
//	cluster: The cluster to uninstall the addon from.
//	logger: Logger for output.
func DeletePrometheusAddon(cluster *types.Cluster, logger *utils.Logger) {
	_, ok := cluster.Addons["prometheus"]
	if !ok {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	releaseName := "kube-prom-stack"
	namespace := "monitoring"
	if err := clusterutils.UninstallHelmRelease(kubeconfig, releaseName, namespace, logger); err != nil {
		logger.LogErr("failed to uninstall Prometheus Helm release: %v", err)
	}
}
