package service

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"time"
)

type AgentService struct {
}

type AgentRegisterRequest struct {
	NodeCode       string `json:"nodeCode" form:"nodeCode"`
	AgentId        string `json:"agentId" form:"agentId"`
	AgentToken     string `json:"agentToken" form:"agentToken"`
	AgentVersion   string `json:"agentVersion" form:"agentVersion"`
	SingboxVersion string `json:"singboxVersion" form:"singboxVersion"`
	PublicKey      string `json:"publicKey" form:"publicKey"`
}

type AgentHeartbeatRequest struct {
	AgentId            string `json:"agentId" form:"agentId"`
	AgentToken         string `json:"agentToken" form:"agentToken"`
	ConfigVersion      uint64 `json:"configVersion" form:"configVersion"`
	AgentStatus        string `json:"agentStatus" form:"agentStatus"`
	SingboxStatus      string `json:"singboxStatus" form:"singboxStatus"`
	AgentVersion       string `json:"agentVersion" form:"agentVersion"`
	SingboxVersion     string `json:"singboxVersion" form:"singboxVersion"`
	ResourceSummaryRaw string `json:"resourceSummary" form:"resourceSummary"`
	AppliedSha256      string `json:"appliedSha256" form:"appliedSha256"`
}

type AgentConfigReportRequest struct {
	AgentId         string `json:"agentId" form:"agentId"`
	AgentToken      string `json:"agentToken" form:"agentToken"`
	ConfigVersionId uint64 `json:"configVersionId" form:"configVersionId"`
	Status          string `json:"status" form:"status"`
	ErrorMessage    string `json:"errorMessage" form:"errorMessage"`
}

type AgentDesiredConfig struct {
	NodeId        uint                     `json:"nodeId"`
	ConfigVersion *model.ConfigVersion     `json:"configVersion"`
	Certificates  []AgentCertificateBundle `json:"certificates"`
}

type AgentCertificateBundle struct {
	Id            uint   `json:"id"`
	Name          string `json:"name"`
	Fingerprint   string `json:"fingerprint"`
	FullchainPEM  string `json:"fullchainPem"`
	PrivateKeyPEM string `json:"privateKeyPem"`
	NotBefore     int64  `json:"notBefore"`
	NotAfter      int64  `json:"notAfter"`
}

func (s *AgentService) Register(req *AgentRegisterRequest) (*model.NodeAgent, error) {
	req.NodeCode = strings.TrimSpace(req.NodeCode)
	req.AgentId = strings.TrimSpace(req.AgentId)
	req.AgentToken = strings.TrimSpace(req.AgentToken)

	if req.NodeCode == "" {
		return nil, fmt.Errorf("node code is required")
	}
	if req.AgentId == "" {
		return nil, fmt.Errorf("agent id is required")
	}
	if req.AgentToken == "" {
		return nil, fmt.Errorf("agent token is required")
	}

	db := database.GetDB()
	node := model.Node{}
	if err := db.Model(model.Node{}).Where("code = ?", req.NodeCode).First(&node).Error; err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	agent := model.NodeAgent{}
	err := db.Model(model.NodeAgent{}).Where("agent_id = ?", req.AgentId).First(&agent).Error
	if err != nil && !database.IsNotFound(err) {
		return nil, err
	}
	if agent.Id > 0 && agent.NodeId != node.Id {
		return nil, fmt.Errorf("agent id is already registered to another node")
	}

	if agent.Id == 0 {
		agent.RegisteredAt = now
	}
	agent.NodeId = node.Id
	agent.AgentId = req.AgentId
	agent.AgentVersion = strings.TrimSpace(req.AgentVersion)
	agent.SingboxVersion = strings.TrimSpace(req.SingboxVersion)
	agent.TokenHash = hashAgentToken(req.AgentToken)
	agent.PublicKey = strings.TrimSpace(req.PublicKey)
	agent.LastSeenAt = now

	if err := db.Save(&agent).Error; err != nil {
		return nil, err
	}

	node.AgentStatus = "registered"
	node.LastSeenAt = now
	_ = db.Save(&node).Error

	return &agent, nil
}

func (s *AgentService) Heartbeat(req *AgentHeartbeatRequest) error {
	agent, node, err := s.authenticate(req.AgentId, req.AgentToken)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	if strings.TrimSpace(req.AgentStatus) == "" {
		req.AgentStatus = "online"
	}

	db := database.GetDB()
	agent.AgentVersion = strings.TrimSpace(req.AgentVersion)
	agent.SingboxVersion = strings.TrimSpace(req.SingboxVersion)
	agent.LastConfigVersion = req.ConfigVersion
	agent.LastSeenAt = now
	if err := db.Save(agent).Error; err != nil {
		return err
	}

	node.AgentStatus = strings.TrimSpace(req.AgentStatus)
	node.SingboxStatus = strings.TrimSpace(req.SingboxStatus)
	if req.AppliedSha256 != "" {
		node.AppliedSha256 = strings.TrimSpace(req.AppliedSha256)
	}
	node.LastSeenAt = now
	if err := db.Save(node).Error; err != nil {
		return err
	}

	heartbeat := model.NodeHeartbeat{
		NodeId:          node.Id,
		AgentId:         agent.AgentId,
		ConfigVersion:   req.ConfigVersion,
		AgentStatus:     node.AgentStatus,
		SingboxStatus:   node.SingboxStatus,
		ResourceSummary: normalizeRawJSON(req.ResourceSummaryRaw),
		CreatedAt:       now,
	}
	return db.Create(&heartbeat).Error
}

