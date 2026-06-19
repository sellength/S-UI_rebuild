package service

import (
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"testing"
)

func TestDistributedInboundServiceSaveAndList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-distributed-inbound-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	node := model.Node{Name: "US-01", Code: "us-01"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	service := DistributedInboundService{}
	inbound := &model.DistributedInbound{
		Enable:     true,
		NodeId:     node.Id,
		Protocol:   "anytls",
		ListenPort: 443,
	}
	if err := service.SaveDistributedInbound(inbound); err != nil {
		t.Fatalf("SaveDistributedInbound() error = %v", err)
	}

	inbounds, err := service.GetDistributedInbounds(node.Id)
	if err != nil {
		t.Fatalf("GetDistributedInbounds() error = %v", err)
	}
	if len(inbounds) != 1 || inbounds[0].Protocol != "anytls" {
		t.Fatalf("unexpected inbounds: %+v", inbounds)
	}
}

func TestDistributedInboundServiceSaveInboundUserRequiresPassword(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-inbound-user-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	service := DistributedInboundService{}
	err := service.SaveInboundUser(&model.InboundUser{InboundId: 1, Name: "alice"})
	if err == nil {
		t.Fatalf("expected missing password error")
	}
}

func TestDistributedInboundServiceRejectsLocalOnlyClient(t *testing.T) {
	t.Skip("Skipping because local client feature has been removed, all clients are treated as cluster clients.")

	dbPath := filepath.Join(t.TempDir(), "s-ui-inbound-user-local-only-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	client := model.Client{Name: "local-only", Enable: true, AccessScope: ClientAccessLocal}
	if err := database.GetDB().Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}

	service := DistributedInboundService{}
	err := service.SaveInboundUser(&model.InboundUser{
		InboundId: 1,
		ClientId:  client.Id,
		Name:      "local-only",
		Password:  "secret",
	})
	if err == nil {
		t.Fatal("expected local-only client to be rejected")
	}
}

func TestDistributedInboundServiceRejectsDuplicateEnabledListenAddressAndPort(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-distributed-inbound-duplicate-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	node := model.Node{Name: "US-01", Code: "us-01"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	service := DistributedInboundService{}
	first := &model.DistributedInbound{
		Enable:     true,
		NodeId:     node.Id,
		Protocol:   "anytls",
		Listen:     "::",
		ListenPort: 443,
	}
	if err := service.SaveDistributedInbound(first); err != nil {
		t.Fatalf("SaveDistributedInbound(first) error = %v", err)
	}

	duplicate := &model.DistributedInbound{
		Enable:     true,
		NodeId:     node.Id,
		Protocol:   "hysteria2",
		Listen:     "",
		ListenPort: 443,
	}
	if err := service.SaveDistributedInbound(duplicate); err == nil {
		t.Fatalf("expected duplicate inbound error")
	}
}

func TestDistributedInboundServiceAllowsDuplicateDisabledListenAddressAndPort(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-distributed-inbound-disabled-duplicate-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	node := model.Node{Name: "US-01", Code: "us-01"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	service := DistributedInboundService{}
	disabled := &model.DistributedInbound{
		Enable:     false,
		NodeId:     node.Id,
		Protocol:   "anytls",
		Listen:     "::",
		ListenPort: 443,
	}
	if err := service.SaveDistributedInbound(disabled); err != nil {
		t.Fatalf("SaveDistributedInbound(disabled) error = %v", err)
	}

	enabled := &model.DistributedInbound{
		Enable:     true,
		NodeId:     node.Id,
		Protocol:   "hysteria2",
		Listen:     "::",
		ListenPort: 443,
	}
	if err := service.SaveDistributedInbound(enabled); err != nil {
		t.Fatalf("SaveDistributedInbound(enabled) error = %v", err)
	}
}
