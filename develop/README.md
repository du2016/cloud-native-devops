# 开发指南

# API扩展

- apiserver-aggregation
允许k8s的开发人员编写一个自己的服务，可以把这个服务注册到k8s的api里面，这样，就像k8s自己的api一样，
你的服务只要运行在k8s集群里面，k8s 的Aggregate通过service名称就可以转发到你写的service里面。

- CRD

CRD是一个内建的API, 它提供了一个简单的方式来创建自定义资源。
部署一个CRD到集群中使Kubernetes API服务端开始为你指定的自定义资源服务。

这使你不必再编写自己的API服务端来处理自定义资源，但是这种实现的一般性意味着比你使用API server aggregation缺乏灵活性。

如果只是想添加资源到集群，可以考虑使用 customer resource define，简称CRD，CRD需要更少的编码和重用，
在[这里](https://kubernetes.io/docs/concepts/api-extension/custom-resources)阅读更多有关自定义资源和扩展api之间的差异


# scheduler

在pod内指定schedulerName可以选择自己的scheduler

# storage

可以通过以下两种方式实现
- flexvolume
- csi

# kubectl

可以实现kubectl的子命令更好的过滤自己需要的资源

# 网络插件

- cni
[网络接口规范](https://github.com/containernetworking/cni/blob/master/SPEC.md)

# custom-metrics-apiserver

custom-metrics-apiserver 可以实现自定义指标HPA
https://github.com/kubernetes-incubator/custom-metrics-apiserver 提供了实现框架