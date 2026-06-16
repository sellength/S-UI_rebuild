package service

import (
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"testing"
)

func initNodeTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "s-ui-node-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
}

func TestNodeServiceSaveRequiresCode(t *testing.T) {
	initNodeTestDB(t)

	service := NodeService{}
	err := service.Save(&model.Node{Name: "US-01"})
	if err == nil {
		t.Fatalf("expected missing node code error")
	}
}

func TestNodeServiceSaveAndList(t *testing.T) {
	initNodeTestDB(t)

	service := NodeService{}
	node := &model.Node{
		Name:       "US-01",
		Code:       "us-01",
		PublicHost: "us.example.com",
	}

	if err := service.Save(node); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	nodes, err := service.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(nodes) != 1 || nodes[0].Code != "us-01" {
		t.Fatalf("unexpected nodes: %+v", nodes)
	}
}
