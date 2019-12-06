块IO控制器

# 总览

cgroup子系统blkio实现了块io控制器。
在存储层次结构中的叶节点以及中间节点处似乎都需要各种类型的IO控制策略（如比例BW，最大BW）。
计划是将相同的基于cgroup的管理接口用于blkio控制器，并根据用户选项在后台切换IO策略。

一种IO控制策略是throttling策略，可用于指定设备上的IO速率上限。
该政策在通用块层实现，可以在叶节点以及更高节点上使用设备映射程序之类的逻辑设备。

# 使用

## 节流/上限策略

- 启用块IO控制器 CONFIG_BLK_CGROUP=y

- 在块层启用节流 CONFIG_BLK_DEV_THROTTLING=y

- 挂载blkio控制器（请参阅cgroups.txt，为什么需要cgroup？）

```
mount -t cgroup -o blkio none /sys/fs/cgroup/blkio
```

- 在特定设备上为根组指定带宽速率。策略格式为'<major>:<minor>  <bytes_per_second>'。

```
echo "8:16  1048576" > /sys/fs/cgroup/blkio/blkio.throttle.read_bps_device
```

上面将对具有主要/次要数字8:16的设备上的根组的读取设置1MB/秒的限制

- 运行dd读取文件，然后查看速率是否已限制为1MB/s。

```
# dd iflag=direct if=/mnt/common/zerofile of=/dev/null bs=4K count=1024
1024+0 records in
1024+0 records out
4194304 bytes (4.2 MB) copied, 4.0001 s, 1.0 MB/s
```

可以使用blkio.throttle.write_bps_device文件设置写入限制。

# 分层的cgroup

节流实现层次结构支持；但是，如果从cgroup端启用了`sane_behavior`，
则启用了节流的层次结构支持，当前这是开发选项，并且不公开。

如果有人创建了如下的层次结构。

			root
			/  \
		     test1 test2
			|
		     test3

对于throttling策略，所有限制都适用于整个子树，
而所有统计信息对于该cgroup中的任务直接生成的IO都是本地的

在没有从cgroup端启用“ sane_behavior”的情况下，进行节流实际上会将所有组视为相同级别，如下所示：

				pivot
			     /  /   \  \
			root  test1 test2  test3

# 各种用户可见的配置选项

- CONFIG_BLK_CGROUP - 块IO控制器。

- CONFIG_BFQ_CGROUP_DEBUG  调试帮助。现在，如果启用了此选项，则cgroup中将显示一些其他统计信息文件。。

CONFIG_BLK_DEV_THROTTLING 在块层启用块设备节流支持。

# cgroup文件的详细信息

## 比例权重策略文件

- blkio.weight

指定每个cgroup的权重。这是所有设备上该组的默认权重，直到且除非被每个设备规则覆盖。（请参阅blkio.weight_device）。当前允许的权重范围是10到1000。

- blkio.weight_device

可以使用该接口为每个设备的每个cgroup指定规则。这些规则将覆盖blkio.weight指定的组权重的默认值。

以下是格式：
```
# echo dev_maj:dev_minor weight > blkio.weight_device
```

在此cgroup的/dev/sdb（8:16）上配置weight = 300

```
# echo 8:16 300 > blkio.weight_device
# cat blkio.weight_device
dev     weight
8:16    300
```

在此cgroup的/dev/sda（8：0）上配置weight = 500：

```
	  # echo 8:0 500 > blkio.weight_device
	  # cat blkio.weight_device
	  dev     weight
	  8:0     500
	  8:16    300
```

在此cgroup中删除/ dev / sda的特定权重：

```
# echo 8:0 0 > blkio.weight_device
# cat blkio.weight_device
dev     weight
8:16    300
```

- blkio.time
	- 每个设备分配给cgroup的磁盘时间（以毫秒为单位）。
	前两个字段指定设备的主编号和次编号，第三个字段指定分配给组的磁盘时间
	（以毫秒为单位）。

- blkio.sectors
	- 组从磁盘传输到磁盘或从磁盘传输的扇区数。前两个字段指定设备的主设备号和次设备号，
	第三个字段指定组向设备/从设备传输的扇区数。

- blkio.io_service_bytes
	- 组从磁盘传输到磁盘或从磁盘传输的字节数。这些按操作类型进一步划分-读或写，同步或异步。前两个字段指定设备的主编号和次编号，第三个字段指定操作类型，第四个字段指定字节数。

- blkio.io_serviced
	- 该组向磁盘发出的IO（bio）数量。这些按操作类型进一步划分-读或写，同步或异步。前两个字段指定设备的主编号和次编号，第三个字段指定操作类型，第四个字段指定IO的数量。

- blkio.io_service_time
	- 此cgroup完成的IO的请求分发和请求完成之间的总时间。这在十亿分之一秒内对闪存设备也很有意义。对于队列深度为1的设备，此时间表示实际服务时间。当queue_depth> 1时，这不再是正确的，因为可能会无序地处理请求。这可能会导致给定IO的服务时间包括多个IO的服务时间（无序提供），这可能导致总io_service_time>实际时间过去。此时间进一步除以操作类型-读或写，同步或异步。前两个字段指定设备的主设备号和次设备号，第三个字段指定操作类型，第四个字段指定以ns为单位的io_service_time。

