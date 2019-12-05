CPU提供了一种分配一组CPU和内存到一组任务的机制。

在本文档中，“Memory Node”是指包含内存的在线节点。 
CPUset任务的CPU和内存位置限制为仅任务当前cpuset中的资源。
它们形成了在虚拟文件系统中可见的嵌套层次结构。
这些是在大型系统上管理动态作业放置所需的基本钩子，而不是已经存在的钩子。

 cpuset使用Documentation / cgroup-v1 / cgroups.txt中描述的通用cgroup子系统。

通过使用sched_setaffinity（2）系统调用将CPU包括在其CPU亲和力掩码
中以及使用mbind（2）和set_mempolicy（2）系统调用将内存节点包括在其内存策略中的任务请求
均通过该请求进行过滤任务的cpuset，过滤掉任何
不在该cpuset中的CPU或内存节点。
调度程序将不会在其cpus_allowed向量中不允许的CPU上调度任务，
并且内核页面分配器不会在请求任务的mems_allowed向量中不允许的节点上分配页面。
 
用户级代码可以按名称在cgroup虚拟文件系统中创建和销毁cpuset，管理这些cpuset的属性和权限，以及为每个cpuset分配了哪些CPU和内存节点，
指定并查询任务分配给哪个cpuset，并列出分配给cpuset的任务pid。

1.2为什么需要cpusets?
----------------------------
管理具有许多处理器（CPU）的大型计算机系统的，
复杂的内存高速缓存层次结构和具有不均匀访问时间（NUMA）
的多个内存节点为有效调度进程和内存放置提出了其他挑战。

通常，仅通过让操作系统自动在请求的任务之间共享可用的CPU和内存资源，
便可以以足够高的效率来运行大小较小的系统。

但是大型系统可以通过将作业明确放置在适当大小的子集上而受益，
大型系统得益于精心的处理器和内存放置，以减少内存访问时间和争用，并且通常为客户带来更大的投资。
系统。

这在以下方面尤其有价值：
    *运行同一Web应用程序的多个实例的Web服务器，*运行不同应用程序的服务器（例如，Web服务器和数据库），或*运行大型HPC应用程序且要求很高的NUMA系统
      性能特点。这些子集或“软分区”必须能够随着工作组合的变化而动态调整，而不会同时影响其他子集
执行工作。当更改内存位置时，也可以移动正在运行的作业页面的位置。内核cpuset补丁提供了最低限度的必需内核
有效实现此类子集所需的机制。它利用Linux内核中现有的CPU和内存放置功能，以避免对关键调度程序或内存分配器代码产生任何其他影响。

  1.3 cpusets如何实现？
---------------------------------

Cpuset提供了一种Linux内核机制来约束一个进程或一组进程使用哪个CPU和内存节点。

Linux内核已经有一对机制来指定可以在哪个CPU上调度任务（sched_setaffinity），以及可以在哪个Memory Node上获得内存（mbind，set_mempolicy）。

Cpusets扩展了以下两种机制
- Cpusets是内核已知的允许的CPU和内存节点的集合。
- 系统中的每个任务都通过任务结构中指向引用计数的cgroup结构的指针连接到cpuset。
- 对sched_setaffinity的调用仅过滤到该任务的cpuset中允许的CPU。
- 对mbind和set_mempolicy的调用仅过滤到该任务的cpuset中允许的那些内存节点。
- 根cpuset包含所有系统CPU和内存节点。
- 对于任何cpuset，可以定义包含父CPU和内存节点资源的子集的子cpuset。 
- 可以将cpuset的层次结构安装在/ dev / cpuset上，以便从用户空间进行浏览和操作。
- 一个cpuset可能被标记为“独占”，以确保没有其他cpuset（直接祖先和后代除外）可以包含任何重叠的CPU或内存节点。
- 您可以列出附加到任何cpuset的所有任务（按pid）。

cpusets的实现需要一些简单的钩子连接到内核的其余部分，而在性能关键路径中则没有钩子：

