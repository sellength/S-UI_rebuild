package database

import (
	"path/filepath"
	"s-ui/database/model"
	"testing"
)

func TestInitDBMigratesDistributedTables(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-test.db")

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := GetDB()
	tables := []interface{}{
		&model.Node{},
		&model.NodeAgent{},
		&model.Certificate{},
		&model.CertificateVersion{},
		&model.ProtocolTemplate{},
		&model.DistributedInbound{},
		&model.ConfigVersion{},
		&model.ConfigDeployment{},
		&model.Subscription{},
	}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("expected table for %T to be migrated", table)
		}
	}
}
