package model

import "encoding/json"

type Setting struct {
	Id    uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Key   string `json:"key" form:"key"`
	Value string `json:"value" form:"value"`
}

type Tls struct {
	Id       uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Name     string          `json:"name" form:"name"`
	Inbounds json.RawMessage `json:"inbounds" form:"inbounds"`
	Server   json.RawMessage `json:"server" form:"server"`
	Client   json.RawMessage `json:"client" form:"client"`
}

type InboundData struct {
	Id      uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Tag     string          `json:"tag" form:"tag"`
	Addrs   json.RawMessage `json:"addrs" form:"addrs"`
	OutJson json.RawMessage `json:"outJson" form:"outJson"`
}

type User struct {
	Id         uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Username   string `json:"username" form:"username"`
	Password   string `json:"password" form:"password"`
	LastLogins string `json:"lastLogin"`
}

type Client struct {
	Id          uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable      bool            `json:"enable" form:"enable"`
	Name        string          `json:"name" form:"name"`
	Config      json.RawMessage `json:"config" form:"config"`
	Inbounds    json.RawMessage `json:"inbounds" form:"inbounds"`
	Links       json.RawMessage `json:"links" form:"links"`
	AccessScope string          `json:"accessScope" form:"accessScope"`
	Volume      int64           `json:"volume" form:"volume"`
	Expiry      int64           `json:"expiry" form:"expiry"`
	Down        int64           `json:"down" form:"down"`
	Up          int64           `json:"up" form:"up"`
	Desc        string          `json:"desc" from:"desc"`
}

type Stats struct {
	Id        uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	DateTime  int64  `json:"dateTime"`
	Resource  string `json:"resource"`
	Tag       string `json:"tag"`
	Direction bool   `json:"direction"`
	Traffic   int64  `json:"traffic"`
}

type Changes struct {
	Id       uint64          `json:"id" gorm:"primaryKey;autoIncrement"`
	DateTime int64           `json:"dateTime"`
	Actor    string          `json:"Actor"`
	Key      string          `json:"key" form:"key"`
	Action   string          `json:"action" form:"action"`
	Index    uint            `json:"index" form:"index"`
	Obj      json.RawMessage `json:"obj" form:"obj"`
}

type Node struct {
	Id            uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable        bool            `json:"enable" form:"enable"`
	Name          string          `json:"name" form:"name"`
	Code          string          `json:"code" form:"code" gorm:"uniqueIndex"`
	Region        string          `json:"region" form:"region"`
	Provider      string          `json:"provider" form:"provider"`
	PublicHost    string          `json:"publicHost" form:"publicHost"`
	PublicIP      string          `json:"publicIp" form:"publicIp"`
	AgentStatus   string          `json:"agentStatus" form:"agentStatus"`
	SingboxStatus string          `json:"singboxStatus" form:"singboxStatus"`
	LastSeenAt    int64           `json:"lastSeenAt" form:"lastSeenAt"`
	Desc          string          `json:"desc" form:"desc"`
	Metadata      json.RawMessage `json:"metadata" form:"metadata"`
	CreatedAt     int64           `json:"createdAt" form:"createdAt"`
	UpdatedAt     int64           `json:"updatedAt" form:"updatedAt"`
	DraftSha256     string          `json:"draftSha256" form:"draftSha256" gorm:"type:varchar(64)"`
	PublishedSha256 string          `json:"publishedSha256" form:"publishedSha256" gorm:"type:varchar(64)"`
	AppliedSha256   string          `json:"appliedSha256" form:"appliedSha256" gorm:"type:varchar(64)"`
}

type NodeAgent struct {
	Id                uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	NodeId            uint            `json:"nodeId" form:"nodeId" gorm:"index"`
	AgentId           string          `json:"agentId" form:"agentId" gorm:"uniqueIndex"`
	AgentVersion      string          `json:"agentVersion" form:"agentVersion"`
	SingboxVersion    string          `json:"singboxVersion" form:"singboxVersion"`
	TokenHash         string          `json:"tokenHash" form:"tokenHash"`
	PublicKey         string          `json:"publicKey" form:"publicKey"`
	LastConfigVersion uint64          `json:"lastConfigVersion" form:"lastConfigVersion"`
	Capabilities      json.RawMessage `json:"capabilities" form:"capabilities"`
	RegisteredAt      int64           `json:"registeredAt" form:"registeredAt"`
	LastSeenAt        int64           `json:"lastSeenAt" form:"lastSeenAt"`
}

