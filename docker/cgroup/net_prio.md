网络优先级cgroup
-------------------------

网络优先级cgroup提供了一个界面，使管理员可以
动态设置各种生成的网络流量的优先级
应用领域

名义上，应用程序将通过
SO_PRIORITY套接字选项。但是，这并非总是可能的，因为：

1）可能未对应用程序进行编码以设置此值
2）应用程序流量的优先级通常是特定于站点的管理
   决定，而不是由应用程序定义。

此cgroup允许管理员将进程分配给定义以下内容的组
给定接口上的出口流量优先级。网络优先级组可以
通过首先挂载cgroup文件系统来创建。

＃mount -t cgroup -onet_prio无/ sys / fs / cgroup / net_prio

通过上述步骤，初始组充当母公司会计组
在'/ sys / fs / cgroup / net_prio'上可见。该组中的所有任务
系统。“ / sys / fs / cgroup / net_prio / tasks”列出了此cgroup中的任务。

每个net_prio cgroup包含两个子系统特定的文件

net_prio.prioidx
该文件为只读文件，仅供参考。它包含一个唯一的整数
内核用作此cgroup的内部表示形式的值。

net_prio.ifpriomap
此文件包含分配给源自的流量的优先级的映射
该组中的进程，并在各种接口上退出系统。它
包含格式为<ifname priority>的元组列表。该文件的内容
可以通过使用相同的元组格式将字符串回显到文件中来进行修改。
例如：

回声“ eth0 5”> /sys/fs/cgroups/net_prio/iscsi/net_prio.ifpriomap

此命令将强制源自属于
iscsi net_prio cgroup并在接口eth0上具有优先级的出口
表示流量设置为值5。父会计组也有一个
可写的“ net_prio.ifpriomap”文件，可用于设置系统默认值
优先。

在将帧排队到设备之前立即设置优先级
排队规则（qdisc），因此优先级将在硬件之前分配
正在进行队列选择。

net_prio cgroup的一种用法是使用mqprio qdisc允许应用程序
流量将转向基于硬件/驱动程序的流量类别。这些映射
然后可以由管理员或其他网络协议（例如
DCBX。

一个新的net_prio cgroup继承了父级的配置。