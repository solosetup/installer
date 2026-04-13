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

VERSION="v0.1.3"
BIN_NAME="solosetup-linux-${BIN_ARCH}"
URL="https://github.com/solosetup/installer/releases/download/${VERSION}/${BIN_NAME}"

echo "📥 正在下载 ${BIN_NAME} ..."

# 重试逻辑：最多尝试 5 次，每次间隔 2 秒
RETRIES=5
for i in $(seq 1 $RETRIES); do
    # 使用 curl -f 确保 HTTP 错误时返回非零退出码
    # --retry 3：curl 内部重试 3 次
    # --connect-timeout 30：连接超时 30 秒
    if curl -fSL --retry 3 --connect-timeout 30 "$URL" -o /tmp/solosetup; then
        # 下载成功，校验是否为有效 ELF 文件
        if file /tmp/solosetup 2>/dev/null | grep -q "ELF"; then
            break
        else
            echo "⚠️ 下载的文件无效（可能网络错误），重试 $i/$RETRIES ..."
            sleep 2
        fi
    else
        echo "⚠️ 下载失败，重试 $i/$RETRIES ..."
        sleep 2
    fi
    # 最后一次尝试仍失败，则退出
    if [ $i -eq $RETRIES ]; then
        echo "❌ 下载失败，请检查网络后重试。"
        echo "   也可手动从以下地址下载并放置到 /tmp/solosetup："
        echo "   $URL"
        exit 1
    fi
done

chmod +x /tmp/solosetup

echo "✅ 下载完成！启动安装向导..."
# 根据是否有 /dev/tty 决定标准输入重定向方式，以兼容交互式与非交互式环境
if [ -c /dev/tty ]; then
    exec /tmp/solosetup </dev/stdin
else
    exec /tmp/solosetup
fi