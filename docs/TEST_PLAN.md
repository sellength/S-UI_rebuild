# Distributed S-UI 测试清单

本文档记录当前分布式重构的本地验证步骤。

## 1. 已通过的自动验证

```sh
cd frontend
npm run build
```

```sh
cd backend
GOCACHE=/private/tmp/s-ui-go-cache go test -tags postgres ./...
GOCACHE=/private/tmp/s-ui-go-cache go build -tags postgres -o /private/tmp/s-ui-backend-preview-check .
GOCACHE=/private/tmp/s-ui-go-cache go build -o /private/tmp/s-ui-agent-preview-check ./agent
GOCACHE=/private/tmp/s-ui-go-cache GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64-preview-check ./agent
```

```sh
sh -n scripts/install-control.sh
sh -n scripts/uninstall-control.sh
sh -n scripts/install-agent.sh
sh -n scripts/uninstall-agent.sh
```

## 2. 当前待联网验证项

以下能力依赖真实公网 DNS、证书 CA、远端节点和 Docker Hub，不能只靠本地单元测试覆盖：

- 使用真实 DNS Provider 凭据申请和续签通配符证书。
- 在一台 x86_64 Linux 节点上安装 Agent，验证注册、心跳、拉取配置、下发证书和应用配置。
- 使用真实客户端订阅链接连接 AnyTLS / Hysteria2 / VLESS Reality 等协议。
- 使用 `docker compose -f docker-compose.dev.yml up -d --build s-ui` 从零构建镜像；该步骤需要能访问 Docker Hub 和 GitHub。

## 3. 手动功能测试路径

登录后台后进入：

```text
/app/distributed
```

建议按顺序测试：

1. 创建节点
   - 名称：`US-01`
   - 编码：`us-01`
   - 公开域名：`us.example.com`

2. 创建 DNS Provider
   - 类型：`Cloudflare` / `AliDNS` / `AWS` / `Tencent DNSPod`
   - 按页面模板填写对应 API 凭据

3. 申请证书
   - 方式：`DNS API 自动申请`
   - 域名：`*.example.com, example.com`
   - CA：`letsencrypt`
   - DNS Provider：选择刚创建的凭据

4. 创建服务入口
   - 节点：`US-01`
   - 协议：`AnyTLS` / `Hysteria2` / `VLESS` / `VMess` / `Trojan` / `Shadowsocks`
   - 端口：`443`
   - 证书：刚创建的证书

5. 绑定用户
   - 选择一个现有用户
   - 填入入口用户名和密码

6. 生成和预览服务入口配置
   - 点击服务入口列表中的 `生成配置`
   - 点击 `预览` 查看最终节点 config.json

7. 发布节点配置
   - 在节点列表点击发布
   - 配置版本页应出现 `published` 版本
   - 如果配置内容没有变化，再次发布应复用已有版本号，不应无意义递增

8. 创建订阅
   - 选择用户
   - 生成订阅 token

9. 验证订阅
   - `{subPath}/{token}?format=distributed-json`
   - `{subPath}/{token}?format=distributed-source`

10. 验证 Agent
   - 构建 `s-ui-agent-linux-amd64`
   - 放到节点 `/usr/local/s-ui-agent/s-ui-agent`
   - 使用 `scripts/install-agent.sh` 安装
   - Agent 应注册、心跳、拉取配置和证书

## 4. 已知限制

- ACME DNS-01 已接入 acme.sh 调用链，但需要真实 DNS Provider 和公网 CA 做端到端验证。
- Agent 目前通过可配置 reload command 应用配置，尚未内置 systemd reload 策略。
- Agent 当前心跳会采集基础 `sing-box` 运行状态和版本；Preview 0.1 仍需在真实节点验证 systemd、进程检测、版本命令和 reload command 的组合表现。
- 分布式页面仍处于 Preview UI，证书中心、服务入口和节点概览已经重构，但还需要真实节点验证后继续打磨。
- 节点健康监控仍需完善：
  - Agent 主动上报应补齐 CPU/内存/磁盘/流量和最近错误。
  - Control Plane 应补充外部探测作为辅助，包括 DNS 解析、端口连通、TLS 证书有效期和协议握手结果。
  - UI 需要区分内部状态（Agent / Sing-box / 配置同步）和外部访问状态（DNS / 端口 / TLS / 协议可用性）。
- Docker 开发镜像仍依赖构建期访问 Docker Hub 和 GitHub；如果网络不可用，构建会在基础镜像或 acme.sh 下载阶段失败。
