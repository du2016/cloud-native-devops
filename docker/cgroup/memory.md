
内存资源控制器

注意：本文档已过时，希望您提供完整的文档
      改写。它仍然包含有用的信息，因此我们将其保留
      在这里，但是如果您需要更深入的信息，请确保检查当前代码
      理解。

注意：内存资源控制器通常被称为
      本文档中的内存控制器。不要混淆内存控制器
      与硬件中使用的内存控制器一起使用。

（对于编辑）
在本文件中：
      当我们提到带有内存控制器的cgroup（cgroupfs的目录）时，
      我们称其为“内存cgroup”。当您看到git-log和源代码时，您将
      请参阅补丁的标题和函数名称倾向于使用“ memcg”。
      在本文档中，我们避免使用它。

内存控制器的优点和目的

内存控制器隔离一组任务的内存行为
从系统的其余部分开始。关于LWN的文章[12]提到了一些可能
内存控制器的用途。内存控制器可用于

一种。隔离一个应用程序或一组应用程序
   内存不足的应用程序可以隔离，并限制在较小的应用程序中
   内存量。
b。创建一个内存有限的cgroup；可以使用
   作为用mem = XXXX引导的一种很好的选择。
C。虚拟化解决方案可以控制所需的内存量
   分配给虚拟机实例。
d。CD / DVD刻录机可以控制打印机使用的内存量。
   系统的其余部分，以确保刻录不会因缺少而失败
   可用内存。
e。还有其他几种用例。找到一个或仅使用控制器
   的乐趣（学习和破解VM子系统）。

当前状态：linux-2.6.34-mmotm（2010年4月开发版本）

特征：
 -计算匿名页面，文件缓存，交换缓存的使用并限制它们的使用。
 -页面仅链接到每内存LRU，并且没有全局LRU。
 -可选地，可以考虑并限制内存+交换使用。
 -分级会计
 -软限制
 -移动任务时可以选择移动（充值）帐户。
 -使用量阈值通知程序
 -内存压力通知器
 -oom-killer禁用旋钮和oom-notifier
 -根cgroup没有限制控件。

 内核内存支持尚在开发中，当前版本提供
 基本的功能。（请参阅第2.7节）

控制文件的简要摘要。

 任务＃附加一个任务（线程）并显示线程列表
 cgroup.procs＃显示进程列表
 cgroup.event_control＃event_fd（）的接口
 memory.usage_in_bytes＃显示当前的内存使用情况
				 （有关详细信息，请参见5.5）
 memory.memsw.usage_in_bytes＃显示内存+交换的当前使用情况
				 （有关详细信息，请参见5.5）
 memory.limit_in_bytes＃设置/显示内存使用限制
 memory.memsw.limit_in_bytes＃设置/显示内存限制+交换使用
 memory.failcnt＃显示内存使用次数限制
 memory.memsw.failcnt＃显示内存数量+交换命中次数限制
 memory.max_usage_in_bytes＃显示记录的最大内存使用情况
 memory.memsw.max_usage_in_bytes＃显示最大内存+记录的交换使用情况
 memory.soft_limit_in_bytes＃设置/显示内存使用量的软限制
 memory.stat＃显示各种统计信息
 memory.use_hierarchy＃设置/显示分层帐户已启用
 memory.force_empty＃触发强制页面回收
 memory.pressure_level＃设置内存压力通知
 memory.swappiness＃设置/显示vmscan的swappiness参数
				 （请参阅sysctl的vm.swappiness）
 memory.move_charge_at_immigrate＃设置/显示移动费用控制
 memory.oom_control＃设置/显示oom控件。
 memory.numa_stat＃显示每个numa节点的内存使用量

 memory.kmem.limit_in_bytes＃设置/显示内核内存的硬限制
 memory.kmem.usage_in_bytes＃显示当前内核内存分配
 memory.kmem.failcnt＃显示内核内存使用次数达到限制的数量
 memory.kmem.max_usage_in_bytes＃显示记录的最大内核内存使用率

 memory.kmem.tcp.limit_in_bytes＃设置/显示tcp buf内存的硬限制
 memory.kmem.tcp.usage_in_bytes＃显示当前tcp buf内存分配
 memory.kmem.tcp.failcnt＃显示tcp buf内存使用次数达到限制的数量
 memory.kmem.tcp.max_usage_in_bytes＃显示记录的最大tcp buf内存使用量

1.历史

