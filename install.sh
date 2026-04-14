#!/bin/bash
set -e

echo "🚀 正在安装 solosetup ..."

# ---------- 检测系统架构 ----------
ARCH=$(uname -m)
case $ARCH in
    x86_64)  BIN_ARCH="amd64" ;;
    aarch64) BIN_ARCH="arm64" ;;
    *)       echo "❌ 不支持的架构: $ARCH"; exit 1 ;;
esac

# ---------- 版本配置 ----------
VERSION="v0.1.3"
BIN_NAME="solosetup-linux-${BIN_ARCH}"
URL="https://github.com/solosetup/installer/releases/download/${VERSION}/${BIN_NAME}"

echo "📥 正在下载 ${BIN_NAME} ..."

# ---------- 下载重试与校验 ----------
RETRIES=5
for i in $(seq 1 $RETRIES); do
    # 使用 curl -f 确保 HTTP 错误时返回非零退出码
    # --retry 3：curl 内部自动重试 3 次
    # --connect-timeout 30：连接超时 30 秒
    if curl -fSL --retry 3 --connect-timeout 30 "$URL" -o /tmp/solosetup; then
        # 等待文件系统同步
        sync
        sleep 0.5
        
        # 校验 ELF 魔数（0x7F 'E' 'L' 'F'）
        if [ "$(head -c 4 /tmp/solosetup 2>/dev/null | od -An -t x1 | tr -d ' ')" = "7f454c46" ]; then
            break
        else
            echo "⚠️ 下载的文件无效（可能网络错误），重试 $i/$RETRIES ..."
            sleep 2
        fi
    else
        echo "⚠️ 下载失败，重试 $i/$RETRIES ..."
        sleep 2
    fi
    
    # 最后一次尝试仍失败，退出
    if [ $i -eq $RETRIES ]; then
        echo "❌ 下载失败，请检查网络后重试。"
        echo "   也可手动从以下地址下载并放置到 /tmp/solosetup："
        echo "   $URL"
        exit 1
    fi
done

chmod +x /tmp/solosetup

echo "✅ 下载完成！启动安装向导..."

# ---------- 交互适配 ----------
# 如果存在 /dev/tty，则将标准输入重定向到它，确保管道安装时能正常读取键盘输入
if [ -c /dev/tty ]; then
    exec /tmp/solosetup </dev/tty
else
    exec /tmp/solosetup
fi