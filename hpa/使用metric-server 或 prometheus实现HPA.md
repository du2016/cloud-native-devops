# 核心指标管道

从 Kubernetes 1.8 开始，资源使用指标（如容器 CPU 和内存使用率）通过 Metrics API 在 Kubernetes 中获取。
这些指标可以直接被用户访问(例如通过使用 kubectl top 命令)，或由集群中的控制器使用(例如，Horizontal Pod Autoscale 
可以使用这些指标作出决策)。



# Resource Metrics API

通过 Metrics API，您可以获取指定 node 或 pod 当前使用的资源量。这个 API 不存储指标值，
因此想要获取某个指定 node 10 分钟前的资源使用量是不可能的。

Metrics API 和其他的 API 没有什么不同：

它可以通过与 `/apis/metrics.k8s.io/` 路径下的其他 Kubernetes API 相同的端点来发现
它提供了相同的安全性、可扩展性和可靠性保证
Metrics API 在 [k8s.io/metrics](https://github.com/kubernetes/metrics/blob/master/pkg/apis/metrics/v1beta1/types.go) 仓库中定义。您可以在这里找到关于Metrics API 的更多信息。

注意： Metrics API 需要在集群中部署 Metrics Server。否则它将不可用。


# Metrics Server

Metrics Server 实现了Resource Metrics API

[Metrics Server](https://github.com/kubernetes-incubator/metrics-server) 是集群范围资源使用数据的聚合器。 
从 Kubernetes 1.8 开始，它作为一个 Deployment 对象默认部署在由 kube-up.sh 脚本创建的集群中。
如果您使用了其他的 Kubernetes 安装方法，您可以使用 Kubernetes 1.7+ (请参阅下面的详细信息) 
中引入的 [deployment yamls](https://github.com/kubernetes-incubator/metrics-server/tree/master/deploy) 文件来部署。

Metrics Server 从每个节点上的 Kubelet 公开的 Summary API 中采集指标信息。

通过在主 API server 中注册的 Metrics Server [Kubernetes 聚合器](https://kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/) 来采集指标信息， 这是在 Kubernetes 1.7 中引入的。

在 [设计文档](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/instrumentation/metrics-server.md) 中可以了解到有关 Metrics Server 的更多信息。

# custom metrics api

该API允许消费者访问任意度量描述Kubernetes资源。

API的目的是通过监测管道供应商，在其度量存储解决方案之上实现。

如果你想实现这个API API服务器，请参阅[kubernetes-incubator/custom-metrics-apiserver](https://github.com/kubernetes-incubator/custom-metrics-apiserver)库，其中包含需要建立这样一个API服务器基础设施
其中包含设置这样一个API服务器所需的基本基础设施。

Import Path: `k8s.io/metrics/pkg/apis/custom_metrics.`


# custom metrics apiserver

[custom metrics apiserver](https://github.com/kubernetes-incubator/custom-metrics-apiserver)是为了实现k8s自定义监控指标
的框架。

# HPA

自动伸缩是一种根据资源使用情况自动伸缩工作负载的方法。
自动伸缩在Kubernetes中有两个维度:cluster Autoscaler处理节点扩容操作和Horizontal Pod Autoscaler自动缩放rs或rc中的pod。
cluster Autoscaler和Horizontal Pod Autoscaler一起可用于动态调整的计算能力以及并行性的水平,你的系统需要满足sla。
虽然cluster Autoscaler高度依赖于托管集群的云提供商的底层功能，但是HPA可以独立于您的IaaS/PaaS提供商进行操作。

在Kubernetes v1.1中首次引入了hpa特性，自那时起已经有了很大的发展。
hpa第一个版本基于观察到的CPU利用率，后续版本支持基于内存使用。
在Kubernetes 1.6中引入了一个新的API自定义指标API，它允许HPA访问任意指标。
Kubernetes 1.7引入了聚合层，允许第三方应用程序通过注册为API附加组件来扩展Kubernetes API。
自定义指标API以及聚合层使得像Prometheus这样的监控系统可以向HPA控制器公开特定于应用程序的指标。

hpa 实现了一个控制环，可以周期性的从资源指标API查询特定应用的CPU/MEM信息。

![](https://github.com/stefanprodan/k8s-prom-hpa/raw/master/diagrams/k8s-hpa.png)

# 实战

以下是关于Kubernetes 1.9或更高版本的HPA v2配置的分步指南。您将安装提供核心指标的度量服务器附加组件，
然后您将使用一个演示应用程序来展示基于CPU和内存使用的pod自动伸缩。在指南的第二部分，
您将部署Prometheus和一个自定义API服务器。您将使用聚合器层注册自定义API服务器，然后使用演示应用程序提供的自定义度量配置HPA。

## 前提

- go 1.8+
- clone [k8s-prom-hpa](https://github.com/stefanprodan/k8s-prom-hpa) repo

```
cd $GOPATH
git clone https://github.com/stefanprodan/k8s-prom-hpa
```

## 安装 Metrics Server

Kubernetes Metrics Server是一个集群范围的资源使用数据聚合器，是Heapster的继承者。
metrics服务器通过从kubernet.summary_api收集数据收集节点和pod的CPU和内存使用情况。
summary API是一个内存有效的API，用于将数据从Kubelet/cAdvisor传递到metrics server。

![](https://github.com/stefanprodan/k8s-prom-hpa/raw/master/diagrams/k8s-hpa-ms.png)

如果在v1版本的HPA中，您将需要Heapster提供CPU和内存指标，在HPA v2和Kubernetes 1.8中，
只有度量服务器是需要的，而水平-pod-autoscaler-use-rest-客户机是打开的。
在Kubernetes 1.9中默认启用HPA rest客户端。GKE 1.9附带了预先安装的指标服务器。

在`kube-system`命名空间总部署metrics-server

```
kubectl create -f ./metrics-server
```

一分钟后，度量服务器开始报告节点和荚的CPU和内存使用情况。
查看nodes指标

```
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/nodes" | jq .
```

查看pod指标

```
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/pods" | jq .
```

## 基于CPU和内存使用的自动缩放

你将使用一个基于golang的小程序测试hpa.

部署podinfo到默认命名空间

```
kubectl create -f ./podinfo/podinfo-svc.yaml,./podinfo/podinfo-dep.yaml
```

在`http://<K8S_PUBLIC_IP>:31198`通过nodeport访问`podinfo`

接下来定义一个HPA，保持最小两个副本和最大十个如果CPU平均超过80%或如果内存超过200mi。

```bash
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: podinfo
spec:
  scaleTargetRef:
    apiVersion: extensions/v1beta1
    kind: Deployment
    name: podinfo
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      targetAverageUtilization: 80
  - type: Resource
    resource:
      name: memory
      targetAverageValue: 200Mi
```

创建HPA
```
kubectl create -f ./podinfo/podinfo-hpa.yaml
```

几秒钟之后，HPA控制器与metrics server联系，然后取出CPU和内存使用情况。

```
kubectl get hpa

NAME      REFERENCE            TARGETS                      MINPODS   MAXPODS   REPLICAS   AGE
podinfo   Deployment/podinfo   2826240 / 200Mi, 15% / 80%   2         10        2          5m
```

为了提高CPU使用率、运行`rakyll/hey`进行压力测试

```
#install hey
go get -u github.com/rakyll/hey

#do 10K requests
hey -n 10000 -q 10 -c 5 http://<K8S_PUBLIC_IP>:31198/
```

你可以通过以下命令获取HPA event

```
$ kubectl describe hpa

Events:
  Type    Reason             Age   From                       Message
  ----    ------             ----  ----                       -------
  Normal  SuccessfulRescale  7m    horizontal-pod-autoscaler  New size: 4; reason: cpu resource utilization (percentage of request) above target
  Normal  SuccessfulRescale  3m    horizontal-pod-autoscaler  New size: 8; reason: cpu resource utilization (percentage of request) above target
```

先将`podinfo`移除一会儿，稍后将再次部署：

```
kubectl delete -f ./podinfo/podinfo-hpa.yaml,./podinfo/podinfo-dep.yaml,./podinfo/podinfo-svc.yaml
```

## 安装Custom Metrics Server

为了根据custom metrics进行扩展，您需要有两个组件。一个从应用程序中收集指标并将其存储为Prometheus时间序列数据库的组件。
第二个组件将Kubernetes自定义指标API扩展到由收集的k8s-prometheus-adapter提供的指标。

![](https://github.com/stefanprodan/k8s-prom-hpa/raw/master/diagrams/k8s-hpa-prom.png)

您将在专用命名空间中部署Prometheus和adapter。

创建`monitoring`命名空间

```
kubectl create -f ./namespaces.yaml
```

将 Prometheus v2部署到`monitoring`命名空间:
如果您部署到GKE，您可能会得到一个错误:从服务器(禁止)中出错:创建这个错误将帮助您解决这个问题:[RBAC on GKE](https://github.com/coreos/prometheus-operator/blob/master/Documentation/troubleshooting.md)。


```
kubectl create -f ./prometheus
```

生成由Prometheus adapter所需的TLS证书:

```
make certs
```

部署Prometheus自定义api适配器

```bash
kubectl create -f ./custom-metrics-api
```

列出由prometheus提供的自定义指标：

```
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq .
```

获取`monitoring`命名空间中所有pod的FS信息：

```
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/monitoring/pods/*/fs_usage_bytes" | jq .

```

## 基于自定义指标的自动扩容

创建podinfo nodeport服务并在default命名空间中部署：

```bash
kubectl create -f ./podinfo/podinfo-svc.yaml,./podinfo/podinfo-dep.yaml
```

`podinfo`应用程序的暴露了一个自定义的度量http_requests_total。普罗米修斯适配器删除`_total`后缀标记度量作为一个计数器度量

从自定义度量API获取每秒的总请求数:

```
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pods/*/http_requests" | jq .
```

```
{
  "kind": "MetricValueList",
  "apiVersion": "custom.metrics.k8s.io/v1beta1",
  "metadata": {
    "selfLink": "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pods/%2A/http_requests"
  },
  "items": [
    {
      "describedObject": {
        "kind": "Pod",
        "namespace": "default",
        "name": "podinfo-6b86c8ccc9-kv5g9",
        "apiVersion": "/__internal"
      },
      "metricName": "http_requests",
      "timestamp": "2018-01-10T16:49:07Z",
      "value": "901m"
    },
    {
      "describedObject": {
        "kind": "Pod",
        "namespace": "default",
        "name": "podinfo-6b86c8ccc9-nm7bl",
        "apiVersion": "/__internal"
      },
      "metricName": "http_requests",
      "timestamp": "2018-01-10T16:49:07Z",
      "value": "898m"
    }
  ]
}
```

`m`代表`milli-units`，例如，`901m`意味着`milli-requests`。

创建一个HPA，如果请求数超过每秒10当将扩大podinfo数量：

```
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: podinfo
spec:
  scaleTargetRef:
    apiVersion: extensions/v1beta1
    kind: Deployment
    name: podinfo
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Pods
    pods:
      metricName: http_requests
      targetAverageValue: 10
```

在`default`命名空间部署`podinfo` HPA：

```
kubectl create -f ./podinfo/podinfo-hpa-custom.yaml
```

过几秒钟HPA从标准的API取得`http_requests`的值：

```
kubectl get hpa

NAME      REFERENCE            TARGETS     MINPODS   MAXPODS   REPLICAS   AGE
podinfo   Deployment/podinfo   899m / 10   2         10        2          1m
```

用25每秒请求数给`podinfo`服务加压
```bash
#install hey
go get -u github.com/rakyll/hey

#do 10K requests rate limited at 25 QPS
hey -n 10000 -q 5 -c 5 http://<K8S-IP>:31198/healthz
```

几分钟后，HPA开始扩大部署。

```bash
kubectl describe hpa

Name:                       podinfo
Namespace:                  default
Reference:                  Deployment/podinfo
Metrics:                    ( current / target )
  "http_requests" on pods:  9059m / 10
Min replicas:               2
Max replicas:               10

Events:
  Type    Reason             Age   From                       Message
  ----    ------             ----  ----                       -------
  Normal  SuccessfulRescale  2m    horizontal-pod-autoscaler  New size: 3; reason: pods metric http_requests above target
```

以每秒当前的请求速率，部署将永远无法达到10个荚的最大值。三副本足以让RPS在10每pod.

负载测试结束后，HPA向下扩展部署到初始副本。

```
Events:
  Type    Reason             Age   From                       Message
  ----    ------             ----  ----                       -------
  Normal  SuccessfulRescale  5m    horizontal-pod-autoscaler  New size: 3; reason: pods metric http_requests above target
  Normal  SuccessfulRescale  21s   horizontal-pod-autoscaler  New size: 2; reason: All metrics below target
```

你可能已经注意到，自动定标器不使用峰值立即做出反应。默认情况下，指标每30秒同步一次，
并且扩展/收缩当3-5分钟没有重新扩展发生变化时。在这种方式中，HPA防止快速执行并保留了指标生效时间




# 总结

不是所有的系统都可以依靠CPU/内存使用指标单独满足SLA，大多数Web和移动后端需要以每秒请求处理任何突发流量进行自动缩放。
对于ETL应用程序，可能会由于作业队列长度超过某个阈值而触发自动缩放，等等。
通过prometheus检测你应用程序的正确指，并为自动是很所提供正确指标，您可以微调您的应用程序更好地处理突发和确保高可用性。


欢迎加入QQ群：k8s开发与实践（482956822）一起交流k8s技术