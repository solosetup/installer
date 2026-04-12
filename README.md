# solosetup

🚀 **一行命令，独奏式开发环境搭建**

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/solosetup/installer?include_prereleases)](https://github.com/solosetup/installer/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/solosetup/installer)](https://go.dev/)
[![License](https://img.shields.io/github/license/solosetup/installer)](LICENSE)

`solosetup` 是一个用 Go 语言编写的**零依赖、跨平台**的一行命令安装器。它将复杂的开发环境配置（如 ROS、Docker 等）压缩成一条命令，让开发者从繁琐的手动安装中解放出来。

---

## ⚠️ 当前状态：公开测试版 (Beta)

`solosetup` 目前处于 **v0.1.0** 早期开发阶段。核心框架已经可用，但功能仍在快速迭代中。

- ✅ **已支持**：Ubuntu 22.04 (ARM64) 上的 **ROS 2 Humble** 一键安装
- 🚧 **开发中**：更多 ROS 版本（Noetic、Jazzy）、Docker、VSCode 等插件
- 🚧 **开发中**：卸载功能、自更新、macOS/Windows 兼容性

**欢迎早期试用者反馈问题和建议，但不建议在生产环境或关键任务中完全依赖。**

---

## ✨ 特性

- **零依赖**：编译为单一静态二进制文件，无需安装 Python、Go 或其他运行时。
- **国内镜像源智能择优**：内置阿里云、清华、中科大等镜像站，自动测速并选择最快的可用源。
- **交互式 TUI 菜单**：基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 构建，操作直观。
- **插件化架构**：轻松扩展新软件支持，只需实现标准接口。
- **非交互模式**：支持通过配置文件或 `-y` 参数实现无人值守安装。
- **系统自动探测**：识别 Linux 发行版、版本、架构，自动适配安装策略。

---

## 🚀 快速开始

### 方式一：一行命令安装（推荐）

```bash
curl -sSL https://get.chaixiangyu.cn/install.sh | bash
```

安装脚本会自动检测你的系统架构，下载对应的预编译二进制，并立即启动安装向导。无需任何额外操作。

### 方式二：手动下载预编译二进制

访问 [Releases](https://github.com/solosetup/installer/releases) 页面，根据你的系统架构下载对应版本：

| 架构 | 文件名 |
| :--- | :--- |
| x86_64 (amd64) | `solosetup-linux-amd64` |
| ARM64 (aarch64) | `solosetup-linux-arm64` |

下载后赋予执行权限并运行：

```bash
chmod +x solosetup-linux-*
./solosetup-linux-*
```

### 方式三：从源码编译

```bash
git clone https://github.com/solosetup/installer.git
cd installer
go build -o solosetup ./cmd/installer
./solosetup
```

---

## 📋 已支持的插件

| 插件 | 状态 | 支持系统 |
| :--- | :--- | :--- |
| **ROS 2 Humble** | ✅ 可用 | Ubuntu 22.04 (ARM64) |
| ROS Noetic | 🚧 计划中 | Ubuntu 20.04 |
| Docker | 🚧 计划中 | Ubuntu/Debian |
| VSCode | 🚧 计划中 | Ubuntu/Debian |

> 想要贡献新插件？请参考 [插件开发指南](PLUGIN.md)（即将推出）。

---

## ⌨️ 使用示例

### 交互式安装

```bash
./solosetup
```

然后根据菜单提示选择要安装的软件。

### 非交互式安装（适合脚本化）

```bash
# 自动安装所有兼容的插件
./solosetup -y

# 通过配置文件指定安装内容
./solosetup --config install.yaml
```

示例配置文件 `install.yaml`：

```yaml
global:
  non_interactive: true
  continue_on_error: false
plugins:
  - name: ros
    enabled: true
    options:
      distro: humble
```

---

## 🤝 贡献

我们欢迎任何形式的贡献！无论是提交 Bug 报告、功能建议，还是直接提交代码 PR。

- 如果你发现了问题，请前往 [Issues](https://github.com/solosetup/installer/issues) 提交。
- 如果你想贡献代码，请先阅读 [贡献指南](CONTRIBUTING.md)（即将推出）。

---

## 📄 许可证

本项目采用 [MIT License](LICENSE)。你可以自由使用、修改和分发，但需保留原始版权声明。

---

## 👤 维护者

**path2future**  
- GitHub: [@path2future](https://github.com/path2future)

---

**感谢你的关注！如果这个项目对你有帮助，欢迎给个 Star ⭐️**


完成后，你的项目主页就会展示最新的安装方式，与刚刚发布的 Release 完全同步。
