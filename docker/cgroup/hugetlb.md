HugeTLB控制器
-------------------

HugeTLB控制器允许限制每个控制组和
在页面错误期间强制执行控制器限制。由于HugeTLB没有
支持页面回收，在页面故障时强制执行限制意味着，
如果尝试访问HugeTLB页面，该应用程序将获得SIGBUS信号
超出其极限。这要求应用程序事先知道多少
使用它所需的HugeTLB页面。

可以通过首先安装cgroup文件系统来创建HugeTLB控制器。

＃mount -t cgroup -o hugetlb none / sys / fs / cgroup

通过上述步骤，初始或父HugeTLB组变为
在/ sys / fs / cgroup中可见。在启动时，该组包括以下所有任务
系统。/ sys / fs / cgroup / tasks列出了此cgroup中的任务。

可以在父组/ sys / fs / cgroup下创建新组。

＃cd / sys / fs / cgroup
＃mkdir g1
＃echo $$> g1 /任务

上面的步骤创建了一个新的组g1并移动了当前外壳
处理（重击）。

控制文件摘要

 hugetlb。<hugepagesize> .limit_in_bytes＃设置/显示“ hugepagesize”的巨大限制
 hugetlb。<hugepagesize> .max_usage_in_bytes＃显示记录的最大“ hugepagesize” hugetlb使用情况
 hugetlb。<hugepagesize> .usage_in_bytes＃显示“ hugepagesize”的当前用法
 hugetlb。<hugepagesize> .failcnt＃显示由于HugeTLB限制而导致分配失败的次数

对于支持三个大​​页面大小（64k，32M和1G）的系统，控制
文件包括：

hugetlb.1GB.limit_in_bytes
hugetlb.1GB.max_usage_in_bytes
hugetlb.1GB.usage_in_bytes
hugetlb.1GB.failcnt
hugetlb.64KB.limit_in_bytes
hugetlb.64KB.max_usage_in_bytes
hugetlb.64KB.usage_in_bytes
hugetlb.64KB.failcnt
hugetlb.32MB.limit_in_bytes
hugetlb.32MB.max_usage_in_bytes
hugetlb.32MB.usage_in_bytes
hugetlb.32MB。失败