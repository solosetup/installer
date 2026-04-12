package ros

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"myinstaller/internal/core/mirror"
	"myinstaller/internal/core/system"
	"myinstaller/internal/core/utils"
	"myinstaller/internal/plugin"
)

type RosPlugin struct{}

func init() {
	plugin.Register(&RosPlugin{})
}

func (p *RosPlugin) Name() string          { return "ros" }
func (p *RosPlugin) DisplayName() string   { return "ROS (Robot Operating System)" }
func (p *RosPlugin) Description() string   { return "安装 ROS 1 / ROS 2 开发环境，支持多版本选择" }
func (p *RosPlugin) Dependencies() []string { return []string{} }

func (p *RosPlugin) SupportedSystems() []system.SystemConstraint {
	return []system.SystemConstraint{
		{Platform: "ubuntu", MinVersion: "14.04"},
		{Platform: "debian", MinVersion: "9"},
	}
}

func (p *RosPlugin) Uninstall(sysInfo *system.SystemInfo) error {
	return fmt.Errorf("卸载功能暂未实现")
}

func (p *RosPlugin) Upgrade(sysInfo *system.SystemInfo) error {
	return fmt.Errorf("升级功能暂未实现")
}

func (p *RosPlugin) Install(sysInfo *system.SystemInfo) error {
	fmt.Println("\n========================================")
	fmt.Println("     ROS 一键安装向导")
	fmt.Println("========================================")

	if !p.isSystemCompatible(sysInfo) {
		return fmt.Errorf("当前系统 %s 不支持安装 ROS", sysInfo.String())
	}
	fmt.Printf("✓ 系统兼容: %s\n", sysInfo.String())

	// 获取兼容的 ROS 版本列表
	compatible := GetCompatibleVersions(sysInfo.Platform, sysInfo.Version)
	if len(compatible) == 0 {
		return fmt.Errorf("未找到兼容当前系统的 ROS 版本")
	}

	// 让用户选择
	selected, err := p.selectROSVersion(compatible)
	if err != nil {
		return err
	}
	fmt.Printf("✓ 已选择: ROS %d %s (%s)\n", selected.ROSVersion, selected.Distro, selected.Description)

	fmt.Println("\n[1/5] 准备基础环境...")
	if err := p.prepareEnvironment(sysInfo); err != nil {
		return fmt.Errorf("环境准备失败: %w", err)
	}

	fmt.Println("\n[2/5] 配置 ROS 软件源...")
	if err := p.addROSRepository(sysInfo, &selected); err != nil {
		return fmt.Errorf("软件源配置失败: %w", err)
	}

	fmt.Println("\n[3/5] 安装 ROS 核心包（可能需要较长时间）...")
	if err := p.installROSPackages(sysInfo, &selected); err != nil {
		return fmt.Errorf("ROS 包安装失败: %w", err)
	}

	fmt.Println("\n[4/5] 初始化 rosdep 依赖管理...")
	if err := p.initRosdep(sysInfo, selected.Distro); err != nil {
		fmt.Printf("⚠️ rosdep 初始化失败（非致命错误）: %v\n", err)
	} else {
		fmt.Println("✓ rosdep 初始化完成")
	}

	fmt.Println("\n[5/5] 配置环境变量...")
	if err := p.setupEnvironment(selected.Distro); err != nil {
		return fmt.Errorf("环境变量配置失败: %w", err)
	}

	fmt.Println("\n========================================")
	fmt.Printf("🎉 ROS %s 安装完成！\n", selected.Distro)
	fmt.Println("========================================")
	fmt.Println("请执行以下命令使环境变量生效：")
	fmt.Printf("  source ~/.bashrc\n")
	fmt.Printf("或重新打开终端。\n")
	fmt.Printf("验证安装: roscore\n")
	return nil
}

func (p *RosPlugin) isSystemCompatible(sysInfo *system.SystemInfo) bool {
	for _, constraint := range p.SupportedSystems() {
		if sysInfo.IsCompatible(constraint) {
			return true
		}
	}
	return false
}