- blkio.io_wait_time
	- 此cgroup的IO在调度程序队列中等待服务所花费的总时间。这可以大于所花费的总时间，因为它是所有IO的累积io_wait_time。它不是衡量cgroup等待总时间的方法，而是衡量各个IO的wait_time的方法。对于queue_depth> 1的设备，此指标不包括将IO分派到设备后等待服务的时间，直到它得到实际服务为止（由于设备对请求的重新排序，因此可能会有时间滞后） 。这在十亿分之一秒内对闪存设备也很有意义。此时间进一步除以操作类型-读或写，同步或异步。前两个字段指定设备的主设备号和次设备号，

- blkio.io_merged
	- 合并到属于此cgroup的请求的BIOS /请求总数。这可以通过操作类型进一步划分-读或写，同步或异步。

- blkio.io_queued
	- 在任何给定时刻为此cgroup排队的请求总数。这可以通过操作类型进一步划分-读或写，同步或异步。

- blkio.avg_queue_size
	- 仅在CONFIG_BFQ_CGROUP_DEBUG = y时启用调试辅助。在此cgroup存在的整个时间内，此cgroup的平均队列大小。每当此cgroup的队列之一获得时间片时，就对队列大小进行采样。

- blkio.group_wait_time
	- 仅在CONFIG_BFQ_CGROUP_DEBUG = y时启用调试辅助。这是自从cgroup变得繁忙（即从0到1的请求队列）以来为它的一个队列获取时间片所必须等待的时间。这与io_wait_time不同，io_wait_time是该cgroup中每个IO在调度程序队列中等待的时间的累积总数。以纳秒为单位。如果在cgroup处于等待（用于时间片）状态时读取此内容，则该统计信息将仅报告直到其最后获得时间片为止所累积的group_wait_time，并且将不包括当前增量。

- blkio.empty_time
	- 仅在CONFIG_BFQ_CGROUP_DEBUG = y时启用调试辅助。这是cgroup在不被服务时花费在没有任何未决请求的情况下的时间量，即，它不包括为cgroup的队列之一空闲所花费的时间。以纳秒为单位。如果在cgroup处于空状态时读取了此内容，则该统计信息将仅报告直到上一次有待处理请求为止所累积的empty_time，并且将不包括当前增量。

- blkio.idle_time
	- 仅在CONFIG_BFQ_CGROUP_DEBUG = y时启用调试辅助。这是IO调度程序闲置给定cgroup所花费的时间，以期望比来自其他队列/ cgroup的现有请求更好的请求。以纳秒为单位。如果在cgroup处于空闲状态时读取此消息，则该统计信息将仅报告直到最后一个空闲周期为止所累积的idle_time，并且将不包括当前增量。

- blkio.dequeue
	- 仅在CONFIG_BFQ_CGROUP_DEBUG = y时启用调试辅助。这提供了有关组从设备的服务树中出队的次数的统计信息。前两个字段指定设备的主设备号和次设备号，第三个字段指定组从特定设备出队的次数。

- blkio.*_recursive
	- 各种统计信息的递归版本。这些文件显示与非递归对应文件相同的信息，但包括所有后代cgroup的统计信息。

## 节流/上传 策略文件

- blkio.throttle.read_bps_device
	- 指定设备读取速率的上限。IO速率以每秒字节数指定。规则是针对每个设备的。以下是格式：

```
 echo "<major>:<minor>  <rate_bytes_per_second>" > /cgrp/blkio.throttle.read_bps_device
```

- blkio.throttle.write_bps_device
	- 指定对设备的写入速率的上限。IO速率以每秒字节数指定。规则是针对每个设备的。以下是格式：

```
 echo "<major>:<minor>  <rate_bytes_per_second>" > /cgrp/blkio.throttle.write_bps_device
```

- blkio.throttle.read_iops_device
	- 指定设备读取速率的上限。IO速率以每秒IO为单位指定。规则是针对每个设备的。以下是格式

```
echo "<major>:<minor>  <rate_io_per_second>" > /cgrp/blkio.throttle.read_iops_device
```

- blkio.throttle.write_iops_device
	- 指定对设备的写入速率的上限。IO速率以io每秒为单位指定。规则是针对每个设备的。以下是格式：

```
echo "<major>:<minor>  <rate_io_per_second>" > /cgrp/blkio.throttle.write_iops_device
```

注：如果设备同时指定BW和IOPS的规则，则是IO受到所有约束。

- blkio.throttle.io_serviced
	- 该组向磁盘发出的IO（生物）数量。这些按操作类型进一步划分-读或写，同步或异步。前两个字段指定设备的主编号和次编号，第三个字段指定操作类型，第四个字段指定IO的数量

- blkio.throttle.io_service_bytes
	- 组从磁盘传输到磁盘或从磁盘传输的字节数。这些按操作类型进一步划分-读或写，同步或异步。前两个字段指定设备的主编号和次编号，第三个字段指定操作类型，第四个字段指定字节数。

# 各种政策中的共同文件


- blkio.reset_stats
  - 将int写入此文件将导致重置该cgroup的所有统计信息。