package service

import (
	"encoding/json"
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"testing"
)

func TestPublishNodeConfigVersion(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-config-bundle-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	node := model.Node{Name: "US-01", Code: "us-01", PublicHost: "us.example.com"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	certificate := model.Certificate{Name: "Wildcard Example.com", Source: "manual", Enable: true}
	if err := db.Create(&certificate).Error; err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	certificateVersion := model.CertificateVersion{
		CertificateId: certificate.Id,
		Fingerprint:   "AA:BB:CC:DD:EE:FF",
	}
	if err := db.Create(&certificateVersion).Error; err != nil {
		t.Fatalf("create certificate version: %v", err)
	}
	certificate.ActiveVersionId = certificateVersion.Id
	if err := db.Save(&certificate).Error; err != nil {
		t.Fatalf("save certificate: %v", err)
	}

	inbound := model.DistributedInbound{
		Enable:        true,
		NodeId:        node.Id,
		Protocol:      "anytls",
		ListenPort:    443,
		CertificateId: certificate.Id,
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	user := model.InboundUser{InboundId: inbound.Id, Name: "alice", Password: "secret"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create inbound user: %v", err)
	}
	templateService := ConfigTemplateService{}
	if err := templateService.SaveNodeConfigTemplates(&NodeConfigTemplates{
		Log:          defaultNodeLogRaw(),
		DNS:          defaultNodeDNSRaw(),
		Outbounds:    defaultNodeOutboundsRaw(),
		Route:        defaultNodeRouteRaw(),
		Experimental: json.RawMessage(`{"cache_file":{"enabled":true}}`),
	}); err != nil {
		t.Fatalf("save node config templates: %v", err)
	}

	version, err := PublishNodeConfigVersion(node.Id, "tester")
	if err != nil {
		t.Fatalf("PublishNodeConfigVersion() error = %v", err)
	}
	if version.Version != 1 || version.Status != "published" {
		t.Fatalf("unexpected version: %+v", version)
	}
	sameVersion, err := PublishNodeConfigVersion(node.Id, "tester")
	if err != nil {
		t.Fatalf("second PublishNodeConfigVersion() error = %v", err)
	}
	if sameVersion.Id != version.Id || sameVersion.Version != version.Version {
		t.Fatalf("expected unchanged config to reuse version %+v, got %+v", version, sameVersion)
	}

	config := SingboxNodeConfig{}
	if err := json.Unmarshal(version.ContentJson, &config); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if len(config.Inbounds) != 1 {
		t.Fatalf("expected one inbound, got %d", len(config.Inbounds))
	}
	if len(config.Outbounds) != 1 {
		t.Fatalf("expected one outbound, got %d", len(config.Outbounds))
	}
	if config.DNS["final"] != "google-dns-v6" {
		t.Fatalf("expected ipv6 dns final, got %+v", config.DNS)
	}
	resolver, ok := config.Outbounds[0]["domain_resolver"].(map[string]interface{})
	if !ok || resolver["strategy"] != "prefer_ipv6" {
		t.Fatalf("expected direct outbound prefer_ipv6 resolver, got %+v", config.Outbounds[0])
	}
	defaultResolver, ok := config.Route["default_domain_resolver"].(map[string]interface{})
	if !ok || defaultResolver["strategy"] != "prefer_ipv6" {
		t.Fatalf("expected route prefer_ipv6 resolver, got %+v", config.Route)
	}
	cacheFile, ok := config.Experimental["cache_file"].(map[string]interface{})
	if !ok || cacheFile["enabled"] != true {
		t.Fatalf("expected experimental cache_file enabled, got %+v", config.Experimental)
	}
}

func TestPublishNodeConfigVersionAppliesInboundPolicyOverrides(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-config-bundle-policy-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	node := model.Node{Name: "US-01", Code: "us-01", PublicHost: "us.example.com"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	certificate := model.Certificate{Name: "Wildcard Example.com", Source: "manual", Enable: true}
	if err := db.Create(&certificate).Error; err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	certificateVersion := model.CertificateVersion{CertificateId: certificate.Id, Fingerprint: "11:22:33"}
	if err := db.Create(&certificateVersion).Error; err != nil {
		t.Fatalf("create certificate version: %v", err)
	}
	certificate.ActiveVersionId = certificateVersion.Id
	if err := db.Save(&certificate).Error; err != nil {
		t.Fatalf("save certificate: %v", err)
	}
	inbound := model.DistributedInbound{
		Enable:              true,
		NodeId:              node.Id,
		Protocol:            "anytls",
		ListenPort:          443,
		CertificateId:       certificate.Id,
		PolicyOverridesJson: json.RawMessage(`{"dns":{"servers":[{"type":"tls","server":"8.8.4.4","tag":"google-dns-v4"}],"final":"google-dns-v4"},"outbounds":[{"type":"direct","tag":"direct-v4"}],"route":{"final":"direct-v4"},"experimental":{"cache_file":{"enabled":false}}}`),
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	if err := db.Create(&model.InboundUser{InboundId: inbound.Id, Name: "alice", Password: "secret"}).Error; err != nil {
		t.Fatalf("create inbound user: %v", err)
	}

	version, err := PublishNodeConfigVersion(node.Id, "tester")
	if err != nil {
		t.Fatalf("PublishNodeConfigVersion() error = %v", err)
	}
	config := SingboxNodeConfig{}
	if err := json.Unmarshal(version.ContentJson, &config); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if config.DNS["final"] != "google-dns-v4" {
		t.Fatalf("expected overridden dns final, got %+v", config.DNS)
	}
	if len(config.Outbounds) != 1 || config.Outbounds[0]["tag"] != "direct-v4" {
		t.Fatalf("expected overridden outbounds, got %+v", config.Outbounds)
	}
	if config.Route["final"] != "direct-v4" {
		t.Fatalf("expected overridden route final, got %+v", config.Route)
	}
	cacheFile, ok := config.Experimental["cache_file"].(map[string]interface{})
	if !ok || cacheFile["enabled"] != false {
		t.Fatalf("expected overridden experimental cache_file disabled, got %+v", config.Experimental)
	}
}

func TestSaveDistributedInboundAllowsOnePolicyOverridePerNode(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-inbound-policy-guard-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	node := model.Node{Name: "US-01", Code: "us-01"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	service := DistributedInboundService{}
	if err := service.SaveDistributedInbound(&model.DistributedInbound{
		Enable:              true,
		NodeId:              node.Id,
		Protocol:            "anytls",
		ListenPort:          443,
		PolicyOverridesJson: json.RawMessage(`{"route":{"final":"direct"}}`),
	}); err != nil {
		t.Fatalf("first policy override should save: %v", err)
	}

	err := service.SaveDistributedInbound(&model.DistributedInbound{
		Enable:              true,
		NodeId:              node.Id,
		Protocol:            "hysteria2",
		ListenPort:          8443,
		PolicyOverridesJson: json.RawMessage(`{"route":{"final":"direct-hy2"}}`),
	})
	if err == nil {
		t.Fatalf("expected second policy override on same node to fail")
	}
}
