package sub

import (
	"encoding/json"
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"s-ui/service"
	"testing"
)

func TestDistributedServiceGetDistributedJson(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-distributed-sub-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	client := model.Client{Name: "alice", Enable: true, AccessScope: service.ClientAccessCluster}
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}
	node := model.Node{Name: "US-01", Code: "us-01", PublicHost: "us.example.com", Enable: true}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	inbound := model.DistributedInbound{
		Enable:     true,
		NodeId:     node.Id,
		Protocol:   "anytls",
		ListenPort: 443,
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	inboundUser := model.InboundUser{
		InboundId: inbound.Id,
		ClientId:  client.Id,
		Name:      "alice",
		Password:  "secret",
	}
	if err := db.Create(&inboundUser).Error; err != nil {
		t.Fatalf("create inbound user: %v", err)
	}

	service := DistributedService{}
	result, err := service.GetDistributedJson("alice")
	if err != nil {
		t.Fatalf("GetDistributedJson() error = %v", err)
	}

	config := map[string]interface{}{}
	if err := json.Unmarshal([]byte(*result), &config); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	outbounds, ok := config["outbounds"].([]interface{})
	if !ok || len(outbounds) == 0 {
		t.Fatalf("expected outbounds in result")
	}
}

func TestDistributedServiceResolvesSubscriptionToken(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-distributed-sub-token-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	client := model.Client{Name: "alice", Enable: true, AccessScope: service.ClientAccessCluster}
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}
	node := model.Node{Name: "SG-01", Code: "sg-01", PublicHost: "sg.example.com", Enable: true}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	inbound := model.DistributedInbound{
		Enable:     true,
		NodeId:     node.Id,
		Protocol:   "anytls",
		ListenPort: 443,
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	inboundUser := model.InboundUser{
		InboundId: inbound.Id,
		ClientId:  client.Id,
		Name:      "alice",
		Password:  "secret",
	}
	if err := db.Create(&inboundUser).Error; err != nil {
		t.Fatalf("create inbound user: %v", err)
	}

	subscriptionService := service.SubscriptionService{}
	token, err := subscriptionService.CreateSubscription(client.Id)
	if err != nil {
		t.Fatalf("CreateSubscription() error = %v", err)
	}

	distributedService := DistributedService{}
	result, err := distributedService.GetDistributedJson(token.Token)
	if err != nil {
		t.Fatalf("GetDistributedJson(token) error = %v", err)
	}

	config := struct {
		Outbounds []map[string]interface{} `json:"outbounds"`
	}{}
	if err := json.Unmarshal([]byte(*result), &config); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	found := false
	for _, outbound := range config.Outbounds {
		if outbound["type"] == "anytls" && outbound["server"] == "sg.example.com" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected anytls outbound for sg.example.com, got %+v", config.Outbounds)
	}
}

func TestDistributedServiceIncludesHysteria2Outbound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-distributed-sub-hy2-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	client := model.Client{Name: "bob", Enable: true, AccessScope: service.ClientAccessCluster}
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}
	node := model.Node{Name: "US-01", Code: "us-01", PublicHost: "us.example.com", Enable: true}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	inbound := model.DistributedInbound{
		Enable:             true,
		NodeId:             node.Id,
		Protocol:           "hysteria2",
		PublicHost:         "hy2.example.com",
		ListenPort:         8443,
		RenderedConfigJson: json.RawMessage(`{"type":"hysteria2","tag":"hy2-us-8443","listen":"::","listen_port":8443,"tls":{"enabled":true,"server_name":"hy2.example.com","certificate_path":"/cert/fullchain.pem","key_path":"/cert/privkey.pem"},"users":[{"name":"bob","password":"rendered-secret"}]}`),
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	inboundUser := model.InboundUser{InboundId: inbound.Id, ClientId: client.Id, Name: "bob", Password: "bound-secret"}
	if err := db.Create(&inboundUser).Error; err != nil {
		t.Fatalf("create inbound user: %v", err)
	}

	distributedService := DistributedService{}
	result, err := distributedService.GetDistributedJson("bob")
	if err != nil {
		t.Fatalf("GetDistributedJson() error = %v", err)
	}

	config := struct {
		Outbounds []map[string]interface{} `json:"outbounds"`
	}{}
	if err := json.Unmarshal([]byte(*result), &config); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	found := false
	for _, outbound := range config.Outbounds {
		if outbound["type"] != "hysteria2" {
			continue
		}
		found = true
		if outbound["server"] != "hy2.example.com" {
			t.Fatalf("expected hy2.example.com server, got %+v", outbound)
		}
		if outbound["password"] != "rendered-secret" {
			t.Fatalf("expected rendered user password, got %+v", outbound)
		}
		tls, ok := outbound["tls"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected tls object, got %+v", outbound)
		}
		if tls["certificate_path"] != nil || tls["key_path"] != nil {
			t.Fatalf("client tls should not include server certificate paths, got %+v", tls)
		}
	}
	if !found {
		t.Fatalf("expected hysteria2 outbound, got %+v", config.Outbounds)
	}
}

