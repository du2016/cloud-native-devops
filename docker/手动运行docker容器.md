# 创建文件系统

mkdir ./rootfs/{merged,diff,work}

merged: 挂载点
diff 是upper
work 是work

mount -t overlay overlay -o lowerdir=./layer1:./layer2,upperdir=./rootfs/diff,workdir=./rootfs/work ./rootfs/merged