- 在init/main.c中，在系统引导时初始化根cpuset。
- 在fork和exit中，将任务从其cpuset附加和分离。
- 在sched_setaffinity中，通过该任务的cpuset中允许的内容屏蔽请求的CPU。
- 在sched.c migrate_live_tasks（）中，如果可能的话，在cpuset允许的CPU内继续迁移任务。
- 在mbind和set_mempolicy系统调用中，通过以下方式屏蔽请求的内存节点
该任务的cpuset中允许的内容。
 -在page_alloc.c中，将内存限制为允许的节点。
-在vmscan.c中，将页面恢复限制为当前cpuset

您应该挂载“ cgroup”文件系统类型以便启用
浏览和修改内核当前已知的cpuset。没有为cpuset添加新的系统调用-通过此cpuset文件系统对查询和修改
cpuset的所有支持。

每个任务的/ proc / <pid> / status文件添加了四行，分别以以下两种格式显示任务的cpus_allowed（可以在其上计划其CPU）和mems_allowed（可以在其上获得内存的内存节点）。下面的例子：

Cpus_allowed：ffffffff，ffffffff，ffffffff，ffffffff
Cpus_allowed_list：0-127
Mems_allowed：ffffffff，ffffffff
Mems_allowed_list：0-63

每个cpuset均由cgroup文件系统中的目录表示，该目录包含（在标准cgroup文件之上）以下内容：
描述该cpuset的文件

- cpuset.cpus：该cpuset中的CPU列表-cpuset.mems：该cpuset中的内存节点列表
- cpuset.memory_migrate标志：如果已设置，则将页面移至cpusets节点
- cpuset.cpu_exclusive标志：cpu展示位置是否排他？
- cpuset.mem_exclusive标志：内存位置是否排他？ 
- cpuset.mem_hardwall标志：内存分配是否为硬壁
- cpuset.memory_pressure：测量cpuset中有多少分页压力
- cpuset.memory_spread_page标志：如果设置，则在允许的节点上平均扩展页面缓存
- cpuset.memory_spread_slab标志：如果设置，则在允许的节点上平均扩展slab缓存
- cpuset.sched_load_balance标志：如果设置，则在该cpuset上CPU内的负载平衡
- cpuset.sched_relax_domain_level：迁移任务时的搜索范围

此外，仅根cpuset具有以下文件：
- cpuset.memory_pressure_enabled标志：计算memory_pressure？
 
使用mkdir系统调用或shell命令创建新的cpuset。通过写操作可以修改cpuset的属性，
例如其标志，允许的CPU和内存节点以及附加的任务。

到上述cpusets目录中的相应文件。嵌套cpuset的命名层次结构允许将大型系统划分为嵌套的，可动态更改的“soft-partitions”。

将每个任务的子任务自动继承的每个任务附加到cpuset，可以将系统上的工作负载组织到相关的任务集中，
从而限制每个任务集
使用特定cpuset的CPU和内存节点。如果必要的cpu​​set文件系统目录的权限允许，
则任务可以重新附加到任何其他cpuset。

通过使用sched_setaffinity，mbind和set_mempolicy系统调用，这种“大型”系统的管理与在单个任务和内存区域上完成的详细放置顺利集成。

以下规则适用于每个cpuset
- 其CPU和内存节点必须是其父级的子集。
- 除非其父项是父项，否则不能将其标记为专有。
- 如果其cpu或内存是互斥的，则它们可能不会与任何同级重叠。

这些规则以及cpusets的自然层次结构可以有效执行排他保证，而不必扫描所有
每次更改cpusets时，以确保没有任何内容与专有cpuset重叠。
另外，使用Linux虚拟文件系统（vfs）表示cpuset层次结构可为cpuset提供熟悉的权限和名称空间，同时还需要最少的附加内核代码。

根（top_cpuset）cpuset中的cpus和mems文件是只读的。 cpus文件使用CPU热插拔通知程序和mems文件自动跟踪cpu_online_mask的值
使用cpuset_track_online_nodes（）挂钩自动跟踪node_states [N_MEMORY]（即具有内存的节点）的值。

1.4什么是专用cpusets？
--------------------------------

如果cpuset是cpu或mem独占的，则除了
直接祖先或后代可以共享任何相同的CPU或内存节点。