内存控制器的历史悠久。要求对内存发表评论
控制器由Balbir Singh发表[1]。RFC发布时
内存控制有几种实现方式。的目标
RFC旨在就所需的最少功能建立共识和协议
用于内存控制。第一个RSS控制器由Balbir Singh发表[2]
在2007年2月。PavelEmelianov [3] [4] [5]发布了三个版本的
RSS控制器。在OLS，在资源管理BoF，每个人都建议
我们一起处理页面缓存和RSS。提出了另一个要求
允许用户处理OOM。当前的内存控制器是
在版本6上；它结合了已映射（RSS）和未映射的页面
缓存控制[11]。

2.内存控制

从某种意义上讲，内存是一种独特的资源
量。如果任务需要大量CPU处理，则该任务可能会扩散
在数小时，数天，数月或数年的时间内进行处理，但是
内存，需要重复使用相同的物理内存才能完成任务。

存储器控制器的实现已分为多个阶段。这些
是：

1.内存控制器
2. mlock（2）控制器
3.内核用户内存记帐和平板控制
4.用户映射长度控制器

内存控制器是第一个开发的控制器。

2.1。设计

设计的核心是一个称为page_counter的计数器。的
page_counter跟踪当前内存使用情况以及该组内存的限制
与控制器关联的过程。每个cgroup都有一个内存控制器
与之关联的特定数据结构（mem_cgroup）。

2.2。会计

		+--------------------+
		| mem_cgroup |
		| （page_counter）|
		+--------------------+
		 /            ^      \
		/             |       \
           +---------------+  |        +---------------+
           | mm_struct | | .... | mm_struct |
           |               |  |        |               |
           +---------------+  |        +---------------+
                              |
                              + --------------+
                                              |
           +---------------+           +------+--------+
           | 页面+ ----------> page_cgroup |
           |               |           |               |
           +---------------+           +---------------+

             （图1：会计层次结构）


图1显示了控制器的重要方面

1.每cgroup进行一次记帐
2.每个mm_struct都知道它属于哪个cgroup
3.每个页面都有一个指向page_cgroup的指针，该指针又知道
   所属的cgroup

记帐方式如下：mem_cgroup_charge_common（）被调用以
设置必要的数据结构，并检查正在使用的cgroup
收费已超过上限。如果是，则在cgroup上调用回收。
可以在本文档的“回收”部分中找到更多详细信息。
如果一切顺利，则称为page_cgroup的页面元数据结构为
更新。page_cgroup在cgroup上具有自己的LRU。
（*）page_cgroup结构在启动/内存热插拔时分配。

2.2.1会计明细

所有映射的匿名页面（RSS）和缓存页面（页面缓存）都被考虑在内。
某些页面永远无法回收，也不会出现在LRU上
不占。我们只是在通常的VM管理下处理页面。

RSS页面在page_fault处进行会计处理，除非它们已经被会计过
更早。文件页面在被保存时将被视为页面缓存
插入到inode（基数树）中。当它映射到
流程中，应谨慎避免重复记帐。

完全未映射RSS页时，将无法对其进行说明。PageCache页面是
从基数树中删除时无法说明。即使RSS页面完整
在未映射之前（由kswapd定义），它们在系统中可能作为SwapCache存在，直到它们
真的被释放了 此类SwapCache也考虑在内。
交换的页面在映射之前不会计入。

注意：内核会进行swapin-readahead并一次读取多个交换。
这意味着交换的页面可能包含除任务之外的其他任务的页面
导致页面错误。因此，我们避免在交换I / O时进行记帐。

在页面迁移时，会计信息会保留。

注意：我们只考虑LRU页面，因为我们的目的是控制金额
使用的页面数；LVM上的非on页面倾向于从VM视图失控。

2.3共享页记帐

共享页面基于首次触摸方式进行会计处理。的
首先触摸页面的cgroup占该页面。原则
这种方法的背后是一个积极使用共享资源的cgroup。
该页面最终将为此付费（一旦从
带来它的cgroup-这将在内存压力下发生）。

但是请参阅第8.2节：将任务移至另一个cgroup时，其页面可能会
如果已选择move_charge_at_immigrate，则将其充值到新的cgroup。

例外：如果未使用CONFIG_MEMCG_SWAP。
当您进行交换时，将shmem（tmpfs）换出的页面
被有效地备份到内存中时，页面费用将根据
swapoff的调用者，而不是shmem的用户。

2.4交换扩展（CONFIG_MEMCG_SWAP）

掉期扩展允许您记录掉期费用。交换的页面是
如果可能的话，将充值回原始页面分配器。

