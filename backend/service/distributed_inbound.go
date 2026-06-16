package service

import (
	"encoding/json"
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"time"
)

type DistributedInboundService struct {
}

func (s *DistributedInboundService) GetDistributedInbounds(nodeId uint) ([]model.DistributedInbound, error) {
	db := database.GetDB()
	inbounds := []model.DistributedInbound{}
	query := db.Model(model.DistributedInbound{}).Order("id asc")
	if nodeId > 0 {
		query = query.Where("node_id = ?", nodeId)
	}
	err := query.Scan(&inbounds).Error
	if err != nil {
		return nil, err
	}
	return inbounds, nil
}

func (s *DistributedInboundService) SaveDistributedInbound(inbound *model.DistributedInbound) error {
	if inbound.NodeId == 0 {
		return fmt.Errorf("node id is required")
	}
	inbound.Protocol = strings.TrimSpace(inbound.Protocol)
	if inbound.Protocol == "" {
		return fmt.Errorf("protocol is required")
	}
	if inbound.ListenPort == 0 {
		return fmt.Errorf("listen port is required")
	}
	if err := validateJSONMap(inbound.FormValuesJson, "form values json"); err != nil {
		return err
	}
	if err := validateJSONMap(inbound.AdvancedOverridesJson, "advanced overrides json"); err != nil {
		return err
	}
	if err := validateJSONMap(inbound.PolicyOverridesJson, "policy overrides json"); err != nil {
		return err
	}

	now := time.Now().Unix()
	db := database.GetDB()
	if hasMeaningfulJSON(inbound.PolicyOverridesJson) {
		var count int64
		query := db.Model(model.DistributedInbound{}).
			Where("node_id = ? AND policy_overrides_json IS NOT NULL AND trim(policy_overrides_json) NOT IN ('', '{}', 'null')", inbound.NodeId)
		if inbound.Id > 0 {
			query = query.Where("id <> ?", inbound.Id)
		}
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("node policy overrides already exist on another inbound; preview supports one policy override per node")
		}
	}

	if inbound.Enable {
		conflicts := []model.DistributedInbound{}
		conflictQuery := db.Model(model.DistributedInbound{}).
			Where("enable = ? AND node_id = ? AND listen_port = ?", true, inbound.NodeId, inbound.ListenPort)
		if inbound.Id > 0 {
			conflictQuery = conflictQuery.Where("id <> ?", inbound.Id)
		}
		if err := conflictQuery.Find(&conflicts).Error; err != nil {
			return err
		}
		for _, existing := range conflicts {
			if listenAddressesConflict(inbound.Listen, existing.Listen) {
				return fmt.Errorf("listen address and port already used by another enabled inbound")
			}
		}
	}

	if inbound.Id == 0 {
		inbound.CreatedAt = now
	} else if inbound.CreatedAt == 0 {
		existing := model.DistributedInbound{}
		if err := db.Model(model.DistributedInbound{}).Where("id = ?", inbound.Id).First(&existing).Error; err != nil {
			return err
		}
		inbound.CreatedAt = existing.CreatedAt
	}
	inbound.UpdatedAt = now

	return db.Save(inbound).Error
}

func listenAddressesConflict(left string, right string) bool {
	left = normalizeListenAddress(left)
	right = normalizeListenAddress(right)
	if left == right {
		return true
	}
	return isWildcardListenAddress(left) || isWildcardListenAddress(right)
}

func normalizeListenAddress(listen string) string {
	listen = strings.TrimSpace(strings.ToLower(listen))
	if listen == "" {
		return "::"
	}
	if listen == "[::]" {
		return "::"
	}
	return listen
}

func isWildcardListenAddress(listen string) bool {
	listen = normalizeListenAddress(listen)
	return listen == "::" || listen == "0.0.0.0" || listen == "*"
}