一个cpuset.mem_exclusive *或* cpuset.mem_hardwall的cpuset是“hardwalled”，
也就是说，它限制了内核通常在多个用户之间共享的页面，缓冲区和其他数据的内核分配。所有cpuset（无论是否为硬墙）都限制用户空间的内存分配。这样可以配置系统，以便几个独立的
作业可以共享公用内核数据，例如文件系统页面，同时在其自己的cpuset中隔离每个作业的用户分配。为此，请构造一个大型的mem_exclusive cpuset以容纳所有作业，并为每个单独的作业构造子级非mem_exclusive cpuset。
即使是mem_exclusive cpuset，也仅允许将少量的典型内核内存（例如来自中断处理程序的请求）带到外部。

1.5什么是memory_pressure？
-----------------------------

cpuset的memory_pressure提供了一个简单的每个cpuset指标
cpuset中的任务试图释放使用的内存的速率的百分比
在cpuset的节点上满足其他内存请求。

这使批处理管理器可以监视在专用cpuset中运行的作业，以有效地检测该作业造成的内存压力级别。

这在运行大量提交作业的紧密管理的系统上很有用，它们可以选择终止或重新分配那些试图使用比分配给它们的节点更多的内存的优先级的作业，以及紧密耦合，长期运行的作业，大规模并行科学
如果计算作业开始使用超出其允许范围的内存，将大大无法达到所需的性能目标。这种机制为批处理管理器提供了一种非常经济的方式
监视CPU内存不足迹象。由批处理管理器或其他用户代码决定如何处理并采取措施。

==>除非通过在特殊文件/dev/cpuset/memory_pressure_enabled中写入'1'来启用此功能，否则__alloc_pages（）的重新平衡代码中针对该度量的钩子将减少为仅通知cpuset_memory_pressure_enabled标志为零。所以只有
    启用此功能的系统将计算指标。为什么要按每CPU运行平均值：
    由于此计量表是按CPU而不是按任务或mm的，因此在大型系统上，监视此指标的批处理调度程序所施加的系统负载将大大减少，因为可以避免在每组查询上都扫描任务列表。
     因为此仪表是运行平均值，而不是累加计数器，所以批处理调度程序可以通过一次读取来检测内存压力，而不必读取和累加结果
    一段时间。因为此仪表是按CPU而不是按任务或mm的，所以批处理调度程序可以获得关键信息，内存
    只需读取一次即可在cpuset中产生压力，而不必查询和累积cpuset中所有（动态变化的）任务集的结果。
如果每个cpuset简单数字滤波器（输入同步（直接）页面回收代码），则该数字滤波器会保留（并需要一个自旋锁和每个cpuset 3个数据字），并由该cpuset附带的任何任务更新。
每个cpuset文件提供一个整数，以整数表示每秒由cpuset中的任务引起的直接页面回收的比率（半衰期为10秒），以每秒尝试回收的次数乘以1000。
  1.6什么是内存扩散？
---------------------------

每个cpuset有两个布尔标志文件，它们控制内核在哪里为文件系统缓冲区分配页面以及与内核数据结构相关的页面。它们分别称为“ cpuset.memory_spread_page”和“ cpuset.memory_spread_slab”。
 如果设置了每个cpuset布尔标志文件'cpuset.memory_spread_page'，则内核将在允许该故障任务使用的所有节点上平均分配文件系统缓冲区（页面缓存）。
倾向于将这些页面放在运行任务的节点上。如果设置了每个CPU的布尔标志文件'cpuset.memory_spread_slab'，则内核将散布一些与文件系统相关的slab缓存，
例如在允许错误任务使用的所有节点上平均分配inode和dentries，而不是将这些页面放在任务正在运行的节点上。
这些标志的设置不会影响任务的匿名数据段或堆栈段页面。默认情况下，两种内存扩展都处于关闭状态，并且内存
页可能会在任务运行所在的本地节点上分配，除非由任务的NUMA内存或cpuset配置修改，只要有足够的可用内存页即可。
创建新的cpuset时，它们会继承其父级的内存扩展设置。设置内存扩展会为受影响的页面分配资源
或平板缓存来忽略任务的NUMA记忆，而是进行分散。使用mbind（）或set_mempolicy（）调用来设置NUMA内存的任务将不会注意到这些调用中的任何更改，因为它们包含任务的内存扩展设置。如果内存扩散
如果将其关闭，则当前指定的NUMA内存将再次应用于内存页分配。 'cpuset.memory_spread_page'和'cpuset.memory_spread_slab'均为布尔值标志
文件。默认情况下，它们包含“ 0”，表示该cpuset的功能已关闭。如果将“ 1”写入该文件，则将打开命名功能。
实现很简单。设置标志'cpuset.memory_spread_page'会为该cpuset中或随后的每个任务打开每个进程的标志PFA_SPREAD_PAGE
加入那个cpuset。修改了对页面缓存的页面分配调用，以对此PFA_SPREAD_PAGE任务标志执行内联检查，如果设置了该调用，则对新例程cpuset_mem_spread_node（）的调用将返回希望进行分配的节点。
 同样，设置'cpuset.memory_spread_slab'将打开标志PFA_SPREAD_SLAB，并适当地标记slab缓存w
