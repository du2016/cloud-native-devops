CPU 账户控制器

CPU账户控制器用于使用cgroup对任务进行分组并计算这些任务组的CPU使用率。

CPU账户控制器支持多层次结构组。审计组会累积其所有子组的CPU使用率，直接存在于该组中的任务。

可以通过首先挂载cgroup文件系统来创建计费组。

```
mount -t cgroup -ocpuacct none /sys/fs/cgroup
```
通过上面的步骤，初始帐户组或父帐户组
在/sys/fs/cgroup中可见。在启动时，该组包括系统中的所有任务
。/sys/fs/cgrouptasks列出了此cgroup中的任务。
/sys/fs/cgroup/cpuacct.usage给出
该组获得的CPU时间（以纳秒为单位），本质
上是系统中所有任务获得的CPU时间。

可以在父组/sys/fs/cgroup下创建新的会计组。

```
＃cd /sys/fs/cgroup
＃mkdir g1 
＃echo $$> g1/tasks 
```

以上步骤创建了一个新的g1组，并将当前的shell 
进程（bash）移入其中。
可以从g1/cpuacct.usage获得此bash及其子进程消耗的CPU时间，该时间也累积在
/sys/fs/cgroup/cpuacct.usage中。

cpuacct.stat文件列出了一些统计信息，这些统计信息进一步将
cgroup获得的CPU时间划分为用户时间和系统时间。目前
支持以下统计信息：

用户：cgroup任务在用户模式下花费的时间。
系统：内核模式下cgroup任务花费的时间。

用户和系统位于USER_HZ单元中。

# 副作用
cpuacct控制器使用percpu_counter接口收集用户和
系统时间。这有两个副作用：

- 理论上，用户和系统时间可能会看到错误的值。
  这是因为32位系统上的percpu_counter_read()不能
  防止并发写入。
- 由于percpu_counter的批处理性质，可能会看到用户和系统时间有些过时的值。