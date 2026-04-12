package mirror

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// CheckMirror 检测单个镜像源的可用性和延迟
func CheckMirror(m *MirrorSource) {
	client := http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	resp, err := client.Head(m.URL)
	m.LastCheck = time.Now()

	if err != nil {
		m.Available = false
		m.Latency = 0
		return
	}
	defer resp.Body.Close()

	m.Available = resp.StatusCode >= 200 && resp.StatusCode < 400
	m.Latency = time.Since(start)
}

// VerifyRelease 验证镜像源是否包含指定发行版代号的 Release 文件
// baseURL 为镜像基础 URL，如 https://mirrors.aliyun.com/ros/ubuntu
// codename 为发行版代号，如 jammy
func VerifyMirrorRelease(baseURL, codename string) bool {
	releaseURL := fmt.Sprintf("%s/dists/%s/Release", baseURL, codename)
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head(releaseURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// CheckAllMirrors 并发检测所有镜像源
func CheckAllMirrors(mirrors []*MirrorSource) {
	var wg sync.WaitGroup
	for _, m := range mirrors {
		wg.Add(1)
		go func(m *MirrorSource) {
			defer wg.Done()
			CheckMirror(m)
		}(m)
	}
	wg.Wait()
}

// SelectBestMirror 从可用源中选择最优的一个（原有逻辑保留）
func SelectBestMirror(mirrors []*MirrorSource, sourceType string) *MirrorSource {
	candidates := filterAvailableByType(mirrors, sourceType)
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		latI := candidates[i].Latency
		latJ := candidates[j].Latency
		if latI < latJ-50*time.Millisecond {
			return true
		}
		if latJ < latI-50*time.Millisecond {
			return false
		}
		return candidates[i].Priority > candidates[j].Priority
	})
	return candidates[0]
}

// GetCandidatesByType 返回所有可用且支持指定类型的镜像源，按优先级和延迟排序
func GetCandidatesByType(mirrors []*MirrorSource, sourceType string) []*MirrorSource {
	candidates := filterAvailableByType(mirrors, sourceType)
	sort.Slice(candidates, func(i, j int) bool {
		// 优先考虑优先级，再考虑延迟
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		return candidates[i].Latency < candidates[j].Latency
	})
	return candidates
}

func filterAvailableByType(mirrors []*MirrorSource, sourceType string) []*MirrorSource {
	var candidates []*MirrorSource
	for _, m := range mirrors {
		if !m.Available {
			continue
		}
		if !supportsType(m, sourceType) {
			continue
		}
		candidates = append(candidates, m)
	}
	return candidates
}

func supportsType(m *MirrorSource, sourceType string) bool {
	for _, t := range m.Types {
		if t == sourceType {
			return true
		}
	}
	return false
}

// GetMirrorURL 获取指定类型的最优镜像源 URL
func GetMirrorURL(mirrors []*MirrorSource, sourceType string) string {
	best := SelectBestMirror(mirrors, sourceType)
	if best == nil {
		return ""
	}
	return best.URL
}

// WarmupMirrors 预热镜像源检测
func WarmupMirrors(mirrors []*MirrorSource) {
	CheckAllMirrors(mirrors)
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			CheckAllMirrors(mirrors)
		}
	}()
}