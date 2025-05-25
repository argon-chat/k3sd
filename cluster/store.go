// Package cluster provides functions for loading and saving cluster configuration from JSON files.
package cluster

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadClusters loads cluster configuration from a JSON file.
//
// Parameters:
//
//	path: Path to the JSON file.
//
// Returns:
//
//	[]Cluster: List of clusters.
//	error: Error if loading or parsing fails.
func LoadClusters(path string) ([]Cluster, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open cluster config: %w", err)
	}
	defer file.Close()

	var clusters []Cluster
	err = json.NewDecoder(file).Decode(&clusters)
	if err != nil {
		return nil, fmt.Errorf("decode cluster config: %w", err)
	}
	return clusters, nil
}

// SaveClusters saves cluster configuration to a JSON file.
//
// Parameters:
//
//	path: Path to the JSON file.
//	clusters: List of clusters to save.
//
// Returns:
//
//	error: Error if writing fails.
func SaveClusters(path string, clusters []Cluster) error {
	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cluster config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
