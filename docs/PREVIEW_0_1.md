# S-UI Distributed v0.1.0-preview.1

本文档说明 `v0.1.0-preview.1` 的发布范围、启动方式、验证路径和已知限制。

## 定位

`v0.1.0-preview.1` 是分布式 Sing-box 管理平台的早期预览版，不是稳定生产版。

核心目标：

- Control Plane 运行在中心服务器，负责用户、节点、证书、服务入口、配置版本和订阅。
- Data Plane 运行在远端节点，通过 Agent 主动连接 Control Plane。
- Control Plane 生成节点配置，Agent 拉取配置、写入证书并应用 sing-box 配置。

当前版本适合：

- 本地 Docker 功能验证
- 远端 Linux 节点 Agent 联调
- AnyTLS / Hysteria2 / VLESS Reality 等协议配置生成链路验证
- 证书中心 DNS-01 签发流程验证

当前版本不适合：

- 直接作为稳定生产面板暴露到公网
- 承载支付、复杂多租户 SaaS 或 Kubernetes 依赖场景
- 在未完成真实节点验证前承诺可用性

## 已包含模块

| 模块 | 状态 |
|---|---|
| PostgreSQL 开发环境 | 已接入 |
| 分布式节点管理 | 已接入 |
| DNS Provider 管理 | 已接入 |
| 证书中心 | 已接入 |
| 多协议服务入口 | 已接入 |
| 入口级 JSON 覆盖 | 已接入 |
| 全局默认配置预览 | 已接入 |
| 配置版本生成 / 发布 | 已接入 |
| 用户绑定 | 已接入 |
| 分布式订阅 | 已接入 |
| Agent 注册 / 心跳 / 拉取配置 | 已接入 |
| Agent 写入证书和配置 | 已接入 |
| Agent 基础 sing-box 状态和版本采集 | 已接入 |
| Control Plane 外部探测 | 待完善 |
| 支付系统 | 非目标 |
| 复杂多租户 SaaS | 非目标 |
| Kubernetes | 非目标 |

## 本地启动

推荐使用 Docker，避免在本机留下 PostgreSQL 或后端运行时残留。

```sh
docker compose -f docker-compose.dev.yml up -d --build s-ui
```

访问：

```text
http://localhost:2095/app/
```

订阅端口：

```text
http://localhost:2096/sub/
```

如果构建卡在 Docker Hub 镜像 metadata 或 GitHub 下载 acme.sh，通常是网络或镜像源问题，不是业务代码失败。

## 本地验证命令

前端：

```sh
cd frontend
npm run build
```

后端：

```sh
cd backend
GOCACHE=/private/tmp/s-ui-go-cache go test -tags postgres ./...
GOCACHE=/private/tmp/s-ui-go-cache go build -tags postgres -o /private/tmp/s-ui-backend-preview-check .
```

Agent：

```sh
cd backend
GOCACHE=/private/tmp/s-ui-go-cache go build -o /private/tmp/s-ui-agent-preview-check ./agent
GOCACHE=/private/tmp/s-ui-go-cache GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64-preview-check ./agent
```

脚本：

```sh
sh -n scripts/install-control.sh
sh -n scripts/uninstall-control.sh
sh -n scripts/install-control-native.sh
sh -n scripts/uninstall-control-native.sh
sh -n scripts/install-agent.sh
sh -n scripts/uninstall-agent.sh
```

## 推荐功能验证路径

1. 创建节点
   - 名称：`US`
   - 节点代号：`us-01`
   - 公开域名：`us.example.com`

2. 创建 DNS Provider
   - Cloudflare：`CF_Token`
   - AliDNS：`Ali_Key`、`Ali_Secret`
   - AWS：`AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY`
   - Tencent DNSPod：`DP_Id`、`DP_Key`

3. 申请证书
   - 方式：`DNS API 自动申请`
   - 域名示例：`*.example.com, example.com`
   - CA：`letsencrypt`
   - DNS Provider：选择刚创建的凭据

4. 创建服务入口
   - 选择节点
   - 选择协议
   - 选择证书
   - 如需 Reality、IPv6 优先、特殊 DNS / outbounds / route，使用入口级 JSON 覆盖

5. 绑定用户
   - 用户必须具备集群代理权限
   - 一个用户可以绑定到一个或多个服务入口

6. 生成并预览配置
   - 在服务入口列表点击生成配置
   - 点击预览查看最终节点 `config.json`

7. 发布节点配置
   - 在节点概览点击发布
   - 若内容未变化，系统会复用现有配置版本，不应无意义递增版本号

8. 生成订阅
   - 在订阅页为用户生成 token
   - 使用 `distributed-json` 或 `distributed-source` 格式验证输出

