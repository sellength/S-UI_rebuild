package service

import (
	"encoding/json"
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"testing"
)

func TestRenderAnyTLSInbound(t *testing.T) {
	raw, err := RenderAnyTLSInbound(AnyTLSInboundInput{
		Tag:        "anytls-us-01-443",
		ListenPort: 443,
		Users: []AnyTLSUser{
			{Name: "alice", Password: "secret"},
		},
		ServerName: "us.example.com",
		CertPath:   "/usr/local/s-ui-agent/certs/fullchain.pem",
		KeyPath:    "/usr/local/s-ui-agent/certs/privkey.pem",
	})
	if err != nil {
		t.Fatalf("RenderAnyTLSInbound() error = %v", err)
	}

	config := map[string]interface{}{}
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatalf("unmarshal inbound: %v", err)
	}
	if config["type"] != "anytls" {
		t.Fatalf("expected anytls type, got %+v", config["type"])
	}
	if config["listen"] != "::" {
		t.Fatalf("expected default listen ::, got %+v", config["listen"])
	}
}

func TestRenderAnyTLSOutboundDefaultsServerName(t *testing.T) {
	raw, err := RenderAnyTLSOutbound(AnyTLSOutboundInput{
		Tag:        "US-01",
		Server:     "us.example.com",
		ServerPort: 443,
		Password:   "secret",
	})
	if err != nil {
		t.Fatalf("RenderAnyTLSOutbound() error = %v", err)
	}

	config := struct {
		TLS struct {
			ServerName string `json:"server_name"`
		} `json:"tls"`
	}{}
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatalf("unmarshal outbound: %v", err)
	}
	if config.TLS.ServerName != "us.example.com" {
		t.Fatalf("expected server_name us.example.com, got %q", config.TLS.ServerName)
	}
}

func TestAnyTLSCertificatePaths(t *testing.T) {
	certPath, keyPath := AnyTLSCertificatePaths("/usr/local/s-ui-agent/certs", "Wildcard Example.com", 1, "AA:BB:CC:DD:EE:FF")
	expectedCertPath := "/usr/local/s-ui-agent/certs/wildcard-example.com-aabbccddeeff/fullchain.pem"
	expectedKeyPath := "/usr/local/s-ui-agent/certs/wildcard-example.com-aabbccddeeff/privkey.pem"

	if certPath != expectedCertPath {
		t.Fatalf("expected cert path %q, got %q", expectedCertPath, certPath)
	}
	if keyPath != expectedKeyPath {
		t.Fatalf("expected key path %q, got %q", expectedKeyPath, keyPath)
	}
}

func TestRenderDistributedAnyTLSInbound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-anytls-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	node := model.Node{Name: "US-01", Code: "us-01", PublicHost: "us.example.com"}
	if err := database.GetDB().Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	certificate := model.Certificate{Name: "Wildcard Example.com", Source: "manual", Enable: true}
	if err := database.GetDB().Create(&certificate).Error; err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	certificateVersion := model.CertificateVersion{CertificateId: certificate.Id, Fingerprint: "AA:BB:CC:DD:EE:FF"}
	if err := database.GetDB().Create(&certificateVersion).Error; err != nil {
		t.Fatalf("create certificate version: %v", err)
	}
	certificate.ActiveVersionId = certificateVersion.Id
	if err := database.GetDB().Save(&certificate).Error; err != nil {
		t.Fatalf("save certificate: %v", err)
	}

	inbound := model.DistributedInbound{
		Enable:        true,
		NodeId:        node.Id,
		Protocol:      "anytls",
		ListenPort:    443,
		CertificateId: certificate.Id,
	}
	if err := database.GetDB().Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	user := model.InboundUser{InboundId: inbound.Id, Name: "alice", Password: "secret"}
	if err := database.GetDB().Create(&user).Error; err != nil {
		t.Fatalf("create inbound user: %v", err)
	}

	raw, err := RenderDistributedAnyTLSInbound(inbound.Id)
	if err != nil {
		t.Fatalf("RenderDistributedAnyTLSInbound() error = %v", err)
	}

	config := struct {
		Tag string `json:"tag"`
		TLS struct {
			ServerName      string `json:"server_name"`
			CertificatePath string `json:"certificate_path"`
		} `json:"tls"`
	}{}
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatalf("unmarshal rendered inbound: %v", err)
	}
	if config.Tag != "anytls-us-01-443" {
		t.Fatalf("expected generated tag, got %q", config.Tag)
	}
	if config.TLS.ServerName != "us.example.com" {
		t.Fatalf("expected node public host, got %q", config.TLS.ServerName)
	}
	if config.TLS.CertificatePath == "" {
		t.Fatalf("expected certificate path")
	}
}
