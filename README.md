# S-UI Distributed Preview

S-UI Distributed 是基于 [alireza0/s-ui](https://github.com/alireza0/s-ui) 重构的 Sing-box 分布式管理面板实验版本。它的目标不是只管理本机 sing-box，而是把一台服务器作为 Control Plane，把多台远端 VPS 作为 Data Plane，通过 Agent 自动同步配置、证书和节点状态。

> 免责声明：本项目仅供个人学习与技术交流使用，请勿用于非法用途。当前版本是 Preview，不是稳定生产版。

## 当前版本

目标版本：`v0.1.0-preview.1`

这个版本主要验证下面这条主链路：

```text
Control Plane 面板 -> 远端 Agent -> sing-box 配置文件 -> 客户端订阅
```

已包含：

- PostgreSQL 开发环境
- 集群节点管理
- Agent 注册、心跳、配置拉取和基础状态上报
- DNS Provider 管理
- 证书中心
- 多协议服务入口
- 入口级 JSON 策略覆盖
- 全局默认配置预览
- 配置版本生成与发布
- 用户绑定与分布式订阅
- 证书 PEM、证书私钥和 DNS Provider 凭据字段级加密

仍处于 Preview：

- 真实服务器端到端联调仍需要继续验证
- Control Plane 外部探测仍未完成
- Docker 多架构镜像发布流程仍需完善
- 证书签发和续签需要真实 DNS Provider 验证
- Agent 资源监控、失败回滚、日志摘要仍是后续增强项

详细状态见 [docs/PREVIEW_0_1.md](docs/PREVIEW_0_1.md)。

## 架构

```text
                                  +----------------------+
                                  |      Sub-Store        |
                                  |   可选：订阅整合       |
                                  +----------^-----------+
                                             |
+-----------------------+        +-----------+-----------+
|  Admin Browser        |        |    S-UI Control Plane |
|  管理员浏览器          +------->+    用户 / 节点 / 证书   |
+-----------------------+        |    配置 / 订阅 / API    |
                                 +-----+-------------+---+
                                       |             |
                       Agent heartbeat |             | subscription
                       config pull     |             |
                                       v             v
                         +-------------+--+     +----+----------------+
                         | Data Plane Node |     | Client Apps         |
                         | s-ui-agent      |     | sing-box / Clash 等 |
                         | sing-box        |     +---------------------+
                         +-----------------+
```

Control Plane 负责：

- 管理用户、节点、证书、服务入口和配置版本
- 生成 sing-box 配置
- 生成订阅
- 接收 Agent 心跳
- 下发证书和配置

Data Plane 负责：

- 运行 `s-ui-agent`
- 运行 `sing-box`
- 主动向 Control Plane 注册和心跳
- 拉取配置、写入证书、校验并应用配置
- 上报 sing-box 状态、版本和当前配置版本

## 目录说明

```text
backend/                 Go 后端、API、订阅、Agent
frontend/                Vue 前端
scripts/install-agent.sh Agent systemd 安装脚本
docker-compose.dev.yml   本地开发 compose，包含 PostgreSQL
docker-compose.yml       服务器部署 compose 示例
docs/                    设计、开发、测试和 Preview 文档
```

## 部署前准备

服务端建议：

- Linux x86_64 VPS
- 2 核 CPU / 2GB 内存以上，推荐 4GB 内存
- Docker 24+ 和 Docker Compose v2
- 已解析到面板服务器的域名
- 如果要签发证书，需要可用 DNS Provider API Token

节点端建议：

- Linux x86_64 VPS
- systemd
- sing-box 已安装，或后续由你的部署脚本安装
- 能访问 Control Plane 的 Agent API
- 节点域名已经解析到节点公网 IP

端口：

- `2095`：面板 Web UI 默认端口，安装脚本可交互修改
- `2096`：订阅服务默认端口，安装脚本可交互修改
- `443` 或自定义端口：节点上的代理协议入口
- Agent 默认不需要公网入站端口，它主动连接 Control Plane

安全变量：

- `SUI_SECRET_KEY`：用于加密 DNS Provider 凭据、证书 PEM 和证书私钥。生产必须设置，建议 32 字符以上，并妥善备份。
- `SUI_AGENT_REGISTER_TOKEN`：Agent 注册令牌。控制端和 Agent 安装参数必须一致。

## 快速本地启动

本地开发推荐使用 `docker-compose.dev.yml`，它会启动 PostgreSQL 和 Control Plane。

```sh
docker compose -f docker-compose.dev.yml up -d --build s-ui
```

访问：

```text
http://localhost:2095/app/
```

订阅服务：

```text
http://localhost:2096/sub/
```

停止：

```sh
docker compose -f docker-compose.dev.yml down
```

清理本地开发数据库：

```sh
docker compose -f docker-compose.dev.yml down -v
```

如果 Docker 构建卡在基础镜像 metadata、Go/NPM registry 或 acme.sh 下载，通常是网络问题。可以稍后重试，或先使用本地测试命令验证代码。

## 服务端部署

### 方式一：一键脚本

在服务器上执行：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/install-control.sh)
```

脚本会进入交互向导，让你确认：

- 面板 Web UI 端口，默认 `2095`
- 订阅服务端口，默认 `2096`
- PostgreSQL 本机端口，默认 `54329`，只绑定 `127.0.0.1`
- 面板公网访问地址，例如 `https://panel.example.com` 或 `http://服务器IP:2095`

