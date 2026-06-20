package service

import (
	"encoding/json"
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"time"
)

const (
	NodeLogTemplateProtocol       = "node-log"
	NodeDNSTemplateProtocol       = "node-dns"
	NodeOutboundsTemplateProtocol = "node-outbounds"
	NodeRouteTemplateProtocol     = "node-route"
	NodeExperimentalProtocol      = "node-experimental"
	defaultNodeTemplateName       = "default"
)

type NodeConfigTemplates struct {
	Log          json.RawMessage `json:"log"`
	DNS          json.RawMessage `json:"dns"`
	Outbounds    json.RawMessage `json:"outbounds"`
	Route        json.RawMessage `json:"route"`
	Experimental json.RawMessage `json:"experimental"`
}

type ConfigTemplateService struct {
}

func (s *ConfigTemplateService) GetNodeConfigTemplates() (*NodeConfigTemplates, error) {
	log, err := s.getNodeTemplate(NodeLogTemplateProtocol, defaultNodeLogRaw())
	if err != nil {
		return nil, err
	}
	dns, err := s.getNodeTemplate(NodeDNSTemplateProtocol, defaultNodeDNSRaw())
	if err != nil {
		return nil, err
	}
	outbounds, err := s.getNodeTemplate(NodeOutboundsTemplateProtocol, defaultNodeOutboundsRaw())
	if err != nil {
		return nil, err
	}
	route, err := s.getNodeTemplate(NodeRouteTemplateProtocol, defaultNodeRouteRaw())
	if err != nil {
		return nil, err
	}
	experimental, err := s.getNodeTemplate(NodeExperimentalProtocol, defaultNodeExperimentalRaw())
	if err != nil {
		return nil, err
	}
	return &NodeConfigTemplates{Log: log, DNS: dns, Outbounds: outbounds, Route: route, Experimental: experimental}, nil
}

func (s *ConfigTemplateService) SaveNodeConfigTemplates(templates *NodeConfigTemplates) error {
	if templates == nil {
		return fmt.Errorf("config templates are required")
	}
	if err := validateJSONObject(templates.Log, "log template"); err != nil {
		return err
	}
	if err := validateJSONValue(templates.DNS, "dns template"); err != nil {
		return err
	}
	if err := validateJSONArray(templates.Outbounds, "outbounds template"); err != nil {
		return err
	}
	if err := validateJSONValue(templates.Route, "route template"); err != nil {
		return err
	}
	if err := validateJSONObject(templates.Experimental, "experimental template"); err != nil {
		return err
	}
	if err := s.saveNodeTemplate(NodeLogTemplateProtocol, templates.Log); err != nil {
		return err
	}
	if err := s.saveNodeTemplate(NodeDNSTemplateProtocol, templates.DNS); err != nil {
		return err
	}
	if err := s.saveNodeTemplate(NodeOutboundsTemplateProtocol, templates.Outbounds); err != nil {
		return err
	}
	if err := s.saveNodeTemplate(NodeRouteTemplateProtocol, templates.Route); err != nil {
		return err
	}
	return s.saveNodeTemplate(NodeExperimentalProtocol, templates.Experimental)
}

func (s *ConfigTemplateService) getNodeTemplate(protocol string, fallback json.RawMessage) (json.RawMessage, error) {
	db := database.GetDB()
	template := model.ProtocolTemplate{}
	err := db.Model(model.ProtocolTemplate{}).
		Where("protocol = ? AND name = ? AND enable = ?", protocol, defaultNodeTemplateName, true).
		First(&template).Error
	if database.IsNotFound(err) {
		return fallback, nil
	}
	if err != nil {
		return nil, err
	}
	if len(template.TemplateJson) == 0 {
		return fallback, nil
	}
	return template.TemplateJson, nil
}

func (s *ConfigTemplateService) saveNodeTemplate(protocol string, raw json.RawMessage) error {
	db := database.GetDB()
	template := model.ProtocolTemplate{}
	err := db.Model(model.ProtocolTemplate{}).
		Where("protocol = ? AND name = ?", protocol, defaultNodeTemplateName).
		First(&template).Error
	if err != nil && !database.IsNotFound(err) {
		return err
	}

	now := time.Now().Unix()
	if template.Id == 0 {
		template.Enable = true
		template.Protocol = protocol
		template.Name = defaultNodeTemplateName
		template.Version = "1"
		template.CreatedAt = now
	}
	template.TemplateJson = raw
	template.UpdatedAt = now
	return db.Save(&template).Error
}

func validateJSONValue(raw json.RawMessage, label string) error {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("%s json: %w", label, err)
	}
	return nil
}

func validateJSONObject(raw json.RawMessage, label string) error {
	var value map[string]interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("%s json object: %w", label, err)
	}
	return nil
}

func validateJSONArray(raw json.RawMessage, label string) error {
	var values []interface{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return fmt.Errorf("%s json array: %w", label, err)
	}
	return nil
}

func defaultNodeLogRaw() json.RawMessage {
	return mustCompactJSON(`{
	  "level": "info"
	}`)
}

func defaultNodeDNSRaw() json.RawMessage {
	return mustCompactJSON(`{
	  "servers": [
	    {
	      "type": "tls",
	      "server": "2001:4860:4860::8888",
	      "tag": "google-dns-v6"
	    },
	    {
	      "type": "tls",
	      "server": "8.8.8.8",
	      "tag": "google-dns-v4"
	    }
	  ],
	  "final": "google-dns-v4"
	}`)
}

func defaultNodeOutboundsRaw() json.RawMessage {
	return mustCompactJSON(`[
	  {
	    "type": "direct",
	    "tag": "direct",
	    "domain_resolver": {
	      "server": "google-dns-v4",
	      "strategy": "prefer_ipv4"
	    }
	  }
	]`)
}

func defaultNodeRouteRaw() json.RawMessage {
	return mustCompactJSON(`{
	  "default_domain_resolver": {
	    "server": "google-dns-v4",
	    "strategy": "prefer_ipv4"
	  },
	  "rules": [
	    {
	      "action": "sniff",
	      "timeout": "1s"
	    }
	  ],
	  "final": "direct"
	}`)
}

func defaultNodeExperimentalRaw() json.RawMessage {
	return mustCompactJSON(`{}`)
}

func mustCompactJSON(input string) json.RawMessage {
	var value interface{}
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		panic(err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(strings.TrimSpace(string(data)))
}
