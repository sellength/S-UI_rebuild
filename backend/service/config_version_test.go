package service

import (
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"testing"
)

func initConfigVersionTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "s-ui-config-version-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
}

func TestConfigVersionServiceDefaultsStatus(t *testing.T) {
	initConfigVersionTestDB(t)

	service := ConfigVersionService{}
	version := &model.ConfigVersion{
		Version: 1,
		Scope:   "node",
		NodeId:  1,
	}

	if err := service.SaveConfigVersion(version); err != nil {
		t.Fatalf("SaveConfigVersion() error = %v", err)
	}
	if version.Status != "draft" {
		t.Fatalf("expected draft status, got %q", version.Status)
	}
}

func TestConfigVersionServiceRequiresDeploymentNode(t *testing.T) {
	initConfigVersionTestDB(t)

	service := ConfigVersionService{}
	err := service.SaveConfigDeployment(&model.ConfigDeployment{ConfigVersionId: 1})
	if err == nil {
		t.Fatalf("expected missing node id error")
	}
}
