package types

import "fmt"

// AddonConfig represents the configuration for a built-in addon.
type AddonConfig struct {
	Enabled bool              `json:"enabled"`
	Path    string            `json:"path,omitempty"`
	Subs    map[string]string `json:"subs,omitempty"`
}

// CustomAddonConfig represents a user-defined custom addon.
type CustomAddonConfig struct {
	Enabled  bool            `json:"enabled"`
	Helm     *HelmConfig     `json:"helm,omitempty"`
	Manifest *ManifestConfig `json:"manifest,omitempty"`
}

// HelmConfig holds Helm chart installation details for a custom addon.
type HelmConfig struct {
	Chart      string   `json:"chart"`
	Repo       HelmRepo `json:"repo"`
	Version    string   `json:"version"`
	ValuesFile string   `json:"valuesFile"`
	Namespace  string   `json:"namespace"`
}

// HelmRepo holds Helm repository details.
type HelmRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ManifestConfig holds manifest installation details for a custom addon.
type ManifestConfig struct {
	Path string            `json:"path"`
	Subs map[string]string `json:"subs,omitempty"`
}

// Cluster represents a K3s cluster configuration, including master and worker nodes, domain, and optional Gitea config.
//
// Fields:
//
//	Worker:      Embedded master node configuration (inherits Worker fields)
//	Domain:      Domain for cluster-issuer and ingress
//	Gitea:       Gitea configuration (see Gitea struct)
//	PrivateNet:  If true, workers are installed from master
//	Workers:     List of worker nodes
//	Addons:      Map of built-in addon configs (e.g. gitea, cert-manager)
//	CustomAddons: Map of user-defined custom addons
//
// Cluster represents a K3s cluster configuration, including master and worker nodes, domain, context, and optional addons.
type Cluster struct {
	Worker
	Domain       string                       `json:"domain"`
	Context      string                       `json:"context"`
	PrivateNet   bool                         `json:"privateNet"`
	Workers      []Worker                     `json:"workers"`
	Addons       map[string]AddonConfig       `json:"addons,omitempty"`
	CustomAddons map[string]CustomAddonConfig `json:"customAddons,omitempty"`
}

// Worker represents a node in the cluster (master or worker).
//
// Fields:
//
//	Address:   IP or hostname
//	User:      SSH username
//	Password:  SSH password
//	NodeName:  Kubernetes node name
//	Labels:    Node labels
//	Done:      Internal flag for install status
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
// Returns:
//
//	string: Comma-separated key=value pairs for all labels.
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
