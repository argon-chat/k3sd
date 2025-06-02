package cluster

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/argon-chat/k3sd/pkg/types"
)

// LoadClusters loads a list of clusters from the specified JSON file.
//
// Parameters:
//
//	path: Path to the clusters.json file.
//
// Returns:
//
//	Slice of Cluster objects and error if loading or decoding fails.
func LoadClusters(path string) ([]types.Cluster, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open cluster config: %w", err)
	}
	defer func() { _ = file.Close() }()

	var clusters []types.Cluster
	err = json.NewDecoder(file).Decode(&clusters)
	if err != nil {
		return nil, fmt.Errorf("decode cluster config: %w", err)
	}
	return clusters, nil
}

// SaveClusters saves the list of clusters to the specified JSON file.
//
// Parameters:
//
//	path: Path to the clusters.json file.
//	clusters: Slice of Cluster objects to save.
//
// Returns:
//
//	Error if marshalling or writing fails.
func SaveClusters(path string, clusters []types.Cluster) error {
	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cluster config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
