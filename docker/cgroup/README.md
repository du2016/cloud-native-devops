# 介绍

cgroup是Linux内核允许将流程组织为分层的功能，然后可以限制其使用各种类型资源的组
并进行监控。内核的cgroup接口通过伪文件系统，称为cgroupfs。分组在核心cgroup内核代码，
而资源跟踪和限制是在一组每个资源类型的子系统（内存，CPU，等等）。

# 术语

cgroup是绑定到一组的进程的集合。通过cgroup文件系统定义的限制或参数

子系统是一个内核组件，可修改cgroup中的进程，
已经实现了各种子系统，使诸如限制CPU 时间和可以使用的内存，
占用CPU时间由cgroup使用，并冻结和恢复执行cgroup中的进程。
子系统有时也称为资源控制者。

控制器的cgroup按层次结构排列。
通过创建，删除和重命名cgroup 文件系统来定义层次结构
在每个级别的层次中，可以定义属性（例如限制）,
cgroup提供的限制，控制和计费通常在定义属性的cgroup之下的整个子层次结构中有效。 
因此，例如，子级cgroup不能超过层次结构中较高级别的cgroup上的限制

# V1和V2历史

cgroups实现的最初版本是在Linux中2.6.24。随着时间的推移，
各种cgroup控制器已添加到允许管理各种类型的资源。
然而这些控制器的开发在很大程度上是不协调的，
结果是控制器和控制器之间出现了许多不一致之处
cgroup层次结构的管理变得相当复杂。

由于最初的cgroups实现存在问题（cgroups版本1）从Linux 3.10开始，开始了新的工作，
正交实施以解决这些问题。最初标记实验性的，并且隐藏在-o __DEVEL__sane_behavior挂载后面
选项，最终制作了新版本（cgroups版本2）正式发布Linux 4.5。

尽管cgroups v2旨在替代cgroups v1，但是较旧的系统继续存在（出于兼容性原因，
不太可能被删除）。目前，cgroups v2仅实现 cgroups v1中可用的控制器子集。两个系统
已实现，因此v1控制器和v2控制器都可以安装在同一系统上。因此，例如，可以使用
在版本2下受支持的控制器，同时使用版本2尚不支持的版本1控制器这些控制器。唯一的限制是控制器
不能同时在cgroups v1层次结构和在cgroups v2层次结构中。


# 子系统

cgroup分为以下子系统：

- cpuacct 子系统，可以统计 cgroups 中的进程的 cpu 使用报告。
- cpuset 子系统，可以为 cgroups 中的进程分配单独的 cpu 节点或者内存节点。
- memory 子系统，可以限制进程的 memory 使用量。
- blkio 子系统，可以限制进程的块设备 io。
- devices 子系统，可以控制进程能够访问某些设备。
- net_cls 子系统，可以标记 cgroups 中进程的网络数据包，然后可以使用 tc 模块（traffic control）对数据包进行控制。
- net_prio — 这个子系统用来设计网络流量的优先级
- freezer 子系统，可以挂起或者恢复 cgroups 中的进程。
- hugetlb — 这个子系统主要针对于HugeTLB系统进行限制，这是一个大页文件系统。
- perf_event - 对cgroup中的进程组进行性能监控 v2
- net_prio  - 对每个网络接口指定优先级
- pids - 限制在cgroup中创建的进程数量
- rdma - cgroup中使用的rdma


# 实践

## 升级内核
```
rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org
rpm -Uvh http://www.elrepo.org/elrepo-release-7.0-2.el7.elrepo.noarch.rpm
yum --enablerepo=elrepo-kernel install kernel-ml kernel-ml-devel
grub2-set-default 0
```

## 修改内核参数禁用V1

```
/etc/default/grub
GRUB_CMDLINE_LINUX添加cgroup_no_v1=all
grub2-mkconfig -o /boot/grub2/grub.cfg
```

## 限制进程IO

```
mkdir /cgroup2
# 挂载cgroup
mount -t cgroup2 nodev /cgroup2
# 为子树添加io子系统
echo "+io" > /cgroup2/cgroup.subtree_control
# 验证是否开启
cat /cgroup2/cg2/cgroup.controllers
# 查看 文件系统设备号
lsblk
NAME   MAJ:MIN RM SIZE RO TYPE MOUNTPOINT
sr0     11:0    1  41M  0 rom
vda    253:0    0  50G  0 disk
└─vda1 253:1    0  50G  0 part /

# 限制设备IO
echo "253:0 wbps=1048576" > /cgroup2/cg2/io.max
# 测试 这里要设置一个比较大的文件，不然看不出效果
dd if=/dev/zero of=/tmp/file1 bs=512M count=1
```
>注意：上面步骤我们可以看到我们对设备进行了限制，cgroup本身无法对分区进行限制，
>凡是我们可以通过lvm使用分区创建LV，从而对LV进行限制


# 参考

[cgroup-v1文档](https://www.kernel.org/doc/Documentation/cgroup-v1/)

[cgroup V2设计草案](https://www.kernel.org/doc/Documentation/cgroup-v2.txt)

[Cgroup V2 and writeback support](http://hustcat.github.io/cgroup-v2-and-writeback-support/)