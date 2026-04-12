package plugin

import "myinstaller/internal/core/system"

// InstallerPlugin 所有安装插件必须实现的接口
type InstallerPlugin interface {
	// Name 返回插件唯一标识（如 "ros", "docker"）
	Name() string
	// DisplayName 返回在菜单中显示的友好名称
	DisplayName() string
	// Description 返回简短描述
	Description() string
	// Dependencies 返回依赖的其他插件 Name 列表
	Dependencies() []string
	// SupportedSystems 返回支持的操作系统约束（空切片表示支持所有）
	SupportedSystems() []system.SystemConstraint
	// Install 执行安装逻辑
	Install(sysInfo *system.SystemInfo) error
	// Uninstall 可选，卸载逻辑
	Uninstall(sysInfo *system.SystemInfo) error
	// Upgrade 可选，升级逻辑
	Upgrade(sysInfo *system.SystemInfo) error
}