无法从cpuset_mem_spread_node（）返回的节点分配页面。
 cpuset_mem_spread_node（）例程也很简单。它使用每个任务转子cpuset_mem_spread_rotor的值来选择当前任务的mems_allowed中的下一个节点，以偏爱该节点。
 这种内存放置策略在其他情况下也称为循环或交错。
对于需要将线程本地数据放置在相应节点上但需要访问大型文件系统数据集的作业，该策略可以提供实质性的改进，这些大型文件系统数据集需要分布在作业cpuset中的多个节点上才能适应。没有这个
策略，特别是对于可能在数据集中读取一个线程的作业，作业cpuset中节点之间的内存分配可能变得非常不均匀。

1.7什么是sched_load_balance？
--------------------------------

内核调度程序（kernel / sched / core.c）自动进行负载平衡
任务。如果未充分利用一个CPU，则在该CPU上运行的内核代码将在cpusets和sched_setaffinity之类的放置机制的约束下，在其他负载更重的CPU上查找任务并将这些任务移至自身。
 负载均衡的算法成本及其对关键共享内核数据结构（如任务列表）的影响，随着所均衡的CPU数量的增加，线性增加的幅度更大。所以调度器
支持将系统CPU划分为多个预定域，以便仅在每个预定域内进行负载平衡。每个调度域都覆盖系统中CPU的某些子集。没有两个预定域重叠；某些CPU可能不在任何计划中
域，因此不会达到负载平衡。简而言之，在两个较小的sched域之间进行平衡所花费的成本比一个大的域要少，但是这样做意味着其中一个过载。
两个域将不会与另一个域保持负载平衡。默认情况下，有一个预定域覆盖所有CPU，包括那些使用内核启动时间“ isolcpus =”参数标记为隔离的CPU。然而，
除非明确分配，否则隔离的CPU将不会参与负载平衡，并且不会在其上运行任务。所有CPU上的默认负载均衡均不适用于

以下两种情况：

1）在大型系统上，跨多个CPU的负载平衡非常昂贵。如果使用cpusets管理系统以将独立的作业放置在不同的CPU组上，则不需要完全的负载平衡。
 
2）在某些CPU上支持实时的系统需要使这些CPU上的系统开销最小化，包括避免不必要的任务负载平衡。

