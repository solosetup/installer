package plugin

var registry = make(map[string]InstallerPlugin)

// Register 由插件在 init() 中调用，将自己注册到全局注册表
func Register(p InstallerPlugin) {
	registry[p.Name()] = p
}

// GetAllPlugins 返回所有已注册的插件
func GetAllPlugins() []InstallerPlugin {
	plugins := make([]InstallerPlugin, 0, len(registry))
	for _, p := range registry {
		plugins = append(plugins, p)
	}
	return plugins
}

// GetPlugin 根据名称获取插件
func GetPlugin(name string) (InstallerPlugin, bool) {
	p, ok := registry[name]
	return p, ok
}