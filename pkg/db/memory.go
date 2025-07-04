package db

import (
	"encoding/json"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	"github.com/argon-chat/k3sd/pkg/types"
)

// DbCtx is the global GORM database context for k3sd.
var DbCtx *gorm.DB

// OpenGormDB opens a GORM database connection to the specified path.
//
// If the path is empty, it uses the default path from GetDBPath().
// The function also auto-migrates the ClusterRecord schema.
//
// Parameters:
//   - path: Path to the SQLite database file.
//
// Returns:
//   - *gorm.DB: The opened GORM database instance.
//   - error: Error if opening or migrating fails.
func OpenGormDB(path string) (*gorm.DB, error) {
	if path == "" {
		path = GetDBPath()
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&ClusterRecord{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// InsertCluster inserts a new cluster record into the database, incrementing the version.
//
// Parameters:
//   - cluster: Pointer to the Cluster object to insert.
//
// Returns:
//   - int: The previous maximum version for the cluster.
//   - error: Error if marshalling or database insertion fails.
func InsertCluster(cluster *types.Cluster) (int, error) {
	var maxVersion int
	DbCtx.Model(&ClusterRecord{}).
		Where("address = ? AND node_name = ?", cluster.Address, cluster.NodeName).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion)

	b, err := json.Marshal(cluster)
	if err != nil {
		return 0, err
	}
	rec := &ClusterRecord{
		Address:  cluster.Address,
		NodeName: cluster.NodeName,
		Version:  maxVersion + 1,
		Cluster:  string(b),
	}
	return maxVersion, DbCtx.Create(rec).Error
}

// GetClusterVersion retrieves a specific version of a cluster from the database.
//
// Parameters:
//   - cluster: Pointer to the Cluster object (address and node name used for lookup).
//   - version: The version number to retrieve.
//
// Returns:
//   - *types.Cluster: The cluster object for the specified version, or nil if not found.
//   - error: Error if retrieval or unmarshalling fails.
func GetClusterVersion(cluster *types.Cluster, version int) (*types.Cluster, error) {
	if version < 1 {
		return nil, nil
	}
	var record ClusterRecord
	err := DbCtx.Where("address = ? AND node_name = ? AND version = ?", cluster.Address, cluster.NodeName, version).
		First(&record).Error
	if err != nil {
		return nil, err
	}

	var result types.Cluster
	err = json.Unmarshal([]byte(record.Cluster), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteClusterRecords(cluster *types.Cluster) error {
	return DbCtx.Where("address = ? AND node_name = ?", cluster.Address, cluster.NodeName).Delete(&ClusterRecord{}).Error
}

// ClusterRecord represents a versioned cluster record stored in the database.
//
// Fields:
//   - ID: Primary key for the record.
//   - Address: Cluster address (indexed).
//   - NodeName: Node name (indexed).
//   - Version: Version number (indexed).
//   - Cluster: JSON-encoded cluster data.
type ClusterRecord struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Address  string `gorm:"index:idx_address" json:"address"`
	NodeName string `gorm:"index:idx_nodename" json:"node_name"`
	Version  int    `gorm:"index:idx_version" json:"version"`
	Cluster  string `gorm:"type:json" json:"cluster"`
}

// TODO: create a function to retrieve the calculated latest cluster record for a given address and node name
func deepMerge(dst, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		if vMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := dst[k].(map[string]interface{}); ok {
				dst[k] = deepMerge(dstMap, vMap)
			} else {
				dst[k] = deepMerge(make(map[string]interface{}), vMap)
			}
		} else {
			dst[k] = v
		}
	}
	return dst
}