当考虑交换时，将添加以下文件。
 -memory.memsw.usage_in_bytes。
 -memory.memsw.limit_in_bytes。

memsw表示内存+交换。内存+交换的使用受到以下限制：
memsw.limit_in_bytes。

示例：假设系统具有4G交换。分配6G内存的任务
（错误）在2G内存限制下将使用所有交换。
在这种情况下，设置memsw.limit_in_bytes = 3G可以防止错误使用交换功能。
通过使用memsw限制，可以避免由交换引起的系统OOM
短缺。

*为什么选择“内存+交换”而不是交换。
全局LRU（kswapd）可以换出任意页面。交换手段
将帐户从内存移至交换......的使用没有变化
内存+交换。换句话说，当我们想限制交换的使用而没有
影响全局LRU，内存+交换限制比仅限制从
操作系统的观点。

*当cgroup命中memory.memsw.limit_in_bytes时会发生什么
当cgroup命中memory.memsw.limit_in_bytes时，换出是没有用的
在这个cgroup中。然后，换出将不会由cgroup例程和文件完成
缓存被丢弃。但是如上所述，全局LRU可以进行内存换出
从中以确保系统内存管理状态的完整性。你不能禁止
通过cgroup。

2.5回收

每个cgroup维护每个cgroup LRU，其结构与
全局VM。当cgroup超过其限制时，我们首先尝试
从cgroup回收内存，以便为新的磁盘空间
cgroup触摸过的页面。如果回收失败，
调用OOM例程以选择并杀死该任务中最大的任务
cgroup。（请参阅下面的10. OOM控制。）

回收算法尚未针对cgroup进行修改，除了
选择回收的页面来自每个cgroup的LRU
清单。

注意：回收不适用于根cgroup，因为我们无法设置任何
对根cgroup的限制。

注意2：当panic_on_oom设置为“ 2”时，整个系统都会出现紧急情况。

注册oom事件通知程序后，事件将被传递。
（请参见oom_control部分）

2.6锁定

   在以下情况下不应调用lock_page_cgroup（）/ unlock_page_cgroup（）
   i_pages锁。

   其他锁定顺序如下：
   PG_锁定。
   mm-> page_table_lock
       pgdat-> lru_lock
	  lock_page_cgroup。
  在许多情况下，仅调用lock_page_cgroup（）。
  每个区域每cgroup LRU（cgroup的私有LRU）仅受到以下保护
  pgdat-> lru_lock，它没有自己的锁。

2.7内核内存扩展（CONFIG_MEMCG_KMEM）

使用内核内存扩展，内存控制器可以限制
系统使用的内核内存量。内核内存从根本上讲
与用户内存不同，因为它不能被换出，因此
通过消耗过多的宝贵资源来对系统进行DoS。

默认情况下，为所有内存cgroup启用内核内存记帐。但
可以通过将cgroup.memory = nokmem传递给内核来在系统范围内禁用它
在启动时。在这种情况下，根本不会考虑内核内存。

不对根cgroup施加内核内存限制。根的用法
cgroup可能会也可能不会被考虑。所使用的内存累积到
memory.kmem.usage_in_bytes，或在有意义的情况下放在单独的计数器中。
（当前仅适用于tcp）。
主“ kmem”计数器被馈送到主计数器，因此kmem费用将
也可以从用户计数器中看到。

当前，内核内存尚未实现软限制。这是未来的工作
达到这些限制时触发板坯回收。

2.7.1当前内核内存资源占

*堆栈页面：每个进程都消耗一些堆栈页面。通过会计入
内核内存，我们可以防止在内核时创建新进程
内存使用率太高。

*平板页面：跟踪由SLAB或SLUB分配器分配的页面。复印件
每次第一次触摸缓存时，都会创建每个kmem_cache的
从memcg内部。创作是懒惰完成的，因此某些对象仍然可以
在创建缓存时跳过。平板页面中的所有对象均应
属于同一个memcg。仅当任务迁移到
在缓存分配页面期间使用不同的memcg。

*套接字内存压力：某些套接字协议具有内存压力
阈值。内存控制器允许对它们进行单独控制
每个cgroup，而不是全局。

* tcp内存压力：套接字tcp协议的内存压力。

2.7.2常见用例

