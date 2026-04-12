package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCommand 执行命令并实时输出到终端
func RunCommand(name string, args ...string) error {
	fmt.Printf("> %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunCommandWithSudo 以 sudo 权限执行命令
func RunCommandWithSudo(name string, args ...string) error {
	sudoArgs := append([]string{name}, args...)
	return RunCommand("sudo", sudoArgs...)
}

// RunCommandWithOutput 执行命令并返回输出内容
func RunCommandWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunShellCommand 执行一段 shell 脚本
func RunShellCommand(script string) error {
	return RunCommand("sh", "-c", script)
}

// RunShellCommandWithSudo 以 sudo 权限执行 shell 脚本
func RunShellCommandWithSudo(script string) error {
	return RunCommandWithSudo("sh", "-c", script)
}