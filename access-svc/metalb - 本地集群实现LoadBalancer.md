# 介绍 

Kubernetes没有为裸机集群提供网络负载平衡器的实现（svc 类型为loadbalance）,Kubernetes附带的Network LB的实现都是调用各种IaaS平台（GCP，AWS，Azure等）的粘合代码。如果您未在受支持的IaaS平台（GCP，AWS，Azure等）上运行，
则LoadBalancers在创建时将无限期保持pending状态

metalb解决了这种问题，使得裸机集群也能使用svc 类型为loadbalance


# 依赖

- k8s 1.13.0+,没有其他的loadbalancer
- 集群网络配置可以与metalb共存
- 给metalb 的IP地址
- 根据模式不同可能需要支持BGP的路由器


# 安装

```
kubectl apply -f https://raw.githubusercontent.com/google/metallb/v0.8.3/manifests/metallb.yaml
```

# 配置

## 二层

二层只需要配置IP地址段即可
```
cat >> metallb < EOF
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - 10.10.8.200-10.10.8.205
EOF
```

## bgp

bgp模式需要配置以下信息
- MetalLB应该连接的路由器IP地址，
- 路由器的AS号，
- MetalLB应该使用的AS号，
- 以CIDR前缀表示的IP地址范围。

示例配置：
```
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    peers:
    - peer-address: 10.0.0.1
      peer-asn: 64501
      my-asn: 64500
    address-pools:
    - name: default
      protocol: bgp
      addresses:
      - 192.168.10.0/24
```

### 广播配置

缺省情况下，BGP模式将每个分配的IP通告给配置的对等方，而没有其他BGP属性。
对于每个svc ip对等路由器将受到一个32位掩码的路由信息，BGP localpref设置为零且没有BGP Community。

通过bgp-advertisements可以自定义广播配置，
除了可以配置localpref和Community之外还可以配置聚合路由，aggregation-length广播参数扩大32位掩码
结合多种广播配置，这使您可以创建与BGP网络其余部分互操作的广播

假设您租用/24了公共IP空间，并且已将其分配给MetalLB。默认情况下，MetalLB将每个IP通告为/32，
但您的IP提供商拒绝路由/24意外的路由。因此，您需要以某种方式/24向您的传输提供商进行广播发布，但仍然可以在内部进行每个IP路由。

配置如下：
```
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    peers:
    - peer-address: 10.0.0.1
      peer-asn: 64501
      my-asn: 64500
    address-pools:
    - name: default
      protocol: bgp
      addresses:
      - 198.51.100.0/24
      bgp-advertisements:
      - aggregation-length: 32
        localpref: 100
        communities:
        - no-advertise
      - aggregation-length: 24
    bgp-communities:
      no-advertise: 65535:65282

```

### 将对等体限制到某些节点

通过使用node-selectors 配置中的对等方属性将对等方限制为某些节点

使用主机名hostA或hostB有rack=frontend标签，但没有标签network-speed=slow使用该配置
```
peers:
- peer-address: 10.0.0.1
  peer-asn: 64501
  my-asn: 64500
  node-selectors:
  - match-labels:
      rack: frontend
    match-expressions:
    - key: network-speed
      operator: NotIn
      values: [slow]
  - match-expressions:
    - key: kubernetes.io/hostname
      operator: In
      values: [hostA, hostB]
```

### 处理buggy网络

由于错误的Smurf_attack保护， 一些旧的用户网络设备错误地阻止了以.0和结尾的IP地址。.255

如果您的用户或网络遇到此问题，则可以`avoid-buggy-ips: true在`地址池上进行设置 以将.0和标记.255 为不可用。

# 应用loadbalance


修改type: LoadBalancer查看状态

```
# kubectl get svc -n kube-system kube-state-metrics
NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
kube-state-metrics   LoadBalancer   200.0.210.165   10.10.8.203   8080:31527/TCP,8081:32312/TCP   4s
```

## 分配IP

- 分配指定IP - 指定spec.loadBalancerIP 字段
- 分配IP池 - metadata.annotations.metallb.universe.tf/address-pool
- IP地址共享  metadata.annotations.metallb.universe.tf/allow-shared-ip: "共享秘钥"，只有共享秘钥相同的svc才能共享IP



# 总结

metallb 赋予了我们使用本地网络直接访问K8S内部服务的能力而不需要依赖于nodeport或者host port,但是当前为beta，不推荐生产使用


扫描关注我:

![微信](http://q08i5y6c2.bkt.clouddn.com/qrcode_for_gh_7457c3b1bfab_258.jpg)