由于“ kmem”计数器被馈送到主用户计数器，因此内核内存可以
永远不会完全不受用户内存限制。说“ U”是用户
限制，内核限制为“ K”。有三种可能的限制方式
组：

    U！= 0，K =无限：
    这是kmem之前已经存在的标准memcg限制​​机制
    会计。内核内存被完全忽略。

    U！= 0，K <U：
    内核内存是用户内存的子集。此设置在以下方面很有用
    过度使用每个cgroup的内存总量的部署。
    绝对不建议过度使用内核内存限制，因为
    框仍会耗尽不可回收的内存。
    在这种情况下，管理员可以设置K，以便所有组的总和为
    决不大于总内存，并以其为代价自由设置U
    QoS。
    警告：在当前的实现中，不会回收内存
    当cgroup保持在U以下时，当其击中K时触发该cgroup
    此设置不切实际。

    U！= 0，K> = U：
    由于kmem费用也将被馈送到用户计数器并进行回收
    为两种内存的cgroup触发。此设置为
    管理内存的统一视图，这对于那些只是
    想跟踪内核内存使用情况。

3.用户界面

3.0。组态

一种。启用CONFIG_CGROUPS
b。启用CONFIG_MEMCG
C。启用CONFIG_MEMCG_SWAP（以使用交换扩展名）
d。启用CONFIG_MEMCG_KMEM（以使用kmem扩展名）

3.1。准备cgroup（请参阅cgroups.txt，为什么需要cgroup？）
＃mount -t tmpfs none / sys / fs / cgroup
＃mkdir / sys / fs / cgroup /内存
＃mount -t cgroup none / sys / fs / cgroup / memory -o内存

3.2。组成新小组并将bash移入其中
＃mkdir / sys / fs / cgroup / memory / 0
＃echo $$> / sys / fs / cgroup / memory / 0 / tasks

由于现在我们位于0 cgroup中，因此我们可以更改内存限制：
＃echo 4M> /sys/fs/cgroup/memory/0/memory.limit_in_bytes

注意：我们可以使用后缀（k，K，m，M，g或G）表示以千克为单位的值，
兆或千兆字节。（此处，Kilo，Mega和Giga是Kibibytes，Mebibytes和Gibibytes。）

注意：我们可以写“ -1”来重置* .limit_in_bytes（无限制）。
注意：我们不能再在根cgroup上设置限制。

＃cat /sys/fs/cgroup/memory/0/memory.limit_in_bytes
4194304

我们可以检查用法：
＃cat /sys/fs/cgroup/memory/0/memory.usage_in_bytes
1216512

成功写入此文件并不能保证成功设置
此限制为写入文件的值。这可能是由于
因素数量，例如四舍五入到页面边界或总数
系统上的内存可用性。要求用户重新阅读
写入后保证该文件由内核提交。

＃回声1> memory.limit_in_bytes
＃cat memory.limit_in_bytes
4096

memory.failcnt字段提供了cgroup限制为
超出。

memory.stat文件提供了记帐信息。现在，数量
显示缓存，RSS和活动页面/非活动页面。

4.测试

有关测试功能和实现，请参阅memcg_test.txt。

性能测试也很重要。要查看纯内存控制器的开销，
在tmpfs上进行测试将为您带来许多小的开销。
示例：在tmpfs上执行内核make。

页面故障可伸缩性也很重要。在平行测量
页面错误测试，多进程测试可能比多线程更好
进行测试，因为它具有共享对象/状态的噪音。

但是以上两个正在测试极端情况。
在内存控制器下尝试常规测试总是很有帮助的。

4.1故障排除

有时用户可能会发现cgroup下的应用程序是
被OOM杀手终止。造成这种情况的原因有很多：

1. cgroup限制太低（太低而无法执行任何有用的操作）
2.用户正在使用匿名内存，并且交换已关闭或太低

同步后回显1> / proc / sys / vm / drop_caches将有助于摆脱
cgroup中缓存的某些页面（页面缓存页面）。

要知道会发生什么，请按照“ 10. OOM控制”（如下）禁用OOM_Kill并
看看会发生什么会有所帮助。

4.2任务迁移

当任务从一个cgroup迁移到另一个cgroup时，其费用不会
默认情况下结转。从原始cgroup分配的页面仍然
保持收费，释放页面或
收回。

您可以随任务迁移一起移动任务的费用。
请参阅8.“在任务迁移时收费”

4.3删除cgroup

可以通过rmdir删除cgroup，但是正如第4.1和4.2节所述，
即使所有的cgroup都有一些费用
任务已从其迁移。（因为我们对网页收费，而不是
针对任务。）

