# S-UI 重构版 (S-UI Rebuild)
**基于 SagerNet/Sing-Box 构建的高级 Web 面板 • 针对原版 Bug 深度修复与功能优化**

本项目是基于原始项目 [alireza0/s-ui](https://github.com/alireza0/s-ui) 的重构与优化版本。针对原项目在中文环境下存在的配置保存 Bug 进行了彻底修复，优化了前端 UI 界面细节，并增加了对部分新协议特性的支持，旨在提供一个更稳定、易用的 Sing-Box 面板管理工具。

> **免责声明：** 本项目仅供个人学习与技术交流使用，请勿用于非法用途，请勿在生产环境或商业场景中使用。

**如果本项目对您有所帮助，欢迎点一个** :star2:

---

## 快速概览

| 功能 | 是否支持 | 状态/优化项 |
| -------------------------------------- | :----------------: | :--- |
| 多协议支持 (VLESS, VMess, Trojan, TUIC, Hysteria2, Shadowsocks等) | :heavy_check_mark: | 基础协议全支持，新增 AnyTLS 交互细节优化 |
| 多语言界面切换 (中文、英文、波斯文等) | :heavy_check_mark: | **[重磅修复]** 解决原版切换中文后配置无法保存的致命 Bug |
| 多客户端 / 多入站 (Inbounds) 管理 | :heavy_check_mark: | 支持在单个入站端口下添加并管理多个客户端 |
| 高级流量路由与分流规则管理 | :heavy_check_mark: | 可视化管理路由规则集与分流策略 |
| 客户端与系统状态监控 (流量、CPU、内存等) | :heavy_check_mark: | 修复了部分资源指示器和 Toast 信息反馈异常 |
| 订阅服务 (链接订阅 / JSON订阅 / 流量及有效期提示) | :heavy_check_mark: | 完善了订阅分发逻辑 |
| 深色 / 浅色主题切换 | :heavy_check_mark: | 曜石黑/半透明磨砂玻璃微调，提升了视觉精致度 |

---

## 默认安装信息

- **面板监听端口**：`2095`
- **面板访问路径**：`/app/`
- **订阅监听端口**：`2096`
- **订阅访问路径**：`/sub/`
- **默认用户名/密码**：`admin` / `admin`
  > **注意**：
  > 1. 为了安全性，如果使用一键脚本全新安装（未检测到已有数据库文件），脚本会**随机生成登录账号和密码**，并在安装结束时在终端中输出。
  > 2. 若您遗忘了登录密码，可在服务器终端输入 `s-ui` 命令，使用内置菜单功能进行密码查看或重置。

---

## 安装与升级方法 (非 Docker 安装方式)

在您的 Linux/macOS 服务器上执行以下一键安装脚本：

```sh
bash <(curl -Ls https://raw.githubusercontent.com/sellength/S-UI_rebuild/main/install.sh)
```

### 安装指定版本

如果您需要安装指定的旧版本，可以在命令后指定版本号（例如安装 `v1.0.0`）：

```sh
bash <(curl -Ls https://raw.githubusercontent.com/sellength/S-UI_rebuild/main/install.sh) v1.0.0
```

---

## Docker 部署方式

如果您更倾向于使用容器部署，可以使用以下方法。我们已提供由 **GitHub Actions** 自动编译的极速、免 CGO 闪退的最新成品容器镜像。

### 方式一：Docker / Podman Compose (推荐)

#### 1. 使用 Docker Compose
创建并进入部署目录：
```shell
mkdir s-ui && cd s-ui
wget -q https://raw.githubusercontent.com/sellength/S-UI_rebuild/main/docker-compose.yml
docker compose up -d
```

#### 2. 使用 Podman Compose (适用于 CentOS / AlmaLinux 等)
```shell
mkdir s-ui && cd s-ui
wget -q https://raw.githubusercontent.com/sellength/S-UI_rebuild/main/docker-compose.yml
podman-compose up -d
```

> [!TIP]
> **如何快速更新或升级镜像：**
> 以后当代码更新时，您只需要在服务器部署目录下执行以下命令即可一键平滑升级：
> ```shell
> # Docker 环境：
> docker compose pull && docker compose up -d
> 
> # Podman 环境（若提示数据卷残留，可配合 prune 一并清理）：
> podman-compose down
> podman volume prune -f
> podman-compose pull && podman-compose up -d
> ```

### 方式二：手动运行 Docker 容器

若要使容器与宿主机网络完全打通（免去繁琐的端口映射，直接在安全组/防火墙中开放对应节点端口即可），推荐使用 `host` 网络模式：

```shell
docker run -itd \
    --network host \
    -v $PWD/db/:/usr/local/s-ui/db/ \
    -v $PWD/cert/:/usr/local/s-ui/cert/ \
    --name s-ui --restart=unless-stopped \
    ghcr.io/sellength/s-ui_rebuild:latest
```

---

## 🛠️ 版本更新日志 (Changelog)

### v1.4.2-rebuild (当前版本)
- **[修复]** 修复了在系统设置中切换为“中文 (Simplified)”后，设置无法持久化保存到数据库的 Bug。
- **[修复]** 修正了 UI 界面中多个操作触发后，Toast 提示状态未及时刷新或误提示的逻辑缺陷。
- **[优化]** 对面板和订阅的资源路径处理（`webPath`, `subPath`）做出了更健壮的输入过滤，避免缺少 `/` 导致的 404 问题。
- **[迁移]** 将一键安装脚本与系统控制命令的自动更新指向迁移到 `sellength/S-UI_rebuild` 新仓库，确保了脚本维护与更新的独立性。
- **[特性]** 优化了前端对 AnyTLS 协议设置的支持与 UI 表单控制。
- **[配置]** 默认 Docker Compose 配置增加了测试端口映射的注释模板，提供更好的自定义端口指导。

---

## 系统管理命令

安装完成后，您可以在服务器终端直接运行 `s-ui` 命令来调出管理菜单。支持的子命令如下：

```text
s-ui              - 显示管理脚本菜单
s-ui start        - 启动面板与 Sing-Box
s-ui stop         - 停止面板与 Sing-Box
s-ui restart      - 重启面板与 Sing-Box
s-ui status       - 查看运行状态
s-ui enable       - 设置面板开机自启
s-ui disable      - 取消开机自启
s-ui log          - 查看面板运行日志
s-ui update       - 强制重新安装最新版本
s-ui install      - 执行全新安装
s-ui uninstall    - 卸载面板与 Sing-Box
```

## 贡献与本地开发

如果您想对本项目进行二次开发或贡献代码，请参考以下编译说明。

### 1. 前端开发 (Vuetify + Vite)
前端源码位于 `frontend` 文件夹：
```shell
cd frontend
npm install
npm run dev
```
前端将在 `3000` 端口启动开发服务器，并通过代理将 API 请求转发至后端的 `2095` 端口。

### 2. 后端开发与编译 (Go)
编译前请确保前端已完成打包（即拥有最新的 `frontend/dist` 资源）：
```shell
cd backend
rm -fr web/html/*
cp -R ../frontend/dist/ web/html/
go build -o ../sui main.go
```
编译完成后，会在根目录下生成 `sui` 二进制程序，在根目录下直接执行 `./sui` 即可运行。

---

## 支持的操作系统

- Ubuntu 20.04+
- Debian 11+
- CentOS 8+
- Fedora 36+
- Arch Linux / Manjaro
- AlmaLinux 9+ / Rocky Linux 9+
