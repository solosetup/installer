package mirror

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	// DefaultMirrorListURL 默认远程镜像列表地址（需替换为你自己的服务器）
	DefaultMirrorListURL = "https://raw.githubusercontent.com/yourname/myinstaller/main/mirrors.json"
)

var (
	// 全局镜像源缓存和读写锁
	cachedMirrors   []*MirrorSource
	cacheExpiry     time.Time
	cacheTTL        = 6 * time.Hour // 缓存有效期
	mu              sync.RWMutex
	remoteListURL   = DefaultMirrorListURL
)

// SetRemoteListURL 设置远程镜像列表的 URL
func SetRemoteListURL(url string) {
	remoteListURL = url
}

// RemoteMirrorConfig 远程镜像源配置结构（JSON 格式）
type RemoteMirrorConfig struct {
	Version   string         `json:"version"`   // 配置版本号
	UpdatedAt string         `json:"updated_at"` // 更新时间
	Mirrors   []*MirrorSource `json:"mirrors"`   // 镜像源列表
}

// FetchRemoteMirrors 从远程服务器获取最新的镜像源列表
func FetchRemoteMirrors() ([]*MirrorSource, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(remoteListURL)
	if err != nil {
		return nil, fmt.Errorf("请求远程列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("远程服务器返回错误状态码: %d", resp.StatusCode)
	}

	var config RemoteMirrorConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("解析远程 JSON 失败: %w", err)
	}

	return config.Mirrors, nil
}

// GetMirrors 获取镜像源列表（优先返回缓存，若过期或无效则返回内置默认）
func GetMirrors() []*MirrorSource {
	mu.RLock()
	// 缓存有效则直接返回
	if cachedMirrors != nil && time.Now().Before(cacheExpiry) {
		defer mu.RUnlock()
		return cachedMirrors
	}
	mu.RUnlock()

	// 缓存无效，尝试远程获取
	mu.Lock()
	defer mu.Unlock()

	// 双重检查，防止并发重复拉取
	if cachedMirrors != nil && time.Now().Before(cacheExpiry) {
		return cachedMirrors
	}

	// 尝试从远程拉取
	remote, err := FetchRemoteMirrors()
	if err == nil && len(remote) > 0 {
		cachedMirrors = remote
		cacheExpiry = time.Now().Add(cacheTTL)
		return cachedMirrors
	}

	// 远程拉取失败，回退到内置默认列表
	cachedMirrors = GetDefaultMirrors()
	cacheExpiry = time.Now().Add(1 * time.Hour) // 失败时缓存时间短一些，便于后续重试
	return cachedMirrors
}

// RefreshMirrors 强制刷新镜像源缓存（从远程拉取最新）
func RefreshMirrors() error {
	mu.Lock()
	defer mu.Unlock()

	remote, err := FetchRemoteMirrors()
	if err != nil {
		return err
	}

	cachedMirrors = remote
	cacheExpiry = time.Now().Add(cacheTTL)
	return nil
}

// StartBackgroundUpdater 启动后台定期更新器
// 每隔一段时间自动从远程拉取最新的镜像列表
func StartBackgroundUpdater(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			_ = RefreshMirrors()
		}
	}()
}