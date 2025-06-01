package types

import "fmt"

// Cluster represents a K3s cluster configuration, including master and worker nodes, domain, and optional Gitea config.
//
// Fields:
//
//	Worker:      Embedded master node configuration (inherits Worker fields)
//	Domain:      Domain for cluster-issuer and ingress
//	Gitea:       Gitea configuration (see Gitea struct)
//	PrivateNet:  If true, workers are installed from master
//	Workers:     List of worker nodes
type Cluster struct {
	Worker
	Domain     string   `json:"domain"`
	Gitea      Gitea    `json:"gitea"`
	PrivateNet bool     `json:"privateNet"`
	Workers    []Worker `json:"workers"`
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

// Gitea holds the PostgreSQL configuration for Gitea.
//
// Fields:
//
//	Pg: PostgreSQL configuration for Gitea (see Pg struct)
type Gitea struct {
	Pg Pg `json:"pg"`
}

// Pg contains PostgreSQL credentials for Gitea.
//
// Fields:
//
//	Username: PostgreSQL username
//	Password: PostgreSQL password
//	DbName:   PostgreSQL database name
type Pg struct {
	Username string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"db"`
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
