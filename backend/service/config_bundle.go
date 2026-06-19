package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"time"
)

type SingboxNodeConfig struct {
	Log          map[string]interface{}   `json:"log,omitempty"`
	DNS          map[string]interface{}   `json:"dns,omitempty"`
	Inbounds     []json.RawMessage        `json:"inbounds"`
	Outbounds    []map[string]interface{} `json:"outbounds"`
	Route        map[string]interface{}   `json:"route,omitempty"`
	Experimental map[string]interface{}   `json:"experimental,omitempty"`
}

func renderNodeConfigJson(nodeId uint) ([]byte, error) {
	db := database.GetDB()
	node := model.Node{}
	if err := db.Model(model.Node{}).Where("id = ?", nodeId).First(&node).Error; err != nil {
		return nil, err
	}

	inbounds := []model.DistributedInbound{}
	if err := db.Model(model.DistributedInbound{}).
		Where("node_id = ? AND enable = ?", nodeId, true).
		Order("id asc").
		Scan(&inbounds).Error; err != nil {
		return nil, err
	}

	renderedInbounds := make([]json.RawMessage, 0, len(inbounds))
	for _, inbound := range inbounds {
		rendered := inbound.RenderedConfigJson
		if len(rendered) == 0 || string(rendered) == "null" {
			raw, err := RenderDistributedInbound(inbound.Id)
			if err != nil {
				return nil, err
			}
			rendered = raw
		}
		renderedInbounds = append(renderedInbounds, rendered)
	}

	templateService := ConfigTemplateService{}
	templates, err := templateService.GetNodeConfigTemplates()
	if err != nil {
		return nil, err
	}
	log, dns, outbounds, route, experimental, err := decodeNodeConfigTemplates(templates)
	if err != nil {
		return nil, err
	}
	for _, inbound := range inbounds {
		log, dns, outbounds, route, experimental, err = applyInboundPolicyOverrides(inbound.PolicyOverridesJson, log, dns, outbounds, route, experimental)
		if err != nil {
			return nil, fmt.Errorf("%s:%d policy overrides: %w", inbound.Protocol, inbound.ListenPort, err)
		}
	}

	config := SingboxNodeConfig{
		Log:          log,
		DNS:          dns,
		Inbounds:     renderedInbounds,
		Outbounds:    outbounds,
		Route:        route,
		Experimental: experimental,
	}

	return json.Marshal(config)
}

func PublishNodeConfigVersion(nodeId uint, actor string) (*model.ConfigVersion, error) {
	if nodeId == 0 {
		return nil, fmt.Errorf("node id is required")
	}

	content, err := renderNodeConfigJson(nodeId)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(content)
	sha256Hex := hex.EncodeToString(sum[:])

	db := database.GetDB()
	latest := model.ConfigVersion{}
	latestErr := db.Model(model.ConfigVersion{}).
		Where("node_id = ?", nodeId).
		Order("version desc").
		First(&latest).Error

	var maxVersion uint64
	if err := db.Model(model.ConfigVersion{}).
		Where("node_id = ?", nodeId).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error; err != nil {
		return nil, err
	}

	var version *model.ConfigVersion
	if latestErr == nil && latest.Sha256 == sha256Hex {
		version = &latest
	} else {
		newVersion := model.ConfigVersion{
			Version:     maxVersion + 1,
			Scope:       "node",
			NodeId:      nodeId,
			Sha256:      sha256Hex,
			Status:      "published",
			ContentJson: content,
			CreatedBy:   actor,
			CreatedAt:   time.Now().Unix(),
		}
		if err := db.Create(&newVersion).Error; err != nil {
			return nil, err
		}
		version = &newVersion
	}

	// 自动更新 Node 表中的已发布和暂存哈希指纹
	node := model.Node{}
	if err := db.Model(model.Node{}).Where("id = ?", nodeId).First(&node).Error; err == nil {
		node.PublishedSha256 = sha256Hex
		node.DraftSha256 = sha256Hex
		db.Save(&node)
	}

	return version, nil
}

func CalculateNodeDraftSha256(nodeId uint) (string, error) {
	if nodeId == 0 {
		return "", fmt.Errorf("node id is required")
	}

	content, err := renderNodeConfigJson(nodeId)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(content)
	sha256Hex := hex.EncodeToString(sum[:])

	db := database.GetDB()
	node := model.Node{}
	if err := db.Model(model.Node{}).Where("id = ?", nodeId).First(&node).Error; err != nil {
		return "", err
	}

	node.DraftSha256 = sha256Hex
	if err := db.Save(&node).Error; err != nil {
		return "", err
	}

	return sha256Hex, nil
}

func applyInboundPolicyOverrides(
	raw json.RawMessage,
	log map[string]interface{},
	dns map[string]interface{},
	outbounds []map[string]interface{},
	route map[string]interface{},
	experimental map[string]interface{},
) (map[string]interface{}, map[string]interface{}, []map[string]interface{}, map[string]interface{}, map[string]interface{}, error) {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" || text == "{}" {
		return log, dns, outbounds, route, experimental, nil
	}
	payload := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if value, ok := payload["log"]; ok && len(value) > 0 {
		next := map[string]interface{}{}
		if err := json.Unmarshal(value, &next); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("log: %w", err)
		}
		log = next
	}
	if value, ok := payload["dns"]; ok && len(value) > 0 {
		next := map[string]interface{}{}
		if err := json.Unmarshal(value, &next); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("dns: %w", err)
		}
		dns = next
	}
	if value, ok := payload["outbounds"]; ok && len(value) > 0 {
		next := []map[string]interface{}{}
		if err := json.Unmarshal(value, &next); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("outbounds: %w", err)
		}
		outbounds = next
	}
	if value, ok := payload["route"]; ok && len(value) > 0 {
		next := map[string]interface{}{}
		if err := json.Unmarshal(value, &next); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("route: %w", err)
		}
		route = next
	}
	if value, ok := payload["experimental"]; ok && len(value) > 0 {
		next := map[string]interface{}{}
		if err := json.Unmarshal(value, &next); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("experimental: %w", err)
		}
		experimental = next
	}
	return log, dns, outbounds, route, experimental, nil
}

func decodeNodeConfigTemplates(templates *NodeConfigTemplates) (map[string]interface{}, map[string]interface{}, []map[string]interface{}, map[string]interface{}, map[string]interface{}, error) {
	log := map[string]interface{}{}
	if err := json.Unmarshal(templates.Log, &log); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("log template json: %w", err)
	}
	dns := map[string]interface{}{}
	if err := json.Unmarshal(templates.DNS, &dns); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("dns template json: %w", err)
	}
	outbounds := []map[string]interface{}{}
	if err := json.Unmarshal(templates.Outbounds, &outbounds); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("outbounds template json: %w", err)
	}
	route := map[string]interface{}{}
	if err := json.Unmarshal(templates.Route, &route); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("route template json: %w", err)
	}
	experimental := map[string]interface{}{}
	if len(templates.Experimental) > 0 {
		if err := json.Unmarshal(templates.Experimental, &experimental); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("experimental template json: %w", err)
		}
	}
	return log, dns, outbounds, route, experimental, nil
}
