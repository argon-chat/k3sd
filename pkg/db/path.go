package db

import (
	"os"
	"path/filepath"

	"github.com/argon-chat/k3sd/pkg/utils"
)

const defaultDBName = "k3sd.db"

func GetDBPath() string {
	dbPath := utils.DBPath
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		dbPath = filepath.Join(home, ".k3sd", defaultDBName)
	}
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0700)
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		f, err := os.Create(dbPath)
		if err == nil {
			f.Close()
		}
	}
	return dbPath
}
