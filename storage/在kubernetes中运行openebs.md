
# 什么是OpenEBS？

现在，OpenEBS是kubernetes下与容器原生和容器附加存储类型相关通用的领先开源项目之一。
通过为每个工作负载指定专用的存储控制器，OpenEBS遵循容器附加存储或CAS的脚步。
为了向用户提供更多功能，OpenEBS具有精细的存储策略和隔离功能，
可帮助用户根据工作负载选择存储。该项目不依赖Linux内核模块，而是在用户空间中运行。
它属于Cloud Native Computing Foundation沙箱，在各种情况下都非常有用，例如在公共云中运行的群集，
在隔离环境中运行的无间隙群集以及本地群集。

# 什么是CAS？

首先，CAS是`Container Attached Storage`的缩写。通常，Kubernetes存储在集群环境之外维护。
无论共享文件系统如何，存储设施始终与外部资源相关，包括Amazon EBS，GCE PD，NFS，Gluster FS和Azure
磁盘等存储巨头。在大多数情况下，存储通常以OS内核模块的形式与节点相关。这也适用于永久卷，在永久卷中，
它们与模块紧密耦合，因此显示为旧版资源和整体式。CAS提供的是Kubernetes使用诸如微服务之类的存储实体的便利。
总体而言，CAS分为两个元素，即数据平面和控制平面。

另一方面，控制平面控制一组CRD或`Custom Resource Definitions`，并涉及低级别的存储实体。
数据平面和控制平面之间的这种清晰的分离为用户提供了与Kubernetes中的微服务相同的优势。
这种独特的架构通过使存储实体与持久性脱钩，从而有助于工作负载的可移植性。这种体系结构的另一个好处是，
它允许操作员和管理员根据工作量动态调整卷的大小。这也称为横向扩展功能。

这种结构将计算（Pod）和数据（P）置于超融合模式，在这种模式下，它们具有较高的容错能力和良好的吞吐量。

# 是什么使OpenEBS与其他存储解决方案不同？

使OpenEBS与传统存储引擎大不相同的一些品质是：

- 就像它所服务的应用程序一样，OpenEBS具有构建的微服务架构。在部署OpenEBS时，
它们作为容器安装到Kubernetes的工作程序节点。此外，该系统管理其组件并使用Kubernetes进行编排。

- 可移植性是OpenEBS作为开放源代码存储选项的优良品质，因为它是完全内置的用户空间。
这使其容易出现跨平台问题。

- Kubernetes的使用使该系统非常有意图驱动，因为它遵循有助于提高客户可用性的原则。

- 谈到可用性功能，使用OpenEBS的另一个好处是，它允许用户从各种存储引擎中进行选择。
这意味着一个人可以使用与其应用程序的设计和目标兼容的存储引擎。无论引擎的类型如何，
OpenEBS都提供了一个强大的框架，该框架具有良好的可管理性，快照，可用性和克隆。例如，
Cassandra是需要低延迟写入的分布式应用程序。因此，它可以使用本地PV引擎。同样，
建议将ZFS引擎用于需要弹性的整体式应用程序，例如PostgreSQL和MySQL。对于流应用程序
，专业人士经常建议使用称为MayaStor的NVMe引擎，该引擎可保证最佳性能。

部署OpenEBS之后，您可以获得许多存储服务，包括：

- 在连接到Kubernetes的工作节点上，使存储管理自动化。这将使您可以使用该存储来动态配置本地PV和OpenEBS PV。

- 跨节点的数据持久性得到了改善，这有助于用户节省通常在重建时浪费的时间。例如，Cassandra ring。
- 云提供商和可用性区域之间的数据将正确同步。此类功能有助于提高所需数据的可用性，并减少连接和分离时间。
- 这有点像用户的通用层，因此他们可以体验到相同级别的存储服务以及良好的开发人员和布线设施。是否使用裸机，AKS，AWS或GKE都没有关系。
- 由于OpenEBS属于Kubernetes原生解决方案，因此管理员与开发人员之间进行交互的机会更大，
这有助于管理OpenEBS。他们可以使用Helm，Prometheus，Kubectl，Grafana和Weave Scope等各种工具。
- 正确管理与S3和其他目标之间的来回分层过程。