默认安装到：

```text
/opt/s-ui-distributed
```

脚本会自动完成：

- 检查 Docker 和 Docker Compose
- 下载 `docker-compose.yml`
- 下载 `.env.example` 并生成 `.env`
- 自动生成 `SUI_POSTGRES_PASSWORD`
- 自动生成 `SUI_SECRET_KEY`
- 自动生成 `SUI_AGENT_REGISTER_TOKEN`
- 写入面板端口和订阅端口
- 创建持久化目录
- 启动 Control Plane

如果你要无人值守安装，可以用参数跳过交互：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/install-control.sh) \
  --non-interactive \
  --panel-port 2095 \
  --sub-port 2096 \
  --postgres-port 54329 \
  --panel-url https://panel.example.com
```

也可以通过环境变量指定：

```sh
SUI_PANEL_PORT=2095 \
SUI_SUB_PORT=2096 \
SUI_POSTGRES_PORT=54329 \
SUI_PANEL_DOMAIN=https://panel.example.com \
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/install-control.sh) --non-interactive
```

安装后查看：

```sh
cd /opt/s-ui-distributed
docker compose --env-file .env ps
docker compose --env-file .env logs -f s-ui
```

卸载但保留数据：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/uninstall-control.sh)
```

卸载并删除数据：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/uninstall-control.sh) --purge
```

如果你的系统默认没有 `bash`，可以先安装 `bash`，或下载脚本后用 `sh` 执行。

### 方式二：Docker Compose

当前 Preview 推荐用 Docker Compose 部署 Control Plane。服务器上准备目录：

```sh
mkdir -p /opt/s-ui-distributed
cd /opt/s-ui-distributed
```

下载或复制仓库中的：

```text
docker-compose.yml
.env.example
```

创建生产环境变量：

```sh
cp .env.example .env
```

编辑 `.env`：

```text
TZ=Asia/Shanghai
SUI_POSTGRES_PORT=54329
SUI_PANEL_PORT=2095
SUI_SUB_PORT=2096
SUI_POSTGRES_PASSWORD=replace-with-a-long-random-postgres-password
SUI_SECRET_KEY=replace-with-a-long-random-secret-at-least-32-chars
SUI_AGENT_REGISTER_TOKEN=replace-with-a-long-random-agent-register-token
SUI_PANEL_DOMAIN=https://panel.example.com
```

启动：

```sh
docker compose --env-file .env up -d
```

查看状态：

```sh
docker compose ps
docker compose logs -f s-ui
```

访问：

```text
http://你的服务器IP:你的面板端口/app/
```

如果你放在反向代理后面，建议使用 HTTPS，并把面板域名配置成：

```text
https://panel.example.com
```

停止：

```sh
docker compose down
```

升级：

```sh
docker compose pull
docker compose up -d
```

备份至少包含：

```text
./db
./cert
./certificates
./acme
.env
```

其中 `.env` 里的 `SUI_SECRET_KEY` 必须备份。丢失后，已保存的 DNS API 凭据和证书私钥无法解密。`./db` 是 PostgreSQL 数据目录，备份和迁移时不要只复制单个数据库文件。

### 方式三：旧服务端脚本说明

仓库根目录保留了原 S-UI 的 `install.sh`、`s-ui.sh`、`docker-run.sh` 等脚本，但它们主要面向旧的单机 S-UI 流程，不建议作为分布式 Preview 的服务器安装入口。

当前 Preview 的服务端主路径是 `scripts/install-control.sh` 或手动 Docker Compose。

## Agent 部署

Agent 安装在每台远端节点上。它主动连接 Control Plane，不需要在节点上额外开放 Agent 入站端口。

### 1. 一键安装 Agent

节点上执行：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/install-agent.sh) \
  --panel-agent-url https://panel.example.com/app/agent \
  --node-code us-01 \
  --agent-id agent-us-01 \
  --agent-token 'replace-with-node-agent-token' \
  --register-token 'same-as-SUI_AGENT_REGISTER_TOKEN' \
  --interval 30s \
  --reload-command 'systemctl restart sing-box'
```

