# 开发文档

本文档面向继续开发 Preview 0.1 之后功能的人。

## 项目结构

```text
backend/                 Go 后端
backend/agent/           远端节点 Agent
backend/api/             管理端 API 和 Agent API
backend/database/        数据库初始化和 PostgreSQL/SQLite 适配
backend/database/model/  GORM 模型
backend/service/         分布式节点、证书、配置、订阅等服务
backend/sub/             订阅输出
frontend/                Vue + Vuetify 前端
frontend/src/views/      页面入口
frontend/src/assets/     前端静态资源
docs/                    产品、测试、Agent 和开发文档
scripts/                 Agent 安装脚本
```

## 后端

后端使用 Go + Gin + GORM。

普通 SQLite 兼容构建：

```sh
cd backend
go test ./...
go build -o ../sui main.go
```

PostgreSQL 构建：

```sh
cd backend
go test -tags postgres ./...
go build -tags postgres -o /private/tmp/s-ui-backend-preview-check .
```

新增后端服务时，优先放在 `backend/service/`，API 层只做参数读取、权限判断和响应包装。

## 前端

前端使用 Vue + Vuetify。

开发：

```sh
cd frontend
npm install
npm run dev
```

构建：

```sh
cd frontend
npm run build
```

构建后如果需要让 Go 后端加载最新前端：

```sh
rm -rf backend/web/html/*
cp -R frontend/dist/* backend/web/html/
```

## 本地 Docker

开发优先使用 Docker Compose：

```sh
docker compose -f docker-compose.dev.yml up -d --build s-ui
```

停止：

```sh
docker compose -f docker-compose.dev.yml down
```

清理数据库卷：

```sh
docker compose -f docker-compose.dev.yml down -v
```

本地开发服务：

```text
Web UI:      http://localhost:2095/app/
Sub API:     http://localhost:2096/sub/
PostgreSQL:  localhost:54329
```

## 分布式核心模型

### Node

节点是远端服务器的抽象。

关键字段：

- 名称
- 节点代号
- 地区
- 服务商
- 公开域名
- 公网 IP
- Agent 状态
- sing-box 状态
- 当前配置版本

### Certificate

证书资产由 Control Plane 统一管理。

来源：

- DNS API 自动申请
- acme.sh 托管

手动上传在 Preview 0.1 不作为主路径。

### DNS Provider

DNS Provider 保存 DNS API 凭据。

Preview 0.1 的 DNS Provider 凭据会写入 `credentialsEncrypted`，证书 PEM/私钥会写入 `certificate_versions` 的加密字段，均使用 `SUI_SECRET_KEY` 做 AES-GCM 字段级加密。管理员登录密码已改为 bcrypt 哈希存储，旧明文密码会在首次登录成功后自动迁移。开发和测试时不要提交真实 token；生产部署必须设置长度至少 32 个字符的 `SUI_SECRET_KEY`，并妥善备份。后续稳定版仍需要补齐密钥轮换、外部 Secret 管理和备份加密策略。

不同 provider 的字段不同：

```json
{
  "CF_Token": "Cloudflare API Token"
}
```

```json
{
  "Ali_Key": "Aliyun AccessKey ID",
  "Ali_Secret": "Aliyun AccessKey Secret"
}
```

```json
{
  "AWS_ACCESS_KEY_ID": "AWS Access Key ID",
  "AWS_SECRET_ACCESS_KEY": "AWS Secret Access Key"
}
```

```json
{
  "DP_Id": "DNSPod ID",
  "DP_Key": "DNSPod Key"
}
```

域名不写在 DNS Provider 中，域名只在证书申请中填写。

### Distributed Inbound

服务入口是某个节点上的一个 sing-box inbound。

一个节点可以有多个服务入口，例如：

```text
us-01:
  - anytls: us.example.com:443
  - hysteria2: hy2-us.example.com:8443

sg-01:
  - anytls: sg.example.com:443
  - vless reality: sg.example.com:443
```

入口表单只管理常用字段。特殊字段通过入口级 JSON 覆盖。

### Config Version

发布节点配置时，系统会把全局默认配置、服务入口、节点级策略覆盖和证书路径融合成最终 `config.json`。

如果新生成配置和最新配置的 SHA256 完全一致，系统复用已有版本，不创建无意义的新版本。

## 配置合成规则

最终节点配置：

```text
final config.json = global defaults + node service entrances + node policy overrides
```

全局默认配置包含：

- `log`
- `dns`
- `outbounds`
- `route`
- `experimental`

服务入口自动生成：

- `inbounds`
- TLS 证书路径
- 入口用户
- tag
- listen / listen_port

节点级策略覆盖可用于：

- 节点走 IPv6 优先
- 节点使用特殊 DNS
- 节点使用特殊 outbound
- 单个协议启用 Reality / TLS / ECH 等高级字段

Preview 阶段每个节点只允许一份策略覆盖。多个服务入口共享同一份 `dns/outbounds/route/experimental`，不要把它理解成每个协议入口独立拥有一套全局出口策略。

## Agent 开发

Agent 位于：

```text
backend/agent
```

构建本机 Agent：

```sh
cd backend
go build -o /private/tmp/s-ui-agent-preview-check ./agent
```

构建 Linux x86_64 Agent：

```sh
cd backend
GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64-preview-check ./agent
```

Agent 当前流程：

1. 注册到 Control Plane
2. 定时 heartbeat
3. 获取 desired config
4. 写入证书 bundle
5. 写入 `pending.json`
6. 可选执行 `sing-box check`
7. 提升为 `current.json`
8. 可选执行 reload command
9. 上报配置应用结果

后续重点：

- 上报最近错误
- 上报 CPU / 内存 / 磁盘 / 流量
- 增加 systemd reload 默认策略
- 增加失败回滚

## UI 工作流

涉及 UI 方案讨论时，先参考：

```text
docs/UI_WORKFLOW.md
```

不要用通用 AI 生图替代产品 UI 设计图。需要图示时，使用项目风格的 SVG wireframe 或直接在前端实现后截图。

## 发布前检查

```sh
cd frontend
npm run build

cd ../backend
go test -tags postgres ./...
go build -tags postgres -o /private/tmp/s-ui-backend-preview-check .
GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64-preview-check ./agent

cd ..
sh -n scripts/install-agent.sh
git diff --check
```

发布前检查是否误提交：

```sh
git status --short
```

不应提交：

- `.codegraph/`
- `.local/`
- `mnt/`
- `dns-provider-icons.zip`
- 本地构建二进制
- 真实证书私钥和 API Token
