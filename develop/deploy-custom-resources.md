# 自定义资源

本页阐释了自定义资源的概念，它是对Kubernetes API的扩展。

# 介绍

一种资源就是Kubernetes API中的一个端点，它存储着某种API 对象的集合。 
例如，内建的pods资源包含Pod对象的集合。

自定义资源是对Kubernetes API的一种扩展，它对于每一个Kubernetes集群不一定可用。
换句话说，它代表一个特定Kubernetes的定制化安装。

在一个运行中的集群内，自定义资源可以通过动态注册出现和消失，集群管理员可以独立于集群本身更新自定义资源。
一旦安装了自定义资源，用户就可以通过kubectl创建和访问他的对象，就像操作内建资源pods那样。

# 自定义控制器

自定义资源本身让你简单地存储和索取结构化数据。
只有当和控制器结合后，他们才成为一种真正的declarative API。 控制器将结构化数据解释为用户所期望状态的记录，并且不断地采取行动来实现和维持该状态。

定制化控制器是用户可以在运行中的集群内部署和更新的一个控制器，它独立于集群本身的生命周期。 定制化控制器可以和任何一种资源一起工作，当和定制化资源结合使用时尤其有效。

## apiserver-aggregation

聚合层允许Kubernetes使用额外的API进行扩展，超出了Kubernetes核心API所提供的范围。

如果想要建立一个api server扩展，可能需要使用 apiserver聚合层，或者构建一个独立的k8s风格的 api server

详细介绍参看[如何实现自己的crd](./how-to-use-crd.md)

- [service-calalog](https://github.com/kubernetes-incubator/service-catalog/blob/master/README.md)

service-calalog 用于集成open service broker api到k8s生态体系，[详细介绍](https://github.com/kubernetes-incubator/service-catalog/blob/master/README.md).

- [apiserver-builder](https://github.com/kubernetes-incubator/apiserver-builder)

如果你想建立一个扩展API 服务器，可以考虑使用[apiserver-builder](https://github.com/kubernetes-incubator/apiserver-builder)
替代本repo,Apiserver-builder 是一个完整用于产生apiserver的框架、客户端库、和安装程序

- sample-apiserver

使用该仓库，首先fork该库，修改添加自己的类型，然后周期性的在该仓库上做变更、为该apiserver跟进更新和bug fix


## CRD

CRD是一个内建的API, 它提供了一个简单的方式来创建自定义资源。
部署一个CRD到集群中使Kubernetes API服务端开始为你指定的自定义资源服务。

这使你不必再编写自己的API服务端来处理自定义资源，但是这种实现的一般性意味着比你使用API server aggregation缺乏灵活性。

如果只是想添加资源到集群，可以考虑使用 customer resource define，简称CRD，CRD需要更少的编码和重用，
在[这里](https://kubernetes.io/docs/concepts/api-extension/custom-resources)阅读更多有关自定义资源和扩展api之间的差异
