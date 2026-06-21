# Changelog

All notable changes to this project will be documented in this file.

## [v0.2.0-preview.1] - 2026-06-21

### Added
- **流量统计系统增强 (Traffic Stats)**: 新增了 Agent 心跳中对 `sing-box` 流量统计数据（基于 `v2ray_api`）的自动收集与上报功能。
- **动态配置注入**: 主控端在生成节点配置时，新增了动态注入 `v2ray_api` 统计模块的逻辑，仅对在全局配置中开启了统计的用户进行流量追踪。
- **安全与端口优化**: 默认将节点端 `sing-box` 的 `v2ray_api` 统计监听端口改为 `10080`（仅本地 `127.0.0.1` 监听），不向公网暴露，极大增强了系统安全性。
- **文档更新**:
  - 在 `README.md` 端口说明中补充了节点本地 `10080` 统计端口的用途和安全注意事项。
  - 在 `README.md` 中新增了 FAQ 章节，解答了为什么主控端与节点 Agent 安装脚本是解耦独立的，以及流量统计在各种状态下的排查步骤。
  - 在 `README.md` 中新增了 “Docker Compose 部署 Agent (双容器 Sidecar 模式)” 章节，提供完整的 Docker 容器化 Agent 部署及 reload 交互方式说明。
  - 在项目根目录中添加了 `docker-compose.agent.yml` 作为 Agent 容器化部署的官方 Compose 模版文件，方便用户在节点上一键拉取，并在 `README.md` 中补充了使用 `sed` 命令行一键生成并配置环境变量的极速部署说明。
- **Agent 容器化支持**: 新增了 `Dockerfile.agent` 用以将 S-UI Agent 打包为 Docker 镜像，并在 GitHub Actions 工作流中集成了 `sellength/s-ui_agent` 双架构镜像（amd64/arm64）的编译与推送。同时在 Agent 中实现了对同宿主机上 `sing-box` 容器运行状态与版本的自动感知（通过 docker.sock 自动获取 `running` 状态和执行版本查询）并向上报给主控。

### Fixed
- **策略覆盖 GORM/SQL 报错**: 修复了在 PostgreSQL 数据库下，由于 `policy_overrides_json` 字段类型为 `bytea`，在执行 SQL 校验时调用 `trim(policy_overrides_json)` 产生 `pg_catalog.btrim(bytea) does not exist (SQLSTATE 42883)` 的保存报错问题。我们将 SQL 过滤改为了在 Go 内存中校验 outstanding 策略覆盖，彻底解决跨数据库方言不兼容的问题。
- **策略覆盖 Fallback 逻辑**: 优化了服务入口“策略覆盖”编辑弹窗的初始化回退逻辑。若该节点尚未发布或生成过最新配置版本时，策略覆盖配置会优先尝试继承和读取主控端的**全局默认配置模板**（`nodeTemplateForm.config`），只有在全局模板也为空时才回退至前端硬编码默认值。
- **流量统计零值显示**: 修复了前端页面中，当用户有绑定但未产生流量时，流量统计渲染为 `-` 的问题。现在会正确渲染并显示为 `0 B`。
- **Postgres 批量写入限制**: 优化了后台 API 的批量记录插入逻辑，改用 Batch 分批写入，避免因单次插入参数过多触发 PostgreSQL 的 `65535 parameter limit` 写入报错。
- **配置校验与友好报错**: 优化了生成配置时缺少用户绑定或 TLS 证书的校验。若服务入口未绑定任何用户或未配置 TLS 证书，后端会返回包含清晰中文提示的友好报错；同时前端会对绑定用户数为 0 或未选择证书的服务入口进行红色警告高亮提示，并在点击“生成配置”时在前端直接进行拦截，指导用户进行绑定。

---

## [v0.1.0-preview.1] - 2026-06-05

### Added
- PostgreSQL 生产/开发数据库环境适配与 AutoMigrate。
- 集群多节点管理 (Data Plane Node 管理)。
- Agent 注册、心跳、配置拉取与状态自动上报。
- DNS Provider 集成管理与 Let's Encrypt 证书申请、续期。
- 证书中心及多协议服务入口（AnyTLS、Hysteria2、Trojan、VLESS 等）支持。
- 全局默认配置模板管理与配置版本生成与发布。
- 分布式订阅生成及用户绑定管理。
- 字段级加密保护：证书 PEM、证书私钥和 DNS 凭据使用 `SUI_SECRET_KEY` 字段级加密。
