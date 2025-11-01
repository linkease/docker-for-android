
检测硬盘：
* e2fsck -f /dev/block/nvme1n1p2
* blkid /dev/block/nvme1n1p2
* cat /proc/partitions | awk '$4 ~ /nvme/ {printf "%-20s %10d MB\n", $4, $3/1024}'
* dumpsys mount

检测分区是否挂载：
```
sm list-volumes all
private mounted null
public:259,1 unmountable 
public:259,2 mounted 1234-ABCD
public:259,3 unmountable 
emulated;0 mounted null
```

# 备份重要数据后，重新创建文件系统
```
umount /dev/block/nvme1n1p2
mke2fs -t ext4 /dev/block/nvme1n1p2
```

# 然后重新挂载
* mount -t ext4 /dev/block/nvme1n1p2 /mnt/nvme_large

# 格式化 public:259,3
* sm format public:259,3

# 查看你关心的 nvme1n1p2 分区
ls -l /dev/block/nvme1n1p2
cat /proc/partitions | grep 259,5

# 查看为什么无法挂载
dmesg | grep nvme1n1p2
blkid /dev/block/nvme1n1p2

# 查看存储卷的详细状态
sm has-disk 259,5
sm get-disk 259,5

# 创建新分区，使用所有剩余空间
sgdisk --new 0:0:0 /dev/block/nvme1n1

# 查看分区表确认
sgdisk --print /dev/block/nvme1n1

# 自动创建最大可能的新分区（最简单的方法）
sgdisk --largest-new 0 /dev/block/nvme1n1

# 查看结果
sgdisk --print /dev/block/nvme1n1

sgdisk --info 1 /dev/block/nvme1n1
mkfs.ext4 /dev/block/nvme1n1p4

# 强制重新扫描
sm set-force-adoptable true
sm forget all

# 永久解决方案：添加分区标签
tune2fs -L "userdata" /dev/block/nvme0n1p4