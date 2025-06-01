package clusterutils

import (
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// WaitForDeploymentReady waits until the specified deployment exists and is ready in the given namespace.
// It checks for deployment existence, then waits for rollout status.
//
// Parameters:
//
//	kubeconfigPath: Path to the kubeconfig file.
//	deployment: Name of the deployment to check.
//	namespace: Kubernetes namespace.
//	logger: Logger for output.
func WaitForDeploymentReady(kubeconfigPath, deployment, namespace string, logger *utils.Logger) {
	for {
		cmd := exec.Command("kubectl", "--kubeconfig", kubeconfigPath, "-n", namespace, "get", "deployment", deployment)
		if out, err := cmd.CombinedOutput(); err == nil {
			logger.Log("Deployment %s exists. Output: %s", deployment, string(out))
			break
		} else {
			logger.Log("Waiting for deployment %s to exist...", deployment)
			time.Sleep(5 * time.Second)
		}
	}
	cmd := exec.Command("kubectl", "--kubeconfig", kubeconfigPath, "-n", namespace, "rollout", "status", "deployment/"+deployment, "--timeout=120s")
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.LogErr("Waiting for deployment %s failed: %v\nOutput: %s", deployment, err, string(out))
	} else {
		logger.Log("Deployment %s is ready. Output: %s", deployment, string(out))
	}
}

func ForEachWorker(workers []types.Worker, fn func(*types.Worker) error) error {
	for i := range workers {
		if err := fn(&workers[i]); err != nil {
			return err
		}
	}
	return nil
}

func EnsureNamespace(kubeconfigPath, namespace string, logger *utils.Logger) {
	if namespace != "default" && namespace != "kube-system" {
		cmd := exec.Command("kubectl", "--kubeconfig", kubeconfigPath, "create", "namespace", namespace)
		_ = cmd.Run()
		logger.Log("Ensured namespace %s exists", namespace)
	}
}

func BuildSubstitutions(pairs ...string) map[string]string {
	subs := make(map[string]string)
	for i := 0; i+1 < len(pairs); i += 2 {
		subs[pairs[i]] = pairs[i+1]
	}
	return subs
}
func ResolveYamlPath(filename string) string {
	if utils.YamlsPath != "" {
		return path.Join(utils.YamlsPath, filename)
	}
	if _, err := os.Stat("yamls/" + filename); err == nil {
		return "yamls/" + filename
	}
	home, err := os.UserHomeDir()
	if err == nil {
		candidate := path.Join(home, ".k3sd", "yamls", filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return filename
}
