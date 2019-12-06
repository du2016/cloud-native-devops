# cpu管理策略

默认情况使用CFS（完全公平的调度程序），CFS情况下指定进程的运行时间计算方式如下,

```
vruntime = 实际运行时间 * 1024 / 进程权重
```

但是有时间我们想让一些性能要求高的应用独占CPU，这个时间该怎么做呢？

# 配置kubelet的cpu管理策略

kubelet 通过`--cpu-manager-policy`参数指定管理策略，
支持两种策略：
- none：默认策略，表示现有的调度行为（即CFS）。
- static：允许为节点上具有某些资源特征的 pod 赋予增强的 CPU 亲和性和独占性。

CPU管理器定期通过 CRI 写入资源更新，以保证内存中 CPU 分配与 cgroupfs 一致。
同步频率通过新增的 Kubelet 配置参数 --cpu-manager-reconcile-period 来设置。 
如不指定，默认与 --node-status-update-frequency 的周期相同。

## none

该none策略显式启用现有的默认CPU关联性方案，不提供OS调度程序自动执行的关联性。 
使用CFS配额实施对Guaranteed Pod的 CPU使用的限制 

## static 

该模式允许配置为整数cpu的`Guaranteed pods`独占CPU，该功能通过`Cgroup cpuset`实现。

资源池： CPU 总量减去通过`--kube-reserved`或`--system-reserved`参数保留的 CPU，
然后按ID升序排列

当符合static策略的pod调度到节点，绑定对应CPU，将CPU从资源池去除，然后其他的运行在资源池
 
- pod为Guaranteed Qos
- 请求核数为整数

# QOS

简单说一下QOS

- Guaranteed

pod中所有容器都必须统一设置limits，并且设置参数都一致，
如果有一个容器要设置requests，那么所有容器都要设置，并设置参数同limits一致，
那么这个pod的QoS就是Guaranteed级别。

- Burstable

pod中只要有一个容器的requests和limits的设置不相同，该pod的QoS即为Burstable。

- Best-Effort

如果对于全部的resources来说requests与limits均未设置，该pod的QoS即为Best-Effort。


QOS优先级如下

```
Guaranteed > Burstable > Best-Effort
```

# 为系统进程和kubelet绑定CPU

因为static模式的资源池是以ID升序排列，所以我们在配置
`--kube-reserved`和`--system-reserved`后，可以通过CPUSET进行绑核心，让系统进程
和容器进程分开运行，

加入操作系统有12个核心，CPU预留策略如下，为系统和kubelet保留1.2个CPU：
```
--kube-reserved cpu=300m 
--kube-reserved-cgroup=/system.slice/kubelet.service
--system-reserved cpu=1 
--system-reserved-cgroup=/system.slice 
```

则可以配置systemd的cpuset为11-12

```
# 因为系统以0开始，所以写10-11
echo '10-11' > /sys/fs/cgroup/cpuset/system.slice
# 也可以为kubelet绑定核心
echo '11' > /sys/fs/cgroup/cpuset/system.slice
```

这个时间我们的容器就运行在前十个CPU，而系统进程则运行在后面两个CPU上面。

# NUMA 场景的问题

不同厂商NUMA架构不一样，笔者所在公司用的惠普服务器配置如下，
多核应用如果对CPU进行绑定，则一定会跨NUMA绑定（因为kubelet是根据ID升序绑核）
可能带来性能问题

```
NUMA 节点0 CPU：    0,2,4,6,8,10,12,14,16,18,20,22,24,26,28,30
NUMA 节点1 CPU：    1,3,5,7,9,11,13,15,17,19,21,23,25,27,29,31
```