启用每个cpuset标志“ cpuset.sched_load_balance”（默认设置）时，它要求该cpuset中允许“ cpuset.cpus”的所有CPU都包含在单个预定域中，以确保负载平衡可以移动任务（未固定，如sched_setaffinity）
从该cpuset中的任何CPU到任何其他CPU。如果禁用了每个cpuset标志“ cpuset.sched_load_balance”，则调度程序将避免该cpuset中各个CPU的负载平衡，
--except--在必要的范围内，因为某些重叠的cpuset启用了“ sched_load_balance”。因此，例如，如果顶部cpuset具有标志“ cpuset.sched_load_balance”
启用后，调度程序将具有一个覆盖所有CPU的调度域，并且其他cpuset中的“ cpuset.sched_load_balance”标志的设置将无关紧要，因为我们已经完全实现了负载平衡。
因此，在以上两种情况下，应禁用顶部的cpuset标志“ cpuset.sched_load_balance”，并且仅某些较小的子cpuset启用此标志。
这样做时，您通常不希望将任何未固定的任务留在可能占用大量CPU的顶级CPU中，因为根据此标志设置的具体情况，此类任务可能被人为地限制在某些CPU子集中在后代cpuset中。即使
这样的任务可能会在其他一些CPU中使用闲置的CPU周期，因此内核调度程序可能不会考虑将该任务与未充分利用的CPU进行负载平衡的可能性。
当然，固定在特定CPU上的任务可以留在cpuset中，以禁用“ cpuset.sched_load_balance”，因为这些任务无论如何都不会进行。
在cpuset和sched域之间，这里存在阻抗不匹配的情况。 CPU集是分层的和嵌套的。预定域是平坦的；它们不重叠，并且每个CPU最多在一个预定的域中。
计划域必须是平坦的，因为部分重叠的CPU组之间的负载平衡可能会带来不稳定的动态变化，这超出了我们的理解。因此，如果两个部分重叠的cpuset中的每一个都启用了标志“ cpuset.sched_load_balance”，那么我们
形成一个单一的预定域，是两者的超集。我们不会将任务移到其cpuset之外的CPU，但是考虑到这种可能性，调度程序负载平衡代码可能会浪费一些计算周期。
这种不匹配就是为什么没有简单的一对一关系b
在哪些CPU启用了“ cpuset.sched_load_balance”标志以及sched域配置之间。如果cpuset启用了该标志，它将在所有CPU之间保持平衡，但是如果禁用了该标志，
如果没有其他重叠的cpuset启用该标志，则只能确保没有负载平衡。如果两个cpuset允许部分重叠的'cpuset.cpus'，并且仅
其中一个启用了此标志，然后另一个仅在重叠的CPU上发现其任务仅部分负载均衡。这只是上面几段给出的top_cpuset示例的一般情况。在一般情况下，例如顶级cpuset情况下，
不要在这样的部分负载均衡的cpuset中保留可能占用大量CPU的任务，因为它们可能由于缺乏对其他CPU的负载均衡而被人为地限制在允许的CPU的某些子集中。
 isolcpus = kernel boot选项将“ cpuset.isolcpus”中的CPU排除在负载平衡之外，无论任何cpuset中的“ cpuset.sched_load_balance”值如何，都将永远不会进行负载平衡。

1.7.1 sched_load_balance的实现细节。
------------------------------------------------

每个cpuset标志'cpuset.sched_load_balance'默认为启用（与大多数cpuset标志相反。）为cpuset启用后，内核将确保它可以在该cpuset中的所有CPU上进行负载平衡（确保所有CPU在该cpuset的cpus_allowed中
如果两个重叠的cpuset都启用了“ cpuset.sched_load_balance”，则它们（必须）都在同一调度域中。
 如果作为默认设置，顶部cpuset启用了'cpuset.sched_load_balance'，则通过上述方法，意味着存在一个覆盖整个系统的调度域，而与其他cpuset设置无关。
 内核致力于用户空间，它将避免在可能的地方进行负载平衡。它会尽可能选择调度域的粒度分区，同时仍为任何集合提供负载平衡
启用了“ cpuset.sched_load_balance”的CPU允许的CPU数量。内部内核cpuset到Scheduler的接口从cpuset代码传递到Scheduler的代码是负载均衡的分区
系统中的CPU。此分区是一组CPU的子集（表示为struct cpumask数组），成对不相交，涵盖必须进行负载平衡的所有CPU。
cpuset代码会构建一个新的分区，并将其传递给调度程序的sched域设置代码，以在必要时重新构建sched域：
 -或启用了此标志的CPU来自或来自cpuset--或启用了该标志的CPU的'cpuset.sched_relax_domain_level'值已更改，或者-启用了该标志的CPU的cpuset与非空CPU的值启用已删除，
 -或cpu离线/在线。此分区准确地定义了调度程序应设置的预定域-计划中每个元素（结构cpumask）的预定域
划分。调度程序会记住当前活动的调度域分区。当调度程序例程partition_sched_domains（）从以下位置调用时
cpuset代码以更新这些计划的域，它将每次请求的新分区与当前分区进行比较，并更新其计划的域，删除旧的并添加新的。

1.8什么是sched_relax_domain_level？
--------------------------------------

在sched域中，调度程序以两种方式迁移任务。定时以及某些计划事件发生时的定期负载平衡。

