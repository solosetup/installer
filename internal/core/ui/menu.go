package ui

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"myinstaller/internal/core/system"
	"myinstaller/internal/plugin"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B7")).
			Padding(1, 0)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(0).
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				PaddingLeft(4)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

type MenuModel struct {
	plugins   []plugin.InstallerPlugin
	sysInfo   *system.SystemInfo
	cursor    int
	selected  map[int]bool
	quitting  bool
	confirmed bool // 是否已确认（按了回车）
	err       error
}

func NewMenuModel(sysInfo *system.SystemInfo) *MenuModel {
	allPlugins := plugin.GetAllPlugins()
	sort.Slice(allPlugins, func(i, j int) bool {
		return allPlugins[i].Name() < allPlugins[j].Name()
	})

	var compatible []plugin.InstallerPlugin
	for _, p := range allPlugins {
		if isCompatible(p, sysInfo) {
			compatible = append(compatible, p)
		}
	}

	return &MenuModel{
		plugins:   compatible,
		sysInfo:   sysInfo,
		cursor:    0,
		selected:  make(map[int]bool),
		confirmed: false,
	}
}

func isCompatible(p plugin.InstallerPlugin, sysInfo *system.SystemInfo) bool {
	constraints := p.SupportedSystems()
	if len(constraints) == 0 {
		return true
	}
	for _, c := range constraints {
		if sysInfo.IsCompatible(c) {
			return true
		}
	}
	return false
}

func (m *MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.plugins)-1 {
				m.cursor++
			}

		case " ":
			// 空格：切换选中
			if len(m.plugins) > 0 {
				if m.selected[m.cursor] {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = true
				}
			}

		case "a":
			// a：全选/反选
			if len(m.selected) == len(m.plugins) {
				m.selected = make(map[int]bool)
			} else {
				for i := range m.plugins {
					m.selected[i] = true
				}
			}

		case "enter":
			// 回车：确认选择，退出
			m.confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *MenuModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("🚀 欢迎使用 solosetup 一键安装工具"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("   当前系统: %s\n\n", m.sysInfo.String()))

	if len(m.plugins) == 0 {
		b.WriteString(errorStyle.Render("⚠️ 没有找到兼容当前系统的安装项"))
		b.WriteString("\n")
		b.WriteString(footerStyle.Render("按 q 退出"))
		return b.String()
	}

	b.WriteString("请选择要安装的软件 (使用 ↑/↓ 移动，空格 选择，a 全选/反选，回车 确认):\n\n")

	for i, p := range m.plugins {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		checked := "[ ]"
		if m.selected[i] {
			checked = "[✓]"
		}

		deps := p.Dependencies()
		depStr := ""
		if len(deps) > 0 {
			depStr = fmt.Sprintf(" (依赖: %s)", strings.Join(deps, ", "))
		}

		line := fmt.Sprintf("%s%s %s", cursor, checked, p.DisplayName())
		if m.cursor == i {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(itemStyle.Render(line))
		}
		b.WriteString("\n")
		b.WriteString(descriptionStyle.Render(p.Description() + depStr))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("按 回车 开始安装，q 退出"))

	return b.String()
}

// GetSelectedPlugins 返回用户选择的插件列表
func (m *MenuModel) GetSelectedPlugins() []plugin.InstallerPlugin {
	var selected []plugin.InstallerPlugin
	for i := range m.selected {
		selected = append(selected, m.plugins[i])
	}
	// 如果没有任何选中，则默认选中当前高亮项
	if len(selected) == 0 && len(m.plugins) > 0 {
		selected = append(selected, m.plugins[m.cursor])
	}
	return selected
}

// RunMenu 运行交互式菜单，返回用户选择的插件列表
func RunMenu(sysInfo *system.SystemInfo) ([]plugin.InstallerPlugin, error) {
	model := NewMenuModel(sysInfo)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("菜单运行失败: %w", err)
	}

	m, ok := finalModel.(*MenuModel)
	if !ok {
		return nil, fmt.Errorf("内部错误：模型类型不匹配")
	}

	// 用户按 q 退出，未确认
	if !m.confirmed {
		return nil, nil
	}

	return m.GetSelectedPlugins(), nil
}

// 以下简化菜单功能保持不变
func SimpleMenu(sysInfo *system.SystemInfo) ([]plugin.InstallerPlugin, error) {
	// 如果环境变量要求非交互，直接返回所有兼容插件
	if os.Getenv("SOLOSETUP_NONINTERACTIVE") == "1" || os.Getenv("CI") == "true" {
		allPlugins := plugin.GetAllPlugins()
		var compatible []plugin.InstallerPlugin
		for _, p := range allPlugins {
			if isCompatible(p, sysInfo) {
				compatible = append(compatible, p)
			}
		}
		if len(compatible) == 0 {
			return nil, fmt.Errorf("没有兼容当前系统的插件")
		}
		fmt.Println("非交互模式：自动选择所有兼容插件")
		return compatible, nil
	}

	allPlugins := plugin.GetAllPlugins()
	var compatible []plugin.InstallerPlugin
	for _, p := range allPlugins {
		if isCompatible(p, sysInfo) {
			compatible = append(compatible, p)
		}
	}

	if len(compatible) == 0 {
		fmt.Println("⚠️ 没有找到兼容当前系统的安装项")
		return nil, nil
	}

	fmt.Println("\n请选择要安装的软件 (输入序号，多个用逗号分隔，直接回车安装全部):")
	for i, p := range compatible {
		fmt.Printf("  [%d] %s - %s\n", i+1, p.DisplayName(), p.Description())
	}
	fmt.Print("\n请输入: ")

	var input string
	fmt.Scanln(&input)

	if input == "" {
		return compatible, nil
	}

	var selected []plugin.InstallerPlugin
	parts := strings.Split(input, ",")
	for _, part := range parts {
		var idx int
		if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &idx); err == nil {
			if idx >= 1 && idx <= len(compatible) {
				selected = append(selected, compatible[idx-1])
			}
		}
	}
	return selected, nil
}

func ShowInstallSummary(plugins []plugin.InstallerPlugin) {
	fmt.Println("\n即将安装以下软件:")
	for _, p := range plugins {
		fmt.Printf("  • %s\n", p.DisplayName())
	}
	fmt.Println()
}

func ConfirmInstallation() bool {
	// 非交互模式直接返回 true
	if os.Getenv("SOLOSETUP_NONINTERACTIVE") == "1" || os.Getenv("CI") == "true" {
		fmt.Println("非交互模式：自动确认安装")
		return true
	}

	// 尝试从 /dev/tty 读取输入，如果失败则回退到 os.Stdin
	var reader *bufio.Reader
	if tty, err := os.Open("/dev/tty"); err == nil {
		defer tty.Close()
		reader = bufio.NewReader(tty)
	} else {
		reader = bufio.NewReader(os.Stdin)
	}

	for {
		fmt.Print("确认开始安装? [y/N]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			// 读取失败，默认返回 false
			return false
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" || input == "" {
			return false
		}
		fmt.Println("无效输入，请输入 y 或 n")
	}
}
