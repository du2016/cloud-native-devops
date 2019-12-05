RDMA控制器
				----------------

内容
--------

1.概述
  1-1。什么是RDMA控制器？
  1-2。为什么需要RDMA控制器？
  1-3。RDMA控制器如何实现？
2.用法示例

1.概述

1-1。什么是RDMA控制器？
-----------------------------

RDMA控制器允许用户限制给定的RDMA / IB特定资源
可以使用的一组过程。这些过程使用RDMA控制器进行分组。

RDMA控制器定义了两个资源，可以限制两个资源
cgroup。

1-2。为什么需要RDMA控制器？
--------------------------------

当前，用户空间应用程序可以轻松删除所有rdma动词
特定资源，例如AH，CQ，QP，MR等。由于其他应用
在其他cgroup或内核空间中，ULP甚至可能没有机会分配任何
rdma资源。这会导致服务不可用。

因此需要使用RDMA控制器来消耗资源
可以限制进程数。通过该控制器，不同的rdma
资源可以考虑。

1-3。RDMA控制器如何实现？
----------------------------------------

RDMA cgroup允许限制资源配置。Rdma CGroup维护
使用资源池结构的每个cgroup，每个设备的资源计费。
每个此类资源池在给定的资源池中最多可限制64个资源
由rdma cgroup提供，如果需要，可以稍后扩展。

此资源池对象链接到cgroup css。通常在那里
在大多数使用情况下，每个设备每个cgroup有0到4个资源池实例。
但没有任何限制可以拥有更多。目前，每个设备有数百个RDMA设备
单个cgroup可能无法得到最佳处理，但是没有
已知用例或此类配置的要求。

由于RDMA资源可以从任何进程分配，并且可以由任何进程释放
共享地址空间的子进程中，rdma资源是
始终由创建者cgroup css拥有。这允许从一个进程迁移
转移到其他cgroup，而没有转移资源所有权的主要复杂性；
因为由于
rdma资源。在css周围链接资源还可以确保cgroups
迁移过程后删除。这也允许进度迁移
活动资​​源，即使这不是主要用例。

每当发生RDMA资源计费时，所有者rdma cgroup都会返回到
呼叫者，召集者。卸载资源时应传递相同的rdma cgroup。
这也允许使用活动RDMA资源迁移的进程进行计费
到新所有者cgroup以获得新资源。它还允许释放资源
从先前收费的cgroup迁移到新的cgroup的过程，
即使这不是主要的用例。

在以下情况下会创建资源池对象。
（a）用户设置限制，并且该设备不存在先前的资源池
对cgroup感兴趣。
（b）没有配置资源限制，但是IB / RDMA堆栈试图
收取资源。这样，当应用程序处于
无限制运行，然后在充电期间强制实施限制时，
否则，使用次数将变为负数。

如果所有资源限制都设置为max和，则销毁资源池
这是最后释放的资源。

如果要删除/取消配置，用户应将所有限制设置为最大值
特定设备的资源池。

IB堆栈遵守rdma控制器强制执行的限制。申请时
查询IB设备的最大资源限制，它返回最小值
用户为给定cgroup配置的功能以及受支持的功能
IB设备。

可以由rdma控制器解释以下资源。
  hca_handle HCA句柄的最大数量
  hca_object HCA对象的最大数量

2.用法示例
-----------------

（a）配置资源限制：
回声mlx4_0 hca_handle = 2 hca_object = 2000> /sys/fs/cgroup/rdma/1/rdma.max
回声ocrdma1 hca_handle = 3> /sys/fs/cgroup/rdma/2/rdma.max

（b）查询资源限制：
猫/sys/fs/cgroup/rdma/2/rdma.max
＃输出：
mlx4_0 hca_handle = 2 hca_object = 2000
ocrdma1 hca_handle = 3 hca_object = max

（c）查询当前使用情况：
猫/sys/fs/cgroup/rdma/2/rdma.current
＃输出：
mlx4_0 hca_handle = 1 hca_object = 20
ocrdma1 hca_handle = 1 hca_object = 23

（d）删除资源限制：
回声回声mlx4_0 hca_handle = max hca_object = max> /sys/fs/cgroup/rdma/1/rdma.max