package db

import (
	"encoding/json"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

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

func InsertCluster(cluster *types.Cluster) error {
	var maxVersion int
	DbCtx.Model(&ClusterRecord{}).
		Where("address = ? AND node_name = ?", cluster.Address, cluster.NodeName).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion)

	b, err := json.Marshal(cluster)
	if err != nil {
		return err
	}
	rec := &ClusterRecord{
		Address:  cluster.Address,
		NodeName: cluster.NodeName,
		Version:  maxVersion + 1,
		Cluster:  string(b),
	}
	return DbCtx.Create(rec).Error
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
