package sub

import (
	"encoding/json"
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"s-ui/service"
	"strings"
)

type DistributedService struct {
	JsonService
	service.SubscriptionService
}

func (s *DistributedService) GetDistributedJson(subId string) (*string, error) {
	client, err := s.getDistributedClientOrToken(subId)
	if err != nil {
		return nil, err
	}

	outbounds, outTags, err := s.getDistributedOutbounds(client.Id)
	if err != nil {
		return nil, err
	}
	s.JsonService.addDefaultOutbounds(&outbounds, &outTags)

	var jsonConfig map[string]interface{}
	if err := json.Unmarshal([]byte(defaultJson), &jsonConfig); err != nil {
		return nil, err
	}
	jsonConfig["outbounds"] = &outbounds
	if err := s.JsonService.addOthers(&jsonConfig); err != nil {
		return nil, err
	}

	result, err := json.MarshalIndent(jsonConfig, " ", "  ")
	if err != nil {
		return nil, err
	}
	resultStr := string(result)
	return &resultStr, nil
}

func (s *DistributedService) GetDistributedRaw(subId string) (*string, error) {
	client, err := s.getDistributedClientOrToken(subId)
	if err != nil {
		return nil, err
	}

	outbounds, _, err := s.getDistributedOutbounds(client.Id)
	if err != nil {
		return nil, err
	}

	lines := ""
	for _, outbound := range outbounds {
		data, err := json.Marshal(outbound)
		if err != nil {
			return nil, err
		}
		lines += string(data) + "\n"
	}
	return &lines, nil
}

func (s *DistributedService) getDistributedClient(subId string) (*model.Client, error) {
	db := database.GetDB()
	client := &model.Client{}
	err := db.Model(model.Client{}).Where("enable = true and name = ?", subId).First(client).Error
	if err != nil {
		return nil, err
	}
	if !service.ClientAllowsCluster(*client) {
		return nil, fmt.Errorf("client is not allowed to use cluster subscriptions")
	}
	return client, nil
}

func (s *DistributedService) getDistributedClientOrToken(subId string) (*model.Client, error) {
	client, err := s.SubscriptionService.ResolveClientByToken(subId)
	if err == nil {
		return client, nil
	}
	return s.getDistributedClient(subId)
}

func (s *DistributedService) getDistributedOutbounds(clientId uint) ([]map[string]interface{}, []string, error) {
	db := database.GetDB()
	inboundUsers := []model.InboundUser{}
	if err := db.Model(model.InboundUser{}).Where("client_id = ?", clientId).Order("id asc").Scan(&inboundUsers).Error; err != nil {
		return nil, nil, err
	}

	outbounds := []map[string]interface{}{}
	outTags := []string{}
	for _, inboundUser := range inboundUsers {
		inbound := model.DistributedInbound{}
		if err := db.Model(model.DistributedInbound{}).Where("id = ? AND enable = ?", inboundUser.InboundId, true).First(&inbound).Error; err != nil {
			if database.IsNotFound(err) {
				continue
			}
			return nil, nil, err
		}
		if inbound.Protocol != "anytls" {
			outbound, tag, err := s.renderDistributedOutbound(inbound, inboundUser)
			if err != nil {
				return nil, nil, err
			}
			if outbound == nil {
				continue
			}
			outbounds = append(outbounds, outbound)
			outTags = append(outTags, tag)
			continue
		}

		outbound, tag, err := s.renderAnyTLSDistributedOutbound(inbound, inboundUser)
		if err != nil {
			return nil, nil, err
		}
		outbounds = append(outbounds, outbound)
		outTags = append(outTags, tag)
	}

	return outbounds, outTags, nil
}

func (s *DistributedService) renderAnyTLSDistributedOutbound(inbound model.DistributedInbound, inboundUser model.InboundUser) (map[string]interface{}, string, error) {
	node, err := enabledNode(inbound.NodeId)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, "", nil
		}
		return nil, "", err
	}

	publicHost := publicInboundHost(inbound, node)
	tag := distributedOutboundTag(node, inbound, inboundUser)

	raw, err := service.RenderAnyTLSOutbound(service.AnyTLSOutboundInput{
		Tag:        tag,
		Server:     publicHost,
		ServerPort: inbound.ListenPort,
		Password:   inboundUser.Password,
		ServerName: publicHost,
	})
	if err != nil {
		return nil, "", err
	}

	outbound := map[string]interface{}{}
	if err := json.Unmarshal(raw, &outbound); err != nil {
		return nil, "", err
	}
	return outbound, tag, nil
}