我们将统计信息移至根目录（如果use_hierarchy == 0）或父目录（如果
use_hierarchy == 1），并且除充电外没有其他费用变化
从孩子。

交换信息中记录的费用在删除cgroup时不会更新。
记录的信息将被丢弃，而使用交换的cgroup（swapcache）
将作为新所有者收取。

关于use_hierarchy，请参阅第6节。

5.其他。接口。

5.1 force_empty
  提供memory.force_empty接口可以使cgroup的内存使用为空。
  当写任何东西

  ＃回声0> memory.force_empty

  cgroup将被回收，并回收尽可能多的页面。

  该接口的典型用例是在调用rmdir（）之前。
  尽管rmdir（）使memcg脱机，但是由于以下原因，该memcg可能仍停留在该位置：
  收费的文件缓存。某些不使用的页面缓存可能会继续收费，直到
  发生内存压力。如果要避免这种情况，force_empty将很有用。

  另外，请注意，在设置memory.kmem.limit_in_bytes时，由于
  内核页面仍然可见。这不被视为失败，并且
  写入仍将返回成功。在这种情况下，预计
  memory.kmem.usage_in_bytes == memory.usage_in_bytes。

  关于use_hierarchy，请参阅第6节。

5.2统计文件

memory.stat文件包括以下统计信息

＃每个内存cgroup的本地状态
cache-页面缓存的字节数。
rss-匿名和交换缓存内存的字节数（包括
		透明的大页面）。
rss_huge-匿名透明大页面的字节数。
appedd_file-映射文件的字节数（包括tmpfs / shmem）
pgpgin-内存cgroup的充电事件数。充电中
		每次将页面计为任一映射时都会发生事件
		匿名页面（RSS）或缓存页面（页面缓存）到cgroup。
pgpgout-向内存cgroup释放事件的次数。正在充电
		每次cgroup中没有页面时，都会发生此事件。
swap-交换使用的字节数
脏-等待被写回到磁盘的字节数。
写回-已排队等待同步的文件/匿名缓存的字节数
		磁盘。
inactive_anon-不活动时匿名和交换缓存内存的字节数
		LRU列表。
active_anon-活动时匿名和交换高速缓存内存的字节数
		LRU列表。
inactive_file-非活动LRU列表上文件支持的内存的字节数。
active_file-活动LRU列表上文件支持的内存的字节数。
不可撤销-无法回收（锁定等）的内存字节数。

＃考虑状态的状态（请参阅memory.use_hierarchy设置）

boundary_memory_limit-关于层次结构的内存限制的字节数
			内存cgroup所在的位置
boundary_memsw_limit-关于的内存+交换限制的字节数
			内存cgroup所在的层次结构。

total_ <counter>-<counter>的＃分层版本，其中
			除了cgroup自身的价值外，还包括
			所有等级儿童的值的总和
			<counter>，即total_cache

＃以下其他统计信息取决于CONFIG_DEBUG_VM。

last_rotated_anon-VM内部参数。（请参阅mm / vmscan.c）
last_rotated_file-VM内部参数。（请参阅mm / vmscan.c）
last_scanned_anon-VM内部参数。（请参阅mm / vmscan.c）
最近扫描的文件-VM内部参数。（请参阅mm / vmscan.c）

备忘录：
	last_rotated表示LRU旋转的最近频率。
	最近扫描的意思是最近对LRU的扫描次数。
	为了更好地显示调试信息，请参见代码以获取含义。

注意：
	仅匿名和交换缓存内存被列为“ rss”状态的一部分。
	请勿将其与真实的“居民人数”或
	cgroup使用的物理内存量。
	'rss + appedd_file”将为您提供cgroup的常驻设置大小。
	（注意：文件和shmem可能在其他cgroup之间共享。在这种情况下，
	 仅当内存cgroup是页的所有者时，才会考虑appedd_file
	 缓存）。

5.3交换性

覆盖特定组的/ proc / sys / vm / swappiness。可调参数
根cgroup中的对应于全局swappiness设置。

请注意，与全局回收不同，限制回收
强制说0 swappiness确实防止了任何交换，即使
有可用的交换存储。这可能会导致memcg OOM杀手
如果没有要回收的文件页面。

5.4失败

内存cgroup提供memory.failcnt和memory.memsw.failcnt文件。
此failcnt（==故障计数）显示使用计数器的次数
达到极限。当内存cgroup达到极限时，failcnt会增加，并且
内存将被回收。

您可以通过将0写入failcnt文件来重置failcnt。
＃回声0> ... / memory.failcnt

