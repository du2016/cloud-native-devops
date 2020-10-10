# 简介

随着istio的发展，istio适配的部署方式也愈加多样化，随着istio1.7的发布，istio目前支持以下部署方式

- 单群集部署
- 单集群使用外部控制平面
- 多集群共享外部istiod
- 多个集群分别使用多个外部istiod


# 部署角色总览

根据与各种服务网格用户和供应商合作的经验，我们认为典型服务网格有3个关键角色：

- Mesh Operator，负责管理服务网格控制平面的安装和升级。

- Mesh Admin，通常称为`平台所有者`，他拥有服务网格平台并为服务所有者定义采用服务网格的总体策略和实现。

- 网格用户，通常称为服务所有者，在网格中拥有一个或多个服务。

在1.7版之前，Istio要求控制平面在一个运行在网格内部的主集群中，导致Mesh Operator和网格管理员之间缺乏分隔。Istio 1.7引入了新的外部控制平面部署模型，使Mesh Operator可以在单独的外部群集上安装和管理网格控制平面。这种部署模型可以使Mesh Operator和网格管理员之间清楚地分开。Istio Mesh Operator现在可以为网格管理员运行Istio控制平面，而网格管理员仍可以控制控制平面的配置，而不必担心安装或管理控制平面。该模型对网格用户透明。

# 外部控制平面部署模型

使用默认安装配置文件安装Istio之后，您将在单个群集中安装Istiod控制平面，如下图所示：

![](http://img.rocdu.top/20200901/single-cluster.png)

使用Istio 1.7中的新部署模型，可以在独立于网格服务的外部群集上运行Istiod，如下图所示。外部控制平面集群由Mesh Operator拥有，而网格管理员拥有运行在网格中部署的服务的集群。Mesh管理员无权访问外部控制平面集群。Mesh Operator可以按照外部istiod单个群集的逐步指南进行探索。(注：在Istio维护人员之间的一些内部讨论中，此模型以前称为`中央istiod`。)

![Single cluster Istio mesh with Istiod in an external control plane cluster](http://img.rocdu.top/20200901/single-cluster-external-Istiod.png)

网格管理员可以将服务网格扩展到多个群集，这些群集由外部群集中运行的同一Istiod管理。没有一个网格集群是主要集群,在这种情况下。他们都是远程集群。但是，除了运行服务之外，其中之一还可以用作Istio配置群集。外部控制平面从读取Istio配置，然后Istiod将配置推送到在配置群集和其他远程群集中运行的数据平面，如下图所示。

![](http://img.rocdu.top/20200901/multiple-clusters-external-Istiod.png)


Mesh operators可以进一步扩展此部署模型，以从运行多个Istiod控制平面的外部集群管理多个Istio控制平面

![外部控制平面集群中具有多个Istiod控制平面的多个单个集群](http://img.rocdu.top/20200901/multiple-external-Istiods.png)

在这种情况下，每个Istiod都管理自己的远程集群。Mesh operators甚至可以在外部控制平面群集中安装自己的Istio网格，并将其istio-ingress网关配置为将流量从远程群集路由到其相应的Istiod控制平面。要了解有关此内容的更多信息，请查看[以下步骤](https://github.com/istio/istio/wiki/External-Istiod-single-cluster-steps#deploy-istio-mesh-on-external-control-plane-cluster-to-manage-traffic-to-istiod-deployments)。