# 本地 Docker 开发环境

本地开发优先使用 Docker，避免在 Mac 上直接安装 PostgreSQL、Go 后端服务或额外运行时。

## 启动

```sh
docker compose -f docker-compose.dev.yml up -d --build s-ui
```

启动后：

- Web UI: http://localhost:2095/app/
- PostgreSQL: `localhost:54329`
- 数据库：`sui`
- 用户：`sui`
- 密码：`sui_dev_password`

开发 compose 会启用：

- `SUI_DB_TYPE=postgres`
- `SUI_POSTGRES_DSN=host=postgres ...`
- 后端使用 `-tags postgres` 构建

普通本地 Go 构建不带 `postgres` tag，仍然走 SQLite 兼容路径。

如果构建卡在 Docker Hub 镜像 metadata 或 GitHub 下载 acme.sh，通常是网络或镜像源问题。可稍后重试，或先使用本地 Go / npm 命令验证代码。

## 清理

停止并删除容器：

```sh
docker compose -f docker-compose.dev.yml down
```

连 PostgreSQL 数据卷一起删除：

```sh
docker compose -f docker-compose.dev.yml down -v
```

## 架构说明

你的 Mac 是 ARM，服务器是 x86-64。开发阶段可以直接在 Mac 上构建 ARM 镜像；发布阶段再使用 buildx 输出多架构镜像：

```sh
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t your-registry/s-ui-control:dev \
  .
```

服务器部署时拉取 `linux/amd64` 镜像。

## 依赖说明

PostgreSQL 支持依赖 `gorm.io/driver/postgres`，只在 `postgres` build tag 下启用。

首次构建开发镜像需要访问 Docker Hub、Go/NPM registry 和 GitHub。如果当前环境没有外网，普通 SQLite 路径仍可通过：

```sh
cd backend
GOCACHE=$PWD/.cache/go-build go test ./...
```

PostgreSQL tag 路径可通过：

```sh
cd backend
go test -tags postgres ./...
```