5.5用法_in_bytes

为了提高效率，内存cgroup与其他内核组件一样进行了一些优化
以避免不必要的缓存行错误共享。usage_in_bytes受
方法，并且不显示内存（和交换）使用情况的“确切”值，这很模糊
有效访问的价值。（当然，在必要时，它是同步的。）
如果您想知道更确切的内存使用情况，则应使用RSS + CACHE（+ SWAP）
在memory.stat中的值（请参阅5.2）。

5.6 numa_stat

这类似于numa_maps，但以每个内存为基础进行操作。这是
可用于提供对其中numa位置信息的可见性
memcg，因为允许从任何物理页面分配页面
节点。用例之一是通过
将此信息与应用程序的CPU分配相结合。

每个memcg的numa_stat文件包括“总计”，“文件”，“匿名”和“无法胜任”
每个节点的页面计数，包括“ hierarchical_ <counter>”，该总数总计
除了memcg自身的价值外，还具有分级的儿童价值。

memory.numa_stat的输出格式为：

total = <总页数> N0 = <节点0页> N1 = <节点1页> ...
file = <总文件页数> N0 = <节点0页数> N1 = <节点1页数> ...
anon = <全部anon页面> N0 = <节点0页> N1 = <节点1页> ...
unevictable = <全部匿名页面> N0 = <节点0页面> N1 = <节点1页面> ...
等级_ <计数器> = <计数器页> N0 = <节点0页> N1 = <节点1页> ...

“总数”计数是文件+负离子+不可胜诉的总和。

6.层次结构支持

内存控制器支持深度层次结构和层次结构记帐。
通过在以下位置创建适当的cgroup来创建层次结构
cgroup文件系统。考虑下面的cgroup文件系统
等级制

	       根
	     /  |   \
            /	|    \
	   公元前
		      | \
		      |  \
		      德

在上图中，启用了分层记帐后，所有内存
e的使用直到其根（即c和root）为止都一直由其祖先承担，
启用了memory.use_hierarchy。如果一位祖先越过
限制，回收算法会从祖先和
祖先的孩子。

6.1启用分层记帐和回收

默认情况下，内存cgroup禁用层次结构功能。支持
可以通过向根cgroup的memory.use_hierarchy文件写入1来启用

＃回声1> memory.use_hierarchy

该功能可以通过以下方式禁用

＃回声0> memory.use_hierarchy

注意1：如果一个cgroup已经有另一个，则启用/禁用将失败。
       在其下创建的cgroup，或者父cgroup具有use_hierarchy
       已启用。

注意2：当panic_on_oom设置为“ 2”时，整个系统将在
       任何cgroup中发生OOM事件的情况。

7.软限制

软限制允许更大的内存共享。软限制背后的想法
是为了允许控制组根据需要使用尽可能多的内存，

一种。没有内存争用
b。他们没有超过硬限制

当系统检测到内存争用或内存不足时，控制组
被推回软极限。如果软限制每个控件
小组很高，他们被尽可能地推回去
确保一个对照组不会饿死其他对照组。

请注意，软限制是尽力而为功能；它带有
没有保证，但是尽最大努力确保何时存储
竞争激烈，根据软限制分配内存
提示/设置。当前基于软限制的回收设置为
它从balance_pgdat（kswapd）调用。

7.1介面

可以使用以下命令来设置软限制（在本示例中，
假设软限制为256 MiB）

＃回显256M> memory.soft_limit_in_bytes

如果要将其更改为1G，我们可以随时使用

＃echo 1G> memory.soft_limit_in_bytes

注意1：软限制会在很长一段时间内生效，因为它们涉及
       回收内存以在内存cgroup之间进行平衡
注意2：建议将软限制设置为始终低于硬限制，
       否则，硬限制将优先。

8.在任务迁移时转移费用

用户可以移动与任务相关的费用以及任务迁移，
是，从旧cgroup卸载任务页面，然后将其充电到新cgroup。
！CONFIG_MMU环境不支持此功能，因为缺少
页表。

8.1界面

默认情况下禁用此功能。可以通过以下方式启用（并再次禁用）
写入目标cgroup的memory.move_charge_at_immigrate。

如果要启用它：

＃回声（一些正值）> memory.move_charge_at_immigrate

注意：move_charge_at_immigrate的每个位对于什么类型都有自己的含义
      费用应予调动。有关详细信息，请参见8.2。
注意：仅当您移动mm-> owner时，电荷才会移动，换句话说，
      线程组的负责人。
