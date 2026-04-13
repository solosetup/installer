#!/bin/bash
set -e

echo "🚀 正在安装 solosetup ..."

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)  BIN_ARCH="amd64" ;;
    aarch64) BIN_ARCH="arm64" ;;
    *)       echo "❌ 不支持的架构: $ARCH"; exit 1 ;;
esac

VERSION="v0.1.2"
BIN_NAME="solosetup-linux-${BIN_ARCH}"
URL="https://github.com/solosetup/installer/releases/download/${VERSION}/${BIN_NAME}"

echo "📥 正在下载 ${BIN_NAME} ..."
# -f 参数：服务器返回 HTTP 错误时让 curl 返回非零退出码
# --retry 3：失败后重试 3 次
curl -fSL --retry 3 "$URL" -o /tmp/solosetup

# 简单校验是否为有效的 ELF 文件（防止下载到 HTML 错误页）
if ! file /tmp/solosetup | grep -q "ELF"; then
    echo "❌ 下载的文件不是有效的可执行程序，可能网络异常，请重试。"
    echo "文件内容预览："
    head -c 200 /tmp/solosetup
    exit 1
fi

chmod +x /tmp/solosetup

echo "✅ 下载完成！启动安装向导..."
# 根据是否有 /dev/tty 决定重定向
if [ -c /dev/tty ]; then
    exec /tmp/solosetup </dev/stdin
else
    exec /tmp/solosetup
fi