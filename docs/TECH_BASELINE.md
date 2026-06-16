# Distributed S-UI 技术基线

本文档记录分布式 S-UI 重构前的组件选择、版本基线和当前仓库对齐点。目标是在动代码之前先明确依赖边界，避免后续在数据库、证书、Agent、订阅格式上反复返工。

核对日期：2026-06-09

## 1. 当前仓库状态

当前项目是 Go 后端 + Vue 前端的单机 S-UI 面板。

- 后端：`backend/go.mod`
- 前端：`frontend/package.json`
- 部署：`docker-compose.yml`、`install.sh`、`s-ui.sh`
- 现有数据库模型：`backend/database/model/model.go`
- 当前 README 定位：单机 Sing-box Web 面板

当前项目已有能力：

- 用户、客户端、入站、出站、TLS、规则、订阅、统计
- Sing-box 配置生成
- Docker/非 Docker 部署脚本
- AnyTLS UI 相关优化痕迹

需要重构的核心：

- 从单机面板改为 Control Plane + Data Plane
- 从 SQLite 默认部署改为 PostgreSQL 生产部署
- 增加节点 Agent
- 增加证书中心
- 增加配置版本和配置下发
- 增加节点健康检查
- 保留现有 UI 风格，仅扩展信息架构

## 2. 组件版本矩阵

| 模块 | 组件 | 当前仓库 | 已核对上游状态 | 建议 |
|---|---:|---:|---:|---|
| Go | Go toolchain | `go 1.25.5` | Go 1.25 系列已有 `go1.25.10` 安全/修复版本 | 升级到 Go 1.25 最新补丁版 |
| Web 后端 | Gin | `v1.12.0` | GitHub latest `v1.12.0` | 保留 |
| ORM | GORM | `v1.31.1` | 当前项目已使用 | 保留 |
| SQLite | glebarez/sqlite | `v1.11.0` | 当前项目已使用 | 仅保留开发/迁移兼容 |
| PostgreSQL | PostgreSQL | 已接入开发环境 | 生产稳定线需发布前重新核对 | 生产默认 PostgreSQL |
| PG 驱动 | gorm postgres driver | 已接入 | pgx v5 是稳定主线 | 保留 |
| 证书 ACME | acme.sh | 已接入调用链 | 发布前重新核对上游版本 | Preview 0.1 主实现 |
| 外部证书 | acme.sh | 已接入调用链 | 发布前重新核对上游版本 | 后续兼容托管/导入 |
| 代理核心 | sing-box | 安装脚本动态拉 latest | GitHub latest `v1.13.13`，AnyTLS 自 `1.12.0` 起支持 | 固定最低版本 `>=1.13.13` 或明确 pin |
| 前端框架 | Vue | `^3.5.35` | GitHub latest `v3.5.35` | 保留 |
| UI 框架 | Vuetify | package: `^3.7.13`，lock: `3.12.8` | GitHub latest 页面显示 `v3.8.4`，npm/lock 需二次核对 | 暂不大升级，先锁定一致版本 |
| Router | vue-router | package: `^5.1.0`，lock: `5.1.0` | GitHub latest 显示 `v5.0.4` | 需要核对 npm dist-tag，避免不稳定版本范围 |
| 状态管理 | Pinia | `^3.0.4` | GitHub latest `v3.0.4` | 保留 |
| HTTP 客户端 | axios | `^1.16.1` | GitHub latest `v1.16.1` | 保留 |
| 图表 | Chart.js | `^4.5.1` | GitHub latest `v4.5.1` | 保留 |
| TypeScript | TypeScript | `^6.0.3` | GitHub latest `v6.0.3` | 保留 |
| 前端构建 | Vite | package: `^8.0.16` | 仓库/registry 信息需以 lock 和 npm audit 为准 | 暂不主动升级 |
| 部署 | Docker Compose | 当前 compose 文件 | docker/compose GitHub latest `v5.1.4` | 文档按 `docker compose` v2+ 命令写 |
| 订阅转换 | Sub-Store | 外部已部署 | 官方项目持续维护 | 不内置，仅集成 |

## 3. 官方来源

- Go Release History: https://go.dev/doc/devel/release
- Gin Releases: https://github.com/gin-gonic/gin/releases
- PostgreSQL: https://www.postgresql.org/
- PostgreSQL Versioning Policy: https://www.postgresql.org/support/versioning/
- pgx: https://github.com/jackc/pgx
- acme.sh Releases: https://github.com/acmesh-official/acme.sh/releases
- sing-box Releases: https://github.com/SagerNet/sing-box/releases
- sing-box AnyTLS inbound: https://sing-box.sagernet.org/configuration/inbound/anytls/
- sing-box AnyTLS outbound: https://sing-box.sagernet.org/configuration/outbound/anytls/
- sing-box TLS: https://sing-box.sagernet.org/configuration/shared/tls/
- Vue Releases: https://github.com/vuejs/core/releases
- Vuetify Releases: https://github.com/vuetifyjs/vuetify/releases
- vue-router Releases: https://github.com/vuejs/router/releases
- Pinia Releases: https://github.com/vuejs/pinia/releases
- Docker Compose Releases: https://github.com/docker/compose/releases
- Sub-Store: https://github.com/sub-store-org/Sub-Store

