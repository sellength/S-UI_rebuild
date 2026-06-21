# High Review 指南

目标：检查当前项目是否可以作为 `v0.2.0-preview.1` 发布到 GitHub。请按 Preview 标准审查，不要按稳定生产版标准要求一次性补齐所有后续能力。

最终请给出明确结论：

```text
Preview OK
```

或：

```text
Preview Not OK
```

并按 P0 / P1 / P2 输出问题。

## 必查文档

- `README.md`
- `docs/PREVIEW_0_2.md`
- `docs/PREVIEW_RELEASE_CHECKLIST.md`
- `docs/TEST_PLAN.md`
- `docs/AGENT_API.md`
- `docs/AGENT_RUNTIME.md`
- `docs/LOCAL_DOCKER_DEV.md`
- `docs/DISTRIBUTED_PRD.md`
- `docs/DEVELOPMENT.md`

## 必查代码

- `backend/api/agent.go`
- `backend/agent/main.go`
- `backend/service/agent.go`
- `backend/service/config_bundle.go`
- `backend/service/config_version.go`
- `backend/service/certificate.go`
- `backend/service/certificate_issue.go`
- `backend/service/distributed_inbound.go`
- `backend/service/subscription.go`
- `backend/service/user.go`
- `backend/sub/distributedService.go`
- `scripts/install-agent.sh`
- `Dockerfile.dev`
- `docker-compose.dev.yml`
- `frontend/src/views/Distributed.vue`
- `frontend/src/layouts/default/Drawer.vue`
- `frontend/src/layouts/modals/Client.vue`

## 审查重点

1. 发布口径
   - 文档是否统一使用 Preview，而不是稳定版或生产版。
   - 是否明确真实节点、真实 DNS Provider、真实客户端连接仍需端到端验证。

2. 安全边界
   - 管理员密码是否 bcrypt 存储。
   - 旧明文密码迁移是否合理。
   - Agent token / register token 是否避免 query 参数和普通 form。
   - DNS Provider 凭据、证书 PEM/私钥是否通过 `SUI_SECRET_KEY` 字段级加密，密钥轮换和外部 Secret 管理仍待加固的限制是否写清楚。
   - `.gitignore` 是否覆盖明显敏感文件和本地构建产物。

3. Agent 主链路
   - 注册、心跳、拉取 desired config、上报配置状态是否闭环。
   - Agent 是否能写入证书和配置。
   - Agent 状态上报是否包含基础 sing-box 状态和版本。
   - systemd 环境变量写入是否做了必要转义。

4. 配置生成
   - 全局默认配置、服务入口、入口级 JSON 覆盖的合并关系是否清楚。
   - 生成配置是否避免无意义版本递增。
   - 同一节点启用入口是否能检查监听地址和端口冲突。
   - VLESS Reality 等协议是否可以通过协议高级 JSON 表达。

5. 证书中心
   - DNS Provider 是否按不同服务商展示不同凭据模板。
   - Cloudflare Token / AliDNS / AWS / DNSPod 的字段说明是否不误导用户。
   - 证书申请域名是否只在证书申请处填写，不混到 Provider 凭据里。
   - acme.sh 路径和 Docker volume 是否能持久化。

6. 订阅与用户
   - 用户是否能区分本地代理权限和集群代理权限。
   - 分布式订阅是否只输出该用户有权限的集群入口。
   - token 生成和格式输出是否可测试。

7. UI 体验
   - 集群管理是否不再把所有表单一次性堆在首屏。
   - 服务入口列表按钮文案是否明确：编辑、配置设置、预览、生成配置。
   - 节点概览是否优先展示 Agent 通信、sing-box 状态与版本、配置版本、最后心跳。
   - 证书中心是否比旧表单堆叠更清晰。

## 建议验证命令

```sh
cd frontend
npm run build
```

```sh
cd backend
go test -tags postgres ./...
go build -tags postgres -o /private/tmp/s-ui-backend-preview-check .
GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64-preview-check ./agent
```

```sh
cd ..
sh -n scripts/install-agent.sh
git diff --check
```

## P0 / P1 / P2 定义

- P0：会导致 Preview 发布后误导用户、泄露敏感信息、核心链路完全不可用，必须发布前修复。
- P1：影响主要功能验证或真实节点联调，建议发布前修复。
- P2：体验、文案、结构或后续硬化问题，可以进入 Preview 已知限制。
