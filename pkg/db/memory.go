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
	b, err := json.Marshal(cluster)
	if err != nil {
		return err
	}
	rec := &ClusterRecord{
		Address:  cluster.Address,
		NodeName: cluster.NodeName,
		Cluster:  string(b),
	}
	return DbCtx.Create(rec).Error
}

func DeleteClusterRecords(cluster *types.Cluster) error {
	return DbCtx.Where("address = ? AND node_name = ?", cluster.Address, cluster.NodeName).Delete(&ClusterRecord{}).Error
}

type ClusterRecord struct {
	ID       uint   `gorm:"primaryKey;index:idx_address_nodename_id" json:"id"`
	Address  string `gorm:"index:idx_address_nodename_id" json:"address"`
	NodeName string `gorm:"index:idx_address_nodename_id" json:"node_name"`
	Cluster  string `gorm:"type:json" json:"cluster"`
}
