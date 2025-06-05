package addons

import (
	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyCustomAddons installs all enabled custom addons (manifest or Helm) for the given cluster.
//
// Parameters:
//
//	cluster: The cluster to apply custom addons to.
//	logger: Logger for output.
func ApplyCustomAddons(cluster *types.Cluster, logger *utils.Logger) {
	for name, addon := range cluster.CustomAddons {
		if !addon.Enabled {
			continue
		}
		if addon.Manifest != nil {
			applyCustomManifestAddon(name, cluster, addon.Manifest, logger)
		}
		if addon.Helm != nil {
			applyCustomHelmAddon(name, cluster, addon.Helm, logger)
		}
	}
}

func applyCustomManifestAddon(name string, cluster *types.Cluster, manifest *types.ManifestConfig, logger *utils.Logger) {
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	manifestPath := manifest.Path
	subs := manifest.Subs
	if manifestPath == "" {
		logger.Log("Custom manifest addon '%s' missing path", name)
		return
	}
	logger.Log("Applying custom manifest addon '%s' from %s", name, manifestPath)
	clusterutils.ApplyComponentYAML(name, kubeconfig, manifestPath, logger, subs)
}

func applyCustomHelmAddon(name string, cluster *types.Cluster, helm *types.HelmConfig, logger *utils.Logger) {
	kubeconfig := clusterutils.KubeConfigPath(cluster, logger)
	if helm.Chart == "" || helm.Repo.URL == "" {
		logger.Log("Custom Helm addon '%s' missing chart or repo URL", name)
		return
	}
	logger.Log("Installing custom Helm addon '%s' (chart: %s, repo: %s)", name, helm.Chart, helm.Repo.URL)
	namespace := helm.Namespace
	if namespace == "" {
		namespace = "default"
	}
	err := clusterutils.InstallHelmChart(
		kubeconfig,
		name,
		namespace,
		helm.Repo.Name,
		helm.Repo.URL,
		helm.Chart,
		helm.Version,
		helm.ValuesFile,
		logger,
	)
	if err != nil {
		logger.LogErr("Helm install failed for custom addon '%s': %v", name, err)
	}
}
