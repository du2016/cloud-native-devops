网络分类器cgroup
-------------------------

网络分类器cgroup提供了一个接口
使用类标识符（classid）标记网络数据包。

流量控制器（tc）可用于分配
来自不同cgroup的数据包的优先级不同。
此外，Netfilter（iptables）可以使用此标记执行
对此类数据包的操作。

创建一个net_cls cgroups实例将创建一个net_cls.classid文件。
此net_cls.classid值初始化为0。

您可以将十六进制值写入net_cls.classid;。这些的格式
值是0xAAAABBBB; AAAA是主要的手柄编号，BBBB
是次要句柄号。
读取net_cls.classid会产生十进制结果。

例：
mkdir / sys / fs / cgroup / net_cls
挂载-t cgroup -onet_cls net_cls / sys / fs / cgroup / net_cls
mkdir / sys / fs / cgroup / net_cls / 0
回声0x100001> /sys/fs/cgroup/net_cls/0/net_cls.classid
	-设置10：1手柄。

猫/sys/fs/cgroup/net_cls/0/net_cls.classid
1048577

配置tc：
tc qdisc添加dev eth0根句柄10：htb

tc class add dev eth0 parent 10：classid 10：1 htb rate 40mbit
 -创建流量类别10：1

tc过滤器添加dev eth0父级10：协议ip prio 10句柄1：cgroup

配置iptables，基本示例：
iptables -A输出-m cgroup！--cgroup 0x100001 -j DROP