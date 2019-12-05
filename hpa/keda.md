# KEDA - 基于Kubernetes事件驱动的自动缩放

事件驱动的计算并不是什么新生事务。数据库世界中的人们使用数据库触发器已有多年了。
这个概念很简单: 每当您添加,更改或删除数据时,都会触发一个事件以执行各种功能。
新的事件是这些类型的事件和触发器在其他领域的应用程序中激增，
例如自动扩展，自动修复，容量规划等。
事件驱动架构的核心是对系统上的各种事件做出反应并采取相应的行动。

自动缩放(以一种或其他方式实现自动化)
已成为几乎所有云平台中不可或缺的组成部分，
微服务又或者容器并不是一种例外。
容器以灵活和解耦设计而闻名最适合自动缩放，因为它们比虚拟机更容易创建。

> 为什么要自动缩放???

![容量扩展—自动扩展](http://q08i5y6c2.bkt.clouddn.com/keda/keda-1.png)

对于基于容器的现代应用程序部署,可伸缩性是要考虑的最重要方面之一。
随着容器编排平台的发展,设计可伸缩性设计解决方案从未如此简单。
KEDA基于Kubernetes的事件驱动自动缩放或KEDA(使用Operator Framework构建)
允许用户在Kubernetes上构建自己以事件驱动的应用程序。
KEDA处理触发器以响应其他服务中发生的事件，并根据需要扩展工作负载。
KEDA使容器可以直接从源使用事件，而不是通过HTTP进行路由。

KEDA可以在任何公共或私有云和本地环境中工作，
包括Azure Kubernetes服务和Red Hat的OpenShift。
借助此功能，开发人员现在还可以采用Microsoft的无服务器平台Azure Functions，
并将其作为容器部署在Kubernetes群集中，包括在OpenShift上。

这可能看起来很简单，但假设每天繁忙处理大量事务，
如下所示真的可以手动管理应用程序的数量(Kubernetes部署)吗？

![在生产中管理自动缩放](http://q08i5y6c2.bkt.clouddn.com/keda/keda-2.png)

KEDA将利用实时度量标准自动检测新部署并开始监视事件源，以推动扩展决策。

# KEDA

KEDA作为Kubernetes上的组件提供了两个关键角色：

- 扩展客户端：用于激活和停用部署以扩展到配置的副本，并在没有事件的情况下将副本缩减回零。

- Kubernetes Metrics Server：一种度量服务器，它公开大量与事件相关的数据，
例如队列长度或流滞后，从而允许基于事件的扩展使用特定类型的事件数据。

Kubernetes Metrics Server与Kubernetes HPA(horizontal pod autoscaler)进行通信，
以从Kubernetes部署副本中扩展规模。然后由部署决定是否直接从源中使用事件。
这样可以保留丰富的事件集成，并使诸如完成或放弃队列消息之类的手势可以立即使用。

![在生产中管理自动缩放](http://q08i5y6c2.bkt.clouddn.com/keda/keda-3.png)


# Scaler

KEDA使用`Scaler`来检测是否应激活或取消激活(缩放)部署，然后将其馈送到特定事件源中。如今，
支持多个`Scaler`,通过特定受支持的触发器,例如(Kafka(trigger: Kafka topic))，
RabbitMQ(trigger: RabbitMQ队列))，并且还会支持更多。

除了这些KEDA，还与Azure Functions工具集成在一起，以本机扩展Azure特定的缩放器，
例如Azure Storage Queues, Azure Service Bus Queues,Azure Service Bus Topics。


# ScaledObject

ScaledObject部署为Kubernetes CRD（自定义资源定义），它具有将部署与事件源同步的功能。


![ScaledObject自定义资源定义](http://q08i5y6c2.bkt.clouddn.com/keda/keda-4.png)

一旦部署为CRD，ScaledObject即可进行以下配置：

![缩放对象规格](http://q08i5y6c2.bkt.clouddn.com/keda/keda-5.png)

如上所述，支持不同的触发器，下面显示了一些示例：

![ScaledObject的触发配置](http://q08i5y6c2.bkt.clouddn.com/keda/keda-6.png)

# 事件驱动的自动伸缩在实践中-本地Kubernetes集群

## KEDA部署在Kubernetes中

![KEDA控制器-Kubernetes部署](http://q08i5y6c2.bkt.clouddn.com/keda/keda-7.png)

## 带有KEDA的RabbitMQ队列缩放器

RabbitMQ是一种称为消息代理或队列管理器的消息队列软件.
简单地说: 这是一个可以定义队列的软件，
应用程序可以连接到队列并将消息传输到该队列上。

![RabbitMQ架构](http://q08i5y6c2.bkt.clouddn.com/keda/keda-8.png)

在下面的示例中，在Kubernetes上将RabbitMQ服务器/发布器部署为“状态集”：

![rabbitmq](http://q08i5y6c2.bkt.clouddn.com/keda/keda-9.png)

RabbitMQ使用者被部署为接受RabbitMQ服务器生成的队列并模拟执行的部署。

![RabbitMQ消费者部署](http://q08i5y6c2.bkt.clouddn.com/keda/keda-10.png)

## 使用RabbitMQ触发器创建ScaledObject
除了上面的部署外，还提供了ScaledObject配置，该配置将由上面创建的KEDA CRD转换，并在Kubernetes上安装KEDA。

![使用RabbitMQ触发器进行ScaledObject配置](http://q08i5y6c2.bkt.clouddn.com/keda/keda-11.png)


![ScaledObject在Kubernetes中](http://q08i5y6c2.bkt.clouddn.com/keda/keda-12.png)

创建ScaledObject后，KEDA控制器将自动同步配置并开始监视上面创建的Rabbitmq-consumer。
KEDA无缝创建具有所需配置的HPA（水平Pod自动缩放器）对象，
并根据通过ScaledObject提供的触发规则(在此示例中，队列长度为`5`)扩展副本。
由于尚无队列，如下所示，rabbitmq-consumer部署副本被设置为零。

![KEDA Controller在Kubernetes中](http://q08i5y6c2.bkt.clouddn.com/keda/keda-13.png)

![KEDA创建的卧式自动定标器](http://q08i5y6c2.bkt.clouddn.com/keda/keda-14.png)

![RabbitMQ使用者—副本:0](http://q08i5y6c2.bkt.clouddn.com/keda/keda-15.png)

通过ScaledObject和HPA配置，KEDA将驱动容器根据从事件源接收的信息进行横向扩展。
使用下面的`Kubernetes-Job`配置发布一些队列，这将产生10个队列：

![Kubernetes-Job将发布队列](http://q08i5y6c2.bkt.clouddn.com/keda/keda-16.png)

KEDA会自动将当前设置为零副本的`rabbitmq-consumer`缩放为`两个`副本，以适应队列。

## 发布10个队列-RabbitMQ Consumer扩展为两个副本：

![10个队列— 2个副本](http://q08i5y6c2.bkt.clouddn.com/keda/keda-17.png)


![缩小为：2 —缩小为：0](http://q08i5y6c2.bkt.clouddn.com/keda/keda-18.png)


## 发布200个队列-RabbitMQ使用者扩展到四十（40）个副本：

![200个队列— 40个副本](http://q08i5y6c2.bkt.clouddn.com/keda/keda-19.png)

![缩小为：40 —缩小为：0](http://q08i5y6c2.bkt.clouddn.com/keda/keda-20.png)

## 发布1000个队列-RabbitMQ Consumer扩展到100个副本，因为最大副本数设置为100：

![1000个队列— 100个副本](http://q08i5y6c2.bkt.clouddn.com/keda/keda-21.png)

![缩小为：100 —缩小为：0](http://q08i5y6c2.bkt.clouddn.com/keda/keda-22.png)


KEDA提供了一个类似于FaaS的事件感知扩展模型，在这种模型中，Kubernetes部署可以基于需求和基于智能，
动态地从零扩展到零，而不会丢失数据和上下文。KEDA还为Azure Functions提供了一个新的托管选项，
可以将其部署为Kubernetes群集中的容器，
从而将Azure Functions编程模型和规模控制器带入云或本地的任何Kubernetes实现中。

KEDA还为Kubernetes带来了更多的事件来源。
随着将来继续添加更多的触发器或为应用程序开发人员根据应用程序的性质设计触发器提供框架，
使KEDA有潜力成为生产级Kubernetes部署中的必备组件，从而使应用程序自动缩放成为应用程序开发中的嵌入式组件。

原文链接： https://itnext.io/keda-kubernetes-based-event-driven-autoscaling-48491c79ec74
扫描关注我:

![微信](http://q08i5y6c2.bkt.clouddn.com/qrcode_for_gh_7457c3b1bfab_258.jpg)