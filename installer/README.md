# Docker for Android Installer

这是一个用于在 Android 系统上安装 Docker 的 Go 语言安装程序。

## 代码结构

程序采用模块化设计，代码分为以下文件：

- **`install-in-docker.go`** - 主程序，控制安装流程
- **`cmd.go`** - 命令行执行模块（硬盘检测、架构识别、脚本执行）
- **`download.go`** - 下载模块（HTTP 下载、SHA256 校验）
- **`extract.go`** - 解压模块（tar.gz 解压、权限设置）

详细的代码架构说明请参考 [ARCHITECTURE.md](ARCHITECTURE.md)。

## 测试

已实现完整的单元测试套件，覆盖所有可在开发机器上测试的功能。

### 运行测试

```bash
# 运行所有测试
go test -v

# 运行性能基准测试
go test -bench=. -benchmem

# 查看测试覆盖率
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 测试覆盖

- ✅ 23 个单元测试
- ✅ 3 个性能基准测试
- ✅ ~70-75% 代码覆盖率
- ✅ 所有测试通过

详细的测试文档请参考 [TESTING.md](TESTING.md)。

## 功能特性

- 自动检测 Android 系统上的外置硬盘挂载点
- 从 CDN 或服务器下载 Docker 安装包
- SHA256 校验确保文件完整性
- 自动解压并部署到正确位置
- 实时显示下载进度和安装日志
- 支持 arm64 和 x86_64 架构

## 构建方法

### 为 Android arm64 构建

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o install-docker-arm64 install-in-docker.go
```

### 为 Android x86_64 构建

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o install-docker-x86_64 install-in-docker.go
```

## 使用方法

1. 将编译好的二进制文件推送到 Android 设备：

```bash
adb push install-docker-arm64 /data/local/tmp/install-docker
adb shell chmod +x /data/local/tmp/install-docker
```

2. 通过 adb shell 进入设备并执行安装：

```bash
adb shell
su  # 需要 root 权限
cd /data/local/tmp
./install-docker
```

## 安装流程

安装程序会自动完成以下步骤：

1. **检测硬盘** - 检查是否有 ext4 格式的外置硬盘挂载
2. **获取版本** - 从服务器获取最新版本信息
3. **下载文件** - 下载 Docker 核心包和架构特定的二进制包
   - 优先从 CDN 下载
   - CDN 失败时自动切换到源服务器
   - 自动进行 SHA256 校验
4. **解压安装** - 解压文件到正确位置
   - Docker 配置和脚本 -> `/data/local/docker/`
   - 二进制文件 -> `/data/local/docker/bin/`
5. **自动部署** - 调用部署脚本完成环境配置和服务启动

## 系统要求

- Android 设备需要 root 权限
- 需要接入并格式化为 ext4 格式的外置硬盘
- 硬盘需要挂载到 `/mnt/media_rw/` 路径
- 网络连接用于下载安装包

## 下载源

- **CDN**: https://fw.kspeeder.com/binary/docker-for-android
- **源服务器**: https://fw.koolcenter.com/binary/docker-for-android

注意：version.txt 文件不会被 CDN 缓存，始终从源获取最新版本信息。

## 文件结构

安装完成后的文件结构：

```
/data/local/docker/          # Docker 主目录
├── bin/                     # Docker 二进制文件
│   ├── docker
│   ├── dockerd
│   ├── containerd
│   └── ...
├── etc/                     # 配置文件
├── deploy-in-android.sh     # 部署脚本
├── start-dockerd.sh         # 启动脚本
└── docker.env               # 环境变量

/mnt/media_rw/xxx/           # 硬盘挂载点
├── opt/dockerd/docker/      # Docker 数据目录
├── Cache/                   # 缓存目录
│   └── Kspeeder/           # Kspeeder 缓存
└── Configs/                 # 配置目录
    └── DPanel/             # DPanel 配置
```

## 故障排查

### 找不到硬盘挂载点

确保：
- 硬盘已正确接入设备
- 硬盘已格式化为 ext4 格式
- 硬盘已被系统挂载到 `/mnt/media_rw/` 下

### 下载失败

- 检查网络连接
- 确认防火墙未阻止访问
- 程序会自动尝试 CDN 和源服务器

### 权限错误

- 确保以 root 权限运行
- 检查 `/data/local/` 目录权限

## 开发说明

installer 程序使用纯 Go 标准库开发，无外部依赖，可以静态编译为单个二进制文件，方便在 Android 环境中使用。

主要模块：
- 硬盘检测
- HTTP 下载（带进度显示）
- SHA256 校验
- tar.gz 解压
- 脚本执行（实时日志输出）
