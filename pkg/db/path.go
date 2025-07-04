package db

import (
	"os"
	"path/filepath"

	"github.com/argon-chat/k3sd/pkg/utils"
)

const defaultDBName = "k3sd.db"

// GetDBPath returns the path to the k3sd database file.
//
// If the path is not set in utils.DBPath, it defaults to ~/.k3sd/k3sd.db.
// The function ensures the directory and file exist, creating them if necessary.
//
// Returns:
//   - string: The absolute path to the database file.
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