9. 部署 Agent 到远端节点
   - 推荐使用 Agent 一键安装脚本
   - 或手动构建 `linux/amd64` Agent 后传入 `--agent-download-url`
   - 验证 Agent 注册、心跳、拉取配置和写入证书

## Agent 部署示例

Control Plane 一键脚本默认会进入交互向导，允许安装者选择面板端口、订阅端口、PostgreSQL 本机端口和公网面板 URL。无人值守部署时可以使用 `--non-interactive` 以及 `--panel-port`、`--sub-port`、`--postgres-port`、`--panel-url` 参数。

安装 systemd 服务并下载最新 Release 里的 Agent：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/install-agent.sh) \
  --panel-agent-url https://panel.example.com/app/agent \
  --node-code us-01 \
  --agent-id agent-us-01 \
  --agent-token 'change-me' \
  --register-token 'same-as-SUI_AGENT_REGISTER_TOKEN'
```

如果 Release 资产还没准备好，可以先手动提供二进制下载地址：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/install-agent.sh) \
  --agent-download-url https://example.com/s-ui-agent-linux-amd64 \
  --panel-agent-url https://panel.example.com/app/agent \
  --node-code us-01 \
  --agent-id agent-us-01 \
  --agent-token 'change-me' \
  --register-token 'same-as-SUI_AGENT_REGISTER_TOKEN'
```

更多 Agent 说明见 [AGENT_RUNTIME.md](AGENT_RUNTIME.md) 和 [AGENT_API.md](AGENT_API.md)。

## 证书说明

DNS Provider 中保存的是 DNS API 凭据，不保存证书域名。

证书域名在申请证书时填写，例如：

```text
*.example.com, example.com
us.example.com, sg.example.com
```

Cloudflare API Token 用于 DNS-01 TXT 记录验证，不等于 Cloudflare Origin Certificate。

AnyTLS / Hysteria2 / VLESS Reality 等远端直连协议建议使用公开信任的 CA 证书，例如 Let's Encrypt。Cloudflare 小云朵代理不等于协议代理，除非明确使用支持 TCP/UDP 代理的产品能力。

## 安全说明

管理员登录密码已改为 bcrypt 哈希存储，并兼容旧明文密码首次登录后自动迁移。

仍需继续硬化：

- DNS Provider 凭据、证书 PEM 和证书私钥会使用 `SUI_SECRET_KEY` 字段级加密后保存。生产部署必须设置长度至少 32 个字符的 `SUI_SECRET_KEY`，并妥善备份；丢失后已保存凭据和证书资产无法解密。
- 后续稳定版仍需要继续引入密钥轮换、外部 Secret 管理和备份加密策略。
- 数据库、备份文件和 Docker volume 必须当作敏感资产处理。
- 不要提交真实 API Token、证书私钥、数据库 dump、SSH key 或账号密码。
- 后续稳定版应引入密钥轮换和/或外部 Secret 管理。

## 已知限制

- Control Plane 容器当前不保证内置可运行的本地 sing-box runtime；分布式节点应由远端 Agent 管理 sing-box。
- `core/sing-box` 在当前仓库中是占位文件，不应理解为生产可用 sing-box 二进制。
- Agent 会尝试采集基础 `sing-box` 运行状态和版本，但资源指标、最近错误和外部可用性探测仍需完善。
- Agent 配置应用依赖 `SUI_AGENT_RELOAD_COMMAND` 或外部 systemd 策略；未配置 reload command 时只会上报 `validated`，不能视为已应用。
- Agent 注册必须配置 `SUI_AGENT_REGISTER_TOKEN`，否则控制端会拒绝注册。
- ACME DNS-01 调用链已接入 acme.sh，但需要真实 DNS Provider、真实域名和公网 CA 做端到端验证。
- Docker 构建需要访问 Docker Hub、Go/NPM registry 和 GitHub。
- UI 仍处于 Preview，尤其是证书中心、节点概览、服务入口高级配置仍会继续调整。

## Preview 发布前检查

```sh
cd frontend && npm run build
cd ../backend && GOCACHE=/private/tmp/s-ui-go-cache go test -tags postgres ./...
cd ..
sh -n scripts/install-control.sh
sh -n scripts/uninstall-control.sh
sh -n scripts/install-control-native.sh
sh -n scripts/uninstall-control-native.sh
sh -n scripts/install-agent.sh
sh -n scripts/uninstall-agent.sh
git diff --check
```

发布前不要提交：

- `.codegraph/`
- `.local/`
- `mnt/`
- `dns-provider-icons.zip`
- 本地构建出的 `sui`、`s-ui-agent`
- 真实 API Token、SSH key、证书私钥
