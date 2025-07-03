package addons

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/argon-chat/k3sd/pkg/clusterutils"
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

var LinkChannel = []*types.Cluster{}

func runStepCertCreate(args []string, logger *utils.Logger) {
	cmd := exec.Command("step", args...)
	outPipe, _ := cmd.StdoutPipe()
	errPipe, _ := cmd.StderrPipe()
	_ = cmd.Start()
	go clusterutils.StreamOutput(outPipe, false, logger)
	go clusterutils.StreamOutput(errPipe, true, logger)
	_ = cmd.Wait()
	logger.Log("Command executed successfully")
}

// ApplyLinkerdAddon installs and configures Linkerd or Linkerd multicluster on the cluster if enabled.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
func ApplyLinkerdAddon(cluster *types.Cluster, logger *utils.Logger) {
	linkerd, ok := cluster.Addons["linkerd"]
	linkerdMC, okMC := cluster.Addons["linkerd-mc"]
	multicluster := okMC && linkerdMC.Enabled
	standard := ok && linkerd.Enabled
	if !multicluster && !standard {
		return
	}
	runLinkerdInstall(cluster, logger, multicluster)
}

func LinkClusters(cluster *types.Cluster, otherClusters *[]types.Cluster, logger *utils.Logger) {
	_, kubeconfig := getLinkerdPaths(logger.Id, cluster.NodeName)
	if len(cluster.LinksTo) == 0 {
		logger.Log("No links to other clusters defined for %s", cluster.NodeName)
	}
	for _, link := range cluster.LinksTo {
		var otherCluster *types.Cluster
		for _, c := range *otherClusters {
			if c.Address == link {
				otherCluster = &c
				break
			}
		}
		_, otherKubeConfig := getLinkerdPaths(logger.Id, otherCluster.NodeName)
		args := []string{
			"link",
			"--kubeconfig", otherKubeConfig,
			"--set", "enableHeadlessServices=true",
			"--log-level=debug",
			fmt.Sprintf("--cluster-name=%s", otherCluster.Context),
			fmt.Sprintf("--api-server-address=https://%s:6443", link),
		}
		logger.Log("Linking to cluster %s with command: linkerd multicluster %v", link, args)
		runLinkerdCmd("multicluster", args, logger, kubeconfig, true)
		logger.Log("Successfully linked to cluster %s", link)
	}
}

func runLinkerdInstall(cluster *types.Cluster, logger *utils.Logger, multicluster bool) {
	dir, kubeconfig := getLinkerdPaths(logger.Id, cluster.NodeName)
	// clusterutils.ApplyComponentYAML("gateway CRDs", kubeconfig, "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml", logger, nil)
	runLinkerdCmd("check", []string{"--pre", "--kubeconfig", kubeconfig}, logger, kubeconfig, false)
	setupLinkerdCertsAndCRDs(dir, kubeconfig, cluster, logger)
	runLinkerdInstallCmd(dir, kubeconfig, cluster, logger, multicluster)
	updateCRDs(kubeconfig, logger)
}

func getLinkerdPaths(loggerId, nodeName string) (string, string) {
	dir := path.Join("./kubeconfigs", loggerId)
	kubeconfig := path.Join(dir, fmt.Sprintf("%s.yaml", nodeName))
	return dir, kubeconfig
}

func setupLinkerdCertsAndCRDs(dir, kubeconfig string, cluster *types.Cluster, logger *utils.Logger) {
	createRootCerts(dir, logger)
	installCRDs(kubeconfig, logger)
	createIssuerCerts(dir, cluster, logger)
}

func runLinkerdInstallCmd(dir, kubeconfig string, cluster *types.Cluster, logger *utils.Logger, multicluster bool) {
	args := []string{
		"--proxy-log-level=linkerd=debug,warn",
		"--cluster-domain=cluster.local",
		"--identity-trust-domain=cluster.local",
		"--identity-trust-anchors-file=" + path.Join(dir, "ca.crt"),
		"--identity-issuer-certificate-file=" + path.Join(dir, fmt.Sprintf("%s-issuer.crt", cluster.NodeName)),
		"--identity-issuer-key-file=" + path.Join(dir, fmt.Sprintf("%s-issuer.key", cluster.NodeName)),
		"--kubeconfig", kubeconfig,
	}
	runLinkerdCmd("install", args, logger, kubeconfig, true)
	logger.Log("Linkerd installed successfully.")
	runLinkerdCmd("check", []string{"--kubeconfig", kubeconfig}, logger, kubeconfig, false)

	if multicluster {
		runLinkerdCmd("multicluster", []string{"install", "--kubeconfig", kubeconfig}, logger, kubeconfig, true)
		logger.Log("Linkerd multicluster installed.")
		runLinkerdCmd("multicluster", []string{"check", "--kubeconfig", kubeconfig}, logger, kubeconfig, false)
	}
}

func runLinkerdCmd(cmd string, args []string, logger *utils.Logger, kubeconfig string, apply bool) {
	parts := append([]string{cmd}, args...)
	c := exec.Command("linkerd", parts...)
	if apply {
		clusterutils.PipeAndApply(c, kubeconfig, logger)
	} else {
		clusterutils.PipeAndLog(c, logger)
	}
}

func updateCRDs(kubeconfig string, logger *utils.Logger) {
	runLinkerdCmd("upgrade", []string{"--crds", "--kubeconfig", kubeconfig}, logger, kubeconfig, true)
}

func installCRDs(kubeconfig string, logger *utils.Logger) {
	run := exec.Command("linkerd", "install", "--crds", "--kubeconfig", kubeconfig)
	clusterutils.PipeAndApply(run, kubeconfig, logger)
}

func createRootCerts(dir string, logger *utils.Logger) {
	caCrt := path.Join(dir, "ca.crt")
	caKey := path.Join(dir, "ca.key")
	if _, errCrt := os.Stat(caCrt); errCrt == nil {
		if _, errKey := os.Stat(caKey); errKey == nil {
			logger.Log("Root CA cert and key already exist, skipping creation.")
			return
		}
	}
	args := []string{
		"certificate", "create",
		"identity.linkerd.cluster.local",
		caCrt,
		caKey,
		"--profile", "root-ca",
		"--no-password", "--insecure", "--force", "--not-after", "438000h",
	}
	runStepCertCreate(args, logger)
}

func createIssuerCerts(dir string, cluster *types.Cluster, logger *utils.Logger) {
	args := []string{
		"certificate", "create",
		"identity.linkerd.cluster.local",
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

// DeleteLinkerdAddon uninstalls Linkerd and Linkerd multicluster from the cluster using the linkerd CLI.
//
// Parameters:
//
//	cluster: The cluster to uninstall the addon from.
//	logger: Logger for output.
func DeleteLinkerdAddon(cluster *types.Cluster, logger *utils.Logger) {
	_, kubeconfig := getLinkerdPaths(logger.Id, cluster.NodeName)
	if mc, ok := cluster.Addons["linkerd-mc"]; ok && mc.Enabled {
		cmd := exec.Command("linkerd", "multicluster", "uninstall", "--kubeconfig", kubeconfig)
		clusterutils.PipeAndApply(cmd, kubeconfig, logger)
	}
	cmd := exec.Command("linkerd", "uninstall", "--kubeconfig", kubeconfig)
	clusterutils.PipeAndApply(cmd, kubeconfig, logger)
}
