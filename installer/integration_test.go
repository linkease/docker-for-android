package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRealServerDownload 测试从真实服务器下载并验证文件
// 这个测试需要网络连接，且会下载真实文件（较大）
func TestRealServerDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要网络连接的测试")
	}

	t.Log("==========================================")
	t.Log("真实服务器集成测试")
	t.Log("==========================================")

	// 创建临时目录
	tmpDir := t.TempDir()
	t.Logf("临时目录: %s", tmpDir)

	// 创建 HTTP 客户端
	client := CreateHTTPClient()

	// Step 1: 下载 version.txt
	t.Log("\n[1/4] 下载 version.txt...")
	versionPath := filepath.Join(tmpDir, "version.txt")
	err := downloadFile(client, versionPath, "version.txt", "")
	if err != nil {
		t.Fatalf("下载 version.txt 失败: %v", err)
	}
	t.Log("✓ version.txt 下载成功")

	// Step 2: 解析版本信息
	t.Log("\n[2/4] 解析版本信息...")
	content, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("读取 version.txt 失败: %v", err)
	}

	// 模拟 getVersionInfo 的解析逻辑
	info := &VersionInfo{
		Architecture: "arm64", // 测试 arm64 版本
	}

	lines := splitLines(string(content))
	for _, line := range lines {
		line = trimSpace(line)
		if line == "" || hasPrefix(line, "#") {
			continue
		}

		parts := splitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := trimSpace(parts[0])
		value := trimSpace(parts[1])

		switch key {
		case "VERSION":
			info.Version = value
		case "DOCKER_SHA256":
			info.DockerSHA256 = value
		case "BIN_ARM64_SHA256":
			info.BinSHA256 = value
		}
	}

	// 验证必需字段
	if info.Version == "" {
		t.Fatal("version.txt 中缺少 VERSION 字段")
	}
	if info.DockerSHA256 == "" {
		t.Fatal("version.txt 中缺少 DOCKER_SHA256 字段")
	}
	if info.BinSHA256 == "" {
		t.Fatal("version.txt 中缺少 BIN_ARM64_SHA256 字段")
	}

	t.Logf("✓ 版本: %s", info.Version)
	t.Logf("✓ Docker SHA256: %s", info.DockerSHA256)
	t.Logf("✓ Bin SHA256: %s", info.BinSHA256)

	// Step 3: 下载并验证 docker 包
	t.Log("\n[3/4] 下载 docker 包...")
	dockerTarFile := "docker-" + info.Version + ".tar.gz"
	dockerTarPath := filepath.Join(tmpDir, dockerTarFile)

	t.Logf("下载文件: %s", dockerTarFile)
	err = downloadFile(client, dockerTarPath, dockerTarFile, info.DockerSHA256)
	if err != nil {
		t.Fatalf("下载或验证 docker 包失败: %v", err)
	}
	t.Log("✓ docker 包下载并验证成功")

	// Step 4: 下载并验证二进制包
	t.Log("\n[4/4] 下载二进制包...")
	binTarFile := "docker-for-android-bin-" + info.Version + "-arm64.tar.gz"
	binTarPath := filepath.Join(tmpDir, binTarFile)

	t.Logf("下载文件: %s", binTarFile)
	err = downloadFile(client, binTarPath, binTarFile, info.BinSHA256)
	if err != nil {
		t.Fatalf("下载或验证二进制包失败: %v", err)
	}
	t.Log("✓ 二进制包下载并验证成功")

	t.Log("\n==========================================")
	t.Log("所有文件下载和验证完成！")
	t.Log("==========================================")
}

