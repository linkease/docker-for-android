# docker-for-android

目前暂时支持 arm64/aarch64 架构的部分 Android 设备。
本项目优先支持 [EasePi-A2](https://www.koolcenter.com/t/topic/5204)

## 本项目使用原生 Docker 二进制

使用原生 Docker 二进制，没对 Docker 二进制进行任何修改，安全可靠。
目前的下载地址为：
* https://download.docker.com/linux/static/stable/aarch64/docker-28.0.1.tgz
 
 其他 Android 系统支持可以讨论：
 * https://github.com/linkease/docker-for-android/issues/1
 * https://www.koolcenter.com/t/topic/5228

## 版本与发布

- 当前版本：[查看版本](VERSION)
- 源服务器（原始）：`https://fw.koolcenter.com/binary/docker-for-android`
- CDN（.txt 不缓存，始终最新）：`https://fw.kspeeder.com/binary/docker-for-android`

### 打包 arm64 二进制

在仓库根目录执行：

```
make arm64
```

会将 `docker/arm64_bin` 打包为 `release/docker-for-android-bin-28.0.1.10-arm64.tar.gz`，并生成对应的 `.sha256` 校验文件。

生成 `version.txt`：

```
make version
```

或者一键构建：

```
make build-release
```

`version.txt` 中会包含版本号与 arm64/x86_64 包名与 sha256，用于脚本在 CDN 上获取并校验下载。

## 从 CDN 更新 arm64_bin

开发调试时，可以使用脚本从 CDN 获取 `version.txt`，并根据其中的信息下载并覆盖 `docker/arm64_bin`：

```
scripts/update_arm64_bin.sh
```

可通过环境变量覆盖默认端点：

```
CDN_URL=https://fw.kspeeder.com/binary/docker-for-android \
ORIGIN_SERVER_URL=https://fw.koolcenter.com/binary/docker-for-android \
scripts/update_arm64_bin.sh
```

脚本会：

- 从 CDN 下载 `version.txt`（.txt 不被 CDN 缓存，始终最新）
- 解析 `ARM64_PACKAGE` 与 `ARM64_SHA256`
- 从 CDN 下载 tar.gz（失败则回退到源服务器）
- 校验 sha256
- 解压到 `docker/` 下，覆盖 `docker/arm64_bin`

## 在 Android 设备上部署

### 方法一：使用自动安装程序（推荐）

最简单的安装方式是使用我们提供的安装程序，它会自动从 CDN/服务器下载并安装所有必需的文件。
[点击下载地址](https://fw0.koolcenter.com/binary/docker-for-android/install-docker-arm64)

也可以查看下面的编译步骤自动编译

#### 前置条件

1. Android 设备已获取 root 权限
2. 已连接 adb
3. **必须已挂载 ext4 格式的外置硬盘**

#### 快速安装

现在把安装软件下载到电脑上，再用 adb 推送到设备上：
* adb connect IP:5555
* adb root
* adb push ./install-docker-arm64 /data/local/
* adb shell

到 shell 里面运行：/data/local/install-docker-arm64

#### 安装程序功能

安装程序会自动完成：

1. **检测硬盘** - 检查 ext4 格式的外置硬盘
2. **下载文件** - 从 CDN/服务器下载最新版本
   - 优先使用 CDN（`https://fw.kspeeder.com/binary/docker-for-android`）
   - CDN 失败时自动切换到源服务器
   - 自动 SHA256 校验确保文件完整性
3. **解压安装** - 解压到正确位置
4. **自动部署** - 执行部署脚本，配置环境并启动服务
5. **安装 DPanel** - 自动部署容器管理面板

整个过程完全自动化，无需手动操作。

### 方法二：手动部署

### 前置条件

1. Android 设备已获取 root 权限
2. 已连接 adb，或者在设备上通过终端执行
3. **必须已挂载 ext4 格式的外置硬盘**（脚本检测不到硬盘会直接退出）

### 部署步骤

下面的教程的版本号请自行根据实际修改

1. 构建 docker 发布包：

```bash
make build-release
```

2. 将 `release/docker-28.0.1.10.tar.gz` 推送到设备：

```bash
adb push release/docker-28.0.1.10.tar.gz /data/local/
```

3. 解压到 `/data/local/docker`：

```bash
adb shell
cd /data/local
tar -xzf docker-28.0.1.10.tar.gz
```

4. 运行部署脚本：

```bash
cd /data/local/docker
sh deploy-in-android.sh
```

### 部署脚本功能

`deploy-in-android.sh` 会自动完成以下操作：

1. **检测硬盘挂载点并设置 DISK_ROOT**
   - 自动检测 ext4 格式的外置硬盘（`/dev/block/vold/public:259,1`）
   - 设置硬盘根目录 `DISK_ROOT`（例如：`/mnt/media_rw/xxx`）
   - 创建目录结构：
     - Docker 数据：`$DISK_ROOT/dockerd/docker`
     - 缓存目录：`$DISK_ROOT/Cache/Kspeeder`
     - DPanel 配置：`$DISK_ROOT/Configs/DPanel`
   - **如果未检测到硬盘，脚本会直接退出**

2. **配置环境变量**
   - 只需更新 `docker.env` 中的 `DISK_ROOT`
   - 其他变量（`DOCKER_DATA_ROOT`、`DISK_CACHE`）会自动派生

3. **配置 kspeeder**
   - 使用环境变量 `${DISK_CACHE}/Kspeeder` 作为缓存路径
   - 缓存路径统一在 `docker.env` 中管理

4. **启动 Docker 并部署 DPanel**
   - 启动 Docker 服务
   - 等待 Docker 就绪（最多 60 秒）
   - 拉取并启动 DPanel 容器管理面板
   - DPanel 监听在 `:8080` 端口
   - DPanel 配置数据保存在 `$DOCKER_DATA_ROOT/../Configs/DPanel`

### 访问 DPanel

部署完成后，在浏览器中访问：

```
http://<Android设备IP>:8080
```

### 环境变量说明

所有环境变量在 `docker.env` 中统一管理，采用统一的硬盘根目录设计：

- `DISK_ROOT`：硬盘根目录（例如：`/mnt/media_rw/xxx`），**唯一需要设置的变量**
- `DOCKER_DATA_ROOT`：Docker 数据目录，**自动派生**为 `${DISK_ROOT}/dockerd/docker`
- `DISK_CACHE`：缓存根目录，**自动派生**为 `${DISK_ROOT}/Cache`
  - kspeeder 缓存：`${DISK_CACHE}/Kspeeder`
- `DOCKER_ROOT`：Docker 安装根目录（`/data/local/docker`）
- `DOCKER_HOST`：Docker socket 路径
- `HOME`：Docker 运行时的 HOME 目录

#### 路径派生规则

**只需设置 `DISK_ROOT`，其他路径自动派生：**

```bash
# 在 docker.env 中
export DISK_ROOT=/mnt/media_rw/xxx
export DOCKER_DATA_ROOT=${DISK_ROOT}/dockerd/docker    # 自动派生
export DISK_CACHE=${DISK_ROOT}/Cache                   # 自动派生
```

实际目录结构：

```
DISK_ROOT=/mnt/media_rw/xxx
├── dockerd/
│   └── docker/          → DOCKER_DATA_ROOT (Docker 数据)
├── Cache/
│   └── Kspeeder/        → kspeeder 缓存
└── Configs/
    └── DPanel/          → DPanel 配置目录
```

这种设计的优势：
- **极简配置**：只需设置一个 `DISK_ROOT` 变量
- **自动派生**：所有路径在 shell 中自动展开
- **统一管理**：所有软件都延续这个硬盘位置，路径统一
- **便于迁移**：整个系统的数据都在 `DISK_ROOT` 下，备份和迁移更方便

