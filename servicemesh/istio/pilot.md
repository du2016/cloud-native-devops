# pilot

## 功能

- 请求路由
- 服务发现负载均衡
- 故障处理
> 客户端设置超时重试："x-envoy-upstream-rq-timeout-ms"和"x-envoy-max-retries"。
> 熔断器当和容错库同时使用时，最终响应内容取决于谁先触发熔断。
- 故障注入

## 配置

- Route Rules/路由规则
- DestinationPolicies/目的地策略
- Egress Rule/出口规则

## Route Rules/路由策略  针对source的策略

### 设置全局请求路由

```
piVersion: config.istio.io/v1alpha2
kind: RouteRule
metadata:
  name: reviews-default
spec:
  destination:    #fqdn 全限定域名
    name: reviews 
    namespace: default  #可省略
    domain: svc.cluster.local #可省略
  route:
  - labels:
      version: v1
    weight: 100
```

### 根据source设置


```
 apiVersion: config.istio.io/v1alpha2
 kind: RouteRule
 metadata:
   name: reviews-to-ratings
 spec:
   destination:
     name: ratings
   match:
     source:
       name: reviews
       labels:
         version: v2
```

### 基于header设置

```
apiVersion: config.istio.io/v1alpha2
kind: RouteRule
metadata:
  name: ratings-jason
spec:
  destination:
    name: reviews
    labels:
      version: v2
  match:
    request:
      headers:
        cookie:
          regex: "^(.*?;)?(user=jason)(;.*)?$"
```

### 设置权重

```
apiVersion: config.istio.io/v1alpha2
kind: RouteRule
metadata:
  name: reviews-v2-rollout
spec:
  destination:
    name: reviews
  route:
  - labels:
      version: v2
    weight: 25
  - labels:
      version: v1
    weight: 75
```

### 设置超时重试

```
apiVersion: config.istio.io/v1alpha2
kind: RouteRule
metadata:
  name: ratings-retry
spec:
  destination:
    name: ratings
  route:
  - labels:
      version: v1
  httpReqRetries:
    simpleRetry:
      attempts: 3   #重试三次
    simpleTimeout:   
      timeout: 10s  # 超时时间十秒
  httpFault:
    delay:
      percent: 10   #百分比10
      fixedDelay: 5s    # 5s延迟
    abort:
      percent: 10   #百分之十
      httpStatus: 400 #返回400
```

### 设置权重

```
apiVersion: config.istio.io/v1alpha2
kind: RouteRule
metadata:
  name: reviews-foo-bar
spec:
  destination:
    name: reviews
  precedence: 2
  match:
    request:
      headers:
        Foo: bar
  route:
  - labels:
      version: v2
---
apiVersion: config.istio.io/v1alpha2
kind: RouteRule
metadata:
  name: reviews-default
spec:
  destination:
    name: reviews
  precedence: 1
  route:
  - labels:
      version: v1
    weight: 100
```

## DestinationPolicies/目的地策略  针对目标地址的策略

可以对负载均衡算法，熔断器配置，健康检查进行配置

源reviews v2目标ratings v1进行轮训
```
apiVersion: config.istio.io/v1alpha2
metadata:
  name: ratings-lb-policy
spec:
  source:
    name: reviews
    labels:
      version: v2
  destination:
    name: ratings
    labels:
      version: v1
  loadBalancing:
    name: ROUND_ROBIN
```

熔断
限制reviews v1的最大连接数100
```
apiVersion: config.istio.io/v1alpha2
metadata:
  name: reviews-v1-cb
spec:
  destination:
    name: reviews
    labels:
      version: v1
  circuitBreaker:
    simpleCb:
       maxConnections: 100
```

## Egress Rule/出栈策略  访问外部服务配置

现在istio只支持http访问外部服务，若要访问https，则需要让sidecar通过https访问外部服务

    http            https
app------->sidecar-------->egress api
```
apiVersion: config.istio.io/v1alpha2
kind: EgressRule
metadata:
  name: foo-egress-rule
spec:
  destination:
    service: *.foo.com
  ports:
    - port: 80
      protocol: http
    - port: 443
      protocol: https
```



~~~yaml
  cat <<EOF | istioctl create -f -
  apiVersion: config.istio.io/v1beta1
  kind: DestinationPolicy
  metadata:
  name: httpbin-circuit-breaker
  spec:
  destination:
    name: httpbin
    labels:
      version: v1
  circuitBreaker:
    simpleCb:
      maxConnections: 1
      httpMaxPendingRequests: 1
      sleepWindow: 3m
      httpDetectionInterval: 1s
      httpMaxEjectionPercent: 100
      httpConsecutiveErrors: 1
      httpMaxRequestsPerConnection: 1
  EOF
~~~