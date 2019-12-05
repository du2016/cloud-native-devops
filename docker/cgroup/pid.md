流程编号控制器
						   =========================

抽象
--------

进程号控制器用于允许cgroup层次结构停止任何
达到一定限制后，从fork（）或clone（）开始新任务。

由于轻松达到任务限制而没有达到任何kmemcg限制​​，
位置，PID是基本资源。因此，PID耗尽必须为
通过允许资源限制可以在cgroup层次结构的范围内避免
cgroup中的任务数。

用法
-----

为了使用`pids`控制器，请在中设置最大任务数
pids.max（出于显而易见的原因，该文件在根cgroup中不可用）。的
cgroup中当前的进程数由pids.current给出。

组织操作不受cgroup策略的阻止，因此有可能
具有pids.current> pids.max。可以通过将限制设置为
小于pids.current，或将足够的进程附加到cgroup，例如
pids.current> pids.max。但是，不可能违反cgroup
通过fork（）或clone（）制定策略。如果（）的fork（）和clone（）返回-EAGAIN
创建新进程将导致违反cgroup策略。

要将cgroup设置为无限制，请将pids.max设置为“ max”。这是默认设置
所有新的cgroup（注意PID限制是分层的，因此最严格
遵循层次结构中的限制）。

pids.current跟踪所有子cgroup层次结构，因此parent / pids.current是一个
父/子/pids.current的超集。

pids.events文件包含事件计数器：
  -max：由于达到限制而导致分叉失败的次数。

例
-------

首先，我们安装pids控制器：
＃mkdir -p / sys / fs / cgroup / pids
＃mount -t cgroup -o pids none / sys / fs / cgroup / pids

然后，我们创建一个层次结构，设置限制并将流程附加到该层次结构：
＃mkdir -p / sys / fs / cgroup / pids / parent / child
＃回声2> /sys/fs/cgroup/pids/parent/pids.max
＃echo $$> /sys/fs/cgroup/pids/parent/cgroup.procs
＃猫/sys/fs/cgroup/pids/parent/pids.current
2
#

应该注意的是，尝试克服设置的限制（在这种情况下为2）将会
失败：

＃猫/sys/fs/cgroup/pids/parent/pids.current
2
＃（/ bin / echo“这是为您准备的一些程序。” | cat）
sh：fork：资源临时不可用
#

即使我们迁移到子cgroup（没有设置限制），我们也会
无法克服层次结构中最严格的限制（在这种情况下，
父母）：

＃echo $$> /sys/fs/cgroup/pids/parent/child/cgroup.procs
＃猫/sys/fs/cgroup/pids/parent/pids.current
2
＃猫/sys/fs/cgroup/pids/parent/child/pids.current
2
＃猫/sys/fs/cgroup/pids/parent/child/pids.max
最高
＃（/ bin / echo“这是为您准备的一些程序。” | cat）
sh：fork：资源临时不可用
#

我们可以设置一个小于pids.current的限制，它将停止任何新的
从根本上分叉的过程（请注意，shell本身对
pids.current）：

＃回声1> /sys/fs/cgroup/pids/parent/pids.max
＃/ bin / echo“我们现在甚至无法生成一个进程。”
sh：fork：资源临时不可用
＃回声0> /sys/fs/cgroup/pids/parent/pids.max
＃/ bin / echo“我们现在甚至无法生成一个进程。”
sh：fork：资源临时不可用
#