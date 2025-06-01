package clusterutils

import (
	"github.com/argon-chat/k3sd/pkg/utils"
)

// ApplyComponentYAML applies a YAML manifest for a given component to the cluster.
// It logs the operation and any errors encountered.
//
// Parameters:
//
//	component: Name of the component being applied.
//	kubeconfigPath: Path to the kubeconfig file.
//	manifest: Path or URL to the manifest YAML.
//	logger: Logger for output.
//	substitutions: Map of string substitutions to apply to the manifest.
func ApplyComponentYAML(component, kubeconfigPath, manifest string, logger *utils.Logger, substitutions map[string]string) {
	logger.Log("Applying %s...", component)
	if err := ApplyYAMLManifest(kubeconfigPath, manifest, logger, substitutions); err != nil {
		logger.Log("%s error: %v", component, err)
	}
}