脚本会自动完成：

- 下载或复用 `/usr/local/s-ui-agent/s-ui-agent`
- 写入 `/usr/local/s-ui-agent/agent.env`
- 创建配置、证书和日志目录
- 创建并启动 `s-ui-agent.service`

默认下载地址是 GitHub 最新 Release：

```text
https://github.com/sellength/S-UI_rebuild/releases/latest/download/s-ui-agent-linux-amd64
```

如果你还没有发布 Release，可以先手动上传二进制，或指定下载地址：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/install-agent.sh) \
  --agent-download-url https://example.com/s-ui-agent-linux-amd64 \
  --panel-agent-url https://panel.example.com/app/agent \
  --node-code us-01 \
  --agent-id agent-us-01 \
  --agent-token 'replace-with-node-agent-token' \
  --register-token 'same-as-SUI_AGENT_REGISTER_TOKEN'
```

卸载 Agent：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/uninstall-agent.sh)
```

卸载但保留 Agent 数据：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/uninstall-agent.sh) --keep-data
```

### 2. 手动构建 Linux x86_64 Agent

在开发机或 CI 上构建：

```sh
cd backend
GOCACHE=/private/tmp/s-ui-go-cache GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64 ./agent
```

如果已经有发布产物，可以直接下载对应平台的 `s-ui-agent`。

### 3. 上传到节点

把二进制上传到节点：

```text
/usr/local/s-ui-agent/s-ui-agent
```

并设置可执行权限：

```sh
sudo mkdir -p /usr/local/s-ui-agent
sudo cp s-ui-agent-linux-amd64 /usr/local/s-ui-agent/s-ui-agent
sudo chmod +x /usr/local/s-ui-agent/s-ui-agent
```

### 4. 用脚本安装 systemd 服务

如果你不想用 `curl | bash`，也可以把仓库中的 `scripts/install-agent.sh` 上传到节点，然后执行：

```sh
sudo sh scripts/install-agent.sh \
  --panel-agent-url https://panel.example.com/app/agent \
  --node-code us-01 \
  --agent-id agent-us-01 \
  --agent-token 'replace-with-node-agent-token' \
  --register-token 'same-as-SUI_AGENT_REGISTER_TOKEN' \
  --interval 30s \
  --reload-command 'systemctl restart sing-box'
```

参数说明：

- `--panel-agent-url`：Control Plane 的 Agent API 地址，通常是 `https://面板域名/app/agent`
- `--node-code`：面板里创建的节点代号，例如 `us-01`
- `--agent-id`：当前 Agent 的唯一 ID，例如 `agent-us-01`
- `--agent-token`：当前节点自己的 Agent Token
- `--register-token`：控制端的 `SUI_AGENT_REGISTER_TOKEN`
- `--interval`：心跳和配置检查间隔
- `--reload-command`：配置应用后执行的 reload/restart 命令

脚本会创建：

```text
/usr/local/s-ui-agent/agent.env
/usr/local/s-ui-agent/configs/
/usr/local/s-ui-agent/certs/
/usr/local/s-ui-agent/state.json
/etc/systemd/system/s-ui-agent.service
```

常用命令：

```sh
sudo systemctl status s-ui-agent
sudo journalctl -u s-ui-agent -f
sudo systemctl restart s-ui-agent
```

更多细节见 [docs/AGENT_RUNTIME.md](docs/AGENT_RUNTIME.md) 和 [docs/AGENT_API.md](docs/AGENT_API.md)。

## 使用流程

### 1. 登录面板

启动 Control Plane 后访问：

```text
http://你的服务器IP:2095/app/
```

首次使用建议立即修改管理员账号和密码。

### 2. 创建节点

进入“集群管理”，创建节点：

- 名称：例如 `us`
- 节点代号：例如 `us-01`
- 地区：例如 `US`
- 服务商：例如 `AWS`
- 公开域名：例如 `us.example.com`
- 公网 IP：节点真实公网 IP

节点代号要和 Agent 安装脚本中的 `--node-code` 保持一致。

