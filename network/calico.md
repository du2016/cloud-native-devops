[calico](https://github.com/projectcalico)是一个安全的 L3 网络和网络策略提供者。

calico使用bgp的原因：

[why bgp not ospf](https://www.projectcalico.org/why-bgp/)

有关BGP [rr的介绍](http://ccie.edufly.cn/CCIEziliao/6368.html)

# 安装方式

## 标准托管安装（ETCD存储）

- 需要提前安装etcd集群

```
# 创建calico连接etcd的secret
kubectl create secret generic calico-etcd-secrets \
--from-file=etcd-key=/etc/kubernetes/ssl/kubernetes-key.pem \
--from-file=etcd-cert=/etc/kubernetes/ssl/kubernetes.pem \
--from-file=etcd-ca=/etc/kubernetes/ssl/ca.pem

# 部署
kubectl create -f https://docs.projectcalico.org/v3.0/getting-started/kubernetes/installation/hosted/calico.yaml

# rbac
kubectl apply -f https://docs.projectcalico.org/v3.0/getting-started/kubernetes/installation/rbac.yaml
```

## kubeadm 托管部署

### 依赖

- k8s1.7+
- 没有其他cni插件
- --pod-network-cidr参数需要和calico ip pool保持一致
- --service-cidr 不能和calico ip pool重叠


```
kubectl apply -f https://docs.projectcalico.org/v3.0/getting-started/kubernetes/installation/hosted/kubeadm/1.7/calico.yaml
```

## Kubernetes 数据存储托管安装(不需要etcd)

### 依赖

- 暂时不支持ipam,推荐使用 host-local ipam与pod cidr结合使用
- 默认使用node-to-node mesh模式
- k8s1.7+
- 没有其他cni插件
- --pod-network-cidr参数需要和calico ip pool保持一致
- --service-cidr 不能和calico ip pool重叠

```
# rbac
kubectl create -f  https://docs.projectcalico.org/v3.0/getting-started/kubernetes/installation/hosted/rbac-kdd.yaml

# 部署
kubectl create -f https://docs.projectcalico.org/v3.0/getting-started/kubernetes/installation/hosted/kubernetes-datastore/calico-networking/1.7/calico.yaml
```

# 配置

## typha


```
超过50各节点推荐启用typha
修改typha_service_name "none"改为"calico-typha"。
```

## 禁用snat

```
calicoctl get ipPool -o yaml | sed 's/natOutgoing: true/natOutgoing: false/g' | calicoctl apply -f -
```

## 关闭node-to-node mesh (节点网络全互联)

```
cat << EOF
apiVersion: projectcalico.org/v3
kind: BGPConfiguration
metadata:
  name: default
spec:
  logSeverityScreen: Info
  nodeToNodeMeshEnabled: false
  asNumber: 64512
EOF | calicoctl apply -f -

calicoctl node status
```

## 创建IP Pool

```
calicoctl get ippool default-ipv4-ippool -o yaml
```


# 配置bird服务

```
yum install bird
service bird start

cat >> /etc/bird.conf < EOF
log syslog { debug, trace, info, remote, warning, error, auth, fatal, bug };
log stderr all;

# Override router ID
router id 172.26.6.1;


filter import_kernel {
if ( net != 0.0.0.0/0 ) then {
   accept;
   }
reject;
}

# Turn on global debugging of all protocols
debug protocols all;

# This pseudo-protocol watches all interface up/down events.
protocol device {
  scan time 2;    # Scan interfaces every 2 seconds
}

protocol bgp {
  description "172.26.6.2";
  local as 64512;
  neighbor 172.26.6.2 as 64512;
  multihop;
  rr client;
  graceful restart;
  import all;
  export all;
}
protocol bgp {
  description "172.26.6.3";
  local as 64512;
  neighbor 172.26.6.3 as 64512;
  multihop;
  rr client;
  graceful restart;
  import all;
  export all;
}
EOF
```


## IP-IN-IP

```
calicoctl get ippool default-ipv4-ippool -o yaml > pool.yaml
# 修改Off/Always/CrossSubnet
calicoctl apply -f pool.yaml
例：
# 所有工作负载

$ calicoctl apply -f - << EOF
apiVersion: projectcalico.org/v3
kind: IPPool
metadata:
  name: ippool-ipip-1
spec:
  cidr: 192.168.0.0/16
  ipipMode: Always
  natOutgoing: true
EOF

# CrossSubnet

$ calicoctl apply -f - << EOF
apiVersion: projectcalico.org/v3
kind: IPPool
metadata:
  name: ippool-cs-1
spec:
  cidr: 192.168.0.0/16
  ipipMode: CrossSubnet
  natOutgoing: true
EOF


#通过修改配置文件环境变量
CALICO_IPV4POOL_IPIP 参数值 Off, Always, CrossSubnet
如果您的网络结构执行源/目标地址检查，并在未识别这些地址时丢弃流量，则可能需要启用工作负载间流量的IP-in-IP封装
```

## bgp peer

查看状态

```
calicoctl node status
```

配置全局 [bgp peer(rr)](https://docs.projectcalico.org/v3.0/usage/routereflector/bird-rr-config)

```
cat << EOF | calicoctl create -f -
apiVersion: projectcalico.org/v3
kind: BGPPeer
metadata:
  name: bgppeer-global-3040
spec:
  peerIP: 172.26.6.1
  asNumber: 64567
EOF

# 删除
$ calicoctl delete bgpPeer 172.26.6.1
```

特定 BGP peer

```
$ cat << EOF | calicoctl create -f -
apiVersion: projectcalico.org/v3
kind: BGPPeer
metadata:
  name: bgppeer-node-aabbff
spec:
  peerIP: aa:bb::ff
  node: node1
  asNumber: 64514
EOF

calicoctl delete bgpPeer aa:bb::ff --scope=node --node=node1
calicoctl get bgpPeer
```