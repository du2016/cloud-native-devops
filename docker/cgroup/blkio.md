块IO控制器

# 总览

cgroup子系统blkio实现了块io控制器。有各种IO控制策略（例如比例带宽，最大带宽）
在存储层次结构中的叶节点和中间节点都可以。计划对blkio控制器使用相同的基于cgroup的管理界面
并根据用户选项在后台切换IO策略。

一种IO控制策略是限制策略，可用于指定设备上的IO速率上限。
该政策在通用块层，可以在叶节点以及更高节点上使用设备映射程序之类的逻辑设备。

# 使用

## 节流/上限策略

- 启用块IO控制器 CONFIG_BLK_CGROUP = y

- 在块层启用节流 CONFIG_BLK_DEV_THROTTLING = y

- 挂载blkio控制器（请参阅cgroups.txt，为什么需要cgroup？）

mount -t cgroup -o blkio none /sys/fs/cgroup/blkio

- 在特定设备上为根组指定带宽速率。格式
  策略为'<major>:<minor>  <bytes_per_second>'。

echo "8:16  1048576" > /sys/fs/cgroup/blkio/blkio.throttle.read_bps_device

上面将根组的读取限制为每秒1MB在具有主要/次要编号8:16的设备上。

- 运行dd读取文件，然后查看速率是否已限制为1MB / s。

＃dd iflag =直接if = / mnt / common / zerofile of = / dev / null bs = 4K count = 1024
中有1024 + 0条记录
1024 + 0条记录
已复制4194304字节（4.2 MB），4.0001 s，1.0 MB / s

可以使用blkio.throttle.write_bps_device文件设置写入限制。

# 分层的cgroup

限流实现了多层次结构；然而，
限流的层次支持时当且仅当"sane_behavior"是
从cgroup端启用，目前这是一个开发选项，
不公开。

如果有人创建了如下的层次结构。

			root
			/  \
		     test1 test2
			|
		     test3

用"sane_behavior"进行限流将处理层次结构正确。对于限流，所有限制都适用
到整个子树，而所有统计信息都在IO本地
由该cgroup中的任务直接生成。

从cgroup中的限流没有启动“sane_behavior”
几乎在同等水平上对待所有group,如下。

				pivot
			     /  /   \  \
			root  test1 test2  test3

各种用户可见的配置选项
===================================
CONFIG_BLK_CGROUP
	-块IO控制器。

CONFIG_DEBUG_BLK_CGROUP
	- 调试帮助。将在cgroup中显示了一些其他统计文件
	  如果启用此选项。

CONFIG_BLK_DEV_THROTTLING
	- 在块层启用块设备节流支持。

cgroup文件的详细信息
=======================
比例权重策略文件
--------------------------------
- blkio.weight
	- 指定每个cgroup的权重。这是组的默认权重
	  在所有设备上，直到并且除非被每个设备规则覆盖。
	  （请参阅blkio.weight_device）。
	  当前允许的权重范围是10到1000。

- blkio.weight_device
	- 可以使用此接口为每个设备、每个cgroup指定规则。
	  这些规则将覆盖指定的组通过blkio.weight设置权重的默认值

	  以下是格式。
```
	  ＃echo dev_maj:dev_minor weight > blkio.weight_device
	  在此cgroup的/dev/sdb（8:16）上配置weight = 300
	  # echo 8:16 300 > blkio.weight_device
      # cat blkio.weight_device
      dev     weight
      8:16    300

	  Configure weight=500 on /dev/sda (8:0) in this cgroup
	  # echo 8:0 500 > blkio.weight_device
	  # cat blkio.weight_device
	  dev     weight
	  8:0     500
	  8:16    300

	  Remove specific weight for /dev/sda in this cgroup
	  # echo 8:0 0 > blkio.weight_device
	  # cat blkio.weight_device
	  dev     weight
	  8:16    300
```

- blkio.leaf_weight [_device]
	- 等效于blkio.weight [_device]
          确定给定cgroup中有多少权重任务
          与cgroup的子cgroup竞争。

- blkio.time
	-每个设备分配给cgroup的磁盘时间（以毫秒为单位）。前两个字段指定设备的主要和次要编号，第三个字段指定分配给组的磁盘时间毫秒。

- blkio.sectors
	- 组从磁盘传输到磁盘或从磁盘传输的扇区数。前两个字段指定设备的主要和次要编号，
	  第三个字段指定设备的扇区数。

- blkio.io_service_bytes
	- 组从磁盘传输到磁盘或从磁盘传输的字节数。这些
	  按操作类型进一步划分-读取或写入，同步
	  或异步。前两个字段指定
	  设备，第三个字段指定操作类型，第四个字段
	  指定字节数。

- blkio.io_serviced
	- 该组向磁盘发布的IO（bio）数量。这些
	  按操作类型进一步划分-读取或写入，同步
	  或异步。前两个字段指定
	  设备，第三个字段指定操作类型，第四个字段
	  指定IO的数量。

- blkio.io_service_time
	- 从请求分派到请求完成之间的总时间
	  对于此cgroup完成的IO。这是十亿分之一秒
	  对闪存设备也有意义。对于队列深度为1的设备，
	  该时间代表实际服务时间。当queue_depth> 1时
	  这不再是正确的，因为可能无法按顺序提供请求。这个
	  可能导致给定IO的服务时间包括服务时间
	  多个IO出现故障时可能导致总数
	  io_service_time>实际经过的时间。这段时间再除以
	  操作类型-读取或写入，同步或异步。前两个领域
	  指定设备的主要和次要编号，第三个字段
	  指定操作类型，第四个字段指定
	  io_service_time（以ns为单位）。