func validateJSONMap(raw json.RawMessage, name string) error {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" {
		return nil
	}
	value := map[string]interface{}{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func hasMeaningfulJSON(raw json.RawMessage) bool {
	text := strings.TrimSpace(string(raw))
	return text != "" && text != "null" && text != "{}"
}

func (s *DistributedInboundService) DeleteDistributedInbound(id uint) error {
	if id == 0 {
		return fmt.Errorf("distributed inbound id is required")
	}
	db := database.GetDB()
	return db.Where("id = ?", id).Delete(model.DistributedInbound{}).Error
}

func (s *DistributedInboundService) GetInboundUsers(inboundId uint) ([]model.InboundUser, error) {
	db := database.GetDB()
	users := []model.InboundUser{}
	query := db.Model(model.InboundUser{}).Order("id asc")
	if inboundId > 0 {
		query = query.Where("inbound_id = ?", inboundId)
	}
	err := query.Scan(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *DistributedInboundService) SaveInboundUser(user *model.InboundUser) error {
	if user.InboundId == 0 {
		return fmt.Errorf("inbound id is required")
	}
	user.Name = strings.TrimSpace(user.Name)
	user.Password = strings.TrimSpace(user.Password)
	if user.Name == "" || user.Password == "" {
		return fmt.Errorf("inbound user name and password are required")
	}

	now := time.Now().Unix()
	db := database.GetDB()
	if user.ClientId > 0 {
		client := model.Client{}
		if err := db.Model(model.Client{}).Where("id = ?", user.ClientId).First(&client).Error; err != nil {
			return err
		}
		if !clientAllowsCluster(client) {
			return fmt.Errorf("client is not allowed to use cluster access")
		}
	}
	if user.Id == 0 {
		user.CreatedAt = now
	} else if user.CreatedAt == 0 {
		existing := model.InboundUser{}
		if err := db.Model(model.InboundUser{}).Where("id = ?", user.Id).First(&existing).Error; err != nil {
			return err
		}
		user.CreatedAt = existing.CreatedAt
	}
	user.UpdatedAt = now

	return db.Save(user).Error
}

func (s *DistributedInboundService) DeleteInboundUser(id uint) error {
	if id == 0 {
		return fmt.Errorf("inbound user id is required")
	}
	db := database.GetDB()
	return db.Where("id = ?", id).Delete(model.InboundUser{}).Error
}

func RenderDistributedInbound(inboundId uint) (json.RawMessage, error) {
	if inboundId == 0 {
		return nil, fmt.Errorf("inbound id is required")
	}

	db := database.GetDB()
	inbound := model.DistributedInbound{}
	if err := db.Model(model.DistributedInbound{}).Where("id = ?", inboundId).First(&inbound).Error; err != nil {
		return nil, err
	}

	switch inbound.Protocol {
	case "anytls":
		return RenderDistributedAnyTLSInbound(inboundId)
	default:
		return renderDistributedCustomInbound(&inbound)
	}
}

func renderDistributedCustomInbound(inbound *model.DistributedInbound) (json.RawMessage, error) {
	if inbound == nil {
		return nil, fmt.Errorf("inbound is required")
	}
	raw := strings.TrimSpace(string(inbound.AdvancedOverridesJson))
	if raw == "" || raw == "null" || raw == "{}" {
		return nil, fmt.Errorf("%s inbound requires advanced JSON template", inbound.Protocol)
	}

	config := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return nil, fmt.Errorf("%s inbound advanced json: %w", inbound.Protocol, err)
	}
	if config["type"] == nil || strings.TrimSpace(fmt.Sprint(config["type"])) == "" {
		config["type"] = inbound.Protocol
	}
	if strings.TrimSpace(fmt.Sprint(config["type"])) != inbound.Protocol {
		return nil, fmt.Errorf("advanced json type must match selected protocol %q", inbound.Protocol)
	}
	if config["tag"] == nil || strings.TrimSpace(fmt.Sprint(config["tag"])) == "" {
		config["tag"] = defaultInboundTag(inbound)
	}
	if config["listen"] == nil && inbound.Listen != "" {
		config["listen"] = inbound.Listen
	}
	if config["listen_port"] == nil && inbound.ListenPort > 0 {
		config["listen_port"] = inbound.ListenPort
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	inbound.RenderedConfigJson = data
	inbound.UpdatedAt = time.Now().Unix()
	if err := database.GetDB().Save(inbound).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func defaultInboundTag(inbound *model.DistributedInbound) string {
	node := model.Node{}
	if err := database.GetDB().Model(model.Node{}).Where("id = ?", inbound.NodeId).First(&node).Error; err != nil {
		return fmt.Sprintf("%s-%d", inbound.Protocol, inbound.ListenPort)
	}
	if node.Code != "" {
		return fmt.Sprintf("%s-%s-%d", inbound.Protocol, node.Code, inbound.ListenPort)
	}
	return fmt.Sprintf("%s-%d", inbound.Protocol, inbound.ListenPort)
}
