package mirror

import "time"

// 支持的软件源类型常量
const (
	TypeSystem = "system" // 系统源 (apt/yum)
	TypeROS    = "ros"    // ROS 源
	TypeRosdep = "rosdep" // rosdep 源
	TypeDocker = "docker" // Docker 源
)

// MirrorSource 定义一个镜像源
type MirrorSource struct {
	Name     string        // 镜像站名称，如 "清华 tuna"
	URL      string        // 基础 URL
	Types    []string      // 支持的源类型列表
	Priority int           // 手动设定的优先级，值越大越优先
	Region   string        // 区域标识，如 "cn", "global"
	// 运行时状态（测速后填充）
	Available bool          // 上次检查是否可用
	Latency   time.Duration // 上次检查的延迟
	LastCheck time.Time     // 上次检查时间
}

// GetDefaultMirrors 返回内置的国内镜像源列表
// 优先保证可访问性，再考虑速度
func GetDefaultMirrors() []*MirrorSource {
	return []*MirrorSource{
		// 企业级镜像站（稳定性好，带宽充足）
		{
			Name:     "阿里云",
			URL:      "https://mirrors.aliyun.com",
			Types:    []string{TypeSystem, TypeROS, TypeDocker},
			Priority: 100,
			Region:   "cn",
		},
		{
			Name:     "腾讯云",
			URL:      "https://mirrors.cloud.tencent.com",
			Types:    []string{TypeSystem, TypeROS},
			Priority: 95,
			Region:   "cn",
		},
		{
			Name:     "华为云",
			URL:      "https://mirrors.huaweicloud.com",
			Types:    []string{TypeSystem, TypeROS, TypeDocker},
			Priority: 95,
			Region:   "cn",
		},
		// 高校镜像站（教育网内速度快）
		{
			Name:     "清华 tuna",
			URL:      "https://mirrors.tuna.tsinghua.edu.cn",
			Types:    []string{TypeSystem, TypeROS, TypeRosdep, TypeDocker},
			Priority: 90,
			Region:   "cn",
		},
		{
			Name:     "中科大 ustc",
			URL:      "https://mirrors.ustc.edu.cn",
			Types:    []string{TypeSystem, TypeROS, TypeRosdep},
			Priority: 90,
			Region:   "cn",
		},
		{
			Name:     "上海交大 sjtu",
			URL:      "https://mirror.sjtu.edu.cn",
			Types:    []string{TypeSystem, TypeROS},
			Priority: 85,
			Region:   "cn",
		},
		{
			Name:     "南京大学 nju",
			URL:      "https://mirrors.nju.edu.cn",
			Types:    []string{TypeSystem, TypeROS},
			Priority: 85,
			Region:   "cn",
		},
		{
			Name:     "北京外国语大学 bfsu",
			URL:      "https://mirrors.bfsu.edu.cn",
			Types:    []string{TypeSystem, TypeROS},
			Priority: 80,
			Region:   "cn",
		},
		// 官方源（作为最后备选，国内访问可能较慢）
		{
			Name:     "ROS 官方源",
			URL:      "http://packages.ros.org",
			Types:    []string{TypeROS},
			Priority: 10,
			Region:   "global",
		},
		{
			Name:     "Docker 官方源",
			URL:      "https://download.docker.com",
			Types:    []string{TypeDocker},
			Priority: 10,
			Region:   "global",
		},
	}
}

// GetMirrorsByType 根据源类型筛选可用镜像站
func GetMirrorsByType(mirrors []*MirrorSource, sourceType string) []*MirrorSource {
	var result []*MirrorSource
	for _, m := range mirrors {
		for _, t := range m.Types {
			if t == sourceType {
				result = append(result, m)
				break
			}
		}
	}
	return result
}

// FilterAvailable 过滤出可用的镜像站
func FilterAvailable(mirrors []*MirrorSource) []*MirrorSource {
	var result []*MirrorSource
	for _, m := range mirrors {
		if m.Available {
			result = append(result, m)
		}
	}
	return result
}