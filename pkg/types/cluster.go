package types

import "fmt"

// AddonConfig represents the configuration for a built-in addon.
//
// Fields:
//
//	Enabled: bool, whether the addon is enabled
//	Path: string, optional path to a manifest or values file
//	Subs: map[string]string, optional substitutions for templating
type AddonConfig struct {
	Enabled bool              `json:"enabled"`
	Path    string            `json:"path,omitempty"`
	Subs    map[string]string `json:"subs,omitempty"`
}

// CustomAddonConfig represents a user-defined custom addon.
//
// Fields:
//
//	Enabled: bool, whether the custom addon is enabled
//	Helm: *HelmConfig, Helm chart configuration (optional)
//	Manifest: *ManifestConfig, manifest configuration (optional)
type CustomAddonConfig struct {
	Enabled  bool            `json:"enabled"`
	Helm     *HelmConfig     `json:"helm,omitempty"`
	Manifest *ManifestConfig `json:"manifest,omitempty"`
}

// HelmConfig holds Helm chart installation details for a custom addon.
//
// Fields:
//
//	Chart: string, Helm chart name
//	Repo: HelmRepo, Helm repository details
//	Version: string, chart version
//	ValuesFile: string, path to values.yaml
//	Namespace: string, Kubernetes namespace
type HelmConfig struct {
	Chart      string   `json:"chart"`
	Repo       HelmRepo `json:"repo"`
	Version    string   `json:"version"`
	ValuesFile string   `json:"valuesFile"`
	Namespace  string   `json:"namespace"`
}

// HelmRepo holds Helm repository details.
//
// Fields:
//
//	Name: string, repository name
//	URL: string, repository URL
type HelmRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ManifestConfig holds manifest installation details for a custom addon.
//
// Fields:
//
//	Path: string, path to manifest file
//	Subs: map[string]string, substitutions for templating
type ManifestConfig struct {
	Path string            `json:"path"`
	Subs map[string]string `json:"subs,omitempty"`
}

// Cluster represents a K3s cluster configuration, including master and worker nodes, domain, and optional addons.
//
// Fields:
//
//	Worker: Worker, embedded master node configuration
//	Domain: string, domain for cluster-issuer and ingress
//	Context: string, kubeconfig context name
//	PrivateNet: bool, if true, workers are installed from master
//	Workers: []Worker, list of worker nodes
//	Addons: map[string]AddonConfig, built-in addon configs
//	CustomAddons: map[string]CustomAddonConfig, user-defined custom addons
//	LinksTo: []string, list of clusters to link for multicluster
type Cluster struct {
	Worker
	Domain       string                       `json:"domain"`
	Context      string                       `json:"context"`
	PrivateNet   bool                         `json:"privateNet"`
	Workers      []Worker                     `json:"workers"`
	Addons       map[string]AddonConfig       `json:"addons,omitempty"`
	CustomAddons map[string]CustomAddonConfig `json:"customAddons,omitempty"`
	LinksTo      []string                     `json:"linksTo,omitempty"`
}

// Worker represents a node in the cluster (master or worker).
//
// Fields:
//
//	Address: string, IP or hostname
//	User: string, SSH username
//	Password: string, SSH password
//	NodeName: string, Kubernetes node name
//	Labels: map[string]string, node labels
//	Done: bool, internal flag for install status
type Worker struct {
	Address  string            `json:"address"`
	User     string            `json:"user"`
	Password string            `json:"password"`
	NodeName string            `json:"nodeName"`
	Labels   map[string]string `json:"labels"`
	Done     bool              `json:"done"`
}

// GetLabels returns a comma-separated string of labels for the worker.
//
// Parameters:
//
//	(worker): the Worker receiver
//
// Returns:
//
//	string: comma-separated key=value pairs for all labels
func (worker *Worker) GetLabels() string {
	labels := ""
	for k, v := range worker.Labels {
		labels += fmt.Sprintf("%s=%s,", k, v)
	}
	if len(labels) > 0 {
		labels = labels[:len(labels)-1]
	}
	return labels
}