func (p *RosPlugin) selectROSVersion(compatible []ROSVersion) (ROSVersion, error) {
	fmt.Println("\n请选择要安装的 ROS 版本:")
	for i, v := range compatible {
		rec := ""
		if v.Recommended {
			rec = "     推荐版本"
		}
		fmt.Printf("  [%d] %s (ROS %d) - %s%s\n", i+1, v.Distro, v.ROSVersion, v.Description, rec)
	}
	fmt.Printf("请输入数字选择 (1-%d): ", len(compatible))

	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(compatible) {
		return ROSVersion{}, fmt.Errorf("无效的选择")
	}
	return compatible[choice-1], nil
}

func (p *RosPlugin) prepareEnvironment(sysInfo *system.SystemInfo) error {
	pkgMgr := system.NewPackageManager(sysInfo)
	if pkgMgr == nil {
		return fmt.Errorf("不支持的包管理器")
	}

	// 清理可能残留的错误格式源文件，避免干扰 apt update
	badFiles := []string{
		"/etc/apt/sources.list.d/ros.list",
		"/etc/apt/sources.list.d/ros-latest.list",
		"/etc/apt/sources.list.d/ros2.list",
	}
	for _, f := range badFiles {
		// 检查文件是否存在，如果存在则删除
		if _, err := os.Stat(f); err == nil {
			fmt.Printf("发现残留源文件 %s，正在清理...\n", f)
			if err := utils.RunCommandWithSudo("rm", "-f", f); err != nil {
				fmt.Printf("警告: 删除 %s 失败: %v\n", f, err)
			}
		}
	}

	// 尝试更新源，失败只警告
	if err := pkgMgr.Update(); err != nil {
		fmt.Printf("⚠️ 软件源更新失败，继续尝试安装基础依赖: %v\n", err)
	}
	basePkgs := []string{"curl", "gnupg2", "lsb-release", "ca-certificates"}
	return pkgMgr.Install(basePkgs...)
}

func (p *RosPlugin) addROSRepository(sysInfo *system.SystemInfo, meta *ROSVersion) error {
	// 清理可能残留的旧源文件
	oldFiles := []string{
		"/etc/apt/sources.list.d/ros-latest.list",
		"/etc/apt/sources.list.d/ros.list",
		"/etc/apt/sources.list.d/ros2.list",
	}
	for _, f := range oldFiles {
		os.Remove(f)
	}

	mirrors := mirror.GetMirrors()
	mirror.WarmupMirrors(mirrors)

	// 根据 ROS 版本确定源类型和路径段
	sourceType := mirror.TypeROS
	pathSegment := "ros"
	if meta.ROSVersion == 2 {
		pathSegment = "ros2"
	}

	candidates := mirror.GetCandidatesByType(mirrors, sourceType)
	if len(candidates) == 0 {
		return fmt.Errorf("没有可用的 ROS 镜像源，请检查网络")
	}

	var lastErr error
	for i, m := range candidates {
		baseURL := m.URL + "/" + pathSegment + "/ubuntu"
		fmt.Printf("正在尝试镜像源 [%d/%d]: %s (%s)\n", i+1, len(candidates), m.Name, baseURL)

		if err := p.addAptRepositoryWithURL(sysInfo, baseURL); err != nil {
			fmt.Printf("  ❌ 失败: %v\n", err)
			lastErr = err
			os.Remove("/etc/apt/sources.list.d/ros.list")
		} else {
			fmt.Printf("  ✅ 成功使用镜像源: %s\n", m.Name)
			return nil
		}
	}

	if lastErr != nil {
		return fmt.Errorf("所有镜像源均尝试失败，最后错误: %w", lastErr)
	}
	return fmt.Errorf("所有镜像源均不可用")
}