唤醒任务后，调度程序将尝试在空闲CPU上移动任务。
例如，如果运行在CPU X上的任务A激活了同一CPU X上的另一个任务B，并且如果CPU Y是X的同级并执行空闲，则调度程序将任务B迁移到CPU Y，
以便任务B可以在CPU Y上启动而无需在CPU X上等待任务A。

而且，如果CPU在其运行队列中用尽了所有任务，则该CPU会尝试从其他繁忙的CPU中提取额外的任务，以在它们空闲之前帮助他们。

当然，查找可移动任务和/或空闲CPU会花费一些搜索成本，调度程序可能不会每次都搜索域中的所有CPU。实际上，在某些架构中，搜索范围是
事件限制在CPU所在的同一套接字或节点中，而tick上的负载平衡将全部搜索。

例如，假定CPU Z距离CPU X相对较远。即使CPU Z
当CPU X处于空闲状态且兄弟姐妹处于繁忙状态时，调度程序无法将唤醒的任务B从X迁移到Z，因为它不在其搜索范围内。
结果，CPU X上的任务B需要等待任务A或等待下一个滴答的负载平衡。对于特殊情况下的某些应用，请等待
1个刻度可能太长。 

`cpuset.sched_relax_domain_level`文件可让您请求根据需要更改此搜索范围。该文件的int值是
表示搜索范围的大小，理想情况下为以下级别，否则为初始值-1
表示cpuset没有请求。
- -1：无要求。使用系统默认值或遵循其他人的要求。
- 0：无搜索。
- 1：搜索同级（内核中的超线程)
- 2：在包中搜索核心。
- 3：在节点中搜索cpus [=非NUMA系统上的系统范围]
- 4：在[NUMA系统上]大量节点中搜索节点。
- 5：在[NUMA系统上]搜索整个系统。系统默认值取决于体系结构。系统默认

可以使用Relax_domain_level =引导参数进行更改。

该文件是每个cpuset的文件，会影响cpuset所属的预定域。因此，如果cpuset的标志'cpuset.sched_load_balance'
如果禁用此选项，则`cpuset.sched_relax_domain_level`无效，因为不存在属于cpuset的预定域。如果多个cpuset重叠，因此它们形成一个调度
域中，使用最大值。请注意，如果一个请求为0，而其他请求为-1，则使用0。

请注意，修改此文件将同时带来好与坏的影响，以及是否可以接受取决于您的情况。如果不确定，请不要修改此文件。如果您的情况是：
- 由于您的特殊应用程序的行为或对CPU缓存的特殊硬件支持，可以假设每个cpu之间的迁移成本非常小（对您而言）。
- 搜索成本对您没有影响（或者对您而言）通过管理cpuset使其紧凑等，搜索成本足够小
- 即使它牺牲了高速缓存命中率等，也需要等待时间。然后增加“ sched_relax_domain_level”将使您受益。

1.9如何使用cpusets
--------------------------

为了最大程度地减少cpuset对诸如调度程序之类的关键内核代码的影响，
并且由于内核不支持直接更新另一个任务的内存位置的一个任务这一事实，对更改其cpuset CPU的任务的影响
或“内存节点”放置，或更改任务附加到哪个cpuset上，都是微妙的。

如果cpuset修改了其内存节点，则对于每个附加任务
对于该cpuset，内核在下一次尝试为该任务分配内存页面时，内核将注意到任务cpuset中的更改，
并更新其按任务存储位置以保留在新的cpusets内存位置内。如果任务正在使用
内存MPOL_BIND及其绑定的节点与其新的cpuset重叠，则任务将继续使用新的cpuset中仍允许的MPOL_BIND节点的任何子集。
如果该任务使用的是MPOL_BIND，则现在不允许其任何MPOL_BIND节点
在新的cpuset中，则基本上将任务视为已将MPOL_BIND绑定到新的cpuset（即使由get_mempolicy（）查询的NUMA位置不变）。如果任务从一个cpuset移到另一个，则内核将调整任务的
如上所述，内存放置在内核下一次尝试为该任务分配内存页面时使用。

如果cpuset修改了其“ cpuset.cpus”，则该cpuset中的每个任务
将立即更改其允许的CPU位置。同样，如果将任务的pid写入另一个cpuset的“任务”文件，则其允许的CPU放置将立即更改。
如果使用sched_setaffinity（）调用将此类任务绑定到其cpuset的某个子集，
该任务将被允许在其新cpuset中允许的任何CPU上运行，从而消除了先前sched_setaffinity（）调用的影响。

