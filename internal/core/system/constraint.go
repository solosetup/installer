package system

// SystemConstraint 定义插件支持的操作系统约束
type SystemConstraint struct {
	Platform   string // 发行版 ID，如 "ubuntu", "debian"
	MinVersion string // 最低版本，如 "20.04"（空字符串表示无限制）
	MaxVersion string // 最高版本（空字符串表示无限制）
}