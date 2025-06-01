package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyPrometheusAddon installs and configures the Prometheus stack on the cluster if enabled by flags.
// It uses Helm to install the kube-prometheus-stack chart and applies custom values.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyPrometheusAddon(cluster *types.Cluster, logger *utils.Logger) {
	if !utils.Flags[utils.FlagPrometheus] {
		return
	}
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	applyPrometheus(kubeconfig, logger)
}

// applyPrometheus installs the kube-prometheus-stack Helm chart and applies values.
func applyPrometheus(kubeconfigPath string, logger *utils.Logger) {
	clusterutils.EnsureNamespace(kubeconfigPath, "monitoring", logger)
	clusterutils.InstallHelmChart(
		kubeconfigPath,
		"kube-prom-stack",
		"monitoring",
		"prometheus-community",
		"https://prometheus-community.github.io/helm-charts",
		"kube-prometheus-stack",
		"35.5.1",
		clusterutils.ResolveYamlPath("prom-stack-values.yaml"),
		logger,
	)
}
