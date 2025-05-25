// Package cluster provides utilities for applying Kubernetes manifests to clusters, including YAML parsing and substitutions.
package cluster

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/argon-chat/k3sd/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlserializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// getManifestData fetches manifest data from a file or URL.
//
// Parameters:
//
//	manifestPathOrURL: Path or URL to manifest.
//
// Returns:
//
//	[]byte: Manifest data.
//	error: Error if fetching fails.
func getManifestData(manifestPathOrURL string) ([]byte, error) {
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

// applySubstitutions applies string substitutions to manifest data.
//
// Parameters:
//
//	data: Manifest data.
//	substitutions: Substitutions to apply.
//
// Returns:
//
//	[]byte: Modified data.
func applySubstitutions(data []byte, substitutions map[string]string) []byte {
	if substitutions == nil {
		return data
	}
	content := string(data)
	for k, v := range substitutions {
		content = strings.ReplaceAll(content, k, v)
	}
	return []byte(content)
}

// splitYAMLDocs splits YAML data into individual documents.
//
// Parameters:
//
//	data: YAML data.
//
// Returns:
//
//	[]string: List of YAML documents.
func splitYAMLDocs(data []byte) []string {
	var docs []string
	rawDocs := strings.Split(string(data), "\n---")
	for _, doc := range rawDocs {
		doc = strings.TrimSpace(doc)
		if doc == "" || strings.HasPrefix(doc, "#") {
			continue
		}
		docs = append(docs, doc)
	}
	return docs
}

// applyYAMLManifest applies a YAML manifest to a Kubernetes cluster.
//
// Parameters:
//
//	kubeconfigPath: Path to kubeconfig.
//	manifestPathOrURL: Path or URL to manifest.
//	logger: Logger for output.
//	substitutions: Substitutions to apply.
//
// Returns:
//
//	error: Error if applying fails.
func applyYAMLManifest(kubeconfigPath, manifestPathOrURL string, logger *utils.Logger, substitutions map[string]string) error {
	data, err := getManifestData(manifestPathOrURL)
	if err != nil {
		return err
	}
	data = applySubstitutions(data, substitutions)
	docs := splitYAMLDocs(data)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	decUnstructured := yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	disco, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco))
	for _, doc := range docs {
		obj := &unstructured.Unstructured{}
		_, _, err := decUnstructured.Decode([]byte(doc), nil, obj)
		if err != nil {
			logger.Log("YAML decode error: %v\n---\n%s", err, doc)
			continue
		}
		m := obj.GroupVersionKind()
		mapping, err := mapper.RESTMapping(m.GroupKind(), m.Version)
		if err != nil {
			logger.Log("RESTMapping error: %v", err)
			continue
		}
		ns := obj.GetNamespace()
		if ns == "" {
			ns = "default"
		}
		resource := dyn.Resource(mapping.Resource).Namespace(ns)
		_, err = resource.Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Log("Apply error: %v", err)
		}
	}
	return nil
}
