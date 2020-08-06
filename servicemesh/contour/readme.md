# contour

Contour是开源的Kubernetes入口控制器，
为Envoy边缘和服务代理提供控制平面.
Contour支持动态配置更新和多团队入口委托，同时保持轻量级配置文件。

# 特点

- 内置envoy

Contour是基于Envoy，高性能L7代理和负载均衡器的控制平面

- 灵活的架构

轮廓可以部署为Kubernetes部署或守护程序集

- TLS证书授权

管理员可以安全地委派通配符证书访问

# 架构原理

- Envoy，提供高性能的反向代理。
- Contour，充当Envoy的管理服务器并为其提供配置。

![](http://img.rocdu.top/20200527/archoverview.png)


# 部署

```bash
kubectl apply -f https://projectcontour.io/quickstart/contour.yaml
```

# HTTPProxy

除了支持原生的ingres规则外，因为ingress-nginx 注解很驳杂，不利于使用，
contour还抽象了HTTPProxy概念，

## HTTPProxy的主要优势

- 安全地支持多团队Kubernetes集群，并具有限制哪些命名空间可以配置虚拟主机和TLS凭据的能力。
- 允许包括来自另一个HTTPProxy（可能在另一个命名空间中）的路径或域的路由配置。
- 在一条路由中接受多种服务，并在它们之间负载均衡流量。
- 本机允许定义服务加权和负载平衡策略而无需注释。
- 在创建时验证HTTPProxy对象，并为创建后的有效性进行状态报告。


如下ingress配置
```
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: basic
spec:
  rules:
  - host: foo-basic.bar.com
    http:
      paths:
      - backend:
          serviceName: s1
          servicePort: 80
```

可以用以下httpproxy规则表示

```
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: basic
spec:
  virtualhost:
    fqdn: foo-basic.bar.com
  routes:
    - conditions:
      - prefix: /
      services:
        - name: s1
          port: 80
          protocol: h2c
```

如果要支持gprc等http2流量只需要设置 spec.routes[*].conditions.services[*].protocol

## 多上游配置

contour支持多上游配置，方便的实现多版本流量控制

```
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: weight-shifting
  namespace: default
spec:
  virtualhost:
    fqdn: weights.bar.com
  routes:
    - services:
        - name: s1
          port: 80
          weight: 10
        - name: s2
          port: 80
          weight: 90
```

## 请求和响应头策略

服务或路由还支持操作标头。可以按照以下步骤设置标头或从请求或响应中删除标头

```
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: header-manipulation
  namespace: default
spec:
  virtualhost:
    fqdn: headers.bar.com
  routes:
    - services:
        - name: s1
          port: 80
          requestHeadersPolicy:
            set:
              - name: X-Foo
                value: bar
            remove:
              - X-Baz
          responseHeadersPolicy:
            set:
              - name: X-Service-Name
                value: s1
            remove:
              - X-Internal-Secret
```


## 流量镜像

每个路由都可以将服务指定为镜像。镜像服务将接收发送到任何非镜像服务的读取流量的副本。镜像流量被视为只读，镜像的任何响应都将被丢弃。

该服务对于记录流量以供以后重播或对新部署进行冒烟测试很有用。

```
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: traffic-mirror
  namespace: default
spec:
  virtualhost:
    fqdn: www.example.com
  routes:
    - conditions:
      - prefix: /
      services:
        - name: www
          port: 80
        - name: www-mirror
          port: 80
          mirror: true
```

## 响应超时

可以将每个路由配置为具有超时策略和重试策略，如下所示：

```
# httpproxy-response-timeout.yaml
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: response-timeout
  namespace: default
spec:
  virtualhost:
    fqdn: timeout.bar.com
  routes:
  - timeoutPolicy:
      response: 1s
      idle: 10s
    retryPolicy:
      count: 3
      perTryTimeout: 150ms
    services:
    - name: s1
      port: 80
```

## 负载均衡策略

支持以下负载均衡算法

可供选择的选项：

- RoundRobin：按循环顺序选择每个正常的上游端点（如果未选择，则为默认策略）。
- WeightedLeastRequest：最少请求策略使用O（1）算法，该算法选择两个随机的健康端点，并选择活动请求较少的端点。注意：此算法非常简单，足以进行负载测试。如果需要真正的加权最小请求行为，则不应使用它。
- Random：随机策略选择随机的健康端点。

```
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: lb-strategy
  namespace: default
spec:
  virtualhost:
    fqdn: strategy.bar.com
  routes:
    - conditions:
      - prefix: /
      services:
        - name: s1-strategy
          port: 80
        - name: s2-strategy
          port: 80
      loadBalancerPolicy:
        strategy: WeightedLeastRequest
```

## 会话亲和

会话亲缘关系（也称为粘性会话）是一种负载平衡策略，通过该策略，
来自单个客户端的一系列请求将始终路由到同一应用程序后端。
Contour支持基于的每个路由的会话相似性`loadBalancerPolicy strategy: Cookie`。

```
# httpproxy-sticky-sessions.yaml
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: httpbin
  namespace: default
spec:
  virtualhost:
    fqdn: httpbin.davecheney.com
  routes:
  - services:
    - name: httpbin
      port: 8080
    loadBalancerPolicy:
      strategy: Cookie
```

还有其他的一些功能，详见[httpproxy说明](https://projectcontour.io/docs/v1.4.0/httpproxy/)

# 跨集群流量管理gimbal

通过gimbal可以实现夸集群的流量统一管理，
通过监视单个Kubernetes群集的可用服务和端点并将它们同步到主机Gimbal群集中来实现此目的。 Discoverer将利用Kubernetes API的监视功能来动态接收更改，而不必轮询API。
所有可用的服务和端点都将同步到与源系统匹配的相同名称空间。 发现者将仅负责一次监视单个集群。如果需要监视多个集群，则将需要部署多个发现者。

## 安装


```
# Sample secret creation
$ kubectl create secret generic remote-discover-kubecfg --from-file=./config --from-literal=backend-name=nodek8s -n gimbal-discovery
```


# 总结

contour 做为envoy的控制平面，可以动态下发各种流量管理策略，其实现的功能都是较为常用的功能保证了envoy的高性能
，可以轻松实现一个分布式gateway，但是对于部分功能例如限流，并没有进行支持，在使用中我们自行实现了这部分功能.

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)