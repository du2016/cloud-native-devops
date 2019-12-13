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

> 注意，如果开启了超线程，这里保留的Cpu计算实际会翻倍，例如超线程之后8CCPU，实际上物理核心是4C
> 如果给系统和kube预留2C则实际上将在资源池里去掉0-4，但是因为没有绑定核心，
> 所以系统进程并不一定会在0-4上运行

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

# 为系统进程绑定CPU

因为static模式的资源池是以ID升序排列，所以我们在配置
`--kube-reserved`和`--system-reserved`后，可以通过CPUSET进行绑核心，让系统进程
和容器进程分开运行，

假如操作系统有12个核心，CPU预留策略如下，为系统和kubelet保留1.2个CPU：

```
--system-reserved cpu=1.3
```

计算allocatable时间会向上取整

也就是说需要为systemd的cpuset为2

```
echo '0-1' > /sys/fs/cgroup/cpuset/system.slice/cpuset.cpus
```

这个时间我们的容器就运行在后十个CPU，而系统进程则运行在后面两个CPU上面。


> 在1.17开始也可以通过--reserved-cpus 设置预留的值，该参数将覆盖--system-reserved，--kube-reserved
中配置的CPU

# cpumanager CPU分配

首先讲解几个概念

- Socket 是主板上插CPU的槽的数量
- Core 每个CPU上的核数
- Thread 是每个core上的硬件线程数，即超线程

cpumanager会先获取numa的节点配置，然后先根据这三个维度进行尝试分配

# 测试

创建以下pod
```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx
    resources:
      limits:
        memory: "200Mi"
        cpu: "2"
      requests:
        memory: "200Mi"
        cpu: "2"
```

查看pod的cpuset,发现并未绑定核心

```
cat /sys/fs/cgroup/cpuset/kubepods/pod45d4c470-ff57-4583-a42f-abec3190582e/cpuset.cpus
0-3
```

查看容器的cpuset,已经绑定，并且未使用第一个核，因为预留给了系统

```
cat /sys/fs/cgroup/cpuset/kubepods/pod45d4c470-ff57-4583-a42f-abec3190582e/dc98a52a78835ca6289db0ff9e0414facb939498f4cf92b164edc54fe37891d8/cpuset.cpus
1-2
```