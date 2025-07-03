package db

import (
	"encoding/json"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	"github.com/argon-chat/k3sd/pkg/types"
)

var DbCtx *gorm.DB

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
