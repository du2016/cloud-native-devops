# 介绍

flagger是一个k8s operator,可以基于多种ingress 实现金丝雀升级，以进行流量转移,并使用Prometheus指标进行流量分析。canary分析器可以通过webhooks进行扩展，以运行系统集成/验收测试，负载测试或任何其他自定义验证。


Flagger实现了一个控制环路，该环路逐渐将流量转移到金丝雀，同时测量关键性能指标，例如HTTP请求成功率，请求平均持续时间和Pod运行状况。 基于对KPI的分析，金丝雀会被提升或中止.

![](http://img.rocdu.top/20200807/flagger-canary-overview.png)

# 工作原理

flaager 可以通过自定义canary资源用于自动化发布k8s 工作负载程序

当创建一个canary资源，会将流量切换到 $APP-primary deployment，对应 $APP-primary service，并将$APP deployment 数量设置为0，在进行升级时，金丝雀流量对应 $APP 对应svc为$APP-canary,在金丝雀流量验证完毕后会对$APP-primary deployment进行升级

## canary资源

Canary自定义资源定义了在Kubernetes上运行的应用程序的发布过程，并且可以跨集群，服务网格和入口提供程序进行移植。
对于名为podinfo的部署，可以将具有渐进式流量转移的Canary版本定义为：


```
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: podinfo
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: podinfo
  service:
    port: 9898
  analysis:  # 流量分析
    interval: 1m  # 分析间隔
    threshold: 10 # 分析次数
    maxWeight: 50 # 最大权重
    stepWeight: 5 # 每次增加多少权重
    metrics:
      - name: request-success-rate  # 指标名称
        thresholdRange:
          min: 99    # 请求成功率要 大于99%
        interval: 1m # 指标查询间隔
      - name: request-duration # 指标
        thresholdRange:
          max: 500   # 请求延迟要小于500ms
        interval: 1m
    webhooks:
      - name: load-test
        url: http://flagger-loadtester.test/
        metadata:
          cmd: "hey -z 1m -q 10 -c 2 http://podinfo-canary.test:9898/"
```

当podinfo 应用发布时，flagger将会将流量逐渐转移到金丝雀，同时查询prom中的成功率和延时，满足条件会将请求转移到新的版本

## canary状态

一个通过以下命令查看canary的状态

```
kubectl get canaries --all-namespaces

NAMESPACE   NAME      STATUS        WEIGHT   LASTTRANSITIONTIME
test        podinfo   Progressing   15       2020-08-07T14:05:07Z
```

## canary finalizers

因为flagger会将原始deploy设置为0，并且新建primary deployment，所以在将canary对象删除时如果直接回收primary deployment则服务真实的pod为0，显然是有问题的，我们可以通过以下设置来进行canary删除时对原始deploy的恢复操作。

```
spec:
  revertOnDeletion: true
```

## Canary 分析

金丝雀分析会定期运行，直到达到最大流量权重或到达迭代次数为止。 每次运行时，Flagger都会调用webhooks，检查指标，如果达到失败的检查阈值，则停止分析并回滚canary。 如果配置了警报，则Flagger将使用警报提供程序发布分析结果。


# flagger-contour安装

![](http://img.rocdu.top/20200807/flagger-contour-overview.png)

flagger-contour实现contour httpproxy中svc权重的动态更新，从而实现金丝雀部署

## 安装flagger-contour

- 安装contour

```
kubectl apply -f https://projectcontour.io/quickstart/contour.yaml
```

- 安装flagger-contour

```
kubectl apply -k github.com/weaveworks/flagger//kustomize/contour
```

# 使用contour实现podinfo的金丝雀部署

## 服务部署

```
# 创建命名空间
kubectl create ns test
# 创建流量探测pod
kubectl apply -k github.com/weaveworks/flagger//kustomize/tester
# 创建podinfo服务
kubectl apply -k github.com/weaveworks/flagger//kustomize/podinfo
```

默认contour会使用loadbalance 模式，可以根据情况改为clusterip/nodeport

## 应用金丝雀配置

```
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: podinfo
  namespace: test
spec:
  # deployment reference
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: podinfo
  # HPA reference
  autoscalerRef:
    apiVersion: autoscaling/v2beta1
    kind: HorizontalPodAutoscaler
    name: podinfo
  service:
    # service port
    port: 80
    # container port
    targetPort: 9898
    # Contour request timeout
    timeout: 15s
    # Contour retry policy
    retries:
      attempts: 3
      perTryTimeout: 5s
  # define the canary analysis timing and KPIs
  analysis:
    # schedule interval (default 60s)
    interval: 30s
    # max number of failed metric checks before rollback
    threshold: 5
    # max traffic percentage routed to canary
    # percentage (0-100)
    maxWeight: 50
    # canary increment step
    # percentage (0-100)
    stepWeight: 5
    # Contour Prometheus checks
    metrics:
    - name: request-success-rate
      # minimum req success rate (non 5xx responses)
      # percentage (0-100)
      thresholdRange:
        min: 99
      interval: 1m
    - name: request-duration
      # maximum req duration P99 in milliseconds
      thresholdRange:
        max: 500
      interval: 30s
    # testing
    webhooks:
    - name: acceptance-test
      type: pre-rollout
      url: http://flagger-loadtester.test/
      timeout: 30s
      metadata:
        type: bash
        cmd: "curl -sd 'test' http://podinfo-canary.test/token | grep token"
    - name: load-test
      url: http://flagger-loadtester.test/
      type: rollout
      timeout: 5s
      metadata:
        cmd: "hey -z 1m -q 10 -c 2 -host app.example.com http://envoy.projectcontour"
```

## 创建contour httpproxy对象

在上一步完成之后，flagger-contour会自动创建一个HTTPProxy资源，我们使用下面配置进行关联

```
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: podinfo-ingress
  namespace: test
spec:
  virtualhost:
    fqdn: app.example.com
  includes:
    - name: podinfo
      namespace: test
      conditions:
        - prefix: /
```

现在我们可以看到两个httpproxy资源

```
kubectl -n test get httpproxies

NAME              FQDN                STATUS
podinfo                               valid
podinfo-ingress   app.example.com     valid
```
podinfo由flagger-contour自动创建，podinfo-ingress将app.example.com域名与podinfo进行关联，
在金丝雀部署时，flagger控制器会动态修改podinfo中 podinfo-primarg和podinfo-canary service的流量权重


## 金丝雀发布

金丝雀发布的流程如下
![](http://img.rocdu.top/20200807/flagger-canary-steps.png)

我们只需要简单的更新podinfo deployment的image就可以自动实现金丝雀部署，

```
kubectl -n test set image deployment/podinfo podinfod=stefanprodan/podinfo:3.1.1
```

查看金丝雀的状态

```
kubectl -n test describe canary/podinfo

Status:
  Canary Weight:         0
  Failed Checks:         0
  Phase:                 Succeeded
Events:
 New revision detected! Scaling up podinfo.test
 Waiting for podinfo.test rollout to finish: 0 of 1 updated replicas are available
 Pre-rollout check acceptance-test passed
 Advance podinfo.test canary weight 5
 Advance podinfo.test canary weight 10
 Advance podinfo.test canary weight 15
 Advance podinfo.test canary weight 20
 Advance podinfo.test canary weight 25
 Advance podinfo.test canary weight 30
 Advance podinfo.test canary weight 35
 Advance podinfo.test canary weight 40
 Advance podinfo.test canary weight 45
 Advance podinfo.test canary weight 50
 Copying podinfo.test template spec to podinfo-primary.test
 Waiting for podinfo-primary.test rollout to finish: 1 of 2 updated replicas are available
 Routing all traffic to primary
 Promotion completed! Scaling down podinfo.test
```

## A/B testing

除加权路由外，还可将Flagger配置为根据HTTP匹配条件将流量路由到金丝雀。 在A/B测试方案中，您将使用HTTP标头或cookie来定位用户的特定细分受众群。 这对于需要会话关联的前端应用程序特别有用。

![](http://img.rocdu.top/20200807/flagger-abtest-steps.png)

以下配置将会将包含'X-Canary: insider' header的服务流量代理到新版本。

```
analysis:
  interval: 1m
  threshold: 5
  iterations: 10
  match:
  - headers:
      x-canary:
        exact: "insider"
  webhooks:
  - name: load-test
    url: http://flagger-loadtester.test/
    metadata:
      cmd: "hey -z 1m -q 5 -c 5 -H 'X-Canary: insider' -host app.example.com http://envoy.projectcontour"
```

在应用后部署新服务，我们可以看到httpproxy配置中已经添加了对应的headermatch 策略。

```
spec:
  routes:
  - conditions:
    - header:
        exact: insider
        name: x-canary
      prefix: /
    retryPolicy:
      count: 3
      perTryTimeout: 5s
    services:
    - name: podinfo-primary
      port: 80
      requestHeadersPolicy:
        set:
        - name: l5d-dst-override
          value: podinfo-primary.test.svc.cluster.local:80
      weight: 100
```

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
