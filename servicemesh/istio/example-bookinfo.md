# bookinfo

# 概览

在本示例中，我们将部署一个简单的应用程序，显示书籍的信息，类似于网上书店的书籍条目。在页面上有书籍的描述、详细信息（ISBN、页数等）和书评。

BookInfo 应用程序包括四个独立的微服务：

productpage：productpage(产品页面)微服务，调用 details 和 reviews 微服务来填充页面。
details：details 微服务包含书籍的详细信息。
reviews：reviews 微服务包含书籍的点评。它也调用 ratings 微服务。
ratings：ratings 微服务包含随书评一起出现的评分信息。
有3个版本的 reviews 微服务：

版本v1不调用 ratings 服务。
版本v2调用 ratings ，并将每个评级显示为1到5个黑色星。
版本v3调用 ratings ，并将每个评级显示为1到5个红色星。
应用程序的端到端架构如下所示。

![](noistio.svg)

# 安装

## 安装示例程序

- 自动注入 参考istio安装章节

```
kubectl apply -f samples/bookinfo/kube/bookinfo.yaml
```

- 手动注入

```
kubectl apply -f <(istioctl kube-inject -f samples/bookinfo/kube/bookinfo.yaml)
```

## 验证

```
$ kubectl get services
NAME          TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
details       ClusterIP   10.254.31.158    <none>        9080/TCP   1m
kubernetes    ClusterIP   10.254.0.1       <none>        443/TCP    3d
productpage   ClusterIP   10.254.20.147    <none>        9080/TCP   1m
ratings       ClusterIP   10.254.81.124    <none>        9080/TCP   1m
reviews       ClusterIP   10.254.59.16     <none>        9080/TCP   1m
sleep         ClusterIP   10.254.244.107   <none>        80/TCP     3d

$ kubectl get pods
NAME                              READY     STATUS    RESTARTS   AGE
details-v1-7b97668445-t2fxf       2/2       Running   0          57s
productpage-v1-7bbdd59459-4bpff   2/2       Running   0          57s
ratings-v1-76dc7f6b9-4v9wb        2/2       Running   0          57s
reviews-v1-64545d97b4-h4bpf       2/2       Running   0          57s
reviews-v2-8cb9489c6-7dtjl        2/2       Running   0          57s
reviews-v3-6bc884b456-sq6mq       2/2       Running   0          57s
```

## 访问bookinfo

kubectl get svc istio-ingress --namespace=istio-system
NAME            TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
istio-ingress   NodePort   10.254.83.223   <none>        80:32013/TCP,443:30784/TCP   3d

通过nodeport访问
http://172.26.6.3:32013/productpage