func (s *DistributedService) renderDistributedOutbound(inbound model.DistributedInbound, inboundUser model.InboundUser) (map[string]interface{}, string, error) {
	node, err := enabledNode(inbound.NodeId)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, "", nil
		}
		return nil, "", err
	}

	config, err := distributedInboundConfig(inbound)
	if err != nil {
		return nil, "", err
	}
	if config == nil {
		return nil, "", nil
	}

	publicHost := publicInboundHost(inbound, node)
	tag := distributedOutboundTag(node, inbound, inboundUser)
	userConfig := findRenderedUser(config, inboundUser)
	credential := userCredential(inboundUser, userConfig)

	outbound := map[string]interface{}{
		"type":        inbound.Protocol,
		"tag":         tag,
		"server":      publicHost,
		"server_port": inbound.ListenPort,
	}
	if inbound.ListenPort == 0 {
		if port, ok := uintValue(config["listen_port"]); ok {
			outbound["server_port"] = port
		}
	}

	addClientTLS(outbound, config, publicHost)
	copyOptionalObject(outbound, config, "transport")
	copyOptionalObject(outbound, config, "multiplex")

	switch inbound.Protocol {
	case "hysteria":
		outbound["auth_str"] = credential
		copyOptional(outbound, config, "up_mbps")
		copyOptional(outbound, config, "down_mbps")
		copyOptional(outbound, config, "obfs")
	case "hysteria2":
		outbound["password"] = credential
		copyOptional(outbound, config, "up_mbps")
		copyOptional(outbound, config, "down_mbps")
		copyOptional(outbound, config, "obfs")
	case "tuic":
		outbound["uuid"] = stringFromMap(userConfig, "uuid", credential)
		outbound["password"] = stringFromMap(userConfig, "password", credential)
		copyOptional(outbound, config, "congestion_control")
	case "trojan":
		outbound["password"] = credential
	case "vless":
		outbound["uuid"] = stringFromMap(userConfig, "uuid", credential)
		if flow := stringFromMap(userConfig, "flow", ""); flow != "" {
			outbound["flow"] = flow
		}
	case "vmess":
		outbound["uuid"] = stringFromMap(userConfig, "uuid", credential)
		outbound["security"] = stringValue(config["security"], "auto")
		if alterId, ok := userConfig["alterId"]; ok {
			outbound["alter_id"] = alterId
		} else if alterId, ok := userConfig["alter_id"]; ok {
			outbound["alter_id"] = alterId
		}
	case "naive":
		outbound["username"] = stringFromMap(userConfig, "username", inboundUser.Name)
		outbound["password"] = stringFromMap(userConfig, "password", credential)
	case "shadowsocks":
		outbound["method"] = stringValue(config["method"], "2022-blake3-aes-128-gcm")
		outbound["password"] = credential
		copyOptional(outbound, config, "plugin")
		copyOptional(outbound, config, "plugin_opts")
	case "shadowtls":
		outbound["version"] = valueOrDefault(config["version"], 3)
		outbound["password"] = credential
		if tls, ok := outbound["tls"].(map[string]interface{}); ok {
			if handshake, ok := config["handshake"].(map[string]interface{}); ok {
				if server, ok := handshake["server"]; ok && tls["server_name"] == nil {
					tls["server_name"] = server
				}
			}
		}
	default:
		return nil, "", fmt.Errorf("distributed subscription does not support protocol %q yet", inbound.Protocol)
	}

	return outbound, tag, nil
}

func enabledNode(nodeId uint) (model.Node, error) {
	node := model.Node{}
	err := database.GetDB().Model(model.Node{}).Where("id = ? AND enable = ?", nodeId, true).First(&node).Error
	return node, err
}

