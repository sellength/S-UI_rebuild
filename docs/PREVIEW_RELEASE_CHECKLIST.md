# Preview 发布检查清单

目标版本：`v0.1.0-preview.1`

发布口径：这是 Preview，不是稳定版、生产版或正式版。

## 1. 文档一致性

- README 标题和状态使用 Preview 口径。
- `docs/PREVIEW_0_1.md` 说明范围、启动、验证路径和限制。
- `docs/TEST_PLAN.md` 的命令和限制与 Preview 口径一致。
- `docs/AGENT_API.md` 和 `docs/AGENT_RUNTIME.md` 能解释 Agent 注册、心跳、配置拉取和部署方式。
- `docs/LOCAL_DOCKER_DEV.md` 能让用户用 Docker 启动本地环境。
- 不把当前版本描述成稳定生产可用。

## 2. 安全检查

- README 和 Preview 文档必须说明真实 API Token、证书私钥、数据库 dump、SSH key 不应提交。
- `.gitignore` 应覆盖 `.codegraph/`、`.local/`、`mnt/`、构建产物和本地密钥类文件。
- 管理员密码使用 bcrypt 哈希，并兼容旧明文迁移。
- Agent 注册 token、agent token 不应通过 query 参数或普通 form 传递。
- DNS Provider 凭据、证书 PEM/私钥已通过 `SUI_SECRET_KEY` 字段级加密；密钥轮换和外部 Secret 管理仍待加固的限制必须写清楚。

## 3. 自动验证

```sh
cd frontend
npm run build
```

```sh
cd backend
GOCACHE=/private/tmp/s-ui-go-cache go test -tags postgres ./...
GOCACHE=/private/tmp/s-ui-go-cache go build -tags postgres -o /private/tmp/s-ui-backend-preview-check .
GOCACHE=/private/tmp/s-ui-go-cache GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64-preview-check ./agent
```

```sh
cd ..
sh -n scripts/install-control.sh
sh -n scripts/uninstall-control.sh
sh -n scripts/install-agent.sh
sh -n scripts/uninstall-agent.sh
git diff --check
```

## 4. 手动 UI 验证

- Docker Compose 能启动 PostgreSQL 和 Control Plane。
- 进入 `/app/distributed` 后，左侧菜单能区分集群管理、用户、系统等入口。
- 证书中心能创建 DNS Provider，并按 provider 展示不同凭据模板。
- 证书申请只要求填写域名、CA、DNS Provider 和必要开关。
- 服务入口能创建 AnyTLS / Hysteria2 / VLESS / VMess / Trojan / Shadowsocks。
- 服务入口支持入口级 JSON 覆盖和最终配置预览。
- 节点概览展示 Agent 通信、sing-box 状态和版本、配置版本、最后心跳。
- 订阅 token 能生成，分布式订阅输出格式可访问。

## 5. 真实节点验证

Preview 发布前可以不完成真实节点全链路，但必须在 Release Notes 中标注为待验证项：

- x86_64 Linux 节点安装 Agent。
- Agent 注册、心跳、拉取配置。
- Agent 写入证书和 `config.json`。
- sing-box reload 或 restart。
- AnyTLS / Hysteria2 / VLESS Reality 等协议真实客户端连接。
- DNS Provider + Let's Encrypt DNS-01 真实签发和续签。

## 6. 发布步骤

- 确认 high review 输出 `Preview OK` 后再发。
- 建议 tag：`v0.1.0-preview.1`
- GitHub Release 标题：`S-UI Distributed v0.1.0-preview.1`
- Release Notes 明确写：不建议生产使用，真实服务器验证仍在进行。
- GitHub 仓库 Secrets 需要配置 `DOCKER_HUB_USERNAME` 和 `DOCKER_HUB_TOKEN`。
- `dev` 分支会推送 `sellength/s-ui_dev:dev` 和 `sellength/s-ui_dev:latest`。
- 如需默认 Compose 可直接拉取 sing-box 核心镜像，手动触发 `Sing-box Docker Image CI`，推送 `sellength/s-ui_rebuild-singbox:latest`。
- GitHub Release 需要包含 `s-ui-agent-linux-amd64` 等 Agent 资产，否则 Agent 一键安装脚本需要传入 `--agent-download-url`。
