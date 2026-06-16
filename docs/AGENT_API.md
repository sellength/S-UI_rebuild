# Agent API 草案

Agent API 用于 Data Plane 节点主动连接 Control Plane。该 API 不走管理员登录 session。

当前最小 Agent runtime 见 `docs/AGENT_RUNTIME.md`。

当前挂载路径：

```text
{webPath}/agent
```

如果默认 `webPath=/app/`，则路径为：

```text
/app/agent/register
/app/agent/heartbeat
/app/agent/config/desired
/app/agent/config/report
```

## 认证

注册阶段：

- 必填环境变量：`SUI_AGENT_REGISTER_TOKEN`
- Agent 注册必须提供：
  - Header: `X-Register-Token`

如果控制端没有配置 `SUI_AGENT_REGISTER_TOKEN`，注册接口会拒绝所有 Agent 注册，避免默认部署暴露为可接管状态。

注册完成后：

- Agent 每次请求提供：
  - Header: `X-Agent-Id`
  - Header: `X-Agent-Token`

数据库只保存 `agentToken` 的 SHA-256 hash。

不要把 Agent Token 放到 URL query 或表单字段里，避免被访问日志、代理日志或浏览器历史记录记录。

## POST /register

用途：把 Agent 绑定到已存在的节点。

节点必须先在 Panel 创建，Agent 通过 `nodeCode` 绑定。

字段：

```json
{
  "nodeCode": "us-01",
  "agentId": "agent-us-01",
  "agentToken": "secret",
  "agentVersion": "0.1.0",
  "singboxVersion": "1.13.13",
  "publicKey": ""
}
```

## POST /heartbeat

用途：上报 Agent、sing-box 和当前配置版本状态。

当前 Agent 会尝试通过 `systemctl is-active sing-box`、进程检测和 `sing-box version` 采集基础状态。没有安装 sing-box 或命令不可用时，会上报 `not_installed`、`stopped` 或 `unknown`。

字段：

```json
{
  "configVersion": 1,
  "agentStatus": "online",
  "singboxStatus": "running",
  "agentVersion": "0.1.0",
  "singboxVersion": "1.13.13",
  "resourceSummary": "{\"mem\":128}"
}
```

## GET /config/desired

用途：获取节点期望配置。

返回：

```json
{
  "nodeId": 1,
  "certificates": [
    {
      "id": 1,
      "name": "Wildcard Example.com",
      "fingerprint": "AA:BB:CC",
      "fullchainPem": "-----BEGIN CERTIFICATE-----...",
      "privateKeyPem": "-----BEGIN PRIVATE KEY-----..."
    }
  ],
  "configVersion": {
    "id": 1,
    "version": 1,
    "scope": "node",
    "nodeId": 1,
    "status": "published",
    "contentJson": {}
  }
}
```

如果没有可用配置版本，`configVersion` 为 `null`。

证书 bundle 来自节点启用的 `DistributedInbound.certificateId`，只返回该节点实际需要的证书。

## POST /config/report

用途：上报配置应用结果。

字段：

```json
{
  "configVersionId": 1,
  "status": "applied",
  "errorMessage": ""
}
```

常用状态：

- `pulled`
- `validated`：配置已写入并通过校验，但没有执行 reload/restart，不能视为已应用。
- `applied`
- `failed`
- `rolled_back`

## 管理端相关 action

这些 action 仍然走管理员 session：

```text
GET  {webPath}/api/nodes
POST {webPath}/api/saveNode
POST {webPath}/api/deleteNode

GET  {webPath}/api/dnsProviders
POST {webPath}/api/saveDNSProvider
POST {webPath}/api/deleteDNSProvider

GET  {webPath}/api/certificates
POST {webPath}/api/saveCertificate
POST {webPath}/api/deleteCertificate
GET  {webPath}/api/certificateVersions
POST {webPath}/api/saveCertificateVersion

POST {webPath}/api/renderDistributedAnyTLSInbound
POST {webPath}/api/publishNodeConfig
GET  {webPath}/api/configVersions
GET  {webPath}/api/configDeployments

GET  {webPath}/api/distributedInbounds
POST {webPath}/api/saveDistributedInbound
POST {webPath}/api/deleteDistributedInbound
GET  {webPath}/api/inboundUsers
POST {webPath}/api/saveInboundUser
POST {webPath}/api/deleteInboundUser

GET  {webPath}/api/subscriptions
POST {webPath}/api/createSubscription
```