type NodeGroup struct {
	Id        uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Name      string `json:"name" form:"name"`
	Code      string `json:"code" form:"code" gorm:"uniqueIndex"`
	Desc      string `json:"desc" form:"desc"`
	CreatedAt int64  `json:"createdAt" form:"createdAt"`
	UpdatedAt int64  `json:"updatedAt" form:"updatedAt"`
}

type NodeGroupMember struct {
	Id        uint  `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	GroupId   uint  `json:"groupId" form:"groupId" gorm:"index"`
	NodeId    uint  `json:"nodeId" form:"nodeId" gorm:"index"`
	CreatedAt int64 `json:"createdAt" form:"createdAt"`
}

type DNSProvider struct {
	Id                   uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable               bool            `json:"enable" form:"enable"`
	Name                 string          `json:"name" form:"name"`
	Type                 string          `json:"type" form:"type"`
	CredentialsEncrypted string          `json:"credentialsEncrypted" form:"credentialsEncrypted"`
	Config               json.RawMessage `json:"config" form:"config"`
	CreatedAt            int64           `json:"createdAt" form:"createdAt"`
	UpdatedAt            int64           `json:"updatedAt" form:"updatedAt"`
}

type Certificate struct {
	Id              uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable          bool            `json:"enable" form:"enable"`
	Name            string          `json:"name" form:"name"`
	Source          string          `json:"source" form:"source"`
	Domains         json.RawMessage `json:"domains" form:"domains"`
	Config          json.RawMessage `json:"config" form:"config"`
	Wildcard        bool            `json:"wildcard" form:"wildcard"`
	AutoRenew       bool            `json:"autoRenew" form:"autoRenew"`
	DNSProviderId   uint            `json:"dnsProviderId" form:"dnsProviderId" gorm:"index"`
	ActiveVersionId uint            `json:"activeVersionId" form:"activeVersionId"`
	CreatedAt       int64           `json:"createdAt" form:"createdAt"`
	UpdatedAt       int64           `json:"updatedAt" form:"updatedAt"`
}

type CertificateVersion struct {
	Id                     uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	CertificateId          uint   `json:"certificateId" form:"certificateId" gorm:"index"`
	Fingerprint            string `json:"fingerprint" form:"fingerprint" gorm:"index"`
	FullchainPEMEncrypted  string `json:"fullchainPemEncrypted" form:"fullchainPemEncrypted"`
	PrivateKeyPEMEncrypted string `json:"privateKeyPemEncrypted" form:"privateKeyPemEncrypted"`
	NotBefore              int64  `json:"notBefore" form:"notBefore"`
	NotAfter               int64  `json:"notAfter" form:"notAfter"`
	CreatedAt              int64  `json:"createdAt" form:"createdAt"`
}

type ProtocolTemplate struct {
	Id            uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable        bool            `json:"enable" form:"enable"`
	Protocol      string          `json:"protocol" form:"protocol" gorm:"index"`
	Name          string          `json:"name" form:"name"`
	Version       string          `json:"version" form:"version"`
	TemplateJson  json.RawMessage `json:"templateJson" form:"templateJson"`
	SchemaJson    json.RawMessage `json:"schemaJson" form:"schemaJson"`
	DefaultValues json.RawMessage `json:"defaultValues" form:"defaultValues"`
	CreatedAt     int64           `json:"createdAt" form:"createdAt"`
	UpdatedAt     int64           `json:"updatedAt" form:"updatedAt"`
}

type DistributedInbound struct {
	Id                    uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable                bool            `json:"enable" form:"enable"`
	NodeId                uint            `json:"nodeId" form:"nodeId" gorm:"index"`
	Protocol              string          `json:"protocol" form:"protocol" gorm:"index"`
	Tag                   string          `json:"tag" form:"tag"`
	PublicHost            string          `json:"publicHost" form:"publicHost"`
	Listen                string          `json:"listen" form:"listen"`
	ListenPort            uint            `json:"listenPort" form:"listenPort"`
	TemplateId            uint            `json:"templateId" form:"templateId"`
	TlsProfileId          uint            `json:"tlsProfileId" form:"tlsProfileId"`
	CertificateId         uint            `json:"certificateId" form:"certificateId"`
	FormValuesJson        json.RawMessage `json:"formValuesJson" form:"formValuesJson"`
	AdvancedOverridesJson json.RawMessage `json:"advancedOverridesJson" form:"advancedOverridesJson"`
	PolicyOverridesJson   json.RawMessage `json:"policyOverridesJson" form:"policyOverridesJson"`
	RenderedConfigJson    json.RawMessage `json:"renderedConfigJson" form:"renderedConfigJson"`
	CreatedAt             int64           `json:"createdAt" form:"createdAt"`
	UpdatedAt             int64           `json:"updatedAt" form:"updatedAt"`
}

type InboundUser struct {
	Id        uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	InboundId uint   `json:"inboundId" form:"inboundId" gorm:"index"`
	ClientId  uint   `json:"clientId" form:"clientId" gorm:"index"`
	Name      string `json:"name" form:"name"`
	Password  string `json:"password" form:"password"`
	CreatedAt int64  `json:"createdAt" form:"createdAt"`
	UpdatedAt int64  `json:"updatedAt" form:"updatedAt"`
}

type ConfigVersion struct {
	Id          uint64          `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Version     uint64          `json:"version" form:"version" gorm:"index"`
	Scope       string          `json:"scope" form:"scope" gorm:"index"`
	NodeId      uint            `json:"nodeId" form:"nodeId" gorm:"index"`
	Sha256      string          `json:"sha256" form:"sha256"`
	Status      string          `json:"status" form:"status" gorm:"index"`
	ContentJson json.RawMessage `json:"contentJson" form:"contentJson"`
	CreatedBy   string          `json:"createdBy" form:"createdBy"`
	CreatedAt   int64           `json:"createdAt" form:"createdAt"`
}

