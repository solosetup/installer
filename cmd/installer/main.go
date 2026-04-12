package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"myinstaller/internal/core/config"
	"myinstaller/internal/core/mirror"
	"myinstaller/internal/core/system"
	"myinstaller/internal/core/ui"
	"myinstaller/internal/plugin"

	_ "myinstaller/internal/plugins/ros"
)

var (
	version = "dev"

	configPath     = flag.String("config", "", "指定配置文件路径 (YAML格式)")
	nonInteractive = flag.Bool("y", false, "非交互模式，自动确认所有操作")
	listPlugins    = flag.Bool("list", false, "列出所有可用的插件")
	initConfig     = flag.Bool("init", false, "生成示例配置文件")
	showVersion    = flag.Bool("version", false, "显示版本信息")
	workDir        = flag.String("workdir", "", "工作目录 (默认为系统临时目录)")
	logLevel       = flag.String("log", "info", "日志级别: debug, info, warn, error")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("myinstaller version %s\n", version)
		os.Exit(0)
	}
	if *initConfig {
		fmt.Println(config.GenerateSampleConfig())
		os.Exit(0)
	}

	fmt.Println("正在检测系统环境...")
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		fmt.Printf("系统检测失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("系统: %s\n", sysInfo.String())

	if sysInfo.PackageMgr == "" {
		fmt.Println("错误: 未检测到支持的包管理器 (apt/dnf/yum)")
		os.Exit(1)
	}
	fmt.Printf("包管理器: %s\n", sysInfo.PackageMgr)

	mirrors := mirror.GetMirrors()
	go mirror.WarmupMirrors(mirrors)

	if *listPlugins {
		listAllPlugins(sysInfo)
		os.Exit(0)
	}

	var cfg *config.InstallConfig
	if *configPath != "" {
		cfg, err = config.LoadConfig(*configPath)
		if err != nil {
			fmt.Printf("加载配置文件失败: %v\n", err)
			os.Exit(1)
		}
		if cfg.Global.NonInteractive {
			*nonInteractive = true
		}
		if cfg.Global.WorkDir != "" && *workDir == "" {
			*workDir = cfg.Global.WorkDir
		}
	}

	if *workDir == "" {
		*workDir = filepath.Join(os.TempDir(), "myinstaller")
	}
	os.MkdirAll(*workDir, 0755)

	var selectedPlugins []plugin.InstallerPlugin
	if cfg != nil && len(cfg.GetEnabledPlugins()) > 0 {
		selectedPlugins = loadPluginsFromConfig(cfg)
		fmt.Printf("\n从配置文件加载了 %d 个插件\n", len(selectedPlugins))
	} else {
		selectedPlugins, err = selectPluginsInteractive(sysInfo, *nonInteractive)
		if err != nil {
			fmt.Printf("选择插件失败: %v\n", err)
			os.Exit(1)
		}
	}

	if len(selectedPlugins) == 0 {
		fmt.Println("没有选择任何插件，退出。")
		os.Exit(0)
	}

	ui.ShowInstallSummary(selectedPlugins)

	// 确认安装
	if !*nonInteractive {
		if cfg == nil || !cfg.Global.NonInteractive {
			if !ui.ConfirmInstallation() {
				fmt.Println("已取消安装。")
				os.Exit(0)
			}
		}
	}

	fmt.Println("\n开始安装...")
	successCount := 0
	continueOnError := false
	if cfg != nil {
		continueOnError = cfg.Global.ContinueOnError
	}

	for _, p := range selectedPlugins {
		fmt.Printf("\n========================================\n")
		fmt.Printf("正在安装: %s\n", p.DisplayName())
		fmt.Printf("========================================\n")

		if err := checkDependencies(p, selectedPlugins); err != nil {
			fmt.Printf("依赖检查失败: %v\n", err)
			if !continueOnError && !askContinueOnError(p.DisplayName(), err) {
				fmt.Println("安装中止。")
				os.Exit(1)
			}
			continue
		}

		if err := p.Install(sysInfo); err != nil {
			fmt.Printf("\n❌ 安装 %s 失败: %v\n", p.DisplayName(), err)
			if !continueOnError && !askContinueOnError(p.DisplayName(), err) {
				fmt.Println("安装中止。")
				os.Exit(1)
			}
		} else {
			fmt.Printf("\n✅ %s 安装成功！\n", p.DisplayName())
			successCount++
		}
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("安装完成: %d 成功, %d 失败\n", successCount, len(selectedPlugins)-successCount)
	fmt.Printf("========================================\n")
}

func askContinueOnError(pluginName string, err error) bool {
	fmt.Printf("\n安装 %s 时出错: %v\n", pluginName, err)
	fmt.Print("是否继续安装其他软件? [y/N]: ")
	var answer string
	fmt.Scanln(&answer)
	return strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes"
}

func listAllPlugins(sysInfo *system.SystemInfo) {
	allPlugins := plugin.GetAllPlugins()
	fmt.Println("\n可用插件列表:")
	fmt.Println("--------------")
	for _, p := range allPlugins {
		compatible := "✓"
		constraints := p.SupportedSystems()
		if len(constraints) > 0 {
			isComp := false
			for _, c := range constraints {
				if sysInfo.IsCompatible(c) {
					isComp = true
					break
				}
			}
			if !isComp {
				compatible = "✗"
			}
		}
		fmt.Printf("  [%s] %s - %s\n", compatible, p.Name(), p.Description())
	}
	fmt.Println("\n✓ = 兼容当前系统, ✗ = 不兼容")
}

func selectPluginsInteractive(sysInfo *system.SystemInfo, nonInteractive bool) ([]plugin.InstallerPlugin, error) {
	allPlugins := plugin.GetAllPlugins()
	if len(allPlugins) == 0 {
		return nil, fmt.Errorf("没有注册任何插件")
	}
	if nonInteractive {
		var compatible []plugin.InstallerPlugin
		for _, p := range allPlugins {
			constraints := p.SupportedSystems()
			if len(constraints) == 0 {
				compatible = append(compatible, p)
				continue
			}
			for _, c := range constraints {
				if sysInfo.IsCompatible(c) {
					compatible = append(compatible, p)
					break
				}
			}
		}
		if len(compatible) == 0 {
			return nil, fmt.Errorf("没有兼容当前系统的插件")
		}
		return compatible, nil
	}

	selected, err := ui.RunMenu(sysInfo)
	if err != nil {
		fmt.Printf("TUI 菜单启动失败，使用简化菜单: %v\n", err)
		selected, err = ui.SimpleMenu(sysInfo)
	}
	return selected, err
}

func loadPluginsFromConfig(cfg *config.InstallConfig) []plugin.InstallerPlugin {
	var plugins []plugin.InstallerPlugin
	for _, name := range cfg.GetEnabledPlugins() {
		if p, ok := plugin.GetPlugin(name); ok {
			plugins = append(plugins, p)
		} else {
			fmt.Printf("警告: 配置中指定的插件 '%s' 未找到\n", name)
		}
	}
	return plugins
}

func checkDependencies(p plugin.InstallerPlugin, selected []plugin.InstallerPlugin) error {
	deps := p.Dependencies()
	if len(deps) == 0 {
		return nil
	}
	selectedNames := make(map[string]bool)
	for _, sp := range selected {
		selectedNames[sp.Name()] = true
	}
	var missing []string
	for _, dep := range deps {
		if !selectedNames[dep] {
			if _, ok := plugin.GetPlugin(dep); !ok {
				missing = append(missing, dep)
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少依赖插件: %s", strings.Join(missing, ", "))
	}
	return nil
}