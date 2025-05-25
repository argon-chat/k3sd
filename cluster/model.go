// Package cluster contains types and logic for representing and managing Kubernetes clusters and their nodes.
package cluster

import "fmt"

type Cluster struct {
	// Worker embeds the master node's information.
	Worker
	// Domain is the domain name for the cluster.
	Domain string `json:"domain"`
	// Gitea contains Gitea configuration for the cluster.
	Gitea Gitea `json:"gitea"`
	// PrivateNet indicates if the cluster uses a private network.
	PrivateNet bool `json:"privateNet"`
	// Workers is the list of worker nodes in the cluster.
	Workers []Worker `json:"workers"`
}

// Worker represents a node (master or worker) in the cluster.
type Worker struct {
	// Address is the IP address or hostname of the node.
	Address string `json:"address"`
	// User is the SSH username.
	User string `json:"user"`
	// Password is the SSH password.
	Password string `json:"password"`
	// NodeName is the Kubernetes node name.
	NodeName string `json:"nodeName"`
	// Labels are the node labels.
	Labels map[string]string `json:"labels"`
	// Done indicates whether setup is complete.
	Done bool `json:"done"`
}

// Gitea holds Gitea configuration for the cluster.
type Gitea struct {
	// Pg contains Postgres configuration for Gitea.
	Pg Pg `json:"pg"`
}

// Pg holds Postgres configuration for Gitea.
type Pg struct {
	// Username is the database username.
	Username string `json:"user"`
	// Password is the database password.
	Password string `json:"password"`
	// DbName is the database name.
	DbName string `json:"db"`
}

// GetLabels returns a comma-separated string of the node's labels in the form key=value.
// Example: "role=worker,zone=us-east"
//
// Returns:
//   string: Comma-separated key=value pairs for all labels on the node.
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
