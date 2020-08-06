

Knative有两个组件，可以独立安装或一起使用。为了帮助您挑选适合自己的组件，以下是每个组件的简要说明：

- Serving 为基于无状态请求的服务提供了一种零扩展抽象。
- Eventing提供了抽象来启用绑定事件源（例如Github Webhooks，Kafka）和使用者（例如Kubernetes或Knative Services）的绑定。

Knative还具有一个Observability插件，该插件提供了标准工具，可用于查看Knative上运行的软件的运行状况

本文将安装Serving后运行一个hello world程序

# 先决条件

本指南假定您要在Kubernetes群集上安装上游Knative版本。 越来越多的供应商已经管理Knative产品。 有关完整列表，请参见Knative产品页面。

Knative v0.15.0需要Kubernetes集群v1.15或更高版本，以及兼容的kubectl。 本指南假定您已经创建了Kubernetes集群，并且在Mac或Linux环境中使用bash。 在Windows环境中需要调整一些命令

# 安装Serving组件

1.使用以下命令安装crd

```
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.15.0/serving-crds.yaml
```

2.serving的安装核心组件

```
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.15.0/serving-core.yaml
```

3.安装网络层
    

- 安装contour

```
kubectl apply --filename https://github.com/knative/net-contour/releases/download/v0.15.0/contour.yaml
```

- 安装knative contour controller

```
kubectl apply --filename https://github.com/knative/net-contour/releases/download/v0.15.0/net-contour.yaml
```

- 配置knativeserving使用Contour

```
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
```

- 获取ip

```
kubectl --namespace contour-external edit service envoy
```

4. 配置DNS
  
因为我们使用kind安装此步骤跳过

# 配置服务

```
配置文件指定有关应用程序的元数据，指向要部署的应用程序的托管镜像，并允许deployment可配置，
创建一个名为service.yaml的新文件。

apiVersion: serving.knative.dev/v1 # Current version of Knative
kind: Service
metadata:
  name: helloworld-go # The name of the app
  namespace: default # The namespace the app will use
spec:
  template:
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go # The URL to the image of the app
          env:
            - name: TARGET # The environment variable printed out by the sample app
              value: "Go Sample v1"
```

# 部署应用

kubectl apply --filename service.yaml


现在，您的服务已创建，Knative将执行以下步骤：

- 为此版本的应用程序创建一个新的不变版本。
- 执行网络编程，为您的应用创建路由，ingress，service,load balancer。
- 根据流量自动扩缩Pod，包括将活动Pod调整为零


> 刚部署完pod数量为1，过一段时间后可以看到pod索容为0

## 与应用交互


- 交互
因为我们使用的是kind所以启动一个centos与我们的服务进行交互

```
kubectl run centos --image=centos -- sleep 10d
kubectl exec -it centos bash
curl helloworld-go.default
Hello Go Sample v1!
```


- 查看contour httpproxy

```
kubectl get httpproxy
NAME                                                    FQDN                                      TLS SECRET   STATUS   STATUS DESCRIPTION
helloworld-go-helloworld-go.default                     helloworld-go.default                                  valid    valid HTTPProxy
helloworld-go-helloworld-go.default.example.com         helloworld-go.default.example.com                      valid    valid HTTPProxy
helloworld-go-helloworld-go.default.svc                 helloworld-go.default.svc                              valid    valid HTTPProxy
helloworld-go-helloworld-go.default.svc.cluster.local   helloworld-go.default.svc.cluster.local                valid    valid HTTPProxy
```


- 查看svc

```
 kubectl get svc
NAME                          TYPE           CLUSTER-IP       EXTERNAL-IP                                PORT(S)                             AGE
helloworld-go                 ExternalName   <none>           envoy.contour-internal.svc.cluster.local   <none>                              63s
```

可见knative coutour controller 创建了一个ExternalName类型的svc,cname到了envoy.contour-internal.svc.cluster.local，也就是我们的代理服务，这样我们在内部就可以通过内部域名直接访问我们的服务



扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