func distributedInboundConfig(inbound model.DistributedInbound) (map[string]interface{}, error) {
	raw := strings.TrimSpace(string(inbound.RenderedConfigJson))
	if raw == "" || raw == "null" || raw == "{}" {
		raw = strings.TrimSpace(string(inbound.AdvancedOverridesJson))
	}
	if raw == "" || raw == "null" || raw == "{}" {
		rendered, err := service.RenderDistributedInbound(inbound.Id)
		if err != nil {
			return nil, err
		}
		raw = string(rendered)
	}
	config := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return nil, err
	}
	return config, nil
}

func publicInboundHost(inbound model.DistributedInbound, node model.Node) string {
	publicHost := strings.TrimSpace(inbound.PublicHost)
	if publicHost == "" {
		publicHost = strings.TrimSpace(node.PublicHost)
	}
	return publicHost
}

func distributedOutboundTag(node model.Node, inbound model.DistributedInbound, inboundUser model.InboundUser) string {
	nodeLabel := strings.TrimSpace(node.Name)
	if nodeLabel == "" {
		nodeLabel = strings.TrimSpace(node.Code)
	}
	if nodeLabel == "" {
		nodeLabel = fmt.Sprintf("node-%d", node.Id)
	}
	protocol := strings.TrimSpace(inbound.Protocol)
	if protocol == "" {
		protocol = "inbound"
	}
	return fmt.Sprintf("%s-%s-%s", nodeLabel, protocol, inboundUser.Name)
}

func findRenderedUser(config map[string]interface{}, inboundUser model.InboundUser) map[string]interface{} {
	users, ok := config["users"].([]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	for _, item := range users {
		user, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name := stringValue(user["name"], "")
		if name == "" {
			name = stringValue(user["username"], "")
		}
		if name == inboundUser.Name {
			return user
		}
	}
	return map[string]interface{}{}
}

func userCredential(inboundUser model.InboundUser, userConfig map[string]interface{}) string {
	if password := stringFromMap(userConfig, "password", ""); password != "" && !strings.Contains(password, "change-me") {
		return password
	}
	if uuid := stringFromMap(userConfig, "uuid", ""); uuid != "" && !strings.HasPrefix(uuid, "00000000-") {
		return uuid
	}
	if auth := stringFromMap(userConfig, "auth_str", ""); auth != "" {
		return auth
	}
	return inboundUser.Password
}

func addClientTLS(outbound map[string]interface{}, config map[string]interface{}, publicHost string) {
	tls, ok := config["tls"].(map[string]interface{})
	if !ok {
		return
	}
	clientTLS := map[string]interface{}{}
	for key, value := range tls {
		switch key {
		case "certificate_path", "key_path", "certificate", "key":
			continue
		case "reality":
			if reality, ok := value.(map[string]interface{}); ok {
				clientReality := map[string]interface{}{}
				for realityKey, realityValue := range reality {
					switch realityKey {
					case "private_key":
						continue
					default:
						clientReality[realityKey] = realityValue
					}
				}
				clientTLS[key] = clientReality
				continue
			}
		default:
			clientTLS[key] = value
		}
	}
	if clientTLS["enabled"] == nil {
		clientTLS["enabled"] = true
	}
	if strings.TrimSpace(fmt.Sprint(clientTLS["server_name"])) == "" {
		clientTLS["server_name"] = publicHost
	}
	outbound["tls"] = clientTLS
}

func copyOptional(outbound map[string]interface{}, config map[string]interface{}, key string) {
	if value, ok := config[key]; ok {
		outbound[key] = value
	}
}

func copyOptionalObject(outbound map[string]interface{}, config map[string]interface{}, key string) {
	if value, ok := config[key].(map[string]interface{}); ok {
		outbound[key] = value
	}
}

func stringFromMap(data map[string]interface{}, key string, fallback string) string {
	if data == nil {
		return fallback
	}
	return stringValue(data[key], fallback)
}

func stringValue(value interface{}, fallback string) string {
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" || text == "<nil>" {
		return fallback
	}
	return text
}

func valueOrDefault(value interface{}, fallback interface{}) interface{} {
	if value == nil {
		return fallback
	}
	return value
}

func uintValue(value interface{}) (uint, bool) {
	switch typed := value.(type) {
	case float64:
		if typed > 0 {
			return uint(typed), true
		}
	case int:
		if typed > 0 {
			return uint(typed), true
		}
	case uint:
		return typed, typed > 0
	}
	return 0, false
}
