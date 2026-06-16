package service

import (
	"encoding/json"
	"strings"
)

func redactedDNSProviderConfig(providerType string) json.RawMessage {
	keys := []string{}
	switch strings.ToLower(strings.TrimSpace(providerType)) {
	case "cloudflare":
		keys = []string{"CF_Token"}
	case "aliyun":
		keys = []string{"Ali_Key", "Ali_Secret"}
	case "route53":
		keys = []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}
	case "dnspod":
		keys = []string{"DP_Id", "DP_Key"}
	default:
		keys = []string{}
	}
	values := map[string]string{}
	for _, key := range keys {
		values[key] = dnsCredentialKeepEncrypted
	}
	raw, _ := json.Marshal(values)
	return raw
}

func containsKeepEncryptedPlaceholder(raw json.RawMessage) bool {
	values := map[string]interface{}{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return false
	}
	for _, value := range values {
		if text, ok := value.(string); ok && strings.TrimSpace(text) == dnsCredentialKeepEncrypted {
			return true
		}
	}
	return false
}
