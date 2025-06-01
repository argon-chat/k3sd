package clusterutils

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// KubeConfigPath returns the path to the kubeconfig file for the given cluster and logger session.
//
// Parameters:
//
//	cluster: The cluster for which to get the kubeconfig path.
//	logger: Logger (used for session ID).
//
// Returns:
//
//	Path to the kubeconfig YAML file.
func KubeConfigPath(cluster *types.Cluster, logger *utils.Logger) string {
	return path.Join("./kubeconfigs", logger.Id, fmt.Sprintf("%s.yaml", cluster.NodeName))
}

func PipeAndLog(cmd *exec.Cmd, logger *utils.Logger) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	_ = cmd.Start()
	go StreamOutput(stdout, false, logger)
	go StreamOutput(stderr, true, logger)
	_ = cmd.Wait()
}

func PipeAndApply(cmd *exec.Cmd, kubeconfig string, logger *utils.Logger) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	_ = cmd.Start()

	var yaml strings.Builder
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		yaml.WriteString(scanner.Text() + "\n")
	}
	go StreamOutput(stderr, true, logger)
	_ = cmd.Wait()

	apply := exec.Command("kubectl", "--kubeconfig", kubeconfig, "apply", "-f", "-")
	apply.Stdin = strings.NewReader(yaml.String())
	out, err := apply.CombinedOutput()
	if err != nil {
		logger.LogErr("apply failed: %v\n%s", err, string(out))
		return
	}
	logger.Log("Apply output:\n%s", string(out))
}

func LabelNode(kubeconfigPath, nodeName string, labels string, logger *utils.Logger) error {
	if labels == "" {
		logger.Log("No labels provided for node %s", nodeName)
		return nil
	}
	labelArgs := []string{"label", "node", nodeName}
	for _, label := range strings.Split(labels, ",") {
		if label = strings.TrimSpace(label); label != "" {
			labelArgs = append(labelArgs, label)
		}
	}
	cmd := exec.Command("kubectl", append([]string{"--kubeconfig", kubeconfigPath}, labelArgs...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Log("Failed to label node %s: %v\nOutput: %s", nodeName, err, string(out))
		return err
	}
	logger.Log("Labeled node %s successfully. Output: %s", nodeName, string(out))
	return nil
}

func GetManifestData(manifestPathOrURL string) ([]byte, error) {
	if strings.HasPrefix(manifestPathOrURL, "http://") || strings.HasPrefix(manifestPathOrURL, "https://") {
		resp, err := http.Get(manifestPathOrURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(manifestPathOrURL)
}

func ApplySubstitutions(data []byte, substitutions map[string]string) []byte {
	if substitutions == nil {
		return data
	}
	content := string(data)
	for k, v := range substitutions {
		content = strings.ReplaceAll(content, k, v)
	}
	return []byte(content)
}

func SplitYAMLDocs(data []byte) []string {
	var docs []string
	rawDocs := strings.Split(string(data), "\n---")
	for _, doc := range rawDocs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		allComment := true
		for _, line := range strings.Split(doc, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				allComment = false
				break
			}
		}
		if allComment {
			continue
		}
		docs = append(docs, doc)
	}
	return docs
}
func ApplyYAMLManifest(kubeconfigPath, manifestPathOrURL string, logger *utils.Logger, substitutions map[string]string) error {
	data, err := GetManifestData(manifestPathOrURL)
	if err != nil {
		logger.LogErr("Failed to read manifest from %s: %v\n", manifestPathOrURL, err)
		return err
	}
	data = ApplySubstitutions(data, substitutions)
	tmpFile, err := os.CreateTemp("", "k3sd-manifest-*.yaml")
	if err != nil {
		logger.LogErr("Failed to create temp file for manifest: %v", err)
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	if err := writeManifestData(tmpFile, data, logger); err != nil {
		return err
	}
	return applyManifestWithKubectl(kubeconfigPath, tmpFile.Name(), logger)
}

func writeManifestData(tmpFile *os.File, data []byte, logger *utils.Logger) error {
	if _, err := tmpFile.Write(data); err != nil {
		logger.LogErr("Failed to write manifest to temp file: %v", err)
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		logger.LogErr("Failed to sync temp manifest file: %v", err)
		return err
	}
	return nil
}

func applyManifestWithKubectl(kubeconfigPath, manifestPath string, logger *utils.Logger) error {
	cmd := exec.Command("kubectl", "--kubeconfig", kubeconfigPath, "apply", "-f", manifestPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.LogErr("kubectl apply failed: %v\nOutput: %s", err, string(out))
		return err
	}
	logger.Log("kubectl apply output:\n%s", string(out))
	return nil
}
