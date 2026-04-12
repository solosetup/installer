package ros

// ROSVersion 描述一个 ROS 发行版的完整元数据
type ROSVersion struct {
	Distro      string // 发行版代号，如 "humble", "noetic"
	ROSVersion  int    // 1 或 2
	UbuntuVers  []string // 支持的 Ubuntu 版本列表，如 ["20.04", "22.04"]
	DebianVers  []string // 支持的 Debian 版本列表（可选）
	Description string
	EOL         string // 官方支持结束日期
	PackageBase string // 基础包名，如 "ros-humble-desktop" 或 "ros-noetic-desktop-full"
	Recommended bool   // 是否为当前推荐版本
}

// GetAllROSVersions 返回所有已知 ROS 发行版（ROS 1 和 ROS 2）
func GetAllROSVersions() []ROSVersion {
	return []ROSVersion{
		// === ROS 2 系列 ===
		{
			Distro:      "rolling",
			ROSVersion:  2,
			UbuntuVers:  []string{"24.04", "23.10", "22.04"},
			Description: "ROS 2 Rolling Ridley (滚动更新版)",
			EOL:         "持续更新",
			PackageBase: "ros-rolling-desktop",
			Recommended: false,
		},
		{
			Distro:      "jazzy",
			ROSVersion:  2,
			UbuntuVers:  []string{"24.04"},
			Description: "ROS 2 Jazzy Jalisco (LTS, 推荐)",
			EOL:         "2029-05",
			PackageBase: "ros-jazzy-desktop",
			Recommended: true,
		},
		{
			Distro:      "iron",
			ROSVersion:  2,
			UbuntuVers:  []string{"22.04"},
			Description: "ROS 2 Iron Irwini",
			EOL:         "2024-11",
			PackageBase: "ros-iron-desktop",
			Recommended: false,
		},
		{
			Distro:      "humble",
			ROSVersion:  2,
			UbuntuVers:  []string{"22.04"},
			Description: "ROS 2 Humble Hawksbill (LTS)",
			EOL:         "2027-05",
			PackageBase: "ros-humble-desktop",
			Recommended: true,
		},
		{
			Distro:      "galactic",
			ROSVersion:  2,
			UbuntuVers:  []string{"20.04"},
			Description: "ROS 2 Galactic Geochelone",
			EOL:         "2022-11",
			PackageBase: "ros-galactic-desktop",
			Recommended: false,
		},
		{
			Distro:      "foxy",
			ROSVersion:  2,
			UbuntuVers:  []string{"20.04"},
			Description: "ROS 2 Foxy Fitzroy (LTS)",
			EOL:         "2023-06",
			PackageBase: "ros-foxy-desktop",
			Recommended: false,
		},

		// === ROS 1 系列 ===
		{
			Distro:      "noetic",
			ROSVersion:  1,
			UbuntuVers:  []string{"20.04"},
			DebianVers:  []string{"10", "11"},
			Description: "ROS 1 Noetic Ninjemys (LTS, 最后一代 ROS 1)",
			EOL:         "2025-05",
			PackageBase: "ros-noetic-desktop-full",
			Recommended: true,
		},
		{
			Distro:      "melodic",
			ROSVersion:  1,
			UbuntuVers:  []string{"18.04"},
			DebianVers:  []string{"9"},
			Description: "ROS 1 Melodic Morenia",
			EOL:         "2023-05",
			PackageBase: "ros-melodic-desktop-full",
			Recommended: false,
		},
		{
			Distro:      "kinetic",
			ROSVersion:  1,
			UbuntuVers:  []string{"16.04"},
			Description: "ROS 1 Kinetic Kame",
			EOL:         "2021-04",
			PackageBase: "ros-kinetic-desktop-full",
			Recommended: false,
		},
		{
			Distro:      "jade",
			ROSVersion:  1,
			UbuntuVers:  []string{"15.04"},
			Description: "ROS 1 Jade Turtle",
			EOL:         "2017-05",
			PackageBase: "ros-jade-desktop-full",
			Recommended: false,
		},
		{
			Distro:      "indigo",
			ROSVersion:  1,
			UbuntuVers:  []string{"14.04"},
			Description: "ROS 1 Indigo Igloo",
			EOL:         "2019-04",
			PackageBase: "ros-indigo-desktop-full",
			Recommended: false,
		},
	}
}

// GetCompatibleVersions 根据系统信息筛选可安装的 ROS 版本
func GetCompatibleVersions(platform, version string) []ROSVersion {
	all := GetAllROSVersions()
	var compatible []ROSVersion
	for _, v := range all {
		if platform == "ubuntu" {
			if contains(v.UbuntuVers, version) {
				compatible = append(compatible, v)
			}
		} else if platform == "debian" {
			if contains(v.DebianVers, version) {
				compatible = append(compatible, v)
			}
		}
	}
	return compatible
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetROSVersionByDistro 根据发行版代号查找元数据
func GetROSVersionByDistro(distro string) *ROSVersion {
	for _, v := range GetAllROSVersions() {
		if v.Distro == distro {
			return &v
		}
	}
	return nil
}