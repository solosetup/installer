// internal/core/system/detect_test.go
package system

import (
    "testing"
)

func TestDetectPackageManager(t *testing.T) {
    // 这个函数在真实系统中运行，会检测到 apt（如果你的开发机是 Ubuntu）
    pkgMgr := detectPackageManager()
    
    // 在 Ubuntu 系统上，我们期望结果是 "apt"
    if pkgMgr != "apt" {
        t.Errorf("detectPackageManager() = %s, want %s", pkgMgr, "apt")
    }
}