总而言之，更改了cpuset的任务的内存位置为
在下一次分配该任务的页面时由内核更新，并且处理器位置会立即更新。

通常，一旦分配了页面（给定物理页面
然后，即使cpusets内存放置策略'cpuset.mems'随后发生更改，该页面仍将保留在分配的任何节点上，只要该页面保持分配状态即可。如果cpuset标志文件'cpuset.memory_migrate'设置为true，则何时
任务被附加到该cpuset上，该任务在其先前cpuset中的节点上分配给它的任何页面都将迁移到任务的新cpuset。如果可能，在这些迁移操作期间将保留页面在cpuset中的相对位置。
例如，如果页面位于先前cpuset的第二个有效节点上，则该页面将被放置在新cpuset的第二个有效节点上。

另外，如果将'cpuset.memory_migrate'设置为true，则如果该cpuset的
修改了“ cpuset.mems”文件，分配给该cpuset中的任务的页面（位于先前的“ cpuset.mems”设置中的节点上）将移至新设置的“ mems”中的节点上。
不在任务的先前cpuset或cpuset的页面中的页面
之前的“ cpuset.mems”设置不会被移动。

上面有一个例外。如果使用热插拔功能删除当前分配给cpuset的所有CPU，
那么该cpuset中的所有任务将被移至具有非空cpus的最近祖先。
但是，如果cpuset与另一个对任务附加有限制的cgroup子系统绑定，则某些（或全部）任务的移动可能会失败。在这种情况下，这些任务将保留
在原始cpuset中，内核将自动更新其cpus_allowed以允许所有在线CPU。当可用的用于删除内存节点的内存热插拔功能可用时，类似的例外也将适用于此。通常，内核倾向于
违反了cpuset放置，因为该任务使所有允许的CPU或内存节点脱机。以上是第二个例外。 GFP_ATOMIC请求是
必须立即满足的内核内部分配。如果GFP_ATOMIC分配失败，内核可能会丢弃某些请求，在极少数情况下甚至会出现紧急情况。如果无法在当前任务的cpuset中满足该请求，则我们放宽cpuset，然后寻找
我们可以在任何地方找到它。最好是违反cpuset而不是强调内核。

要开始要包含在cpuset中的新作业，步骤如下：

- mkdir /sys/fs/cgroup/cpuset 
- mount -t cgroup -ocpuset cpuset /sys/fs/cgroup/cpuset 
- 通过执行mkdir并在其中写入（或回显）来创建新的cpuset/sys/fs/cgroup/cpuset虚拟文件系统。
- 开始一项将成为新工作的“founding father”的任务。
- 通过将其pid写入该cpuset的/sys/fs/cgroup/cpuset任务文件中，将该任务附加到新的cpuset。
- 从此创始父任务派生，执行或克隆作业任务。例如，以下命令序列将设置一个名为“ Charlie”的cpuset，其中仅包含CPU 2和3，以及内存节点1，

然后在该cpuset中启动子shell'sh'：
```
  mount -t cgroup -ocpuset cpuset /sys/fs/cgroup/cpuset
  cd /sys/fs/cgroup/cpuset
  mkdir Charlie
  cd Charlie
  /bin/echo 2-3 > cpuset.cpus
  /bin/echo 1 > cpuset.mems
  /bin/echo $$ > tasks
  sh
  # The subshell 'sh' is now running in cpuset Charlie
  # The next line should display '/Charlie'
  cat /proc/self/cpuset
```

有以下几种查询或修改cpuset的方法
-通过cdset，mkdir，echo，
   从外壳程序获取cat，rmdir命令，或从C获得等效命令-通过C库libcpuset。
- 通过C库libcgroup。 （http://sourceforge.net/projects/libcg/）
- 通过python应用程序cset。 （http://code.google.com/p/cpuset/）

sched_setaffinity调用也可以在shell提示符下使用
SGI的runon或Robert Love的任务集。可以使用numactl命令（Andi Kleen的numa软件包的一部分）在shell提示符下完成mbind和set_mempolicy调用。


