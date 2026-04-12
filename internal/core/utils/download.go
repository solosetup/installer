package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DownloadFile 下载文件到指定路径
func DownloadFile(url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// DownloadWithProgress 带进度显示的下载
func DownloadWithProgress(url, filepath, desc string) error {
	fmt.Printf("正在下载 %s ...\n", desc)
	return DownloadFile(url, filepath)
}

// ExtractTarGz 解压 tar.gz 文件到指定目录
func ExtractTarGz(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("创建 gzip reader 失败: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar 头失败: %w", err)
		}

		target := filepath.Join(destDir, header.Name)

		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("非法路径: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("创建目录失败 %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("创建父目录失败: %w", err)
			}
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("创建文件失败 %s: %w", target, err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("写入文件失败 %s: %w", target, err)
			}
			outFile.Close()
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("设置权限失败 %s: %w", target, err)
			}
		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("创建符号链接失败 %s -> %s: %w", target, header.Linkname, err)
			}
		default:
			fmt.Printf("警告: 忽略未知类型 %c 文件 %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

// ExtractZip 解压 zip 文件
func ExtractZip(zipPath, destDir string) error {
	return fmt.Errorf("暂不支持 zip 解压")
}

// IsCommandAvailable 检查系统命令是否可用
func IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}