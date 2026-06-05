# S-UI Rebuild 智能体开发与运维指南 (AGENT_PLAYBOOK.md)

本文件是专为 AI 智能体 (AI Agent) 以及自动化 CI/CD 工具在接管、开发、调试、打包及发布 `S-UI_rebuild` 项目时所制定的 **工作流手册 (Workflow & Skill Playbook)**。

如果您是接手本项目的 AI Agent，在开始任何开发之前，请全文阅读本指南并激活对应的技能包 (Skills)。

---

## 🧭 智能体专属技能地图 (Skills Map)

在处理本项目不同模块的问题时，请根据下表引导激活对应的全局开发技能：

| 任务场景 | 推荐调用的 Skill | 核心执行逻辑 |
| :--- | :--- | :--- |
| **功能开发与 Bug 修复** | `systematic-debugging` | **铁律**：严禁猜测和打补丁。必须遵循：1. 确认错误信息 -> 2. 编写/执行最小复现脚本 -> 3. 顺着调用栈回溯根源 (Root Cause) -> 4. 实施单点修复 -> 5. 验证是否解决。 |
| **前端 UI 美化与重构** | `ui-ux-pro-max` | 维持 Obsidian 暗黑曜石风格。全站组件必须使用磨砂玻璃面板 (Glassmorphism)，按钮与控制控件使用平滑的过渡曲线与微交互。不允许出现 Mock 图片，需要时必须调用 `generate_image` 产生。 |
| **后端 Go 接口与数据库** | `backend-dev-guidelines` | 遵循 Go 的强类型安全，在编写 GORM 数据库操作时务必注意数据库锁与防竞态。处理 IP 解析或配置热重载时应有超时控制。 |
| **部署方案与本地测试** | `kaizen` | 遵循 Kaizen（持续改进）的精简原则。在打包运行环境时，抛弃笨重的 Debian，选用 Alpine 作为基础运行环境，保证体积小、加载快。 |
| **任务成果交付** | `verification-before-completion` | 在任何时候声称“任务完成”前，必须进行自动化或手动的双重验证（如前端执行 build，后端通过编译，Docker 推送成功）。 |

---

## 🛠️ 核心开发工作流 (Standard Workflows)

### 工作流 A：前端重构与编译同步流
当您修改了前端代码（`frontend/`）并准备将其同步到 Go 后端服务中时，请按以下步骤操作：

1. **进入前端目录进行打包**：
   ```bash
   cd frontend
   npm run build
   ```
   *注意：必须保证没有 Typescript 强类型报错，且 Vite 编译零 Warning/Error。*
2. **清除并同步静态资源到 Go 后端**：
   ```bash
   cd ..
   rm -rf backend/web/html/*
   cp -R frontend/dist/ backend/web/html/
   ```
3. **在本地验证后端编译**：
   ```bash
   cd backend
   go build -o sui main.go
   ```

---

### 工作流 B：多平台 Docker 镜像打包与推送
由于本项目默认需要支持在 x86_64（`linux/amd64`，云服务器常用）和 ARM（`linux/arm64`，如 M系列 Mac、树莓派等）平台上原生部署，打包时必须输出多架构原生镜像。

#### 1. 前置准备：配置多平台构建器
默认的本地 Docker 驱动通常无法进行多架构构建，必须先创建一个基于 `docker-container` 驱动的 buildx 实例：
```bash
# 检查当前构建器
docker buildx inspect

# 如果当前构建器不支持多架构，则新建并切换
docker buildx create --name mybuilder --driver docker-container --use
docker buildx inspect --bootstrap
```

#### 2. 双平台镜像并行构建并推送 (Push to DockerHub)
运行构建时，必须通过 `--platform` 传参，且必须加入 `--push` 标记直接上传到 DockerHub 仓库中（这能避开本地 Docker 无法加载异构镜像的问题）：

```bash
# 1. 构建并推送面板本体镜像 (自动执行 Golang 内部交叉编译)
docker buildx build --platform linux/amd64,linux/arm64 -t sellength/s-ui_rebuild:latest --push .

# 2. 构建并推送 sing-box 核心守护镜像
docker buildx build --platform linux/amd64,linux/arm64 -t sellength/s-ui_rebuild-singbox:latest --push ./core
```

*注意：为使 `core/` 目录顺利打包，请在构建前确保 `core/sing-box` 中已存在一个 0 字节的占位文件用于过检。*

---

### 工作流 C：GitHub 源码推送与自更新链接维护
在本项目中，一键安装脚本及控制菜单在服务器上支持在线更新。因此，每次您提交修改到 GitHub 仓库时，**必须确保代码中的 URL 地址全部完成替换**。

#### 1. 关键自更新地址列表
在将代码 push 之前，请通过 `grep` 或文本编辑器确认以下位置的链接是否正确对应：

- **`install.sh`** 内部的获取源：
  - 下载 `s-ui-linux-$(arch).tar.gz` 应当指向：`https://github.com/sellength/S-UI_rebuild/releases/...`
  - 下载自更新 Shell 应当指向：`https://raw.githubusercontent.com/sellength/S-UI_rebuild/main/s-ui.sh`
- **`s-ui.sh`** 内部的自更新源：
  - 更新面板命令应当指向：`https://github.com/sellength/S-UI_rebuild/raw/main/s-ui.sh`
  - 更新一键脚本应当指向：`https://raw.githubusercontent.com/sellength/S-UI_rebuild/main/install.sh`

#### 2. 提交并推送到 GitHub 仓库
```bash
# 1. 确认无调试临时文件混入暂存区
git status

# 2. 执行提交与推送
git add .
git commit -m "feat/chore: [简短具体的修改描述]"
git push origin main
```

---

## ⚠️ 智能体禁忌与安全红线 (Red Lines)

1.  **禁止在前端直接硬编码后端 API 端口/协议**：所有的网络请求必须使用 Vite 代理服务或相对路径映射，以防止端口冲突。
2.  **默认凭证与一键脚本随机密码逻辑**：
    - 数据库的 SQLite 默认值硬编码为 `admin` / `admin`。
    - **但是**，一键安装脚本中，对于全新安装环境为避免爆破漏洞，必须调用 `urandom` 生成强随机密码并输出在终端中。重构时严禁将安装脚本里的随机生成逻辑改回弱密码。
3.  **Docker compose 网络模式**：
    - 在 Docker Compose 配置中，如果需要开放较多代理端口，在 README 中应有指引说明：可以通过修改 Compose 的网络模式为 `host`，以免除冗余的 `ports` 端口配置映射，仅需打开服务器的云防火墙即可。
