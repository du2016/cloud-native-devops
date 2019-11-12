# 介绍

KubeEdge是一个开源系统，用于将本机容器化的应用程序编排功能扩展到Edge上的主机，
它基于kubernetes构建，并为网络，应用程序提供基本的基础架构支持。云和边缘之间的部署和元数据同步。
Kubeedge已获得Apache 2.0的许可。并且完全免费供个人或商业使用。我们欢迎贡献者！

我们的目标是建立一个开放平台，以支持Edge计算，将原生容器化应用程序编排功能扩展到Edge上的主机，该主机基于kubernetes，并为网络，
应用程序部署以及云与Edge之间的元数据同步提供基础架构支持。

# 特点

- 完全开放 - Edge Core和Cloud Core都是开源的。
- 离线模式 - 即使与云断开连接，Edge也可以运行。
- 基于Kubernetes - 节点，群集，应用程序和设备管理。
- 可扩展 - 容器化，微服务
- 资源优化 - 可以在资源不足的情况下运行。边缘云上资源的优化利用。
- 跨平台 - 无感知；可以在私有，公共和混合云中工作。
- 数据与分析 - 支持数据管理，数据分析管道引擎。
- 异构 - 可以支持x86，ARM。
- 简化开发 - 基于SDK的设备加成，应用程序部署等开发
- 易于维护 - 升级，回滚，监视，警报等

# 优势

- 边缘计算 - 通过在Edge上运行的业务逻辑，可以在生成数据的本地保护和处理大量数据。这减少了网络带宽需求以及边缘和云之间的消耗。这样可以提高响应速度，降低成本并保护客户的数据隐私。

- 简化开发 - 开发人员可以编写基于常规http或mqtt的应用程序，对其进行容器化，然后在Edge或Cloud中的任何位置运行它们中的更合适的一个。

- Kubernetes原生支持 - 借助KubeEdge，用户可以在Edge节点上编排应用，管理设备并监视应用和设备状态，就像云中的传统Kubernetes集群一样

- 大量的应用 - 可以轻松地将现有的复杂机器学习，图像识别，事件处理和其他高级应用程序部署和部署到Edge。

# 架构

![微信](http://q08i5y6c2.bkt.clouddn.com/kubeedge-highlevel-arch.png)

kubeedge分为两个可执行程序，cloudcore和edgecore,分别有以下模块

cloudcore：
- CloudHub：云中的通信接口模块。
- EdgeController：管理Edge节点。
- devicecontroller 负责设备管理。

edgecore：
- Edged：在边缘管理容器化的应用程序。
- EdgeHub：Edge上的通信接口模块。
- EventBus：使用MQTT处理内部边缘通信。
- DeviceTwin：它是用于处理设备元数据的设备的软件镜像。
- MetaManager：它管理边缘节点上的元数据。

## edged

EdgeD是管理节点生命周期的边缘节点模块。它可以帮助用户在边缘节点上部署容器化的工作负载或应用程序。
这些工作负载可以执行任何操作，从简单的遥测数据操作到分析或ML推理等。使用kubectl云端的命令行界面，用户可以发出命令来启动工作负载。

Docker容器运行时当前受容器和镜像管理支持。将来应添加其他运行时支持，例如containerd等。

有许多模块协同工作以实现edged的功能。

![edged-overall](http://q08i5y6c2.bkt.clouddn.com/edged-overall.png)

-  pod管理

    用于pod的添加删除修改,它还使用pod status manager和pleg跟踪pod的运行状况。其主要工作如下：
    
    - 从metamanager接收和处理pod添加/删除/修改消息。
    - 处理单独的工作队列以添加和删除容器。
    - 处理工作程序例程以检查工作程序队列以执行pod操作。
    - 分别为config map 和 secrets保留单独的的缓存。
    - 定期清理孤立的pod
    
- Pod生命周期事件生成器
- CRI边缘化
- secret管理
- Probe Management
- ConfigMap Management
- Container GC
- Image GC
- Status Manager
- 卷管理
- MetaClient

## eventbus

Eventbus充当用于发送/接收有关mqtt主题的消息的接口

它支持三种模式：

- internalMqttMode
- externalMqttMode
- bothMqttMode

## metamanager

MetaManager是edged和edgehub之间的消息处理器。它还负责将元数据存储到轻量级数据库（SQLite）或从中检索元数据。

Metamanager根据以下列出的操作接收不同类型的消息：

- Insert
- Update
- Delete
- Query
- Response
- NodeConnection
- MetaSync

## Edgehub


Edge Hub负责与云中存在的CloudHub组件进行交互。它可以使用Web套接字连接或QUIC协议连接到CloudHub 。它支持同步云端资源更新，报告边缘端主机和设备状态更改等功能。

它充当边缘与云之间的通信链接。它将从云接收的消息转发到边缘的相应模块，反之亦然。

edgehub执行的主要功能是：

- Keep Alive
- Publish Client Info
- Route to Cloud
- Route to Edge

## DeviceTwin

DeviceTwin模块负责存储设备状态，处理设备属性，处理设备孪生操作，在边缘设备和边缘节点之间创建成员资格，
将设备状态同步到云以及在边缘和云之间同步设备孪生信息。它还为应用程序提供查询接口。
DeviceTwin由四个子模块（即membership，communication，device和device twin）组成，以执行device twin模块的职责。


## Edge Controller

EdgeController是Kubernetes Api服务器和Edgecore之间的桥梁

## CloudHub

CloudHub是cloudcore的一个模块，是Controller和Edge端之间的中介。它同时支持基于Web套接字的连接以及QUIC协议访问。Edgehub可以选择一种协议来访问cloudhub。CloudHub的功能是启用边缘与控制器之间的通信。

## Device Controller
 
通过k8s CRD来描述设备metadata/status ，devicecontroller在云和边缘之间同步，有两个goroutines: `upstream controller`/`downstream controller`