func (s *AgentService) GetDesiredConfig(agentId string, agentToken string) (*AgentDesiredConfig, error) {
	_, node, err := s.authenticate(agentId, agentToken)
	if err != nil {
		return nil, err
	}

	db := database.GetDB()
	version := model.ConfigVersion{}
	err = db.Model(model.ConfigVersion{}).
		Where("node_id = ? AND status IN ?", node.Id, []string{"published", "pending", "validated", "applied"}).
		Order("version desc, id desc").
		First(&version).Error
	if database.IsNotFound(err) {
		return &AgentDesiredConfig{NodeId: node.Id}, nil
	}
	if err != nil {
		return nil, err
	}

	certificates, err := s.getDesiredCertificates(node.Id)
	if err != nil {
		return nil, err
	}

	return &AgentDesiredConfig{
		NodeId:        node.Id,
		ConfigVersion: &version,
		Certificates:  certificates,
	}, nil
}

func (s *AgentService) ReportConfig(req *AgentConfigReportRequest) error {
	agent, node, err := s.authenticate(req.AgentId, req.AgentToken)
	if err != nil {
		return err
	}
	if req.ConfigVersionId == 0 {
		return fmt.Errorf("config version id is required")
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "reported"
	}

	now := time.Now().Unix()
	db := database.GetDB()

	version := model.ConfigVersion{}
	if err := db.Model(model.ConfigVersion{}).
		Where("id = ? AND node_id = ?", req.ConfigVersionId, node.Id).
		First(&version).Error; err != nil {
		return err
	}

	deployment := model.ConfigDeployment{
		ConfigVersionId: req.ConfigVersionId,
		NodeId:          node.Id,
		Status:          status,
		ErrorMessage:    strings.TrimSpace(req.ErrorMessage),
		StartedAt:       now,
	}
	if status == "applied" || status == "failed" || status == "rolled_back" {
		deployment.FinishedAt = now
	}

	if err := db.Create(&deployment).Error; err != nil {
		return err
	}

	switch status {
	case "pulled":
		version.Status = "pending"
	case "validated":
		version.Status = "validated"
	case "applied":
		version.Status = "applied"
		agent.LastConfigVersion = version.Version
	case "failed":
		version.Status = "failed"
	}
	if err := db.Save(&version).Error; err != nil {
		return err
	}
	if status == "applied" {
		return db.Save(agent).Error
	}
	return nil
}

func (s *AgentService) authenticate(agentId string, agentToken string) (*model.NodeAgent, *model.Node, error) {
	agentId = strings.TrimSpace(agentId)
	agentToken = strings.TrimSpace(agentToken)
	if agentId == "" || agentToken == "" {
		return nil, nil, fmt.Errorf("agent credentials are required")
	}

	db := database.GetDB()
	agent := model.NodeAgent{}
	if err := db.Model(model.NodeAgent{}).Where("agent_id = ?", agentId).First(&agent).Error; err != nil {
		return nil, nil, err
	}
	if subtle.ConstantTimeCompare([]byte(agent.TokenHash), []byte(hashAgentToken(agentToken))) != 1 {
		return nil, nil, fmt.Errorf("invalid agent token")
	}

	node := model.Node{}
	if err := db.Model(model.Node{}).Where("id = ? AND enable = ?", agent.NodeId, true).First(&node).Error; err != nil {
		if database.IsNotFound(err) {
			return nil, nil, fmt.Errorf("node is disabled or not found")
		}
		return nil, nil, err
	}
	return &agent, &node, nil
}

func hashAgentToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *AgentService) getDesiredCertificates(nodeId uint) ([]AgentCertificateBundle, error) {
	db := database.GetDB()
	inbounds := []model.DistributedInbound{}
	if err := db.Model(model.DistributedInbound{}).
		Where("node_id = ? AND certificate_id > 0 AND enable = ?", nodeId, true).
		Scan(&inbounds).Error; err != nil {
		return nil, err
	}

	bundles := []AgentCertificateBundle{}
	seen := map[uint]bool{}
	for _, inbound := range inbounds {
		if seen[inbound.CertificateId] {
			continue
		}
		seen[inbound.CertificateId] = true

		certificate := model.Certificate{}
		if err := db.Model(model.Certificate{}).Where("id = ?", inbound.CertificateId).First(&certificate).Error; err != nil {
			return nil, err
		}
		if certificate.ActiveVersionId == 0 {
			continue
		}

		version := model.CertificateVersion{}
		if err := db.Model(model.CertificateVersion{}).Where("id = ?", certificate.ActiveVersionId).First(&version).Error; err != nil {
			return nil, err
		}
		fullchainPEM, privateKeyPEM, err := certificateVersionPEM(version)
		if err != nil {
			return nil, err
		}

		bundles = append(bundles, AgentCertificateBundle{
			Id:            certificate.Id,
			Name:          certificate.Name,
			Fingerprint:   version.Fingerprint,
			FullchainPEM:  fullchainPEM,
			PrivateKeyPEM: privateKeyPEM,
			NotBefore:     version.NotBefore,
			NotAfter:      version.NotAfter,
		})
	}

	return bundles, nil
}

func normalizeRawJSON(value string) []byte {
	value = strings.TrimSpace(value)
	if value == "" {
		return []byte("{}")
	}
	return []byte(value)
}