2.使用示例和语法
============================

2.1基本用法
---------------

通过cpuset虚拟文件系统可以创建，修改和使用cpuset。
要安装它，请键入：

```
＃mount -t cgroup -o cpuset cpuset /sys/fs/cgroup/cpuset
```

然后，在/sys/fs/cgroup/cpuset下，您可以找到与系统中cpusets的树相对应的树。例如，/sys/fs/cgroup/cpuset是保存整个系统的cpuset。
如果要在/sys/fs/cgroup/cpuset下创建新的cpuset：＃cd / sys / fs / cgroup / cpuset＃mkdir my_cpuset
现在，您想对此cpuset做一些事情。 ＃cd my_cpuset在此目录中，您可以找到几个文件：

```
＃ls
cgroup.clone_children cpuset.memory_pressure 
cgroup.event_control cpuset.memory_spread_page 
cgroup.procs cpuset.memory_spread_slab
cpuset.cpu_exclusive cpuset.mems 
cpuset.cpus cpuset.sched_load_balance
cpuset.mem_exclusive cpuset.sched_relax_domain_level
cpuset.mem_hardwall notify_on_release
cpuset.memory_migrate tasks
```

读取它们将为您提供有关此cpuset状态的信息：它可以使用的CPU和内存节点，正在使用的进程
它，它的属性。通过写入这些文件，您可以操作cpuset。

设置一些标志：

```
# /bin/echo 1 > cpuset.cpu_exclusive
```

添加一些cpus：

```
/bin/echo 0-7 > cpuset.cpus
```

添加一些内存：

```
/bin/echo 0-7 > cpuset.mems
```

现在，将您的shell附加到此cpuset：＃/ bin / echo $$>任务您还可以在此cpuset中使用mkdir创建cpuset。
目录。

```
＃mkdir my_sub_cs
```

要删除cpuset，只需使用rmdir：

```
＃rmdir my_sub_cs
```

如果正在使用cpuset（内部有cpuset，或已附加进程），则此操作将失败。
请注意，由于遗留原因，“ cpuset”文件系统作为cgroup文件系统的包装器存在。

命令
```
 mount -t cpuset X /sys/fs/cgroup/cpuset等效于
 mount -t cgroup -ocpuset,noprefix X /sys/fs/cgroup/cpuset
 echo "/sbin/cpuset_release_agent" > /sys/fs/cgroup/cpuset/release_agent
```

2.2添加/删除cpus
------------------------

这是在cpus或mems文件中写入时使用的语法
在cpuset目录中：

```
# /bin/echo 1-4 > cpuset.cpus		-> set cpus list to cpus 1,2,3,4
# /bin/echo 1,2,3,4 > cpuset.cpus	-> set cpus list to cpus 1,2,3,4
```

 要将CPU添加到CPUset，请写入新的CPU列表，包括CPU
要添加。要将6添加到上述cpuset中：

```
# /bin/echo 1-4,6 > cpuset.cpus	-> set cpus list to cpus 1,2,3,4,6
```

与从cpuset中删除CPU类似，编写新的CPU列表而不将CPU删除。
删除所有CPU：

```
# /bin/echo "" > cpuset.cpus		-> clear cpus list
```

2.3设置标志
-----------------

语法非常简单：

```
# /bin/echo 1 > cpuset.cpu_exclusive 	-> set flag 'cpuset.cpu_exclusive'
# /bin/echo 0 > cpuset.cpu_exclusive 	-> unset flag 'cpuset.cpu_exclusive'
```

2.4附加过程
-----------------------

```
＃# /bin/echo PID > tasks
它是PID，而不是PID。您一次只能附加一个任务。如果要附加多个任务，则必须一个接一个地执行：
# /bin/echo PID1 > tasks
# /bin/echo PID2 > tasks
	...
# /bin/echo PIDn > tasks
```

3.问题
============

问：这个'/bin/echo'是怎么回事？
答：bash的内置“ echo”命令不会检查对write（）的调用错误。如果在cpuset文件系统中使用它，则将无法判断命令是成功还是失败。

问：当我附加流程时，只有第一行才真正被附加！
答：我们每次调用write（）只能返回一个错误代码。因此，您还应该只放置一个pid。 4.联系方式