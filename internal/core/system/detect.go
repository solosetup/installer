package system

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/shirou/gopsutil/v3/host"
)

// SystemInfo 包含探测到的系统信息
type SystemInfo struct {
	OS         string // linux, darwin, windows
	Arch       string // amd64, arm64
	Platform   string // ubuntu, debian, centos
	Version    string // 22.04, 11
	Codename   string // jammy, focal, bullseye
	PackageMgr string // apt, dnf, yum, pacman
}

// GetSystemInfo 探测并返回系统信息
func GetSystemInfo() (*SystemInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("获取系统信息失败: %w", err)
	}

	codename := getCodename(info.Platform, info.PlatformVersion)
	pkgMgr := detectPackageManager()

	return &SystemInfo{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Platform:   info.Platform,
		Version:    info.PlatformVersion,
		Codename:   codename,
		PackageMgr: pkgMgr,
	}, nil
}

// detectPackageManager 检测系统中可用的包管理器
func detectPackageManager() string {
	managers := []string{"apt", "dnf", "yum", "pacman", "zypper"}
	for _, mgr := range managers {
		if _, err := exec.LookPath(mgr); err == nil {
			return mgr
		}
	}
	return ""
}

// getCodename 尝试获取发行版代号
func getCodename(platform, version string) string {
	ubuntuCodenames := map[string]string{
		"20.04": "focal",
		"22.04": "jammy",
		"24.04": "noble",
	}
	if platform == "ubuntu" {
		if codename, ok := ubuntuCodenames[version]; ok {
			return codename
		}
	}
	return ""
}

// IsCompatible 检查当前系统是否满足约束
func (s *SystemInfo) IsCompatible(constraint SystemConstraint) bool {
	if constraint.Platform != "" && s.Platform != constraint.Platform {
		return false
	}
	if constraint.MinVersion != "" && s.Version < constraint.MinVersion {
		return false
	}
	if constraint.MaxVersion != "" && s.Version > constraint.MaxVersion {
		return false
	}
	return true
}

// String 返回友好的系统描述
func (s *SystemInfo) String() string {
	return fmt.Sprintf("%s %s (%s)", s.Platform, s.Version, s.Arch)
}