func TestDistributedServiceIncludesVLESSOutbound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-distributed-sub-vless-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	client := model.Client{Name: "carol", Enable: true, AccessScope: service.ClientAccessCluster}
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}
	node := model.Node{Name: "SG-01", Code: "sg-01", PublicHost: "sg.example.com", Enable: true}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	inbound := model.DistributedInbound{
		Enable:             true,
		NodeId:             node.Id,
		Protocol:           "vless",
		PublicHost:         "vless.example.com",
		ListenPort:         443,
		RenderedConfigJson: json.RawMessage(`{"type":"vless","tag":"vless-sg-443","listen":"::","listen_port":443,"tls":{"enabled":true,"server_name":"vless.example.com"},"transport":{"type":"tcp"},"users":[{"name":"carol","uuid":"11111111-1111-1111-1111-111111111111","flow":"xtls-rprx-vision"}]}`),
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	inboundUser := model.InboundUser{InboundId: inbound.Id, ClientId: client.Id, Name: "carol", Password: "22222222-2222-2222-2222-222222222222"}
	if err := db.Create(&inboundUser).Error; err != nil {
		t.Fatalf("create inbound user: %v", err)
	}

	distributedService := DistributedService{}
	result, err := distributedService.GetDistributedJson("carol")
	if err != nil {
		t.Fatalf("GetDistributedJson() error = %v", err)
	}

	config := struct {
		Outbounds []map[string]interface{} `json:"outbounds"`
	}{}
	if err := json.Unmarshal([]byte(*result), &config); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	found := false
	for _, outbound := range config.Outbounds {
		if outbound["type"] != "vless" {
			continue
		}
		found = true
		if outbound["uuid"] != "11111111-1111-1111-1111-111111111111" {
			t.Fatalf("expected rendered vless uuid, got %+v", outbound)
		}
		if outbound["flow"] != "xtls-rprx-vision" {
			t.Fatalf("expected vless flow, got %+v", outbound)
		}
		if _, ok := outbound["transport"].(map[string]interface{}); !ok {
			t.Fatalf("expected transport object, got %+v", outbound)
		}
	}
	if !found {
		t.Fatalf("expected vless outbound, got %+v", config.Outbounds)
	}
}

func TestDistributedServiceDoesNotExposeRealityPrivateKey(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-distributed-sub-reality-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	db := database.GetDB()
	client := model.Client{Name: "dave", Enable: true, AccessScope: service.ClientAccessCluster}
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}
	node := model.Node{Name: "JP-01", Code: "jp-01", PublicHost: "jp.example.com", Enable: true}
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	inbound := model.DistributedInbound{
		Enable:     true,
		NodeId:     node.Id,
		Protocol:   "vless",
		PublicHost: "reality.example.com",
		ListenPort: 443,
		RenderedConfigJson: json.RawMessage(`{
			"type":"vless",
			"tag":"vless-jp-443",
			"listen":"::",
			"listen_port":443,
			"tls":{
				"enabled":true,
				"server_name":"www.cloudflare.com",
				"reality":{
					"enabled":true,
					"private_key":"server-private-key",
					"short_id":["abcd"]
				}
			},
			"users":[{"name":"dave","uuid":"33333333-3333-3333-3333-333333333333"}]
		}`),
	}
	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	inboundUser := model.InboundUser{InboundId: inbound.Id, ClientId: client.Id, Name: "dave", Password: "33333333-3333-3333-3333-333333333333"}
	if err := db.Create(&inboundUser).Error; err != nil {
		t.Fatalf("create inbound user: %v", err)
	}

	distributedService := DistributedService{}
	result, err := distributedService.GetDistributedJson("dave")
	if err != nil {
		t.Fatalf("GetDistributedJson() error = %v", err)
	}

	config := struct {
		Outbounds []map[string]interface{} `json:"outbounds"`
	}{}
	if err := json.Unmarshal([]byte(*result), &config); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	for _, outbound := range config.Outbounds {
		if outbound["type"] != "vless" {
			continue
		}
		tls, ok := outbound["tls"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected tls object, got %+v", outbound)
		}
		reality, ok := tls["reality"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected reality object, got %+v", tls)
		}
		if reality["private_key"] != nil {
			t.Fatalf("client reality config must not expose private_key, got %+v", reality)
		}
		return
	}
	t.Fatalf("expected vless outbound, got %+v", config.Outbounds)
}