## 4. 版本策略

### 4.1 后端

后端继续使用 Go + Gin + GORM。

原因：

- 当前项目已经是这套结构
- 迁移成本低
- Agent 也适合用 Go 写成单二进制
- Go 对 systemd、文件、进程、HTTP 客户端、TLS 都很适合

调整：

- Go 从 `1.25.5` 升级到 1.25 最新补丁版
- 新增 PostgreSQL driver
- SQLite 降级为开发模式/兼容模式
- Agent 单独作为一个 Go binary，而不是塞进现有面板进程

### 4.2 前端

前端继续使用 Vue + Vuetify + Pinia。

原则：

- 不换 UI 框架
- 不重做视觉风格
- 保持现有深色、玻璃质感、1Panel 风格
- 只扩展导航、表单和状态反馈

需要先做依赖一致性检查：

- `package.json` 和 `package-lock.json` 的 Vuetify 版本不一致
- `vue-router` 的当前范围需要核对是否来自稳定 dist-tag
- Vite 版本以 lock 和实际构建为准，不在第一阶段主动做大版本迁移

### 4.3 sing-box

AnyTLS 需要 sing-box `>=1.12.0`。为了减少协议字段差异，建议分布式版本最低要求直接提高到 `>=1.13.13`。

注意：

- TLS 文档显示 `1.13.0`、`1.14.0` 有字段变更计划或新增字段
- 第一阶段配置生成器只生成稳定字段
- 高级 JSON 编辑器允许用户覆盖，但保存前必须跑 `sing-box check`

### 4.4 Docker

本地开发和服务器部署都用 Docker/Compose。

原因：

- 用户不想在 Mac 留太多残留
- Mac 是 ARM，服务器是 x86-64
- 可以用 buildx 生成 `linux/arm64` 和 `linux/amd64`
- PostgreSQL、Control Plane、fake Agent 都可以容器化测试

本地清理目标：

```sh
docker compose down -v
```

## 5. 证书中心基线

证书中心是分布式 AnyTLS 的核心模块，不作为临时功能处理。

Preview 0.1 主路径支持两种证书来源：

| 来源 | 用途 | 是否推荐 |
|---|---|---|
| Panel 内置 ACME DNS-01 | 新用户主路径 | 推荐 |
| acme.sh 导入 | 用户已有证书自动化 | 推荐作为兼容 |
| 手动上传 | 临时迁移/调试 | 暂不作为主路径 |

### 5.1 Panel 内置 ACME

Preview 0.1 底层使用 `acme.sh` 调用链。若未来切换到 Go 原生 ACME 库，需要重新评估 provider 覆盖、证书续期和凭据格式。

DNS Provider 不写死，设计为 provider adapter：

- Cloudflare
- AWS Route53
- Aliyun DNS
- Tencent Cloud DNS / DNSPod
- Manual DNS
- 后续扩展其他 provider

ACME 推荐方式：

- DNS-01
- 支持 wildcard 证书
- 不使用 HTTP-01 作为默认
- 不使用 TLS-ALPN-01 作为默认

原因：

- DNS-01 不占用 80/443 端口
- 不会和 AnyTLS 业务端口冲突
- 控制面可以集中签发，再下发到节点
- 适合 `*.example.com` 通配符证书

### 5.2 acme.sh 导入

acme.sh 不作为内置主流程，但提供两种接入：

- 从 acme.sh 证书目录导入
- 接收 acme.sh deploy hook 调用

导入后仍然进入统一证书资产模型：

- certificate
- certificate_version
- fingerprint
- not_before
- not_after
- domains
- private_key_encrypted

### 5.3 Cloudflare API 说明

Cloudflare API 在本项目中用于 DNS-01 TXT 记录验证，不等于生成 Cloudflare Origin Certificate。

默认不推荐 Cloudflare Origin Certificate 用于 AnyTLS 直连，因为普通客户端通常不信任 Cloudflare Origin CA。

## 6. 部署基线

### 6.1 本地开发

本地只用 Docker：

- PostgreSQL
- s-ui-control
- fake-agent-us
- fake-agent-sg
- 可选 mock sing-box

Mac ARM 只做功能验证。

### 6.2 生产部署

服务器 A，4GB 内存：

- PostgreSQL
- s-ui-control
- Web UI
- Agent API
- Subscription API
- Certificate Center
- Sub-Store 只做外部集成

服务器 B/C，512MB：

- s-ui-agent.service
- sing-box.service

Agent 不开放公网管理端口，只主动访问 Control Plane。

## 7. README 对齐项

后续 README 需要更新：

- 项目定位改为分布式 Sing-box 管理平台
- 增加架构图
- 增加 Control Plane 部署
- 增加 Agent 部署
- 增加 PostgreSQL 环境变量
- 增加证书中心说明
- 增加 ACME DNS-01 provider 配置说明
- 增加 Sub-Store 集成方式
- 增加本地 Docker 开发流程
- 增加多架构镜像构建说明

当前不建议立刻删除旧单机部署说明。可以保留为 Legacy / Standalone Mode，等分布式能力稳定后再降级展示。

本地 Docker 开发说明见 `docs/LOCAL_DOCKER_DEV.md`。
