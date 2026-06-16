package service

import (
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"testing"
)

func initAgentTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "s-ui-agent-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
}

func TestAgentServiceRegisterRequiresExistingNode(t *testing.T) {
	initAgentTestDB(t)

	service := AgentService{}
	_, err := service.Register(&AgentRegisterRequest{
		NodeCode:   "missing",
		AgentId:    "agent-us-01",
		AgentToken: "secret",
	})
	if err == nil {
		t.Fatalf("expected missing node error")
	}
}

func TestAgentServiceRegisterAndHeartbeat(t *testing.T) {
	initAgentTestDB(t)

	nodeService := NodeService{}
	if err := nodeService.Save(&model.Node{Name: "US-01", Code: "us-01"}); err != nil {
		t.Fatalf("Save node error = %v", err)
	}

	service := AgentService{}
	agent, err := service.Register(&AgentRegisterRequest{
		NodeCode:       "us-01",
		AgentId:        "agent-us-01",
		AgentToken:     "secret",
		AgentVersion:   "0.1.0",
		SingboxVersion: "1.13.13",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if agent.TokenHash == "secret" || agent.TokenHash == "" {
		t.Fatalf("expected agent token to be hashed")
	}

	err = service.Heartbeat(&AgentHeartbeatRequest{
		AgentId:            "agent-us-01",
		AgentToken:         "secret",
		ConfigVersion:      1,
		AgentStatus:        "online",
		SingboxStatus:      "running",
		AgentVersion:       "0.1.1",
		ResourceSummaryRaw: `{"mem":128}`,
	})
	if err != nil {
		t.Fatalf("Heartbeat() error = %v", err)
	}

	desired, err := service.GetDesiredConfig("agent-us-01", "secret")
	if err != nil {
		t.Fatalf("GetDesiredConfig() error = %v", err)
	}
	if desired.NodeId == 0 {
		t.Fatalf("expected desired config to include node id")
	}

	configService := ConfigVersionService{}
	version := &model.ConfigVersion{
		Version: 1,
		Scope:   "node",
		NodeId:  desired.NodeId,
		Status:  "published",
	}
	if err := configService.SaveConfigVersion(version); err != nil {
		t.Fatalf("SaveConfigVersion() error = %v", err)
	}

	desired, err = service.GetDesiredConfig("agent-us-01", "secret")
	if err != nil {
		t.Fatalf("GetDesiredConfig() with version error = %v", err)
	}
	if desired.ConfigVersion == nil || desired.ConfigVersion.Version != 1 {
		t.Fatalf("expected desired config version 1, got %+v", desired.ConfigVersion)
	}

	if err := service.ReportConfig(&AgentConfigReportRequest{
		AgentId:         "agent-us-01",
		AgentToken:      "secret",
		ConfigVersionId: version.Id,
		Status:          "applied",
	}); err != nil {
		t.Fatalf("ReportConfig() error = %v", err)
	}

	updatedVersion := model.ConfigVersion{}
	if err := database.GetDB().Model(model.ConfigVersion{}).Where("id = ?", version.Id).First(&updatedVersion).Error; err != nil {
		t.Fatalf("load config version error = %v", err)
	}
	if updatedVersion.Status != "applied" {
		t.Fatalf("expected config status applied, got %q", updatedVersion.Status)
	}

	updatedAgent := model.NodeAgent{}
	if err := database.GetDB().Model(model.NodeAgent{}).Where("agent_id = ?", "agent-us-01").First(&updatedAgent).Error; err != nil {
		t.Fatalf("load agent error = %v", err)
	}
	if updatedAgent.LastConfigVersion != version.Version {
		t.Fatalf("expected agent last config version %d, got %d", version.Version, updatedAgent.LastConfigVersion)
	}
}

func TestAgentServiceRegisterRejectsAgentIdMove(t *testing.T) {
	initAgentTestDB(t)

	nodeService := NodeService{}
	if err := nodeService.Save(&model.Node{Name: "US-01", Code: "us-01"}); err != nil {
		t.Fatalf("Save us node error = %v", err)
	}
	if err := nodeService.Save(&model.Node{Name: "SG-01", Code: "sg-01"}); err != nil {
		t.Fatalf("Save sg node error = %v", err)
	}

	service := AgentService{}
	if _, err := service.Register(&AgentRegisterRequest{
		NodeCode:   "us-01",
		AgentId:    "shared-agent",
		AgentToken: "secret",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if _, err := service.Register(&AgentRegisterRequest{
		NodeCode:   "sg-01",
		AgentId:    "shared-agent",
		AgentToken: "secret",
	}); err == nil {
		t.Fatalf("expected reused agent id on another node to fail")
	}
}

func TestAgentServiceReportValidatedDoesNotMarkApplied(t *testing.T) {
	initAgentTestDB(t)

	nodeService := NodeService{}
	if err := nodeService.Save(&model.Node{Name: "US-01", Code: "us-01"}); err != nil {
		t.Fatalf("Save node error = %v", err)
	}

	service := AgentService{}
	agent, err := service.Register(&AgentRegisterRequest{
		NodeCode:   "us-01",
		AgentId:    "agent-us-validated",
		AgentToken: "secret",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	version := &model.ConfigVersion{
		Version: 3,
		Scope:   "node",
		NodeId:  agent.NodeId,
		Status:  "published",
	}
	configService := ConfigVersionService{}
	if err := configService.SaveConfigVersion(version); err != nil {
		t.Fatalf("SaveConfigVersion() error = %v", err)
	}

	if err := service.ReportConfig(&AgentConfigReportRequest{
		AgentId:         "agent-us-validated",
		AgentToken:      "secret",
		ConfigVersionId: version.Id,
		Status:          "validated",
	}); err != nil {
		t.Fatalf("ReportConfig() error = %v", err)
	}

	updatedVersion := model.ConfigVersion{}
	if err := database.GetDB().Model(model.ConfigVersion{}).Where("id = ?", version.Id).First(&updatedVersion).Error; err != nil {
		t.Fatalf("load config version error = %v", err)
	}
	if updatedVersion.Status != "validated" {
		t.Fatalf("expected config status validated, got %q", updatedVersion.Status)
	}

	updatedAgent := model.NodeAgent{}
	if err := database.GetDB().Model(model.NodeAgent{}).Where("agent_id = ?", "agent-us-validated").First(&updatedAgent).Error; err != nil {
		t.Fatalf("load agent error = %v", err)
	}
	if updatedAgent.LastConfigVersion != 0 {
		t.Fatalf("validated config should not update applied version, got %d", updatedAgent.LastConfigVersion)
	}

	desired, err := service.GetDesiredConfig("agent-us-validated", "secret")
	if err != nil {
		t.Fatalf("GetDesiredConfig() error = %v", err)
	}
	if desired.ConfigVersion == nil || desired.ConfigVersion.Id != version.Id {
		t.Fatalf("expected validated config to remain desired, got %+v", desired.ConfigVersion)
	}
}

func TestAgentServiceReportRejectsOtherNodeVersionBeforeDeployment(t *testing.T) {
	initAgentTestDB(t)

	nodeService := NodeService{}
	if err := nodeService.Save(&model.Node{Name: "US-01", Code: "us-01"}); err != nil {
		t.Fatalf("Save node error = %v", err)
	}
	if err := nodeService.Save(&model.Node{Name: "SG-01", Code: "sg-01"}); err != nil {
		t.Fatalf("Save node error = %v", err)
	}

	service := AgentService{}
	if _, err := service.Register(&AgentRegisterRequest{
		NodeCode:   "us-01",
		AgentId:    "agent-us-report",
		AgentToken: "secret",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	sgNode := model.Node{}
	if err := database.GetDB().Model(model.Node{}).Where("code = ?", "sg-01").First(&sgNode).Error; err != nil {
		t.Fatalf("load sg node error = %v", err)
	}
	foreignVersion := &model.ConfigVersion{
		Version: 1,
		Scope:   "node",
		NodeId:  sgNode.Id,
		Status:  "published",
	}
	configService := ConfigVersionService{}
	if err := configService.SaveConfigVersion(foreignVersion); err != nil {
		t.Fatalf("SaveConfigVersion() error = %v", err)
	}

	err := service.ReportConfig(&AgentConfigReportRequest{
		AgentId:         "agent-us-report",
		AgentToken:      "secret",
		ConfigVersionId: foreignVersion.Id,
		Status:          "applied",
	})
	if err == nil {
		t.Fatalf("expected cross-node config report to fail")
	}

	var deployments int64
	if err := database.GetDB().Model(model.ConfigDeployment{}).Count(&deployments).Error; err != nil {
		t.Fatalf("count deployments error = %v", err)
	}
	if deployments != 0 {
		t.Fatalf("expected no deployment record for rejected report, got %d", deployments)
	}
}

func TestAgentServiceDesiredConfigIncludesCertificates(t *testing.T) {
	initAgentTestDB(t)
	t.Setenv("SUI_SECRET_KEY", "test-secret-key-at-least-32-characters")

	nodeService := NodeService{}
	if err := nodeService.Save(&model.Node{Name: "US-01", Code: "us-01"}); err != nil {
		t.Fatalf("Save node error = %v", err)
	}

	agentService := AgentService{}
	if _, err := agentService.Register(&AgentRegisterRequest{
		NodeCode:   "us-01",
		AgentId:    "agent-us-cert",
		AgentToken: "secret",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	certService := CertificateService{}
	certificate := &model.Certificate{Name: "Wildcard", Source: "manual", Enable: true}
	if err := certService.SaveCertificate(certificate); err != nil {
		t.Fatalf("SaveCertificate() error = %v", err)
	}
	version := &model.CertificateVersion{
		CertificateId:          certificate.Id,
		Fingerprint:            "abc123",
		FullchainPEMEncrypted:  "fullchain",
		PrivateKeyPEMEncrypted: "privkey",
	}
	if err := certService.SaveCertificateVersion(version); err != nil {
		t.Fatalf("SaveCertificateVersion() error = %v", err)
	}
	certificate.ActiveVersionId = version.Id
	if err := certService.SaveCertificate(certificate); err != nil {
		t.Fatalf("SaveCertificate(active) error = %v", err)
	}

	db := database.GetDB()
	node := model.Node{}
	if err := db.Model(model.Node{}).Where("code = ?", "us-01").First(&node).Error; err != nil {
		t.Fatalf("load node error = %v", err)
	}
	inbound := model.DistributedInbound{
		Enable:        true,
		NodeId:        node.Id,
		Protocol:      "anytls",
		Tag:           "anytls-us",
		CertificateId: certificate.Id,
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound error = %v", err)
	}

	configService := ConfigVersionService{}
	if err := configService.SaveConfigVersion(&model.ConfigVersion{
		Version: 1,
		Scope:   "node",
		NodeId:  node.Id,
		Status:  "published",
	}); err != nil {
		t.Fatalf("SaveConfigVersion() error = %v", err)
	}

	desired, err := agentService.GetDesiredConfig("agent-us-cert", "secret")
	if err != nil {
		t.Fatalf("GetDesiredConfig() error = %v", err)
	}
	if len(desired.Certificates) != 1 {
		t.Fatalf("expected one certificate bundle, got %+v", desired.Certificates)
	}
	if desired.Certificates[0].FullchainPEM != "fullchain" {
		t.Fatalf("unexpected certificate bundle: %+v", desired.Certificates[0])
	}
}
