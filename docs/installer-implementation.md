# Docker for Android - Installer 实现总结

## 项目概述

为 Docker for Android 项目实现了一个完整的自动化安装程序，支持从 CDN/服务器自动下载、校验、解压并部署 Docker 到 Android 设备。

## 实现的功能

### 1. Installer 程序 (`installer/install-in-docker.go`)

一个纯 Go 实现的安装程序，无外部依赖，可静态编译为单个二进制文件。

**主要功能：**

- **硬盘检测** - 自动检测 Android 设备上的 ext4 外置硬盘挂载点
- **版本获取** - 从服务器下载 version.txt 获取最新版本和 SHA256 信息
- **架构识别** - 自动识别 arm64 或 x86_64 架构
- **智能下载** - 从 CDN 或服务器下载文件
  - 优先使用 CDN (`https://fw.kspeeder.com/binary/docker-for-android`)
  - CDN 失败自动回退到源服务器 (`https://fw.koolcenter.com/binary/docker-for-android`)
  - 实时显示下载进度（百分比和 MB）
- **安全校验** - 使用 SHA256 验证下载文件完整性
- **自动解压** - 解压 tar.gz 到正确位置
- **自动部署** - 调用 deploy-in-android.sh 完成环境配置
- **实时日志** - 实时显示部署脚本的执行日志

**支持的下载文件：**
- `docker-{version}.tar.gz` - Docker 核心文件（配置、脚本等）
- `docker-for-android-bin-{version}-arm64.tar.gz` - ARM64 二进制文件
- `docker-for-android-bin-{version}-x86_64.tar.gz` - x86_64 二进制文件

### 2. Makefile 增强

**新增目标：**

```makefile
make installer    # 构建 arm64 和 x86_64 版本的 installer
make version      # 生成 version.txt（包含版本号和 SHA256）
make build-release # 完整构建：包、版本文件、installer
```

**version.txt 格式：**

```ini
VERSION=28.0.1.10
DOCKER_SHA256=<docker包的sha256>
BIN_ARM64_SHA256=<arm64二进制包的sha256>
BIN_X86_64_SHA256=<x86_64二进制包的sha256>
```

### 3. 快速安装脚本 (`scripts/quick-install.sh`)

一键安装脚本，自动：
- 检测 Android 设备连接
- 识别设备架构
- 推送对应的 installer 到设备
- 引导用户完成安装

### 4. 文档

**installer/README.md** - Installer 使用文档
- 构建方法
- 使用方法
- 系统要求
- 故障排查

**installer/TESTING.md** - 测试指南
- 编译测试
- 设备测试
- 错误场景测试
- 调试技巧

**主 README 更新** - 添加自动安装方法说明

## 技术实现细节

### 版本文件设计

**version.txt 特性：**
- `.txt` 后缀不会被 CDN 缓存
- 始终从源服务器获取最新内容
- 包含所有包的 SHA256 校验和

### 下载策略

1. 先尝试 CDN（快速）
2. CDN 失败回退到源服务器（可靠）
3. 每次下载都进行 SHA256 校验
4. 校验失败自动重试下一个源

### 安全性

- 所有下载文件都进行 SHA256 校验
- 使用 HTTPS 协议
- 编译时使用 `-ldflags="-s -w"` 减小体积并移除调试信息

### 用户体验

- 清晰的步骤提示 `[1/5]`, `[2/5]` 等
- 实时进度显示
- 彩色状态标记 `✓`, `✗`, `⏳`
- 详细的错误信息
- 部署脚本日志实时输出

## 使用流程

### 开发者流程

```bash
# 1. 更新 VERSION 文件
echo "DOCKER_VERSION=28.0.1" > VERSION
echo "SUB_VERSION=10" >> VERSION

# 2. 构建发布包
make build-release

# 3. 上传到服务器
# 上传 release/ 目录下的所有文件到：
# https://fw.koolcenter.com/binary/docker-for-android/

# 4. 同步到 CDN（version.txt 除外）
```

### 用户安装流程

```bash
# 方法一：使用快速安装脚本（推荐）
./scripts/quick-install.sh

# 方法二：手动安装
adb push release/install-docker-arm64 /data/local/tmp/install-docker
adb shell chmod +x /data/local/tmp/install-docker
adb shell su -c /data/local/tmp/install-docker
```

## 文件结构

```
docker-for-android/
├── installer/
│   ├── install-in-docker.go    # 安装程序源码
│   ├── go.mod                  # Go 模块文件
│   ├── README.md               # 使用文档
│   └── TESTING.md              # 测试文档
├── scripts/
│   └── quick-install.sh        # 快速安装脚本
├── release/                    # 构建输出目录
│   ├── docker-{ver}.tar.gz
│   ├── docker-for-android-bin-{ver}-arm64.tar.gz
│   ├── docker-for-android-bin-{ver}-x86_64.tar.gz
│   ├── install-docker-arm64
│   ├── install-docker-x86_64
│   ├── version.txt
│   └── *.sha256               # 校验文件
├── Makefile                    # 构建脚本
└── VERSION                     # 版本定义
```

## 安装后的系统结构

```
Android 设备:
/data/local/docker/            # Docker 安装目录
├── bin/                       # 二进制文件（从 {arch}_bin 解压）
│   ├── docker
│   ├── dockerd
│   ├── containerd
│   └── ...
├── etc/                       # 配置文件
├── scripts/                   # 脚本
├── deploy-in-android.sh       # 部署脚本
├── start-dockerd.sh           # 启动脚本
└── docker.env                 # 环境变量

/mnt/media_rw/xxx/             # 外置硬盘（DISK_ROOT）
├── opt/dockerd/docker/        # Docker 数据（DOCKER_DATA_ROOT）
├── Cache/
│   └── Kspeeder/              # Kspeeder 缓存
└── Configs/
    └── DPanel/                # DPanel 配置
```

## 环境变量设计

统一的硬盘根目录设计，所有路径从 `DISK_ROOT` 派生：

```bash
export DISK_ROOT=/mnt/media_rw/xxx              # 唯一需要设置
export DOCKER_DATA_ROOT=${DISK_ROOT}/opt/dockerd/docker  # 自动派生
export DISK_CACHE=${DISK_ROOT}/Cache            # 自动派生
```

**优势：**
- 极简配置（只设置一个变量）
- 路径统一管理
- 便于备份和迁移

## 测试建议

1. **编译测试**
   ```bash
   make installer
   file release/install-docker-arm64
   file release/install-docker-x86_64
   ```

2. **设备测试**
   - 在真实 Android 设备上测试
   - 测试网络异常情况
   - 测试 SHA256 校验
   - 测试无硬盘场景

3. **集成测试**
   - 完整安装流程
   - DPanel 部署验证
   - Docker 功能测试

## 后续优化建议

1. **支持断点续传** - 大文件下载中断后可续传
2. **并行下载** - 同时下载多个文件加快速度
3. **镜像站支持** - 添加更多下载源
4. **安装进度持久化** - 记录安装进度，支持恢复
5. **自动更新检测** - 检测已安装版本，提示更新
6. **卸载功能** - 实现完整的卸载流程

## 依赖项

- Go 1.21+ （开发）
- Android 设备 root 权限（运行）
- ext4 格式外置硬盘（运行）
- adb 工具（部署）

## 兼容性

- **架构**: arm64 (aarch64), x86_64 (amd64)
- **Android**: 需要 root 权限的原生 Android 系统
- **硬盘**: 必须是 ext4 格式

## 许可证

遵循项目原有许可证。