### 3. 配置 DNS Provider

进入“集群管理 -> 证书中心”，添加 DNS Provider。

支持模板：

- Cloudflare：`CF_Token`
- AliDNS：`Ali_Key`、`Ali_Secret`
- AWS：`AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY`
- Tencent DNSPod：`DP_Id`、`DP_Key`

Cloudflare 推荐使用只包含 Zone DNS Edit 权限的 API Token。DNS API 只用于 Let's Encrypt DNS-01 TXT 验证，不要求开启 Cloudflare 小云朵代理。

### 4. 申请证书

在“证书中心”申请证书：

```text
*.example.com, example.com
us.example.com, sg.example.com
```

说明：

- 通配符由域名输入决定，填写 `*.example.com` 即表示申请通配符证书。
- 手动上传证书不负责自动续期，Preview 版本默认不作为主路径。
- AnyTLS、Hysteria2、Trojan、VLESS Reality 等入口可引用证书中心的证书资产。

### 5. 创建服务入口

进入“集群管理 -> 服务入口”，创建协议入口：

- 选择节点
- 选择协议
- 填写公开域名
- 填写监听端口
- 选择证书
- 如需特殊参数，在入口级 JSON 中覆盖

当前集群入口支持：

- AnyTLS
- Hysteria2
- TUIC
- Trojan
- VLESS
- VMess
- Shadowsocks

VLESS + Reality 这类组合应通过 VLESS 入口和协议高级 JSON 配置完成。

### 6. 配置全局默认策略

“全局配置”用于查看所有节点默认继承的运行骨架：

- `log`
- `dns`
- `outbounds`
- `route`
- `experimental`

最终配置生成规则：

```text
最终 config.json = 全局默认配置 + 当前节点服务入口 + 入口级策略覆盖
```

如果某个协议需要 IPv6 优先、特殊 DNS 或特殊 outbound，建议放到该入口的策略覆盖中，而不是修改全局默认配置。

### 7. 绑定用户

进入“用户管理”创建用户，并选择使用范围：

- 本地代理
- 集群代理

集群入口里把用户绑定到服务入口后，该用户的订阅会包含对应节点和协议。

### 8. 发布配置

在“集群管理 -> 概览”点击发布配置。发布后：

- Control Plane 生成节点配置版本
- Agent 心跳时拉取 desired config
- Agent 写入证书和配置文件
- Agent 可选执行 `sing-box check`
- Agent 可选执行 reload/restart
- 面板显示 Agent 通信、sing-box 状态、版本、配置版本和最后心跳

### 9. 获取订阅

订阅服务默认运行在：

```text
http://面板域名或IP:2096/sub/
```

用户生成订阅 token 后，可以用订阅链接导入客户端，也可以交给已经部署好的 Sub-Store 做二次整理和格式转换。

## 证书和密钥安全

Preview 0.1 已经做到：

- 管理员密码使用 bcrypt 哈希保存
- DNS Provider 凭据使用 `SUI_SECRET_KEY` 字段级加密
- 证书 PEM 和私钥使用 `SUI_SECRET_KEY` 字段级加密
- API 列表不会直接返回证书私钥明文
- Agent 只在拉取配置时拿到需要下发的证书 bundle

生产注意事项：

- `SUI_SECRET_KEY` 不要使用默认值
- `SUI_SECRET_KEY` 必须备份
- 不要提交 `.env`、真实 API Token、证书私钥、数据库 dump、SSH key
- DNS API Token 尽量使用最小权限
- 面板建议放在 HTTPS 反向代理后
- Agent 注册令牌和 Agent Token 应使用高强度随机值

仍需后续增强：

- 密钥轮换
- 外部 Secret Manager
- 数据库备份加密
- 更细的审计日志

## Sub-Store 集成

本项目不重写 Sub-Store。推荐方式：

1. S-UI Distributed 负责生成用户级订阅。
2. Sub-Store 读取 S-UI 订阅链接。
3. Sub-Store 负责不同客户端格式转换、策略组整理和分发。

## 本地测试

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

安装脚本语法检查：

```sh
sh -n scripts/install-control.sh
sh -n scripts/uninstall-control.sh
sh -n scripts/install-agent.sh
sh -n scripts/uninstall-agent.sh
```

完整测试路径见 [docs/TEST_PLAN.md](docs/TEST_PLAN.md)。

## 发布前检查

建议发布 GitHub 前执行：

