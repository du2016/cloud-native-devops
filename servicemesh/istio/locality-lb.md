在云上跨区域流量通常需要收费，我们可以借助istio的能力将流量路由到同一区域内的服务将节省大量的出口流量费。

在Kubernetes中，pod的位置是通过其部署节点上的区域和区域的的标签确定的。如果您使用托管的Kubernetes服务，则云提供商应为您配置此服务。如果您正在运行自己的Kubernetes集群，则需要将这些标签添加到您的节点。Kubernetes中不存在分区概念。因此，Istio引入了自定义节点标签topology.istio.io/subzone来定义子区域。


istio使用envoy的Zone aware routing实现本地流量负载均衡，Istio从k8s获取位置信息，下发策略给envoy，这可以将流量路由到最近的容器。这样可以确保低延迟，并减少出口流量费用。

# 先决条件

- k8s 1.16以上版本集群

这里我使用了kind部署了一个三个节点的集群

- istio最新版本，未关闭locality load balancing功能


# 安装部署

## 设置node标签

```
kubectl label nodes cluster2-worker failure-domain.beta.kubernetes.io/region=us-central1
kubectl label nodes cluster2-worker failure-domain.beta.kubernetes.io/zone=us-central1-a
kubectl label nodes cluster2-worker2 failure-domain.beta.kubernetes.io/region=us-central1
kubectl label nodes cluster2-worker failure-domain.beta.kubernetes.io/zone=us-central1-b
kubectl label nodes cluster2-control-plane failure-domain.beta.kubernetes.io/region=us-central1
kubectl label nodes cluster2-control-plane failure-domain.beta.kubernetes.io/zone=us-central1-c
```

## 安装istio

```
istioctl install --set profile=demo
```


## 部署bookinfo

```
kubectl label namespace default istio-injection=enabled
# 这里我们需要修改bookinfo的deployment 分别通过nodeselector到三个节点
kubectl apply -f samples/bookinfo/platform/kube/bookinfo.yaml
```

## 查看运行的服务

可以看到reviews服务分别运行在三个节点上
```
NAME                              READY   STATUS    RESTARTS   AGE   IP           NODE               NOMINATED NODE   READINESS GATES
details-v1-79c697d759-jgh7t       2/2     Running   0          18h   10.244.1.6   cluster2-worker2   <none>           <none>
productpage-v1-65576bb7bf-rvn89   2/2     Running   0          18h   10.244.2.7   cluster2-worker    <none>           <none>
ratings-v1-7d99676f7f-2sznf       2/2     Running   0          18h   10.244.2.8   cluster2-worker    <none>           <none>
reviews-v1-987d495c-8rt8p         2/2     Running   0          18h   10.244.1.7   cluster2-worker2   <none>           <none>
reviews-v2-6c5bf657cf-7fqrx       2/2     Running   0          18h   10.244.2.9   cluster2-worker    <none>           <none>
reviews-v3-5f7b9f4f77-7j6tc       2/2     Running   0          18h   10.244.1.8   cluster2-control-plane   <none>           <none>
```

## 应用virtualservice

需要去掉subset信息

```
cat samples/bookinfo/networking/virtual-service-all-v1.yaml | sed '/subset/d' | kubectl apply -f -
```

## 应用DestinationRule

这里必须设置outlierDetection因为如果未定义异常检测配置，则代理无法确定实例是否正常，即使您启用了本地优先负载均衡，它也会全局路由流量。
```
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: productpage
spec:
  host: productpage
  trafficPolicy:
    outlierDetection:
      consecutiveErrors: 7
      interval: 30s
      baseEjectionTime: 30s

---

apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: reviews
spec:
  host: reviews
  trafficPolicy:
    outlierDetection:
      consecutiveErrors: 7
      interval: 30s
      baseEjectionTime: 30s

---

apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: ratings
spec:
  host: ratings
  trafficPolicy:
    outlierDetection:
      consecutiveErrors: 7
      interval: 30s
      baseEjectionTime: 30s

---

apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: details
spec:
  host: details
  trafficPolicy:
    outlierDetection:
      consecutiveErrors: 7
      interval: 30s
      baseEjectionTime: 30s
```

## 测试

从运行节点查看，可以看到productpage应用和v2版本的Reviews运行在同一个zone,也就是cluster2-worker节点上

```
kubectl port-forward service/productpage 80:9080
```

在浏览器访问http://127.0.0.1/productpage,刷新几次，可见服务一直访问的v2版本的Reviews
![](http://img.rocdu.top/20201023/productpage.png)


# 局部加权负载平衡

大多数用例都可以与本地优先的负载平衡一起使用。但是，在某些用例中，您可能需要将流量分成多个区域。如果所有请求都来自单个区域，则可能不想使一个区域超载。

修改reviews的DestinationRule如下

```
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: reviews
spec:
  host: reviews
  trafficPolicy:
    outlierDetection:
      consecutiveErrors: 7
      interval: 30s
      baseEjectionTime: 30s
    loadBalancer:
      localityLbSetting:
        enabled: true
        distribute:
        - from: us-central1/us-central1-a/*
          to:
            "us-central1/us-central1-a/*": 80
            "us-central1/us-central1-b/*": 20
```

访问productpage可以发现，reviews 80%到v2,20%到v3

扫描关注我:

![微信](http://img.rocdu.top/20201023/qrcode_for_gh_7457c3b1bfab_258.jpg)