func (p *RosPlugin) addAptRepositoryWithURL(sysInfo *system.SystemInfo, baseURL string) error {
	keyURL := "https://raw.githubusercontent.com/ros/rosdistro/master/ros.key"

	// 确保 curl 可用
	if !utils.IsCommandAvailable("curl") {
		if err := utils.RunCommandWithSudo("apt", "install", "-y", "curl"); err != nil {
			return fmt.Errorf("安装 curl 失败: %w", err)
		}
	}

	// 下载密钥到临时文件
	tmpKey := "/tmp/ros.key"
	if err := utils.DownloadFile(keyURL, tmpKey); err != nil {
		return fmt.Errorf("下载 ROS 密钥失败: %w", err)
	}
	defer os.Remove(tmpKey)

	// 使用 apt-key 添加密钥（简单可靠，不会卡住）
	if err := utils.RunCommandWithSudo("apt-key", "add", tmpKey); err != nil {
		return fmt.Errorf("添加 ROS 密钥失败: %w", err)
	}

	// 获取系统架构
	arch := ""
	if out, err := exec.Command("dpkg", "--print-architecture").Output(); err == nil {
		arch = strings.TrimSpace(string(out))
	}

	// 构建源行（注意：不再使用 signed-by，因为密钥已通过 apt-key 管理）
	var repoLine string
	if arch != "" {
		repoLine = fmt.Sprintf("deb [arch=%s] %s %s main", arch, baseURL, sysInfo.Codename)
	} else {
		repoLine = fmt.Sprintf("deb %s %s main", baseURL, sysInfo.Codename)
	}

	// 写入源文件
	repoFile := "/etc/apt/sources.list.d/ros.list"
	cmd := fmt.Sprintf("echo '%s' | sudo tee %s", repoLine, repoFile)
	if err := utils.RunShellCommand(cmd); err != nil {
		return fmt.Errorf("写入源文件失败: %w", err)
	}

	// 更新源
	return utils.RunCommandWithSudo("apt", "update")
}

// 辅助函数：检查 GPG 密钥文件是否有效
func isValidGPGKey(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	// GPG 公钥环至少应该有几百字节
	if info.Size() < 100 {
		return false
	}
	// 可以进一步用 gpg --show-keys 验证，但为了简单，这里只做基本检查
	return true
}

func (p *RosPlugin) installROSPackages(sysInfo *system.SystemInfo, meta *ROSVersion) error {
	pkgMgr := system.NewPackageManager(sysInfo)
	if pkgMgr == nil {
		return fmt.Errorf("不支持的包管理器")
	}
	packages := []string{meta.PackageBase}
	fmt.Printf("正在安装: %s ...\n", meta.PackageBase)
	return pkgMgr.Install(packages...)
}

func (p *RosPlugin) initRosdep(sysInfo *system.SystemInfo, distro string) error {
	pkgMgr := system.NewPackageManager(sysInfo)
	if pkgMgr == nil {
		return fmt.Errorf("不支持的包管理器")
	}
	if err := pkgMgr.Install("python3-rosdep"); err != nil {
		return fmt.Errorf("安装 python3-rosdep 失败: %w", err)
	}
	if utils.IsCommandAvailable("rosdepc") {
		if err := utils.RunCommand("rosdepc", "init"); err != nil {
			return fmt.Errorf("rosdepc init 失败: %w", err)
		}
		return utils.RunCommand("rosdepc", "update")
	}
	if err := utils.RunCommandWithSudo("rosdep", "init"); err != nil {
		fmt.Println("⚠️ rosdep init 可能已经执行过，继续...")
	}
	return utils.RunCommand("rosdep", "update")
}

func (p *RosPlugin) setupEnvironment(distro string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	sourceLine := fmt.Sprintf("source /opt/ros/%s/setup.bash", distro)

	content, err := os.ReadFile(bashrcPath)
	if err == nil && strings.Contains(string(content), sourceLine) {
		fmt.Println("✓ 环境变量已配置，跳过")
		return nil
	}

	f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开 .bashrc 失败: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n# ROS Environment\n" + sourceLine + "\n"); err != nil {
		return fmt.Errorf("写入 .bashrc 失败: %w", err)
	}
	fmt.Println("✓ 环境变量已添加到 ~/.bashrc")
	return nil
}