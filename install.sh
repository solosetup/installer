#!/bin/bash
set -e

echo "🚀 正在安装 solosetup ..."

ARCH=$(uname -m)
case $ARCH in
    x86_64)  BIN_ARCH="amd64" ;;
    aarch64) BIN_ARCH="arm64" ;;
    *)       echo "❌ 不支持的架构: $ARCH"; exit 1 ;;
esac

VERSION="v0.1.0"
BIN_NAME="solosetup-linux-${BIN_ARCH}"
URL="https://github.com/solosetup/installer/releases/download/${VERSION}/${BIN_NAME}"

echo "📥 正在下载 ${BIN_NAME} ..."
curl -sSL "$URL" -o /tmp/solosetup
chmod +x /tmp/solosetup

echo "✅ 安装完成！运行以下命令启动："
echo "   /tmp/solosetup"