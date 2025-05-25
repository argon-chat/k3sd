// Package cluster provides functions for installing and configuring Linkerd and its certificates on Kubernetes clusters.
package cluster

import (
	"fmt"
	"os/exec"
	"path"

	"github.com/argon-chat/k3sd/utils"
)

// runStepCertCreate runs the `step` CLI to create certificates for Linkerd.
//
// Parameters:
//
//	args: Arguments for the `step` command.
//	logger: Logger for output.
func runStepCertCreate(args []string, logger *utils.Logger) {
	cmd := exec.Command("step", args...)
	pipeAndLog(cmd, logger)
}

// runLinkerdInstall installs Linkerd (and optionally multicluster) on a cluster.
//
// Parameters:
//
//	cluster: Cluster information.
//	logger: Logger for output.
//	multicluster: Whether to install multicluster.
func runLinkerdInstall(cluster Cluster, logger *utils.Logger, multicluster bool) {
	dir := path.Join("./kubeconfigs", logger.Id)
	kubeconfig := path.Join(dir, fmt.Sprintf("%s.yaml", cluster.NodeName))

	createRootCerts(dir, logger)
	installCRDs(kubeconfig, logger)
	createIssuerCerts(dir, cluster, logger)
	runLinkerdCmd("install", []string{
		"--proxy-log-level=linkerd=debug,warn",
		"--cluster-domain=cluster.local",
		"--identity-trust-domain=cluster.local",
		"--identity-trust-anchors-file=" + path.Join(dir, "ca.crt"),
		"--identity-issuer-certificate-file=" + path.Join(dir, fmt.Sprintf("%s-issuer.crt", cluster.NodeName)),
		"--identity-issuer-key-file=" + path.Join(dir, fmt.Sprintf("%s-issuer.key", cluster.NodeName)),
		"--kubeconfig", kubeconfig,
	}, logger, kubeconfig, true)

	if multicluster {
		runLinkerdCmd("multicluster", []string{"install", "--kubeconfig", kubeconfig}, logger, kubeconfig, true)
		logger.Log("Linkerd multicluster installed.")
		runLinkerdCmd("multicluster", []string{"check", "--kubeconfig", kubeconfig}, logger, kubeconfig, false)
	} else {
		runLinkerdCmd("check", []string{"--pre", "--kubeconfig", kubeconfig}, logger, kubeconfig, true)
		runLinkerdCmd("check", []string{"--kubeconfig", kubeconfig}, logger, kubeconfig, false)
	}
}

// runLinkerdCmd runs a Linkerd CLI command, optionally applying the output to the cluster.
//
// Parameters:
//
//	cmd: Linkerd subcommand.
//	args: Arguments for the command.
//	logger: Logger for output.
//	kubeconfig: Path to kubeconfig.
//	apply: Whether to apply the output.
func runLinkerdCmd(cmd string, args []string, logger *utils.Logger, kubeconfig string, apply bool) {
	parts := append([]string{cmd}, args...)
	c := exec.Command("linkerd", parts...)
	if apply {
		pipeAndApply(c, kubeconfig, logger)
	} else {
		pipeAndLog(c, logger)
	}
}

// installCRDs installs Linkerd CRDs on the cluster.
//
// Parameters:
//
//	kubeconfig: Path to kubeconfig.
//	logger: Logger for output.
func installCRDs(kubeconfig string, logger *utils.Logger) {
	run := exec.Command("linkerd", "install", "--crds", "--kubeconfig", kubeconfig)
	pipeAndApply(run, kubeconfig, logger)
}

// createRootCerts creates root CA certificates for Linkerd.
//
// Parameters:
//
//	dir: Directory to store certificates.
//	logger: Logger for output.
func createRootCerts(dir string, logger *utils.Logger) {
	args := []string{
		"certificate", "create",
		"identity.linkerd.cluster.local",
		path.Join(dir, "ca.crt"),
		path.Join(dir, "ca.key"),
		"--profile", "root-ca",
		"--no-password", "--insecure", "--force", "--not-after", "438000h",
	}
	runStepCertCreate(args, logger)
}

// createIssuerCerts creates issuer certificates for Linkerd for a specific cluster.
//
// Parameters:
//
//	dir: Directory to store certificates.
//	cluster: Cluster information.
//	logger: Logger for output.
func createIssuerCerts(dir string, cluster Cluster, logger *utils.Logger) {
	args := []string{
		"certificate", "create",
		fmt.Sprintf("identity.linkerd.%s", cluster.Domain),
		path.Join(dir, fmt.Sprintf("%s-issuer.crt", cluster.NodeName)),
		path.Join(dir, fmt.Sprintf("%s-issuer.key", cluster.NodeName)),
		"--ca", path.Join(dir, "ca.crt"),
		"--ca-key", path.Join(dir, "ca.key"),
		"--profile", "intermediate-ca",
		"--not-after", "438000h",
		"--no-password", "--insecure", "--force",
	}
	runStepCertCreate(args, logger)
}
