#!/bin/bash

# ==============================================================================
# s-ui DevOps Docker 一键启动与状态校验脚本
# ==============================================================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;36m'
PLAIN='\033[0m'

# 日志输出函数
log_info() {
    echo -e "${BLUE}[INFO]${PLAIN} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]🎉${PLAIN} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]⚠️${PLAIN} $1"
}

log_error() {
    echo -e "${RED}[ERROR]❌${PLAIN} $1"
}

# 1. 环境校验：检测 Docker 与 Docker Compose
log_info "正在检测 Docker 运行环境..."

if ! command -v docker >/dev/null 2>&1; then
    log_error "未检测到 Docker，请先安装 Docker！"
    exit 1
fi

# 检测 docker compose v2 或 v1
COMPOSE_CMD=""
if docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
    log_info "检测到 Docker Compose V2 插件"
elif command -v docker-compose >/dev/null 2>&1; then
    COMPOSE_CMD="docker-compose"
    log_info "检测到 Docker Compose V1 工具"
else
    log_error "未检测到 Docker Compose，请先安装 Docker Compose！"
    exit 1
fi

# 检测 Docker 守护进程是否在运行
if ! docker info >/dev/null 2>&1; then
    log_error "Docker 守护进程未启动，请先启动 Docker 服务！"
    exit 1
fi

# 2. 检查端口占用情况
check_port() {
    local port=$1
    if command -v lsof >/dev/null 2>&1; then
        lsof -i :"$port" -t >/dev/null 2>&1
        return $?
    elif command -v ss >/dev/null 2>&1; then
        ss -tulnp 2>/dev/null | grep -q -E ":$port\s"
        return $?
    elif command -v netstat >/dev/null 2>&1; then
        netstat -tulnp 2>/dev/null | grep -q -E ":$port\s"
        return $?
    else
        # 最后的 Bash TCP 尝试
        (exec 3<>/dev/tcp/127.0.0.1/"$port") >/dev/null 2>&1
        local res=$?
        exec 3>&- 2>/dev/null
        return $res
    fi
}

PORTS_TO_CHECK=(2095 2096 443)
PORT_CONFLICT=0

log_info "正在进行宿主机端口占用冲突校验..."
for port in "${PORTS_TO_CHECK[@]}"; do
    if check_port "$port"; then
        log_warn "检测到端口 $port 已被宿主机其他进程占用！"
        PORT_CONFLICT=1
    fi
done

if [ $PORT_CONFLICT -eq 1 ]; then
    log_error "端口占用冲突，请先释放上述端口，或修改 docker-compose.yml 中的端口映射后再重试！"
    exit 1
else
    log_success "宿主机端口冲突校验通过，所有核心端口均可用。"
fi

# 3. 创建持久化数据目录
log_info "正在初始化本地持久化卷目录..."
mkdir -p ./db ./cert
chmod 777 ./db ./cert
log_success "本地目录 ./db 与 ./cert 准备就绪。"

# 3.5. 宿主机极速编译与二进制分发 (避免容器内编译导致 OOM/I/O 挂起)
log_info "开始在宿主机进行 Go 后端 (sui) 二进制编译..."
if ! (cd backend && go build -o sui main.go); then
    log_error "宿主机 Go 后端编译失败，请检查 Golang 环境或语法！"
    exit 1
fi
log_success "宿主机 Go 后端编译成功。"

log_info "正在复制宿主机 Sing-Box 带 stats API 的定制二进制到 context..."
if [ -f "/usr/local/s-ui/bin/sing-box" ]; then
    cp -f /usr/local/s-ui/bin/sing-box ./core/sing-box
elif [ -f "/tmp/sing-box-1.13.12/sing-box" ]; then
    cp -f /tmp/sing-box-1.13.12/sing-box ./core/sing-box
else
    log_error "未在宿主机找到已编译的带 stats API 的 sing-box 二进制！请确保宿主机上已成功编译过它。"
    exit 1
fi
chmod +x ./core/sing-box
log_success "Sing-Box 部署二进制就绪。"

# 4. 启动 Docker Compose 容器栈
log_info "正在启动容器栈 (执行 $COMPOSE_CMD up --build -d)..."
if ! $COMPOSE_CMD up --build -d; then
    log_error "容器栈启动失败，请检查 Docker 日志或配置文件！"
    exit 1
fi

# 5. 校验容器运行状态与健康检查
log_info "正在等待容器初始化并进行健康状态校验 (最长等待 30 秒)..."
MAX_RETRIES=15
RETRY_COUNT=0
HEALTHY=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
    
    # 检查 s-ui 容器健康状态
    SUI_STATUS=$(docker inspect --format='{{.State.Health.Status}}' s-ui 2>/dev/null || echo "unknown")
    SUI_RUNNING=$(docker inspect --format='{{.State.Running}}' s-ui 2>/dev/null || echo "false")
    SB_RUNNING=$(docker inspect --format='{{.State.Running}}' sing-box 2>/dev/null || echo "false")
    
    if [ "$SUI_STATUS" = "healthy" ] && [ "$SUI_RUNNING" = "true" ] && [ "$SB_RUNNING" = "true" ]; then
        HEALTHY=1
        break
    fi
    
    # 如果容器直接挂了，提前退出
    if [ "$SUI_RUNNING" = "false" ] || [ "$SB_RUNNING" = "false" ]; then
        log_warn "检测到部分容器已停止运行！"
        break
    fi
    
    log_info "当前健康检查状态：s-ui=${SUI_STATUS} (Running=${SUI_RUNNING}), sing-box (Running=${SB_RUNNING})... [${RETRY_COUNT}/${MAX_RETRIES}]"
done

if [ $HEALTHY -eq 1 ]; then
    log_success "s-ui & sing-box 容器栈全部成功启动，且通过健康校验！"
    echo -e "\n========================================================"
    echo -e "🚀  ${GREEN}s-ui 面板容器化部署成功！${PLAIN}"
    echo -e "--------------------------------------------------------"
    echo -e "📅 部署时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo -e "🌐 IPv4 访问地址: ${BLUE}http://<你的服务器公网IP>:2095${PLAIN}"
    echo -e "🌐 IPv6 访问地址: ${BLUE}http://[<你的服务器IPv6地址>]:2095${PLAIN}"
    echo -e "📂 持久化数据库目录: ${YELLOW}./db/${PLAIN} (已映射到 /usr/local/s-ui/db)"
    echo -e "📂 持久化证书目录: ${YELLOW}./cert/${PLAIN} (已映射到 /usr/local/s-ui/cert)"
    echo -e "⚙️  已注入环境变量: ${BLUE}enable_deprecated_special_outbounds=true${PLAIN}"
    echo -e "========================================================\n"
else
    log_error "容器健康校验超时或失败！请通过以下命令排查："
    echo -e "   - 查看容器状态: docker ps -a"
    echo -e "   - 查看 s-ui 日志: docker logs s-ui"
    echo -e "   - 查看 sing-box 日志: docker logs sing-box"
    exit 1
fi