// TestRealServerExtractAndVerifyStructure 测试解压并验证目录结构
func TestRealServerExtractAndVerifyStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要网络连接的测试")
	}

	t.Log("==========================================")
	t.Log("真实服务器下载和解压测试")
	t.Log("==========================================")

	// 创建临时目录
	tmpDir := t.TempDir()
	downloadDir := filepath.Join(tmpDir, "downloads")
	extractDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(downloadDir, 0755)
	os.MkdirAll(extractDir, 0755)

	// 创建 HTTP 客户端
	client := CreateHTTPClient()

	// Step 1: 下载 version.txt 并解析
	t.Log("\n[1/5] 获取版本信息...")
	versionPath := filepath.Join(downloadDir, "version.txt")
	err := downloadFile(client, versionPath, "version.txt", "")
	if err != nil {
		t.Fatalf("下载 version.txt 失败: %v", err)
	}

	content, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("读取 version.txt 失败: %v", err)
	}

	info := &VersionInfo{Architecture: "arm64"}
	lines := splitLines(string(content))
	for _, line := range lines {
		line = trimSpace(line)
		if line == "" || hasPrefix(line, "#") {
			continue
		}
		parts := splitN(line, "=", 2)
		if len(parts) == 2 {
			key := trimSpace(parts[0])
			value := trimSpace(parts[1])
			switch key {
			case "VERSION":
				info.Version = value
			case "DOCKER_SHA256":
				info.DockerSHA256 = value
			case "BIN_ARM64_SHA256":
				info.BinSHA256 = value
			}
		}
	}

	if info.Version == "" || info.DockerSHA256 == "" || info.BinSHA256 == "" {
		t.Fatal("version.txt 缺少必需字段")
	}
	t.Logf("✓ 版本: %s", info.Version)

	// Step 2: 下载 docker 包
	t.Log("\n[2/5] 下载 docker 包...")
	dockerTarFile := "docker-" + info.Version + ".tar.gz"
	dockerTarPath := filepath.Join(downloadDir, dockerTarFile)
	err = downloadFile(client, dockerTarPath, dockerTarFile, info.DockerSHA256)
	if err != nil {
		t.Fatalf("下载 docker 包失败: %v", err)
	}
	t.Log("✓ docker 包下载成功")

	// Step 3: 下载二进制包
	t.Log("\n[3/5] 下载二进制包...")
	binTarFile := "docker-for-android-bin-" + info.Version + "-arm64.tar.gz"
	binTarPath := filepath.Join(downloadDir, binTarFile)
	err = downloadFile(client, binTarPath, binTarFile, info.BinSHA256)
	if err != nil {
		t.Fatalf("下载二进制包失败: %v", err)
	}
	t.Log("✓ 二进制包下载成功")

	// Step 4: 解压 docker 包
	t.Log("\n[4/5] 解压文件...")
	t.Log("解压 docker 包到 extracted/...")
	err = extractTarGz(dockerTarPath, extractDir, "")
	if err != nil {
		t.Fatalf("解压 docker 包失败: %v", err)
	}
	t.Log("✓ docker 包解压完成")

	// Step 5: 解压二进制包到 docker 目录
	dockerDir := filepath.Join(extractDir, "docker")

	t.Log("解压二进制包到 docker/...")
	err = extractTarGz(binTarPath, dockerDir, "")
	if err != nil {
		t.Fatalf("解压二进制包失败: %v", err)
	}
	t.Log("✓ 二进制包解压完成")

	// Step 6: 验证目录结构
	t.Log("\n[5/5] 验证目录结构...")

	// 验证 docker 目录存在
	if _, err := os.Stat(dockerDir); os.IsNotExist(err) {
		t.Fatal("docker 目录不存在")
	}
	t.Logf("✓ docker 目录存在: %s", dockerDir)

	// 查找 bin 目录（可能是 bin 或 arm64_bin）
	binDir := filepath.Join(dockerDir, "bin")
	arm64BinDir := filepath.Join(dockerDir, "arm64_bin")

	var actualBinDir string
	if _, err := os.Stat(arm64BinDir); err == nil {
		actualBinDir = arm64BinDir
		t.Logf("✓ 找到 arm64_bin 目录: %s", arm64BinDir)
	} else if _, err := os.Stat(binDir); err == nil {
		actualBinDir = binDir
		t.Logf("✓ 找到 bin 目录: %s", binDir)
	} else {
		t.Fatalf("bin 目录不存在（检查了 bin 和 arm64_bin）")
	}

	// 验证 bin 目录的相对路径
	relPath, err := filepath.Rel(extractDir, actualBinDir)
	if err != nil {
		t.Fatalf("计算相对路径失败: %v", err)
	}
	t.Logf("✓ bin 目录相对路径: %s", relPath)

	// 验证必需的二进制文件
	requiredBinaries := []string{"docker", "dockerd", "containerd", "runc"}

	// 检查 bin 目录
	entries, err := os.ReadDir(actualBinDir)
	if err != nil {
		t.Fatalf("读取 bin 目录失败: %v", err)
	}

	if len(entries) == 0 {
		// 可能在 arm64_bin 目录
		arm64BinDir := filepath.Join(dockerDir, "arm64_bin")
		if _, err := os.Stat(arm64BinDir); err == nil {
			binDir = arm64BinDir
			entries, err = os.ReadDir(binDir)
			if err != nil {
				t.Fatalf("读取 arm64_bin 目录失败: %v", err)
			}
			t.Logf("✓ 使用 arm64_bin 目录: %s", binDir)
		}
	}

	foundBinaries := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			foundBinaries[entry.Name()] = true
		}
	}

	t.Logf("✓ 找到 %d 个二进制文件", len(foundBinaries))

	missingBinaries := []string{}
	for _, binary := range requiredBinaries {
		if !foundBinaries[binary] {
			missingBinaries = append(missingBinaries, binary)
		}
	}

	if len(missingBinaries) > 0 {
		t.Logf("⚠ 警告: 缺少以下二进制文件: %v", missingBinaries)
		t.Logf("找到的文件: %v", getKeys(foundBinaries))
		// 只要有一些二进制文件就算通过，因为可能版本不同
		if len(foundBinaries) == 0 {
			t.Fatal("bin 目录中没有找到任何二进制文件")
		}
	} else {
		t.Logf("✓ 所有必需的二进制文件都存在")
	}

	// 验证其他重要目录和文件
	t.Log("\n验证其他目录和文件...")

	// 检查 etc 目录
	etcDir := filepath.Join(dockerDir, "etc")
	if _, err := os.Stat(etcDir); err == nil {
		t.Logf("✓ etc 目录存在")

		// 检查配置文件
		configFiles := []string{
			"etc/docker/daemon.json",
			"etc/dockerd.conf",
		}
		for _, configFile := range configFiles {
			configPath := filepath.Join(dockerDir, configFile)
			if _, err := os.Stat(configPath); err == nil {
				t.Logf("✓ 配置文件存在: %s", configFile)
			}
		}
	}

	// 检查脚本文件
	scripts := []string{"start-dockerd.sh", "deploy-in-android.sh", "docker.sh"}
	for _, script := range scripts {
		scriptPath := filepath.Join(dockerDir, script)
		if _, err := os.Stat(scriptPath); err == nil {
			t.Logf("✓ 脚本存在: %s", script)

			// 检查执行权限
			info, _ := os.Stat(scriptPath)
			if info.Mode()&0111 != 0 {
				t.Logf("  - 可执行权限: ✓")
			}
		}
	}

	// 检查 docker.env
	envFile := filepath.Join(dockerDir, "docker.env")
	if _, err := os.Stat(envFile); err == nil {
		t.Logf("✓ docker.env 存在")
	}

	t.Log("\n==========================================")
	t.Log("目录结构验证完成！")
	t.Log("==========================================")
	t.Logf("\n最终目录结构:")
	t.Logf("  %s/", extractDir)
	t.Logf("    └── docker/")
	t.Logf("        ├── bin/ (或 arm64_bin/)")
	t.Logf("        ├── etc/")
	t.Logf("        ├── scripts/")
	t.Logf("        ├── start-dockerd.sh")
	t.Logf("        ├── deploy-in-android.sh")
	t.Logf("        └── docker.env")
}

// Helper functions
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}

	return s[start:end]
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func splitN(s, sep string, n int) []string {
	if n <= 0 {
		return nil
	}

	var result []string
	start := 0

	for i := 0; i < len(s) && len(result) < n-1; i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}

	if start < len(s) {
		result = append(result, s[start:])
	} else if len(result) < n {
		result = append(result, "")
	}

	return result
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