注意：如果我们在目标cgroup中找不到足够的空间来执行任务，则我们
      尝试通过回收内存来腾出空间。如果我们迁移任务可能会失败
      无法腾出足够的空间。
注意：如果您移动很多电荷，可能需要几秒钟。

如果要再次禁用它：

＃回声0> memory.move_charge_at_immigrate

8.2可以转移的费用类型

move_charge_at_immigrate中的每一位对于什么类型的
费用应予调动。但无论如何，必须注意的是
页面或交换只有在按任务的当前费用计费时才能移动
（旧）内存cgroup。

  位| 什么类型的费用会被转移？
 -----+------------------------------------------------------------------------
   0 | 由目标任务使用的匿名页面（或其交换）的费用。
      | 您必须启用交换扩展（请参阅2.4）才能启用交换费用的移动。
 -----+------------------------------------------------------------------------
   1 | 文件页面收费（普通文件，tmpfs文件（例如ipc共享内存）
      | 和目标任务映射的tmpfs文件交换）。与情况不同
      | 任务映射范围内的匿名页面，文件页面（和交换）
      | 即使任务尚未完成页面错误也将被移动，即它们可能
      | 不是任务的“ RSS”，而是其他映射相同文件的任务的“ RSS”。
      | 并且页面的mapcount被忽略（即使
      | page_mapcount（page）> 1）。您必须启用交换扩展（请参阅2.4）才能
      | 启用掉期费用转移。

8.3全部

-所有移动充电操作均在cgroup_mutex下完成。这不好
  行为将互斥量保持太长时间，因此我们可能需要一些技巧。

9.内存阈值

内存cgroup使用cgroups通知实现内存阈值
API（请参阅cgroups.txt）。它允许注册多个内存和memsw
越过阈值并获得通知。

要注册阈值，应用程序必须：
-使用eventfd（2）创建一个eventfd；
-打开memory.usage_in_bytes或memory.memsw.usage_in_bytes；
-将类似“ <event_fd> <memory.usage_in_bytes> <threshold>”的字符串写入
  cgroup.event_control。

当内存使用量超过时，将通过eventfd通知应用程序
任何方向的阈值。

适用于root和非root cgroup。

10. OOM控制

memory.oom_control文件用于OOM通知和其他控件。

内存cgroup使用cgroup通知实现OOM通知程序
API（请参阅cgroups.txt）。它允许注册多个OOM通知
交付并在发生OOM时获得通知。

要注册通知者，应用程序必须：
 -使用eventfd（2）创建一个eventfd
 -打开memory.oom_control文件
 -将类似“ <event_fd> <memory.oom_control>的fd”的字符串写入
   cgroup.event_control

OOM发生时，将通过eventfd通知该应用程序。
OOM通知不适用于根cgroup。

您可以通过将“ 1”写入memory.oom_control文件来禁用OOM杀手，如下所示：

	#echo 1> memory.oom_control

如果禁用OOM-killer，则cgroup下的任务将挂起/休眠
当他们请求负责的内存时，cgroup的OOM-waitqueue在内存中。

要运行它们，您必须通过以下方式放松内存cgroup的OOM状态：
	*扩大限制或减少使用量。
为了减少使用量，
	*杀死一些任务。
	*通过帐户迁移将一些任务移至其他组。
	*删除一些文件（在tmpfs上？）

然后，已停止的任务将再次起作用。

读取时，显示OOM的当前状态。
	oom_kill_disable 0或1（如果为1，则禁用oom-killer）
	under_oom 0或1（如果为1，则内存cgroup在OOM下，任务可能
				 停下来。）

11.记忆压力

压力水平通知可用于监视内存
分配费用；根据压力，应用程序可以实施
管理内存资源的不同策略。压力
级别定义如下：

“低”级别表示系统正在回收新的内存
分配。监视此回收活动可能对
保持缓存级别。收到通知后，该程序（通常
“活动管理器”）可能会分析vmstat并事先采取行动（即
过早关闭不重要的服务）。

“中等”级别表示系统正在经历中等内存
压力，系统可能正在进行交换，调出活动文件缓存，
等等。根据此事件，应用程序可能决定进一步分析
vmstat / zoneinfo / memcg或内部内存使用情况统计信息并释放任何
可以轻松重建或从磁盘重新读取的资源。

“关键”级别表示系统正在主动运行，这是
即将用尽内存（OOM）甚至内核中的OOM杀手
触发方式。应用程序应尽力帮助
系统。与vmstat或任何其他公司协商可能为时已晚
统计信息，因此建议立即采取行动。

