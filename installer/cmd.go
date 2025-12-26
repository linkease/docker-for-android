package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	ext4Magic           = 0xEF53
	internalStoragePath = "/data/local/docker-storage"
)

// ensureDiskRoot 准备并检查磁盘根路径
// - 确保目录存在
// - 提示文件系统类型（非 ext4 时给出警告）
func ensureDiskRoot(path string) (string, error) {
	root := strings.TrimSpace(path)
	if root == "" {
		return "", fmt.Errorf("磁盘路径为空")
	}

	if err := os.MkdirAll(root, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(root, &stat); err != nil {
		return "", fmt.Errorf("无法获取文件系统信息: %v", err)
	}

	if stat.Type != ext4Magic {
		fmt.Printf("⚠ 警告: %s 的文件系统类型为 0x%x（非 ext4），Docker 可能无法正常运行\n", root, stat.Type)
	}

	return root, nil
}

// promptYesNo 简单的 Y/N 交互
func promptYesNo(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	text, _ := reader.ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(text))
	return answer == "y" || answer == "yes"
}

// getDiskSize 获取指定路径的磁盘大小（KB）
func getDiskSize(path string) (uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	// 计算总大小（字节），然后转换为 KB
	// Blocks * BlockSize / 1024
	totalSizeKB := (stat.Blocks * uint64(stat.Bsize)) / 1024
	return totalSizeKB, nil
}

// detectDiskMount 检测硬盘挂载点
func detectDiskMount() (string, error) {
	// 纯 Go 实现：遍历 /mnt/media_rw 下 ext4 挂载点，选最大空间
	file, err := exec.Command("mount").Output()
	if err != nil {
		return "", fmt.Errorf("无法获取挂载信息: %v", err)
	}
	lines := strings.Split(string(file), "\n")
	var maxSize uint64
	var bestMount string
	for _, line := range lines {
		// 只处理 ext4 且挂载在 /mnt/media_rw 下的
		if !strings.Contains(line, " type ext4 ") || !strings.Contains(line, "/mnt/media_rw/") {
			continue
		}
		// 解析挂载点
		parts := strings.Fields(line)
		var mountPath string
		for i, part := range parts {
			if part == "on" && i+1 < len(parts) {
				mountPath = parts[i+1]
				break
			}
		}
		if mountPath == "" {
			continue
		}
		// 使用纯 Go 获取挂载点空间
		size, err := getDiskSize(mountPath)
		if err != nil {
			continue
		}
		if size > maxSize {
			maxSize = size
			bestMount = mountPath
		}
	}
	if bestMount == "" {
		fmt.Println("⚠ 未检测到外置 ext4 硬盘")
		useInternal := promptYesNo("是否使用内置存储 /data/local/docker-storage 继续安装? (y/N): ")
		if !useInternal {
			return "", fmt.Errorf("未检测到硬盘挂载点")
		}

		confirm := promptYesNo("请确认使用 /data/local/docker-storage 作为存储目录，可能有空间限制，继续吗? (y/N): ")
		if !confirm {
			return "", fmt.Errorf("已取消使用内置存储，请接入外置硬盘后重试")
		}

		if root, err := ensureDiskRoot(internalStoragePath); err == nil {
			fmt.Printf("⚠ 使用内置存储: %s\n", root)
			return root, nil
		}

		return "", fmt.Errorf("无法准备内置存储路径: %s", internalStoragePath)
	}
	return bestMount, nil
}

// detectArchitecture 检测当前系统架构
func detectArchitecture() (string, error) {
	cmd := exec.Command("uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法检测系统架构: %v", err)
	}

	arch := strings.TrimSpace(string(output))
	switch arch {
	case "aarch64", "arm64":
		return "arm64", nil
	case "x86_64", "amd64":
		return "x86_64", nil
	default:
		return "", fmt.Errorf("不支持的架构: %s", arch)
	}
}

// executeScript 执行脚本并实时输出日志，同时传递 diskRoot 参数
func executeScript(scriptPath string, diskRoot string) error {
	cmd := exec.Command("sh", scriptPath, diskRoot)

	// 创建管道来捕获标准输出和标准错误
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建 stdout 管道失败: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建 stderr 管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动脚本失败: %v", err)
	}

	// 实时打印输出
	go streamOutput(stdout)
	go streamOutput(stderr)

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("脚本执行失败: %v", err)
	}

	return nil
}

// streamOutput 流式输出日志
func streamOutput(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

// stopSupervisord 停止正在运行的 supervisord 服务
// 在解压文件之前调用，避免后续服务启动失败
func stopSupervisord() error {
	supervisordPath := "/data/local/docker/bin/supervisord"

	// 检查 supervisord 文件是否存在
	if _, err := os.Stat(supervisordPath); os.IsNotExist(err) {
		// 文件不存在，说明是首次安装，无需停止服务
		fmt.Println("✓ 未检测到已安装的 supervisord，跳过停止服务")
		return nil
	}

	fmt.Println("⏳ 检测到已安装的 supervisord，正在停止服务...")

	// 执行 supervisorctl stop all
	cmd := exec.Command(supervisordPath, "ctl", "stop", "all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果命令执行失败，可能是 supervisord 未运行，继续执行
		fmt.Printf("⚠ supervisorctl stop all 执行警告: %v\n", err)
		fmt.Printf("  输出: %s\n", strings.TrimSpace(string(output)))
	} else {
		fmt.Println("✓ supervisord 服务已停止")
		fmt.Printf("  %s\n", strings.TrimSpace(string(output)))
	}

	// 等待一小段时间，让进程有机会正常退出
	time.Sleep(2 * time.Second)

	// 检查 supervisord 进程是否仍在运行，并强制终止
	fmt.Println("⏳ 检查并终止 supervisord 进程...")
	if err := killSupervisordProcesses(); err != nil {
		fmt.Printf("终止 supervisord 进程失败: %v，但继续运行\n", err)
	}

	fmt.Println("✓ 所有 supervisord 进程已终止")
	return nil
}

// killSupervisordProcesses 使用 pkill 终止所有 supervisord 进程
func killSupervisordProcesses() error {
	// 使用 pkill -9 直接终止所有 supervisord 进程
	// pkill 返回码: 0=成功终止进程, 1=没有匹配的进程, 其他=错误
	cmd := exec.Command("pkill", "-9", "supervisord")
	err := cmd.Run()

	if err != nil {
		// 检查是否是 exit status 1（没有匹配的进程）
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				// 没有找到进程，这也是成功的情况
				fmt.Println("  未检测到 supervisord 进程")
				return nil
			}
		}
		// 其他错误
		return fmt.Errorf("执行 pkill 失败: %v", err)
	}

	// 等待进程完全终止
	time.Sleep(1 * time.Second)

	fmt.Println("  已终止所有 supervisord 进程")
	return nil
}
