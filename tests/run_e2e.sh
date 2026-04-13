#!/bin/bash
# tests/run_e2e.sh - 实时输出版本，最终汇总测试结果

DISTROS=(
    "ubuntu:20.04"
    "ubuntu:22.04"
    "ubuntu:24.04"
    "debian:11"
    "debian:12"
)

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo "========================================"
echo "   solosetup 多发行版兼容性测试"
echo "========================================"

PASS_COUNT=0
FAIL_COUNT=0
FAILED_DISTROS=()

for DISTRO in "${DISTROS[@]}"; do
    echo -e "\n${YELLOW}>>> 正在测试: ${DISTRO}${NC}"
    
    LOG_FILE="/tmp/solosetup-test-${DISTRO//:/_}.log"
    
    # 配置镜像源替换命令（兼容 Debian 12 DEB822 格式）
    if [[ "$DISTRO" == ubuntu* ]]; then
        MIRROR_CMD="sed -i 's@http://.*archive.ubuntu.com@http://mirrors.tuna.tsinghua.edu.cn@g' /etc/apt/sources.list"
    elif [[ "$DISTRO" == debian* ]]; then
        MIRROR_CMD='if [ -f /etc/apt/sources.list ]; then sed -i "s@http://.*debian.org@http://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list; else sed -i "s@http://.*debian.org@http://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list.d/debian.sources; fi'
    else
        MIRROR_CMD=""
    fi
    
    echo -e "${CYAN}   [1/3] 配置镜像源并安装依赖...${NC}"
    
    # 执行容器，实时显示输出，同时写入日志
    if timeout 600 docker run --rm \
        -v "$(pwd)":/work -w /work \
        -e DEBIAN_FRONTEND=noninteractive \
        -e SOLOSETUP_NONINTERACTIVE=1 \
        "$DISTRO" /bin/bash -c "
            set -e
            if [ -n \"$MIRROR_CMD\" ]; then
                echo '>>> 配置国内镜像源...'
                eval \"$MIRROR_CMD\"
            fi
            echo '>>> 更新软件源...'
            apt-get update -qq
            echo '>>> 安装 curl...'
            apt-get install -y curl
            echo '>>> 执行 install.sh...'
            bash install.sh
        " 2>&1 | tee "$LOG_FILE" | grep -E '^(>>>|🚀|📥|✅|❌|⚠️|正在|系统:|包管理器:|非交互|🎉|请选择|确认|即将安装)' --line-buffered; then
        
        echo -e "   ${CYAN}[2/3] 验证二进制...${NC}"
        if docker run --rm -v "$(pwd)":/work -w /work "$DISTRO" /bin/bash -c "
            if [ -f /tmp/solosetup ]; then
                /tmp/solosetup --version 2>/dev/null || /tmp/solosetup -h >/dev/null 2>&1
            else
                exit 1
            fi
        " >> "$LOG_FILE" 2>&1; then
            echo -e "   ${CYAN}[3/3] ${GREEN}✅ 测试通过${NC}"
            ((PASS_COUNT++))
            rm -f "$LOG_FILE"
        else
            echo -e "   ${RED}❌ 二进制验证失败${NC}"
            echo -e "   详细日志: ${LOG_FILE}"
            ((FAIL_COUNT++))
            FAILED_DISTROS+=("$DISTRO")
        fi
    else
        echo -e "   ${RED}❌ 安装脚本执行失败${NC}"
        echo -e "   详细日志: ${LOG_FILE}"
        ((FAIL_COUNT++))
        FAILED_DISTROS+=("$DISTRO")
    fi
done

echo -e "\n========================================"
echo -e "测试结果: ${GREEN}${PASS_COUNT} 通过${NC} / ${RED}${FAIL_COUNT} 失败${NC}"
if [ ${#FAILED_DISTROS[@]} -gt 0 ]; then
    echo -e "失败发行版: ${RED}${FAILED_DISTROS[*]}${NC}"
fi
echo "========================================"

if [ $FAIL_COUNT -eq 0 ]; then
    exit 0
else
    exit 1
fi