默认情况下，事件向上传播，直到事件被处理为止，即
事件不传递。例如，您有三个cgroup：A-> B-> C。现在
您在cgroup A，B和C上设置了一个事件侦听器，并假设C组
遇到一些压力。在这种情况下，只有C组会收到
通知，即A和B组将不会收到该通知。这样做是为了避免
消息的过度“广播”，这会干扰系统，并且
如果我们的内存不足或崩溃，则尤其糟糕。B组，将收到
仅当C组没有事件列表时通知。

有三种指定不同传播行为的可选模式：

 -“默认”：这是上面指定的默认行为。此模式是
   与省略可选模式参数（由向后保留）相同
   兼容性。

 -“层次结构”：事件始终传播到根，类似于默认值
   行为，除了无论是否存在
   每个级别的事件侦听器，都具有“层次结构”模式。在上面
   例如，组A，B和C将收到内存压力通知。

 -“本地”：事件是传递的，即事件仅在以下情况下接收通知
   内存压力在通知所针对的memcg中遇到
   注册。在上面的示例中，如果
   注册以获取“本地”通知，并且该小组可以进行记忆
   压力。但是，B组将永远不会收到通知，无论是否
   如果组B已注册，则是否有C组的事件侦听器
   本地通知。

级别和事件通知模式（如有必要，为“层次结构”或“本地”）
用逗号分隔的字符串指定，即“ low，hierarchy”指定
所有祖先memcg的分层，传递，通知。通知
这是默认的非直通行为，未指定模式。
“ medium，local”指定中等级别的传递通知。

文件memory.pressure_level仅用于设置eventfd。至
注册通知，应用程序必须：

-使用eventfd（2）创建一个eventfd；
-打开memory.pressure_level;
-将字符串写为“ <event_fd> <memory.pressure_level> <level [，mode]>”
  到cgroup.event_control。

当内存不足时，将通过eventfd通知应用程序
具体级别（或更高）。读/写操作
memory.pressure_level未实现。

测试：

   这是一个小脚本示例，它创建一个新的cgroup，设置一个
   内存限制，在cgroup中设置一个通知，然后将其设为子级
   cgroup遇到严重压力：

   ＃cd / sys / fs / cgroup / memory /
   ＃mkdir foo
   ＃cd foo
   ＃cgroup_event_listener memory.pressure_level低，层次结构＆
   ＃回声8000000> memory.limit_in_bytes
   ＃回声8000000> memory.memsw.limit_in_bytes
   ＃echo $$>任务
   ＃dd if = / dev / zero | 读x

   （希望收到一堆通知，最终，oom-killer会
   触发。）

12.全部

1.首先让每组扫描程序收回未共享的页面
2.教控制器说明共享页面
3.当限制为时，在后台开始回收
   尚未击中，但用法越来越近

摘要

总体而言，内存控制器一直是稳定的控制器，并且已经
在社区中进行了广泛的评论和讨论。

参考文献

1.辛格，巴尔比尔。RFC：内存控制器，http：//lwn.net/Articles/206697/
2.辛格，巴尔比尔。内存控制器（RSS控制），
   http://lwn.net/Articles/222762/
3. Emelianov，Pavel。基于进程cgroup的资源控制器
   http://lkml.org/lkml/2007/3/6/198
4. Emelianov，Pavel。基于进程cgroup（v2）的RSS控制器
   http://lkml.org/lkml/2007/4/9/78
5. Emelianov，Pavel。基于进程cgroup（v3）的RSS控制器
   http://lkml.org/lkml/2007/5/30/244
6.保禄，保罗。控制组v10，http：//lwn.net/Articles/236032/
7. Vaidyanathan，Srinivasan，控制组：Pagecache记帐和控制
   子系统（v3），http：//lwn.net/Articles/235534/
8.辛比，巴尔比尔。RSS控制器v2测试结果（lmbench），
   http://lkml.org/lkml/2007/5/17/232
9.辛格，巴尔比尔。RSS控制器v2 AIM9结果
   http://lkml.org/lkml/2007/5/18/1
10.辛格，巴尔比尔。内存控制器v6测试结果，
    http://lkml.org/lkml/2007/8/19/36
11.辛格，巴尔比尔。内存控制器介绍（v6），
    http://lkml.org/lkml/2007/8/17/69
12. Corbet，Jonathan，控制cgroup中的内存使用情况，
    http://lwn.net/Articles/243795/