# 开放式EBS架构

我们已经知道OpenEBS属于CAS或容器附加存储模型。用此模型维护其结构，OpenEBS系统的每个卷都有一个指定
的控制器POD和一组重复的POD。我们已经讨论了CAS系统背后的思想，因此我们可以说CAS系统的整体体系结构有助于
使OpenEBS成为用户友好的系统。人们认为OpenEBS在操作过程中非常简单，就像来自Kubernetes的其他任何云原
生项目一样。

这是OpenEBS架构图。相应地绘制了所有重要功能，我们将对其进行简要讨论。


![](http://img.rocdu.top/20200618/01-OpenEBS-Architecture.png)


OpenEBS系统包含许多组件。总体而言，它们可以分为以下几类。

- 控制平面：包括API服务器，预配器，卷side car和卷exports 。
- 数据平面：Jiva，cStor和LocalPV
- 节点磁盘管理器：监视，发现和管理连接到Kubernetes节点的媒介。
- 云原生工具集成：集成通过Grafana，Jaeger，Prometheus和Fluentd完成。

# OpenEBS的控制平面
OpenEBS集群的另一个术语是"Maya"。控制平面有多种功能。其中一些功能包括配置卷，
与卷关联的操作（如克隆制作，快照快照，存储策略实施，存储策略创建，卷指标的导出），
以便Prometheus/Grafana可以使用它们以及许多其他功能。

![控制平面](http://img.rocdu.top/20200618/02-Control-Plane-of-OpenEBS.jpg)

标准的Kubernetes存储插件是动态预配器，OpenEBS PV预配器的主要任务是根据Kubernetes用于PV实施
规范并启动卷预配。当涉及批量策略管理和批量处理任务时，m-apiserver有助于公开存储REST API。
当我们查看数据平面和控制平面之间的连接时，我们可以看到一个sidecar模式。我们有一些数据平面必
须与控制平面通信时的条件示例。

- 对于吞吐量，延迟和IOPS等卷统计信息。在此，使用了volume-exporter sidecar。
- 在卷副本容器的帮助下进行磁盘或池管理，在卷控制器容器的帮助下执行卷策略。在这里，使用了volume-management sidecar。

让我们谈谈控制平面的上述组件：

![PV Provisioner](http://img.rocdu.top/20200618/03-provisioning-flow-in-openEBS.jpg)

该组件的主要功能是在作为POD运行时做出供应决策。工作机制也非常简单。首先，开发人员提出具有必要体积参数的
声明，然后选择正确的存储类别。最后，他或她在YAML规范上调用Kubelet。在这里，maya-apiserver和
OpenEBS PV供应商相互交互，并创建节点上的卷副本容器和卷控制器容器所需的部署规范。
使用PVC规范中的注释来控制体积容器的调度。根据当前统计，OpenEBS仅支持iSCSI绑定。

![Maya-Apiserver](http://img.rocdu.top/20200618/04-m-apiserver-internals.jpg)

m-apiserver的主要任务是公开OpenEBS REST API，并且它以POD的形式运行。如果需要创建卷容器，
则m-apiserver会生成部署规范所需的文件。然后，根据情况调度pod并调用kube-apiserver。
该过程完成后，将创建对象PV，然后将其安装在应用程序容器上。然后，控制器盒与副本盒的帮助一起托管PV。
副本容器和控制器容器都是数据平面的重要部分。

m-apiserver的另一项主要任务是卷策略的管理。在提供策略时，OpenEBS通常使用精细的规范。
然后，m-apiserver使用YAML规范的这些解释将它们转换为可执行的组件。在那之后，它们通过音量管理侧边栏得
到了强制执行。

## Maya Volume Exporter

每个存储控制器pod，即cStor和Jiva，都有一个称为Maya卷导出器的sidecar。这些sidecar的功能是通过将数
据平面与控制平面相连来帮助检索信息。如果我们查看统计信息的粒度，则它始终处于卷级别。统计信息的一些示例是：

- Volume write latency
- Volume read latency
- Write IOPS
- Read IOPS
- Write block size
- Read block size
- Capacity status

![Volume exporter 数据流](http://img.rocdu.top/20200618/05-volume-exporter-data-flow.jpg)

# volume manauagement sidecar

sidecar的主要功能有两个：一是将卷策略和控制器配置参数传递到卷控制器容器或数据平面。
另一个是传递副本配置参数以及卷副本容器的数据保护参数的副本。

![卷管理sidecar](http://img.rocdu.top/20200618/06-volume-management-side-car.jpg)

# OpenEBS的数据平面

OpenEBS架构最喜欢的一件事是它向用户提供的与存储引擎相关的功能。它为用户提供了根据工作负载的配置
和特征来配置其存储引擎的选择。例如，如果您具有基于IOPS的高数据库，则可以从读取繁重的共享CM
S工作负载中选择其他存储引擎。因此，数据平面为用户提供了三种存储引擎选择：Jiva，cStor和Local PV。

cStor是OpenEBS提供的最受欢迎的存储引擎选项，其中包括丰富的存储引擎和轻量级的功能。这些功能
对于类似HA工作负载的数据库特别有用。您通过此选项获得的功能是企业级的。其中一些是按需容量和性能提升，
高数据弹性，数据一致性，同步数据复制，克隆，快照和精简数据提供。cStor同步复制的单个副本可提供高可用性的
有状态Kubernetes部署。当从应用程序请求数据的高可用性时，cStor会生成3个副本，其中数据以同步顺序写入。
此类复制有助于保护数据丢失。

Jiva是OpenEBS最早的存储引擎，使用非常简单。之所以如此方便，部分原因在于该引擎完全根据用户空间的标准运行，
并具有标准的块存储容量，例如同步复制。如果您的小型应用程序没有添加块存储设备的选项，那么Jiva可能是您的
正确选择。考虑到这一点，事实也恰恰相反，这意味着该引擎对于需要高级存储和高性能功能的工作负载效率不高。

继续使用OpenEBS的最简单的存储引擎是Local PV或Local Persistent Volume。它只是一个直接连接到单个
Kubernetes节点的磁盘。对熟悉的API的这种使用意味着Kubernetes可以在此过程中提取高性能的本地存储。
概括整个概念，OPenEBS的Local PV将帮助用户在节点上创建持久的本地磁盘或路径卷。这对于不需要高级存储
功能（例如克隆，复制和快照）的应用程序（例如云原生应用程序）非常有用。例如，对于基于OpenEBS的本地PV的配置，
可以使用同时处理HA和复制的StatefulSet。

| 快照和克隆支持	| 基本的 | 高级 | 没有 |
| ------------- | ----- |----- | --- |
| 资料一致性 | 	是	| 是 | 	不适用
| 使用Velero备份和还原 | 	是 | 	是 | 	是
| 适合大容量工作负载	 |  	是 | 	是
| 精简配置	  | 	是 | 	没有
| 磁盘池或聚合支持	 |  	是 | 	没有
| 按需扩容	  | 	是 | 是
| 数据弹性（RAID支持） | 	 	是 | 	是*
| 近磁盘性能 | 	没有 | 	没有 | 	是

我们有三个当前可用的存储引擎，但这并不意味着它们是唯一的选择。实际上，OpenEBS社区目前正在开发新引擎。
它们仍然是原型，需要在进入市场之前进行适当的测试。例如，MayaStor是一种数据引擎，可能很快就会投放市场。
它是用Rust编写的，具有低延迟引擎，对于需要API访问以访问块存储和接近磁盘性能的应用程序非常有帮助。此外，
与本地PV相关的问题已经过测试，以ZFS本地PV为名称的变体因克服了这些缺点而获得了一些认可。


# 节点设备管理器

在Kubernetes中工作时，在有状态应用程序的情况下管理持久性存储的任务由各种工具完成。NDM或节点设备管理器
就是一种可以填补这一空白的工具。DevOps架构师必须以保持一致性和弹性的方式提供应用程序开发人员和应用程序本身的基础设施需求。为了做到这一点，存储堆栈的灵活性必须很高，以便云原生生态系统可以轻松使用堆栈。在这种情况下，NDM的功能非常方便，它可以将单独的磁盘组合在一起，并赋予它们将它们分段存储的能力。NDM通过将磁盘标识为Kubernetes对象来实现此目的。它还有助于管理Kubernetes PV供应商（如OpenEBS，Prometheus和其他系统）的磁盘子系统。

![NDM](http://img.rocdu.top/20200618/07-node-device-manager-ndm.jpg)


# 与云原生工具的集成

`Grafana和Prometheus`： Prometheus的安装是在OpenEBS运营商作为微服务的一部分进行的初始设置期间进行的。音量策略负责根据给定的音量控制Prometheus监视。总体而言，Prometheus和Grafana工具共同帮助OpenEBS社区监视持久性数据。

`WeaveScope`：如果需要查看与容器，进程，主机或服务相关的标签，元数据和度量，则使用WeaveScope。因此，在Kubernetes中将它作为云原生可视化解决方案的重要组成部分。对于WeaveScope集成，将启用诸如卷Pod，节点磁盘管理器组件以及与Kubernetes相关的其他类型的存储结构之类的东西。所有这些增强功能都有助于遍历和探索这些组件。

# 数据如何受到保护？

Kubernetes有许多方法可以保护数据。例如，如果IO容器与iSCSI目标一起发生故障，则它会被Kubernetes旋转回去。将相同的原理应用于存储数据的副本容器。OpenEBS可以借助可配置的仲裁或副本的最低要求来保护多个副本。cStor具有其他功能，可以检查静默数据的损坏，并可以在将其隐藏在后台的同时对其进行修复。

# 如何安装和入门

首先要做的是确认iSCSI客户端设置。通过使用必要的iSCSI协议，OpenEBS为用户提供了块卷支持。因此，必须在安装期间所有Kubernetes节点都具有iSCSI启动器。根据您的操作系统，有多种方法可以验证iSCSI客户端安装。如果尚未安装，我们以Ubuntu用户的整个过程为例：

正如我们已经讨论的那样，为使OpenEBS系统正常运行，需要确保iSCSI服务在所有辅助节点上运行。请按照以下步骤在Linux平台（Ubuntu）中启动该过程。

## 配置：

如果您的系统中已经安装了iSCSI启动器，请使用以下给定命令检查启动器名称的配置和iSCSI服务的状态：

```
sudo cat /etc/iscsi/initiatorname.iscsi
systemctl status iscsid
```

成功运行命令后，系统将显示服务是否正在运行。如果状态显示为“非活动”，则键入以下命令以重新启动iscsid服务：

```
sudo systemctl enable iscsid 
sudo systemctl start iscsid 
```

如果提供正确的命令，那么系统将为您提供以下输出：

```
systemctl status iscsid
● iscsid.service - iSCSI initiator daemon (iscsid)
   Loaded: loaded (/lib/systemd/system/iscsid.service; disabled; vendor preset: enabled)
   Active: active (running) since Mon 2019-02-18 11:00:07 UTC; 1min 51s ago
     Docs: man:iscsid(8)
  Process: 11185 ExecStart=/sbin/iscsid (code=exited, status=0/SUCCESS)
  Process: 11170 ExecStartPre=/lib/open-iscsi/startup-checks.sh (code=exited, status=0/SUCCESS)
 Main PID: 11187 (iscsid)
    Tasks: 2 (limit: 4915)
   CGroup: /system.slice/iscsid.service
           ├─11186 /sbin/iscsid
           └─11187 /sbin/iscsid
```

如果您未在节点上安装iSCSI启动器，请在以下命令的帮助下转到“ open-iscsi”软件包：

```
sudo apt-get update
sudo apt-get install open-iscsi
sudo systemctl enable iscsid 
sudo systemctl start iscsid
```

如果在他们的系统上预先安装了Kubernetes环境，则他或她可以借助以下命令轻松地部署

OpenEBS：

```
kubectl apply -f https://openebs.github.io/charts/openebs-operator.yaml
```

之后，您可以开始针对OpenEBS运行工作负载。实际上，有许多工作负载使用OpenEBS的存储类。不，您不必非常特定于存储类。这是简单的方法，但是，花一些时间选择特定的存储类将帮助您节省时间，从长远来看还有助于自定义工作负载。默认的OpenEBS回收策略与K8所使用的相同。“删除”是动态配置的PersistentVolume的默认回收策略。它们在某种意义上是相关的，如果一个人删除了相应的PersistentVolumeClaim，则动态配置的卷将被自动删除。对于cStor卷，则数据随之删除。对于jiva（0.8.0版及更高版本），清理作业将执行数据删除工作。

```
kubectl delete job <job_name> -n <namespace>
```

在配置Jiva和cStor卷之前，您应该做的第一件事是验证iSCSI客户端。话虽这么说，始终建议用户完成iSCSI客户端的设置，并确保iscsid服务运行良好并在每个工作节点上运行。这是正确正确地安装OpenEBS安装程序所必需的。

另外，请记住，如果要安装OpenEBS，则必须具有集群管理员用户上下文。如果您没有集群管理员用户上下文，则创建一个上下文并在该过程中使用它。对于创建，可以使用以下命令。

```
kubectl config set-context NAME [--cluster=cluster_nickname] [--user=user_nickname] [--namespace=namespace]
```


这是上述命令的示例：

```
kubectl config set-context admin-ctx --cluster=gke_strong-eon-153112_us-central1-a_rocket-test2 --user=cluster-admin
```

之后，键入以下命令来设置新创建的上下文或现有的上下文。请参阅以下示例

```
kubectl config use-context admin-ctx
```


## 通过helm安装过程

在启动该过程之前，请检查您的系统中是否安装了helm，并且helm存储库需要任何更新。

对于Helm的v2版本：

首先，运行命令`helm init`，将分till pod安装在“ kube-system”命名空间下，然后按照下面给出的说明为分till设置RBAC。要获取helm已安装版本，用户可以键入以下命令：

```
helm version
```

这是输出示例：

```
Client: &amp;version.Version{SemVer:"v2.16.1", GitCommit:"bbdfe5e7803a12bbdf97e94cd847859890cf4050", GitTreeState:"clean"}
```

如果使用默认模式进行安装，请使用下面给出的命令在“ openebs”命名空间中安装OpenEBS：

```
helm install --namespace openebs --name openebs stable/openebs --version 1.10.0
```

对于Helm v3版本：

您可以在以下命令的帮助下获取helm v3版本的预安装版本：

```
helm version
```

这是输出示例：

```
version.BuildInfo{Version:"v3.0.2", GitCommit:"19e47ee3283ae98139d98460de796c1be1e3975f", GitTreeState:"clean", GoVersion:"go1.13.5"}
```

在helm v3的帮助下，可以通过两种方式安装OpenEBS。让我们一一讨论。

`第一种选择`：在这种方法中，helm从本地kube配置获取当前的名称空间，并在用户决定运行helm命令时稍后使用它。如果不存在，则掌舵将使用默认名称空间。首先，借助以下命令将openebs与openebs命名空间一起安装：

您可以使用以下代码查看当前上下文：

```
kubectl config current-context
```

为当前上下文分配名称openebs并键入以下内容：

```
kubectl config set-context <current_context_name> --namespace=openebs
```

要创建OpenEBS名称空间：

```
kubectl create ns openebs
```

然后以openebs作为图表名称安装OpenEBS。使用以下命令：

```
helm install openebs stable/openebs --version 1.10.0
```

最后，写下以下代码以查看chart：

```
helm ls
```

通过执行上述步骤，您将安装带有openebs名称空间的OpenEBS，该名称空间的图表名称为openebs。

`第二个选项`：第二个选项是关于直接在helm命令中提及名称空间。定期执行以下步骤进行安装。

为OpenEBS创建名称空间 

```
kubectl create ns openebs
```

使用图表名称openebs，安装openebs系统。命令如下：

```
helm install --namespace openebs openebs stable/openebs --version 1.10.0
```

helm install –namespace openebs openebs stable/openebs –version 1.10.0

要查看chart，请使用以下代码：

```
helm ls -n openebs
```

之后，您将获得带有chart名称和名称空间openebs的OpenEBS安装版本。

### 您需要注意一些事项：

- 从Kubernetes的1.12版本开始，容器必须设置其极限值和资源请求，否则容器将被逐出。在安装之前，我们建议读者首先在YAML运算符中将值设置为OpenEBS pod spec。

- 在安装OpenEBS操作员之前，请检查节点上块设备的安装状态。

如果继续使用自定义安装模式，则会遇到以下高级配置：

- 您可以为OpenEBS控制平面pod选择节点。
- 节点选择也可用于OpenEBS存储池。
- 如果不需要磁盘过滤器，则可以简单地排除它们。
- 在OpenEBS运营商YAML中，有一个配置环境变量是可选的。
- 如果您想采用自定义安装方式，则需要下载openebs-operator-1.10.0，更新配置，然后使用“ kubectl”命令。

#### 设置控制平面的节点选择器

如果您有一个很大的Kubernetes集群，则可以故意将OpenEBS控制平面的调度过程限制为仅几个特定节点。对于此过程，应该指定键值对的映射，然后找到所需的群集节点以标签的形式附加相同的键值对。

#### 为准入控制设置节点选择器

准入控制器的作用是在对象持久化之前截取已提交给Kubernetes的API服务器的请求。仅在授权或验证请求后才能执行此操作。为了验证传入请求，openebs准入控制器制定其他自定义准入策略。例如，这是最新版本的两种准入策略。

- 如果存在PersistentVolumeClaim的克隆，则通过删除PersistentVolumeClaim来完成验证。
- 为了验证请求的声明容量，该大小必须变为快照大小，并由Clone PersistentVolumeClaim创建。

我们可以使用节点选择器方法来调度特定节点上的准入控制器容器。

#### 对于节点磁盘管理器节点选择器设置

对于OpenEBS cStorPool的构建，可以使用块设备自定义资源，也可以使用节点磁盘管理器创建块设备。如果您想过滤掉Kubernetes中的集群并且仅将某些节点用于OpenEBS存储，则指定键值对并将相同的键值以标签的形式附加到必要的节点即可。

### 节点磁盘管理器磁盘筛选器设置

NDM的默认功能是分离出下面给出的磁盘模式，然后将在特定节点上发现的其余磁盘模式转换为DISK CR。问题是，不应安装它们。

如果群集中还有其他类型的磁盘尚未过滤掉，您要做的就是将其他磁盘模式包括到排除列表中。该列表位于YAML文件中。

```
"exclude":"loop,/dev/fd0,/dev/sr0,/dev/ram,/dev/dm-"
```

### 配置环境变量

在环境变量主题下，提供了与默认cStor SparsePool，本地PV基本路径，默认存储配置和cStor Target相关的配置。

#### 启用核心转储：

对于NDM守护程序集和cStor池容器，转储核心被禁用为默认设置的一部分。要启用此功能，您需要将ENV变量“ ENABLE_COREDUMP”设置为1。然后您要做的就是在cStor池中部署ENV设置以在cStor池pod中启用转储核心，并将ENV设置放入ndm守护程序规范中daemonset pod核心转储。

```
- name: ENABLE_COREDUMP
value: "1"
```

#### Sparse目录：

SparseDir只是用于查找Sparse文件的hostPath目录。默认情况下，该值设置为“ / var / openebs / sparse”。在应用OpenEBS运算符YAML文件之前，应将某些配置添加为maya-apiserver规范中环境变量的一部分。

```
# environment variable
 - name: SparseDir
   value: "/var/lib/"
```

#### cStorSparsePool默认

根据配置值，OpenEBS安装过程将创建一个默认的cStor SparsePool。该配置完全取决于true和false的值。如果为true，则将配置cStor SparsePool，否则将不会进行配置。配置的默认值始终为false，此SparsePool仅用于测试目的。如果要使用Sparse磁盘安装cStor，则应在Maya-apiserver规范中以环境变量的形式添加此特定配置。这是一个例子：

```
# environment variable
- name: OPENEBS_IO_INSTALL_DEFAULT_CSTOR_SPARSE_POOL
  value: "false"
```

#### TargetDir

目标目录用作目标容器的hostPath，其默认值设置为“ / var / openebs”。该预值将覆盖主机路径，并在maya-apiserver部署中引入OPENEBS_IO_CSTOR_TARGET_DIR ENV。当主机操作系统无法在默认的OpenEBS路径（即（/ var / openebs /））上写入时，通常需要这种类型的配置。与cStor SparsePool一样，应在应用操作员YAML文件之前将某些配置作为环境变量添加到maya-apiserver规范中。这是一个例子：

```
# environment variable
- name: OPENEBS_IO_CSTOR_TARGET_DIR
  value: "/var/lib/overlay/openebs"
```


#### OpenEBS本地PV基本路径

对于基于主机路径的localPV，默认的hospath为/ var / openebs / local。以后可以在OpenEBS operator的安装过程中对此进行更改。您要做的就是传递OPENEBS_IO_BASE_PATH ENV参数。

```
# environment variable
 - name: OPENEBS_IO_BASE_PATH
   value: "/mnt/"
```

### 默认存储配置：

Jiva和本地PV存储类是OpenEBS随附的一些默认存储配置。可以根据需要配置和定制OpenEBS中的存储引擎，并通过关联的自定义资源和存储类来完成。在安装过程之后，您始终可以更改存储的默认配置，但是它会被API服务器覆盖。因此，我们通常建议用户在默认选项的帮助下创建自己的存储配置。如果在安装过程中禁用默认配置，则可以进行自己的存储配置类型。若要正确禁用默认配置，请在Maya-apiserver中将以下代码添加为环境变量。

```
# environment variable
- name: OPENEBS_IO_CREATE_DEFAULT_STORAGE_CONFIG
  value: "false"
```

# 验证安装过程

要使用<openebs>命名空间获取Pod列表，请使用以下代码：

```
kubectl get pods -n openebs
```

如果成功安装了OpenEBS，则很可能会看到以下示例所示的输出：

![](http://img.rocdu.top/20200618/kubectl-get-pods-n-openebs.png)

openebs-ndm引用守护程序集，该守护程序集应在集群的所有节点上运行，或者至少在nodeSelector配置期间选择的节点上运行。同样，maya-apiserver，openebs-snapshot-operator和openebs-provisioner控制平面容器也应该正在运行。如果已经配置了nodeSelector，请确保将它们安排在正确的节点上。为此，请使用“ Kubectly get pods -n openebs -o wide”列出容器。

#### 验证存储类：

首先，通过列出以下内容检查OpenEBS是否已安装默认存储类：

```
kubectl get sc
```

供您参考，以下是成功安装后将看到的输出示例。您将找到创建的给定StorageClasses：

![](http://img.rocdu.top/20200618/kubectl-get-sc.png)

#### 验证块设备CR

对于NDM守护程序集创建的每个块设备CR，发现的节点具有以下两个例外：

- 与排除供应商过滤器和“路径过滤器”匹配的磁盘。
- 节点上已经挂载的磁盘。

要检查CR是否如预期的那样来临，请使用以下命令列出块设备CR。


```
kubectl get blockdevice -n openebs
```

如果以正确的方式进行操作，屏幕上将显示类似的输出：

![](http://img.rocdu.top/20200618/blockdevice-n-openebs.png)


之后，使用以下命令检查节点上的标签集，以找到节点的相应块设备CR。

```
kubectl describe blockdevice <blockdevice-cr> -n openebs
```

#### 验证Jiva默认池

```
kubectl get sp
```

达到上述要求后，您可能会看到以下输出：

![](http://img.rocdu.top/20200618/kubectl-get-sp.png)

# 安装后要考虑的事项
安装后，可以使用以下存储类来简单地测试OpenEBS：

- 要配置Jiva卷，请使用openebs-jiva-default。在这里，使用默认池，并在mnt / openebs_disk目录下创建数据的副本。该目录位于Jiva副本窗格中。
- 要在主机路径上配置本地PV，请使用openebs-host路径。
- 要在设备上配置本地PV，请使用openebs-device。

要使用实际磁盘，必须首先根据要求创建Jiva池，cStorPools或OpenEBS Local PV。之后，创建所需的StorageClasses或使用默认的StorageClasses进行使用。


原文地址： https://goglides.io/running-openebs-in-kubernetes/371/

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