```sh
cd frontend
npm run build

cd ../backend
GOCACHE=/private/tmp/s-ui-go-cache go test -tags postgres ./...
GOCACHE=/private/tmp/s-ui-go-cache go build -tags postgres -o /private/tmp/s-ui-backend-preview-check .
GOCACHE=/private/tmp/s-ui-go-cache GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64-preview-check ./agent

cd ..
sh -n scripts/install-control.sh
sh -n scripts/uninstall-control.sh
sh -n scripts/install-agent.sh
sh -n scripts/uninstall-agent.sh
git diff --check
```

如果你希望 GitHub Actions 自动把镜像构建并推送到 Docker Hub，需要在 GitHub 仓库设置 Secrets：

```text
DOCKER_HUB_USERNAME=sellength
DOCKER_HUB_TOKEN=你的 Docker Hub Access Token
```

这两个变量只给 GitHub Actions 使用，不需要写入服务器 `.env`，也不需要放进 `docker-compose.yml`。服务器使用 `docker compose up -d` 拉取公开镜像时不需要 Docker Hub Token；只有镜像仓库设为私有时，才需要先在服务器上执行 `docker login`。

推送 `dev` 分支会构建并推送：

```text
sellength/s-ui_dev:dev
sellength/s-ui_dev:latest
```

生产 Compose 默认还会拉取 sing-box 运行镜像：

```text
sellength/s-ui_rebuild-singbox:latest
```

首次发布时，请在 GitHub Actions 手动触发 `Sing-box Docker Image CI`。如果你想临时使用其它 sing-box 运行镜像，也可以在 `.env` 中设置 `SUI_SINGBOX_IMAGE`。

发布 GitHub Release 时，Release workflow 会上传：

```text
s-ui-linux-amd64.tar.gz
s-ui-agent-linux-amd64
```

以及其它矩阵平台对应文件。Agent 一键安装脚本默认下载 `releases/latest/download/s-ui-agent-linux-当前架构`。

发布前不要提交：

- `.codegraph/`
- `.local/`
- `mnt/`
- `dns-provider-icons.zip`
- `.env`
- 本地构建出的 `sui`、`s-ui-agent`
- 真实 API Token、SSH key、证书私钥

## 常见问题

### 这是正式生产版吗？

不是。当前建议命名为 `v0.1.0-preview.1`。它适合发布到 GitHub 让别人阅读、试用和 review，但不建议直接暴露在公网生产使用。

### 服务端必须用 PostgreSQL 吗？

Preview 开发环境默认使用 PostgreSQL。旧 SQLite 路径仍存在，但分布式方向建议使用 PostgreSQL。

### Agent 需要单独开端口吗？

不需要。Agent 主动连接 Control Plane。你只需要保证节点能访问面板的 Agent API。

### 一个节点可以有多个协议吗？

可以。一个节点可以创建多个服务入口，例如 AnyTLS、Hysteria2、VLESS。最终发布时会合并到该节点的 `config.json`。

### Reality 是单独协议吗？

不是。Reality 是 TLS 传输安全配置，常见组合是 VLESS + Reality。集群里应通过 VLESS 入口和协议高级 JSON 配置。

### 为什么全局配置不建议频繁修改？

全局配置是所有节点的默认运行骨架。协议差异更适合放在服务入口级策略里，否则一个节点的特殊需求可能影响全部节点。

### Cloudflare API Token 是否需要开启代理模式？

Let's Encrypt DNS-01 不需要开启 Cloudflare 代理。DNS API 只临时写 TXT 记录。AnyTLS 等协议通常建议 DNS 灰云直连，除非你明确使用支持该协议的代理能力。

### Docker 部署时证书和数据库如何持久化？

`docker-compose.yml` 会挂载：

```text
./db
./cert
./certificates
./acme
```

这些目录和 `.env` 都应该纳入备份。

## 文档索引

- [Preview 0.1 发布说明](docs/PREVIEW_0_1.md)
- [Preview 发布检查清单](docs/PREVIEW_RELEASE_CHECKLIST.md)
- [High Review 指南](docs/HIGH_REVIEW_GUIDE.md)
- [开发文档](docs/DEVELOPMENT.md)
- [本地 Docker 开发环境](docs/LOCAL_DOCKER_DEV.md)
- [测试清单](docs/TEST_PLAN.md)
- [Agent Runtime](docs/AGENT_RUNTIME.md)
- [Agent API](docs/AGENT_API.md)
- [分布式产品设计文档](docs/DISTRIBUTED_PRD.md)
- [技术基线](docs/TECH_BASELINE.md)
- [UI 工作流](docs/UI_WORKFLOW.md)
