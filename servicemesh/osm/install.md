# 介绍

Open Service Mesh（OSM）是一种轻量级，可扩展的Cloud Native服务网格，它使用户能够统一管理，保护和获得针对高度动态微服务环境的开箱即用的可观察性功能。

# 安装osm


```
wget https://github.com/openservicemesh/osm/releases/download/v0.1.0/osm-v0.1.0-darwin-amd64.tar.gz
tar xf osm-v0.1.0-darwin-amd64.tar.gz

# pre check
./darwin-amd64/osm check --pre-install

# 查看服务状态
kubectl get pods -n osm-system
NAME                              READY   STATUS    RESTARTS   AGE
osm-controller-6cc9f4866c-28hg8   1/1     Running   0          6m3s
osm-grafana-58ff65dfb7-tprvl      1/1     Running   0          6m3s
osm-prometheus-5756769877-mdmsv   1/1     Running   0          6m3s
zipkin-6df4b57677-4qkfq           1/1     Running   0          6m3s
```

# 安装demo程序

接下来我们将会安装一个demo服务

服务调用关系如下
![](http://img.rocdu.top/20200806/graph.png)

- Bookbuyer和bookthief不断向bookstore发出HTTP GET请求以购买书和访问github.com来验证出口流量。
 
- bookstore是由两个服务器支持的服务：bookstore-v1和bookstore-v2。 无论何时出售一本书，它都会向bookstore发出HTTP POST请求以进行补货。

- 下载代码仓库，进行配置

```
git clone https://github.com/openservicemesh/osm
cd osm

修改 .env文件 以下变量
export CTR_REGISTRY=openservicemesh
export CTR_TAG=latest
```

- 创建ns

```
./demo/configure-app-namespaces.sh
```

- 部署demo应用

```
demo/deploy-apps.sh
```

- HTTPRouteGroup规则

```
./demo/deploy-traffic-specs.sh
```

- TrafficSplit规则

```
demo/deploy-traffic-split.sh
```

- TrafficTarget规则

```
demo/deploy-traffic-target.sh
```

- 配置注入的命名空间，等同于istio的自动注入

```
demo/join-namespaces.sh
```

- 重启容器

```
demo/rolling-restart.sh
``` 

# 流量观测

- zipkin 

执行port forward
```
./scripts/port-forward-zipkin.sh
```

在浏览器打开http://127.0.0.1:9411/可以看到服务间的调用信息

![](http://img.rocdu.top/20200806/osm-zipkin.png)

- grafana

./scripts/port-forward-grafana.sh

![](http://img.rocdu.top/20200806/osm-grafana.png)

# ingress管理

通过ingress暴露一个http服务
```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: bookstore-v1
  namespace: bookstore-ns
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - host: bookstore-v1.bookstore-ns.svc.cluster.local
    http:
      paths:
      - path: /books-bought
        backend:
          serviceName: bookstore-v1
          servicePort: 80
```


暴露一个https服务
```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: bookstore-v1
  namespace: bookstore-ns
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    nginx.ingress.kubernetes.io/configuration-snippet: |
      proxy_ssl_name "bookstore-v1.bookstore-ns.svc.cluster.local";
    nginx.ingress.kubernetes.io/proxy-ssl-secret: "osm-system/osm-ca-bundle"
    nginx.ingress.kubernetes.io/proxy-ssl-server-name: "on" # optional
    nginx.ingress.kubernetes.io/proxy-ssl-verify: "on"
spec:
  rules:
  - host: bookstore-v1.bookstore-ns.svc.cluster.local
    http:
      paths:
      - path: /books-bought
        backend:
          serviceName: bookstore-v1
          servicePort: 80
```

# egress管理

获取集群内的CIDR
```
./scripts/get_mesh_cidr.sh
```

可以在安装时通过参数指定CIDR范围从而避免出口流量受到sidecar影响

```
osm install --enable-egress --mesh-cidr "10.244.0.0/16,10.96.0.0/12" --mesh-name='osm'
```

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