type ConfigDeployment struct {
	Id              uint64 `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	ConfigVersionId uint64 `json:"configVersionId" form:"configVersionId" gorm:"index"`
	NodeId          uint   `json:"nodeId" form:"nodeId" gorm:"index"`
	Status          string `json:"status" form:"status" gorm:"index"`
	ErrorMessage    string `json:"errorMessage" form:"errorMessage"`
	StartedAt       int64  `json:"startedAt" form:"startedAt"`
	FinishedAt      int64  `json:"finishedAt" form:"finishedAt"`
}

type NodeHeartbeat struct {
	Id              uint64          `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	NodeId          uint            `json:"nodeId" form:"nodeId" gorm:"index"`
	AgentId         string          `json:"agentId" form:"agentId"`
	ConfigVersion   uint64          `json:"configVersion" form:"configVersion"`
	AgentStatus     string          `json:"agentStatus" form:"agentStatus"`
	SingboxStatus   string          `json:"singboxStatus" form:"singboxStatus"`
	ResourceSummary json.RawMessage `json:"resourceSummary" form:"resourceSummary"`
	CreatedAt       int64           `json:"createdAt" form:"createdAt" gorm:"index"`
}

type NodeMetric struct {
	Id        uint64          `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	NodeId    uint            `json:"nodeId" form:"nodeId" gorm:"index"`
	DateTime  int64           `json:"dateTime" form:"dateTime" gorm:"index"`
	Metrics   json.RawMessage `json:"metrics" form:"metrics"`
	CreatedAt int64           `json:"createdAt" form:"createdAt"`
}

type Subscription struct {
	Id         uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable     bool            `json:"enable" form:"enable"`
	ClientId   uint            `json:"clientId" form:"clientId" gorm:"index"`
	TokenHash  string          `json:"tokenHash" form:"tokenHash" gorm:"uniqueIndex"`
	Policy     json.RawMessage `json:"policy" form:"policy"`
	LastUsedAt int64           `json:"lastUsedAt" form:"lastUsedAt"`
	CreatedAt  int64           `json:"createdAt" form:"createdAt"`
	UpdatedAt  int64           `json:"updatedAt" form:"updatedAt"`
}

type SubStoreIntegration struct {
	Id        uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable    bool            `json:"enable" form:"enable"`
	Name      string          `json:"name" form:"name"`
	BaseURL   string          `json:"baseUrl" form:"baseUrl"`
	Config    json.RawMessage `json:"config" form:"config"`
	CreatedAt int64           `json:"createdAt" form:"createdAt"`
	UpdatedAt int64           `json:"updatedAt" form:"updatedAt"`
}

type AuditLog struct {
	Id         uint64          `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	DateTime   int64           `json:"dateTime" form:"dateTime" gorm:"index"`
	Actor      string          `json:"actor" form:"actor"`
	Resource   string          `json:"resource" form:"resource"`
	ResourceId string          `json:"resourceId" form:"resourceId"`
	Action     string          `json:"action" form:"action"`
	Detail     json.RawMessage `json:"detail" form:"detail"`
}
