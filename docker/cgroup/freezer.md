
cgroup冷冻机对于启动批处理作业管理系统很有用
并停止任务集以安排机器资源
根据系统管理员的需求。这种程序
通常在HPC群集上使用，以计划对整个群集的访问
。cgroup冷冻机使用cgroups描述要执行的任务集
由批处理作业管理系统启动/停止。它还提供
一种开始和停止组成工作的任务的方法。

cgroup冷冻机对于检查运行组也很有用
的任务。冷冻机允许检查点代码获得一致的
通过尝试将cgroup中的任务强制为
静态。任务静止后，另一个任务可以
walk / proc或调用内核接口来收集有关
停顿的任务。有检查点的任务可以稍后重新启动
发生可恢复的错误。这也使检查点任务可以
通过复制收集的信息在集群中的节点之间迁移
到另一个节点并在那里重新启动任务。

SIGSTOP和SIGCONT的序列并不总是足以停止
和恢复用户空间中的任务。这两个信号都是可观察到的
从我们希望冻结的任务中。虽然无法捕获SIGSTOP，
阻止或忽略它可以通过等待或跟踪父任务来看到。
SIGCONT特别不合适，因为它可能会被任务捕获。任何
用于监视SIGSTOP和SIGCONT的程序可能会被破坏
尝试使用SIGSTOP和SIGCONT停止和继续任务。我们可以
使用嵌套的bash shell演示此问题：

	$ echo $$
	16644
	$ bash
	$ echo $$
	16690

	从另一个无关的bash shell中：
	$ kill -SIGSTOP 16690
	$ kill -SIGCONT 16690

	<此时16690退出并导致16644也退出>

发生这种情况是因为bash可以观察两个信号并选择其方式
回应他们。

捕获并响应这些问题的程序的另一个示例
信号是gdb。实际上，任何旨在使用ptrace的程序都可能
这种停止和恢复任务的方法存在问题。

相反，cgroup冻结器使用内核冻结器代码来
防止冻结/解冻循环对任务可见
被冻结。这允许上面的bash示例和gdb以
预期。

cgroup冷冻机是分层的。冻结cgroup将冻结所有
属于cgroup及其所有后代cgroup的任务。每
cgroup具有其自己的状态（自我状态），并且该状态是从
父级（父级）。如果两个状态都已解冻，则cgroup为
解冻。

以下cgroupfs文件由cgroup freezer创建。

* freezer.state：可读写。

  读取时，返回cgroup的有效状态-“ THAWED”，
  “冻结”或“冻结”。这是自身状态和父状态的组合。
  如果有任何冻结，则cgroup处于冻结状态（FREEZING或FROZEN）。

  当所有任务都冻结时，cgroup转换为冻结状态
  属于cgroup的子孙及其后代将被冻结。注意
  添加新任务后，cgroup从冻结状态恢复为冻结状态
  到cgroup或其后代cgroup之一，直到新任务为
  冻结的。

  编写时，设置cgroup的自状态。两个值是
  允许-“冻结”和“解冻”。如果写了FROZEN，则cgroup，
  如果尚未冻结，则进入冻结状态
  后代cgroups。

  如果写入THAWED，则cgroup的自身状态将更改为
  解冻。请注意，如果发生以下情况，则有效状态可能不会更改为THAWED
  母国仍处于冻结状态。如果cgroup的有效状态
  变为解冻，其所有后代由于
  cgroup也会退出冻结状态。

* freezer.self_freezing：只读。

  显示自我状态。如果自身状态为THAWED，则为0；否则为0。否则1。
  如果最后一次写入freezer.state是“ FROZEN”，则此值为1。

* freezer.parent_freezing：只读。

  显示父状态。如果cgroup的祖先都不是，则为0
  冻结 否则1。

根cgroup是不可冻结的，并且上面的接口文件没有
存在。

*用法示例：

   ＃mkdir / sys / fs / cgroup / freezer
   ＃mount -t cgroup -ofreezer冷冻机/ sys / fs / cgroup / freezer
   ＃mkdir / sys / fs / cgroup / freezer / 0
   ＃echo $ some_pid> / sys / fs / cgroup / freezer / 0 / tasks

获取冷冻机子系统的状态：

   ＃cat /sys/fs/cgroup/freezer/0/freezer.state
   解冻

冻结容器中的所有任务：

   ＃echo冻结> /sys/fs/cgroup/freezer/0/freezer.state
   ＃cat /sys/fs/cgroup/freezer/0/freezer.state
   冷冻
   ＃cat /sys/fs/cgroup/freezer/0/freezer.state
   冷冻

解冻容器中的所有任务：

   ＃回显THAWED> /sys/fs/cgroup/freezer/0/freezer.state
   ＃cat /sys/fs/cgroup/freezer/0/freezer.state
   解冻

这是对用户空间任务应该做正确的事情的基本机制
在一个简单的场景中。