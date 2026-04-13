#!/bin/bash

echo "🚀 正在安装 solosetup ..."

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)  BIN_ARCH="amd64" ;;
    aarch64) BIN_ARCH="arm64" ;;
    *)       echo "❌ 不支持的架构: $ARCH"; exit 1 ;;
esac

# 版本号（与 GitHub Release 对应）
VERSION="v0.1.2"
BIN_NAME="solosetup-linux-${BIN_ARCH}"
URL="https://github.com/solosetup/installer/releases/download/${VERSION}/${BIN_NAME}"

echo "📥 正在下载 ${BIN_NAME} ..."
curl -sSL "$URL" -o /tmp/solosetup
chmod +x /tmp/solosetup

echo "✅ 下载完成！启动安装向导..."
# 如果 /dev/tty 可用，则重定向标准输入以支持交互；否则直接执行
if [ -c /dev/tty ]; then
    exec /tmp/solosetup </dev/stdin
else
    exec /tmp/solosetup
fi
