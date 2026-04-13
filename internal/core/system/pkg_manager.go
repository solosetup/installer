package system

import (
	"fmt"
	"os"
	"os/exec"
)

// PackageManager 定义包管理器统一接口
type PackageManager interface {
	Update() error
	Install(pkgs ...string) error
	Remove(pkgs ...string) error
}

// NewPackageManager 根据系统信息创建对应的包管理器实例
func NewPackageManager(sysInfo *SystemInfo) PackageManager {
	switch sysInfo.PackageMgr {
	case "apt":
		return &AptManager{}
	case "dnf":
		return &DnfManager{}
	case "yum":
		return &YumManager{}
	default:
		return nil
	}
}

// AptManager Debian/Ubuntu 系列
type AptManager struct{}

func (a *AptManager) Update() error {
	return runCommand("apt", "update")
}

func (a *AptManager) Install(pkgs ...string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	return runCommand("apt", args...)
}

func (a *AptManager) Remove(pkgs ...string) error {
	args := append([]string{"remove", "-y"}, pkgs...)
	return runCommand("apt", args...)
}

// DnfManager Fedora/RHEL 系列
type DnfManager struct{}

func (d *DnfManager) Update() error {
	return runCommand("dnf", "check-update")
}

func (d *DnfManager) Install(pkgs ...string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	return runCommand("dnf", args...)
}

func (d *DnfManager) Remove(pkgs ...string) error {
	args := append([]string{"remove", "-y"}, pkgs...)
	return runCommand("dnf", args...)
}

// YumManager 旧版 RHEL/CentOS
type YumManager struct{}

func (y *YumManager) Update() error {
	return runCommand("yum", "check-update")
}

func (y *YumManager) Install(pkgs ...string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	return runCommand("yum", args...)
}

func (y *YumManager) Remove(pkgs ...string) error {
	args := append([]string{"remove", "-y"}, pkgs...)
	return runCommand("yum", args...)
}

// isRoot 检查当前进程是否以 root 用户运行
func isRoot() bool {
	return os.Geteuid() == 0
}

// runCommand 执行命令并实时输出，自动根据是否为 root 决定是否添加 sudo
func runCommand(name string, args ...string) error {
	var cmd *exec.Cmd
	if isRoot() {
		cmd = exec.Command(name, args...)
	} else {
		cmd = exec.Command("sudo", append([]string{name}, args...)...)
	}
	cmd.Stdout = &cmdOutput{}
	cmd.Stderr = &cmdOutput{}
	return cmd.Run()
}

// cmdOutput 实现 io.Writer，用于实时打印命令输出
type cmdOutput struct{}

func (c *cmdOutput) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}

// AddRepository 添加 APT 仓库（仅 Ubuntu/Debian 适用）
func AddRepository(sysInfo *SystemInfo, repoLine, keyURL string) error {
	if sysInfo.PackageMgr != "apt" {
		return fmt.Errorf("当前系统不支持 APT 仓库")
	}

	if keyURL != "" {
		keyPath := "/tmp/repo_key.gpg"
		if err := runCommand("curl", "-fsSL", keyURL, "-o", keyPath); err != nil {
			return fmt.Errorf("下载密钥失败: %w", err)
		}
		if err := runCommand("apt-key", "add", keyPath); err != nil {
			return fmt.Errorf("添加密钥失败: %w", err)
		}
	}

	repoFile := "/etc/apt/sources.list.d/custom.list"
	var cmd string
	if isRoot() {
		cmd = fmt.Sprintf("echo '%s' | tee %s", repoLine, repoFile)
	} else {
		cmd = fmt.Sprintf("echo '%s' | sudo tee %s", repoLine, repoFile)
	}
	return runCommand("sh", "-c", cmd)
}
