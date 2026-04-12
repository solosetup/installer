package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// InstallConfig 定义安装配置文件的结构
type InstallConfig struct {
	// 全局配置
	Global GlobalConfig `yaml:"global"`
	// 要安装的插件列表
	Plugins []PluginConfig `yaml:"plugins"`
}

// GlobalConfig 全局安装选项
type GlobalConfig struct {
	// 是否非交互模式（无人值守）
	NonInteractive bool `yaml:"non_interactive"`
	// 安装失败时是否继续
	ContinueOnError bool `yaml:"continue_on_error"`
	// 日志级别: debug, info, warn, error
	LogLevel string `yaml:"log_level"`
	// 工作目录（用于下载临时文件）
	WorkDir string `yaml:"work_dir"`
}

// PluginConfig 单个插件的安装配置
type PluginConfig struct {
	// 插件名称（对应 InstallerPlugin.Name()）
	Name string `yaml:"name"`
	// 是否启用安装
	Enabled bool `yaml:"enabled"`
	// 插件特定的配置项（由插件自行解析）
	Options map[string]interface{} `yaml:"options,omitempty"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *InstallConfig {
	return &InstallConfig{
		Global: GlobalConfig{
			NonInteractive:  false,
			ContinueOnError: false,
			LogLevel:        "info",
			WorkDir:         filepath.Join(os.TempDir(), "myinstaller"),
		},
		Plugins: []PluginConfig{},
	}
}

// LoadConfig 从 YAML 文件加载配置
func LoadConfig(path string) (*InstallConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("配置文件不存在: %s", path)
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config InstallConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析 YAML 配置失败: %w", err)
	}

	// 设置默认值
	if config.Global.WorkDir == "" {
		config.Global.WorkDir = filepath.Join(os.TempDir(), "myinstaller")
	}
	if config.Global.LogLevel == "" {
		config.Global.LogLevel = "info"
	}

	// 创建工作目录
	if err := os.MkdirAll(config.Global.WorkDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作目录失败: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存配置到 YAML 文件
func SaveConfig(config *InstallConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

// GetEnabledPlugins 返回所有 Enabled 为 true 的插件名称列表
func (c *InstallConfig) GetEnabledPlugins() []string {
	var names []string
	for _, p := range c.Plugins {
		if p.Enabled {
			names = append(names, p.Name)
		}
	}
	return names
}

// GetPluginOptions 获取指定插件的配置选项
func (c *InstallConfig) GetPluginOptions(pluginName string) map[string]interface{} {
	for _, p := range c.Plugins {
		if p.Name == pluginName {
			return p.Options
		}
	}
	return nil
}

// IsPluginEnabled 检查指定插件是否启用
func (c *InstallConfig) IsPluginEnabled(pluginName string) bool {
	for _, p := range c.Plugins {
		if p.Name == pluginName {
			return p.Enabled
		}
	}
	return false
}

// GenerateSampleConfig 生成示例配置文件内容（用于 --init 命令）
func GenerateSampleConfig() string {
	cfg := DefaultConfig()
	cfg.Plugins = []PluginConfig{
		{Name: "ros", Enabled: true, Options: map[string]interface{}{"version": "humble"}},
		{Name: "docker", Enabled: false},
	}
	data, _ := yaml.Marshal(cfg)
	return string(data)
}