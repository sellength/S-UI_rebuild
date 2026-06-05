#!/bin/bash

# ==============================================================================
# s-ui 本地原生部署安装辅助脚本 (Systemd 守护运行)
# ==============================================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
PLAIN='\033[0m'

# check root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}Error: Please run this script as root!${PLAIN}"
   exit 1
fi

echo -e "${YELLOW}正在清理旧版本服务...${PLAIN}"
systemctl stop s-ui || true
systemctl stop sing-box || true

echo -e "${YELLOW}正在创建本地目录...${PLAIN}"
mkdir -p /usr/local/s-ui
mkdir -p /usr/local/s-ui/bin
mkdir -p /usr/local/s-ui/db

echo -e "${YELLOW}正在部署 s-ui 主二进制程序...${PLAIN}"
cp -f /home/ec2-user/s-ui2/backend/sui /usr/local/s-ui/sui
chmod +x /usr/local/s-ui/sui

echo -e "${YELLOW}正在部署 Sing-Box 核心守护脚本...${PLAIN}"
cp -f /home/ec2-user/s-ui2/core/runSingbox.sh /usr/local/s-ui/bin/runSingbox.sh
chmod +x /usr/local/s-ui/bin/runSingbox.sh
# 强行在该脚本中注入向下兼容环境变量垫片
sed -i 's|runSingbox(){|runSingbox(){\n  export enable_deprecated_special_outbounds=true|g' /usr/local/s-ui/bin/runSingbox.sh

echo -e "${YELLOW}正在获取 Sing-Box 最新稳定版本...${PLAIN}"
latest_sb_tag=$(curl -Ls "https://api.github.com/repos/SagerNet/sing-box/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [[ ! -n "$latest_sb_tag" ]]; then
    latest_sb_tag="v1.13.12"
fi
sb_version=${latest_sb_tag#v}
sb_filename="sing-box-${sb_version}-linux-amd64"
sb_url="https://github.com/SagerNet/sing-box/releases/download/${latest_sb_tag}/${sb_filename}.tar.gz"

echo -e "最新 Sing-Box 版本: ${latest_sb_tag}"
echo -e "下载链接: ${sb_url}"

wget -O /tmp/sing-box-latest.tar.gz "${sb_url}"
tar -zxf /tmp/sing-box-latest.tar.gz -C /tmp/
cp -f "/tmp/${sb_filename}/sing-box" "/usr/local/s-ui/bin/sing-box"
chmod +x "/usr/local/s-ui/bin/sing-box"
rm -rf /tmp/sing-box-latest.tar.gz "/tmp/${sb_filename}"
echo -e "${GREEN}Sing-Box v${sb_version} 二进制部署成功！${PLAIN}"

echo -e "${YELLOW}正在配置 Systemd 服务单元...${PLAIN}"
cp -f /home/ec2-user/s-ui2/s-ui.service /etc/systemd/system/s-ui.service
cp -f /home/ec2-user/s-ui2/sing-box.service /etc/systemd/system/sing-box.service

echo -e "${YELLOW}正在初始化 SQLite 数据库与模型迁移...${PLAIN}"
/usr/local/s-ui/sui migrate

# 设置默认登录信息以防随机生成无法登录
echo -e "${YELLOW}正在设置默认管理员账号凭证 (admin / admin123)...${PLAIN}"
/usr/local/s-ui/sui admin -username admin -password admin123

echo -e "${YELLOW}正在重载 Systemd 并启动服务...${PLAIN}"
systemctl daemon-reload
systemctl enable s-ui
systemctl enable sing-box
systemctl start s-ui
systemctl start sing-box

echo -e "${YELLOW}正在等待服务启动并校验端口状态...${PLAIN}"
sleep 3

if systemctl is-active --quiet s-ui && systemctl is-active --quiet sing-box; then
    echo -e "\n========================================================"
    echo -e "🚀  ${GREEN}s-ui 原生 Systemd 部署成功！${PLAIN}"
    echo -e "--------------------------------------------------------"
    echo -e "📅 部署时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo -e "👤 默认用户名: ${GREEN}admin${PLAIN}"
    echo -e "🔑 默认密码: ${GREEN}admin123${PLAIN}"
    echo -e "🌐 面板端口: 2095"
    echo -e "🌐 订阅端口: 2096"
    echo -e "🤖 Sing-Box 版本: ${latest_sb_tag}"
    echo -e "========================================================\n"
else
    echo -e "${RED}部分服务启动失败，请检查 journalctl -u s-ui/sing-box 日志！${PLAIN}"
    exit 1
fi