- blkio.io_wait_time
	- 此cgroup的IO花费在等待中的总时间
	  服务的调度程序队列。这可以大于总时间
	  由于它是所有IO的累积io_wait_time，因此已过去。这不是一个
	  cgroup等待总时间的量度，而是
	  其各个IO的wait_time。对于queue_depth> 1的设备
	  此指标不包括等待服务一次所花费的时间
	  IO已分派到设备，但直到它得到实际维修为止
	  （由于对请求的重新排序，此处可能会有时间滞后
	  设备）。这是十亿分之一秒，使其对闪存有意义
	  设备。该时间进一步除以操作类型-
	  读取或写入，同步或异步。前两个字段指定专业和
	  设备的次编号，第三个字段指定操作类型
	  第四个字段以ns为单位指定io_wait_time。

- blkio.io_merged
	- 合并到属于此请求的BIOS /请求的总数
	  cgroup。除以操作类型-读或
	  写入，同步或异步。

- blkio.io_queued
	- 为此cgroup在任何给定时刻排队的请求总数。由操作类型决定-读或
	  写入，同步或异步。

- blkio.avg_queue_size
	- 仅在CONFIG_DEBUG_BLK_CGROUP = y时启用调试辅助。
	  此cgroup在整个过程中的平均队列大小
	  cgroup的存在。每次
	  此cgroup的队列获取时间片。

- blkio.group_wait_time
	- 仅在CONFIG_DEBUG_BLK_CGROUP = y时启用调试辅助。
	  这是cgroup自变得繁忙以来必须等待的时间
	  （即，从0到1排队的请求）以获取以下项之一的时间片
	  它的队列。这与io_wait_time不同，后者是
	  该cgroup中每个IO花费的时间总计
	  在调度程序队列中等待。以纳秒为单位。如果这是
	  在cgroup处于等待（用于时间片）状态时读取stat
	  只会报告直到最后一次累积的group_wait_time
	  有时间片，将不包括当前增量。

- blkio.empty_time
	- 仅在CONFIG_DEBUG_BLK_CGROUP = y时启用调试辅助。
	  这是cgroup在没有任何待处理的情况下花费的时间
	  不提供服务时请求，即不包含任何时间
	  空闲了cgroup的队列之一。这是在
	  纳秒。如果在cgroup处于空状态时读取了此内容，
	  统计信息将仅报告直到最后一次为止的empty_time
	  它有一个待处理的请求，并且将不包括当前增量。

- blkio.idle_time
	-仅在CONFIG_DEBUG_BLK_CGROUP = y时启用调试辅助。
	  这是IO调度程序在空闲状态下花费的时间
	  给cgroup期望比现有请求更好的请求
	  来自其他队列/ cgroup。以纳秒为单位。如果这读
	  当cgroup处于空闲状态时，统计信息将仅报告
	  idle_time累积到最后一个空闲时间，将不包括
	  当前增量。

- blkio.dequeue
	-仅在CONFIG_DEBUG_BLK_CGROUP = y时启用调试辅助。这个
	  提供有关组出队多少次的统计信息
	  从设备的服务树中。前两个字段指定专业
	  设备的次编号，第三个字段指定编号
	  从特定设备出队的次数。

-blkio。* _ recursive
	-各种统计信息的递归版本。这些文件显示了
          与非递归对应的信息相同，但
          包括所有后代cgroup的统计信息。

节流/上传 策略文件
-----------------------------------
- blkio.throttle.read_bps_device
	- 指定设备读取速率的上限。IO率为
	  以每秒字节数指定。规则是针对每个设备的。以下是
	  格式。

  echo "<major>:<minor>  <rate_bytes_per_second>" > /cgrp/blkio.throttle.read_bps_device

- blkio.throttle.write_bps_device
	- 指定写入设备的速率上限。IO率为
	  以每秒字节数指定。规则是针对每个设备的。以下是
	  格式。

  echo "<major>:<minor>  <rate_bytes_per_second>" > /cgrp/blkio.throttle.write_bps_device

- blkio.throttle.read_iops_device
	- 指定设备读取速率的上限。IO率为
	  每秒IO中指定的值。规则是针对每个设备的。以下是
	  格式。

  echo "<major>:<minor>  <rate_io_per_second>" > /cgrp/blkio.throttle.read_iops_device

- blkio.throttle.write_iops_device
	- 指定写入设备的速率上限。IO率为
	  以io每秒指定。规则是针对每个设备的。以下是
	  格式。

  echo "<major>:<minor>  <rate_io_per_second>" > /cgrp/blkio.throttle.write_iops_device

注：如果设备同时指定BW和IOPS的规则，则是IO受到所有约束。

- blkio.throttle.io_serviced
	- 该组向磁盘发布的IO（生物）数量。这些
	  按操作类型进一步划分-读取或写入，同步
	  或异步。前两个字段指定
	  设备，第三个字段指定操作类型，第四个字段
	  指定IO的数量。

- blkio.throttle.io_service_bytes
	- 组从磁盘传输到磁盘或从磁盘传输的字节数。这些
	  按操作类型进一步划分-读取或写入，同步
	  或异步。前两个字段指定
	  设备，第三个字段指定操作类型，第四个字段
	  指定字节数。

各种政策中的共同文件
-----------------------------------
-blkio.reset_stats
	- 写入整数重置该cgroup的所有统计信息