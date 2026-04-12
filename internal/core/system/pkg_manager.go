package system

import (
	"fmt"
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
	return runCommand("sudo", "apt", "update")
}

func (a *AptManager) Install(pkgs ...string) error {
	args := append([]string{"apt", "install", "-y"}, pkgs...)
	return runCommand("sudo", args...)
}

func (a *AptManager) Remove(pkgs ...string) error {
	args := append([]string{"apt", "remove", "-y"}, pkgs...)
	return runCommand("sudo", args...)
}

// DnfManager Fedora/RHEL 系列
type DnfManager struct{}

func (d *DnfManager) Update() error {
	return runCommand("sudo", "dnf", "check-update")
}

func (d *DnfManager) Install(pkgs ...string) error {
	args := append([]string{"dnf", "install", "-y"}, pkgs...)
	return runCommand("sudo", args...)
}

func (d *DnfManager) Remove(pkgs ...string) error {
	args := append([]string{"dnf", "remove", "-y"}, pkgs...)
	return runCommand("sudo", args...)
}

// YumManager 旧版 RHEL/CentOS
type YumManager struct{}

func (y *YumManager) Update() error {
	return runCommand("sudo", "yum", "check-update")
}

func (y *YumManager) Install(pkgs ...string) error {
	args := append([]string{"yum", "install", "-y"}, pkgs...)
	return runCommand("sudo", args...)
}

func (y *YumManager) Remove(pkgs ...string) error {
	args := append([]string{"yum", "remove", "-y"}, pkgs...)
	return runCommand("sudo", args...)
}

// runCommand 执行命令并实时输出
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
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
		if err := runCommand("sudo", "curl", "-fsSL", keyURL, "-o", keyPath); err != nil {
			return fmt.Errorf("下载密钥失败: %w", err)
		}
		if err := runCommand("sudo", "apt-key", "add", keyPath); err != nil {
			return fmt.Errorf("添加密钥失败: %w", err)
		}
	}

	repoFile := "/etc/apt/sources.list.d/custom.list"
	cmd := fmt.Sprintf("echo '%s' | sudo tee %s", repoLine, repoFile)
	return runCommand("sh", "-c", cmd)
}