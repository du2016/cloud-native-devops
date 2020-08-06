# 介绍
Knative Serving项目提供了中间件原语，这些原语可实现：

- 快速部署无服务器容器
- 自动放大和缩小到零
- Istio组件的路由和网络编程
- 部署的代码和配置的时间点快照

# Serving 资源
Knative Serving将一组对象定义为Kubernetes自定义资源定义（CRD）。这些对象用于定义和控制无服务器工作负载在集群上的行为：

- 服务：service.serving.knative.dev资源自动管理您的工作负载的整个生命周期。它控制其他对象的创建，以确保您的应用程序具有针对服务的每次更新的路由，配置和新修订版。可以将服务定义为始终将流量路由到最新修订版或固定修订版。
- 路线：route.serving.knative.dev资源将网络端点映射到一个或多个修订版。您可以通过多种方式管理流量，包括部分流量和命名路由。
- 配置：configuration.serving.knative.dev资源保持部署所需的状态。它在代码和配置之间提供了清晰的分隔，并遵循了十二要素应用程序方法。修改配置会创建一个新修订。
- 修订：revision.serving.knative.dev资源是对工作负载进行的每次修改的代码和配置的时间点快照。修订是不可变的对象，可以保留很长时间。可以根据传入流量自动缩放“服务提供修订”。


![](https://github.com/knative/serving/raw/master/docs/spec/images/object_model.png)



# 组件

## Serving：activator

激活器负责接收和缓冲非活动修订的请求，并向自动定标器报告指标。在自动缩放器根据报告的指标缩放修订版本后，它还会重试对修订的请求。

## Serving：autoscaler
自动缩放器接收请求指标并调整处理流量负载所需的Pod数量。

## Serving：controller

控制器服务协调所有公共Knative对象和自动缩放CRD。当用户向Kubernetes API应用Knative服务时，这将创建配置和路由。它将配置转换为修订版，将修订转换为部署和Knative Pod自动缩放器（KPA）。

## Serving：webhook

Webhook拦截所有Kubernetes API调用以及所有CRD插入和更新。它设置默认值，拒绝不一致和无效的对象，并验证和更改Kubernetes API调用。

## Deployment: networking-certmanager
证书管理器将群集入口协调为证书管理器对象。

## Deployment：networking-istio

