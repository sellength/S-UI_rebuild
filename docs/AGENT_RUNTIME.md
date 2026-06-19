# Agent Runtime 草案

当前 Agent 最小版本位于 `backend/agent`。

它目前覆盖远端节点的基础运行链路：

- 注册到 Control Plane
- 定时心跳
- 拉取 desired config
- 发现新配置版本后上报 `pulled`
- 将配置写入 `pending.json`
- 可选执行 `sing-box check`
- 校验成功后提升为 `current.json`
- 覆盖前备份旧配置到 `backup.json`
- 可选执行 reload 命令并上报 `applied`
- 保存 `state.json`，避免重启后重复处理同一配置版本
- 写入证书 bundle 到 `certs/{name}-{fingerprint}/fullchain.pem`
- 写入证书私钥到 `certs/{name}-{fingerprint}/privkey.pem`
- 通过 `systemctl`、进程检测和 `sing-box version` 上报基础 sing-box 状态和版本
- 上报已成功应用的配置版本

Preview 0.1 暂未完整实现：

- 配置失败回滚
- CPU / 内存 / 磁盘 / 流量等完整资源指标
- 最近错误日志摘要
- 默认 systemd reload 策略

## 构建

```sh
cd backend
GOCACHE=/private/tmp/s-ui-go-cache go build -o /private/tmp/s-ui-agent-preview-check ./agent
```

交叉构建 x86-64 Linux：

```sh
cd backend
GOCACHE=/private/tmp/s-ui-go-cache GOOS=linux GOARCH=amd64 go build -o /private/tmp/s-ui-agent-linux-amd64-preview-check ./agent
```

## systemd 安装草案

仓库提供了初版脚本：

```sh
scripts/install-agent.sh
```

默认情况下，该脚本会按服务器架构从 GitHub Release 下载 `s-ui-agent`：

```text
https://github.com/sellength/S-UI_rebuild/releases/latest/download/s-ui-agent-linux-当前架构
```

如果你已经手动上传二进制，脚本也会复用：

```text
/usr/local/s-ui-agent/s-ui-agent
```

然后生成：

```text
/usr/local/s-ui-agent/agent.env
/etc/systemd/system/s-ui-agent.service
```

示例：

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/sellength/S-UI_rebuild/dev/scripts/install-agent.sh) \
  --panel-agent-url https://panel.example.com/app/agent \
  --node-code us-01 \
  --agent-id agent-us-01 \
  --agent-token 'change-me' \
  --register-token 'same-as-SUI_AGENT_REGISTER_TOKEN'
```

如 Release 资产还没有发布，可以显式指定下载地址：

```sh
sudo scripts/install-agent.sh \
  --agent-download-url https://example.com/s-ui-agent-linux-amd64 \
  --panel-agent-url https://panel.example.com/app/agent \
  --node-code us-01 \
  --agent-id agent-us-01 \
  --agent-token 'change-me' \
  --register-token 'same-as-SUI_AGENT_REGISTER_TOKEN'
```

## 运行

```sh
./s-ui-agent \
  --base-url https://panel.example.com/app/agent \
  --node-code us-01 \
  --agent-id agent-us-01 \
  --agent-token 'change-me' \
  --register-token 'same-as-SUI_AGENT_REGISTER_TOKEN'
```

也可以使用环境变量：

```text
SUI_AGENT_BASE_URL
SUI_NODE_CODE
SUI_AGENT_ID
SUI_AGENT_TOKEN
SUI_AGENT_REGISTER_TOKEN
SUI_AGENT_INTERVAL
SUI_AGENT_CONFIG_DIR
SUI_AGENT_CERT_DIR
SUI_SINGBOX_BIN
SUI_AGENT_CHECK_CONFIG
SUI_AGENT_RELOAD_COMMAND
SUI_AGENT_STATE_PATH
```

`SUI_AGENT_INTERVAL` 支持 Go duration 格式，例如 `30s`、`1m`。

`SUI_AGENT_REGISTER_TOKEN` 是必填项，必须和控制端环境变量保持一致。控制端未配置该变量时会拒绝 Agent 注册。
> **获取方式**：
> * 在主控端机器 `/opt/s-ui-distributed/.env` (Docker) 或 `/usr/local/s-ui/s-ui.env` (实体机) 的 `SUI_AGENT_REGISTER_TOKEN` 中配置或查看。
> * 也可登录主控端服务器终端执行 `cat /proc/$(pgrep -f "/usr/local/s-ui/sui")/environ 2>/dev/null | tr '\0' '\n' | grep SUI_AGENT_REGISTER_TOKEN | cut -d '=' -f 2-` 动态捕获当前正在生效的 Token。

`SUI_AGENT_CHECK_CONFIG=false` 可以在没有 sing-box 二进制的开发环境跳过配置检查。

如果设置 `SUI_AGENT_RELOAD_COMMAND`，Agent 会在配置校验并提升为 `current.json` 后执行该命令。命令成功时上报 `applied`；未设置时只上报 `validated`。

Agent 状态采集优先使用系统服务名 `sing-box`。如果你的节点使用了不同的 systemd unit 名称，Preview 0.1 暂时会退回到进程检测；后续应把服务名做成可配置项。
