package service

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"s-ui/config"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
)

type AnyTLSUser struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type AnyTLSInboundInput struct {
	Tag           string
	Listen        string
	ListenPort    uint
	Users         []AnyTLSUser
	PaddingScheme json.RawMessage
	ServerName    string
	CertPath      string
	KeyPath       string
}

type AnyTLSOutboundInput struct {
	Tag        string
	Server     string
	ServerPort uint
	Password   string
	ServerName string
}

func RenderAnyTLSInbound(input AnyTLSInboundInput) (json.RawMessage, error) {
	input.Tag = strings.TrimSpace(input.Tag)
	input.Listen = strings.TrimSpace(input.Listen)
	input.ServerName = strings.TrimSpace(input.ServerName)
	input.CertPath = strings.TrimSpace(input.CertPath)
	input.KeyPath = strings.TrimSpace(input.KeyPath)

	if input.Tag == "" {
		return nil, fmt.Errorf("tag is required")
	}
	if input.Listen == "" {
		input.Listen = "::"
	}
	if input.ListenPort == 0 {
		return nil, fmt.Errorf("listen port is required")
	}
	if input.ServerName == "" {
		return nil, fmt.Errorf("server name is required")
	}
	if input.CertPath == "" || input.KeyPath == "" {
		return nil, fmt.Errorf("certificate and key paths are required")
	}
	if len(input.Users) == 0 {
		return nil, fmt.Errorf("at least one anytls user is required")
	}

	users := make([]AnyTLSUser, 0, len(input.Users))
	for _, user := range input.Users {
		user.Name = strings.TrimSpace(user.Name)
		user.Password = strings.TrimSpace(user.Password)
		if user.Name == "" || user.Password == "" {
			return nil, fmt.Errorf("anytls user name and password are required")
		}
		users = append(users, user)
	}

	paddingScheme := json.RawMessage(`[]`)
	if len(input.PaddingScheme) > 0 {
		paddingScheme = input.PaddingScheme
	}

	config := map[string]interface{}{
		"type":           "anytls",
		"tag":            input.Tag,
		"listen":         input.Listen,
		"listen_port":    input.ListenPort,
		"users":          users,
		"padding_scheme": paddingScheme,
		"tls": map[string]interface{}{
			"enabled":          true,
			"server_name":      input.ServerName,
			"certificate_path": input.CertPath,
			"key_path":         input.KeyPath,
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func RenderAnyTLSOutbound(input AnyTLSOutboundInput) (json.RawMessage, error) {
	input.Tag = strings.TrimSpace(input.Tag)
	input.Server = strings.TrimSpace(input.Server)
	input.Password = strings.TrimSpace(input.Password)
	input.ServerName = strings.TrimSpace(input.ServerName)

	if input.Tag == "" {
		return nil, fmt.Errorf("tag is required")
	}
	if input.Server == "" {
		return nil, fmt.Errorf("server is required")
	}
	if input.ServerPort == 0 {
		return nil, fmt.Errorf("server port is required")
	}
	if input.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if input.ServerName == "" {
		input.ServerName = input.Server
	}

	config := map[string]interface{}{
		"type":        "anytls",
		"tag":         input.Tag,
		"server":      input.Server,
		"server_port": input.ServerPort,
		"password":    input.Password,
		"tls": map[string]interface{}{
			"enabled":     true,
			"server_name": input.ServerName,
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func AnyTLSCertificatePaths(certDir string, certificateName string, certificateId uint, fingerprint string) (string, string) {
	dirName := certificateDirName(certificateId, certificateName, fingerprint)
	certPath := filepath.Join(certDir, dirName, "fullchain.pem")
	keyPath := filepath.Join(certDir, dirName, "privkey.pem")
	return certPath, keyPath
}

func RenderDistributedAnyTLSInbound(inboundId uint) (json.RawMessage, error) {
	if inboundId == 0 {
		return nil, fmt.Errorf("inbound id is required")
	}

	db := database.GetDB()
	inbound := model.DistributedInbound{}
	if err := db.Model(model.DistributedInbound{}).Where("id = ?", inboundId).First(&inbound).Error; err != nil {
		return nil, err
	}
	if inbound.Protocol != "anytls" {
		return nil, fmt.Errorf("inbound protocol must be anytls")
	}

	node := model.Node{}
	if err := db.Model(model.Node{}).Where("id = ?", inbound.NodeId).First(&node).Error; err != nil {
		return nil, err
	}

	certificate := model.Certificate{}
	if err := db.Model(model.Certificate{}).Where("id = ?", inbound.CertificateId).First(&certificate).Error; err != nil {
		return nil, err
	}
	if certificate.ActiveVersionId == 0 {
		return nil, fmt.Errorf("certificate has no active version")
	}

	certificateVersion := model.CertificateVersion{}
	if err := db.Model(model.CertificateVersion{}).Where("id = ?", certificate.ActiveVersionId).First(&certificateVersion).Error; err != nil {
		return nil, err
	}

	inboundUsers := []model.InboundUser{}
	err := db.Model(&model.InboundUser{}).
		Joins("JOIN clients ON clients.id = inbound_users.client_id").
		Where("inbound_users.inbound_id = ? AND clients.enable = ?", inbound.Id, true).
		Order("inbound_users.id asc").
		Scan(&inboundUsers).Error
	if err != nil {
		return nil, err
	}

	users := make([]AnyTLSUser, 0, len(inboundUsers))
	for _, user := range inboundUsers {
		users = append(users, AnyTLSUser{
			Name:     user.Name,
			Password: user.Password,
		})
	}

	publicHost := strings.TrimSpace(inbound.PublicHost)
	if publicHost == "" {
		publicHost = node.PublicHost
	}
	tag := strings.TrimSpace(inbound.Tag)
	if tag == "" {
		tag = fmt.Sprintf("anytls-%s-%d", node.Code, inbound.ListenPort)
	}

	certPath, keyPath := AnyTLSCertificatePaths(config.GetAgentCertDir(), certificate.Name, certificate.Id, certificateVersion.Fingerprint)
	raw, err := RenderAnyTLSInbound(AnyTLSInboundInput{
		Tag:        tag,
		Listen:     inbound.Listen,
		ListenPort: inbound.ListenPort,
		Users:      users,
		ServerName: publicHost,
		CertPath:   certPath,
		KeyPath:    keyPath,
	})
	if err != nil {
		return nil, err
	}
	raw, err = mergeInboundOverrides(raw, inbound.AdvancedOverridesJson)
	if err != nil {
		return nil, err
	}

	inbound.Tag = tag
	inbound.PublicHost = publicHost
	inbound.RenderedConfigJson = raw
	if err := db.Save(&inbound).Error; err != nil {
		return nil, err
	}

	return raw, nil
}

func mergeInboundOverrides(baseRaw json.RawMessage, overrideRaw json.RawMessage) (json.RawMessage, error) {
	overrideText := strings.TrimSpace(string(overrideRaw))
	if overrideText == "" || overrideText == "null" || overrideText == "{}" {
		return baseRaw, nil
	}
	base := map[string]interface{}{}
	if err := json.Unmarshal(baseRaw, &base); err != nil {
		return nil, err
	}
	overrides := map[string]interface{}{}
	if err := json.Unmarshal(overrideRaw, &overrides); err != nil {
		return nil, fmt.Errorf("anytls advanced json: %w", err)
	}
	merged := mergeJSONObjects(base, overrides)
	data, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func mergeJSONObjects(base map[string]interface{}, overrides map[string]interface{}) map[string]interface{} {
	for key, value := range overrides {
		overrideMap, overrideIsMap := value.(map[string]interface{})
		baseMap, baseIsMap := base[key].(map[string]interface{})
		if overrideIsMap && baseIsMap {
			base[key] = mergeJSONObjects(baseMap, overrideMap)
			continue
		}
		base[key] = value
	}
	return base
}

func certificateDirName(id uint, name string, fingerprint string) string {
	safeName := safeFileName(name)
	if safeName == "" {
		safeName = fmt.Sprintf("cert-%d", id)
	}
	if strings.TrimSpace(fingerprint) == "" {
		return safeName
	}
	return fmt.Sprintf("%s-%s", safeName, shortFingerprint(fingerprint))
}

func safeFileName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	builder := strings.Builder{}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			continue
		}
		if r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('-')
	}
	return strings.Trim(builder.String(), "-")
}

func shortFingerprint(value string) string {
	value = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), ":", ""))
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}
