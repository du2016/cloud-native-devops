# kube-router 实战

[官网](https://www.kube-router.io/)

kube-router[官方文档](https://github.com/cloudnativelabs/kube-router/tree/master/Documentation)
中文版[文档](https://rocdu.io/2017/12/%E8%AF%91kube-router-documentation/)

#### 介绍
kube-router 
- 使用iptables实现网络策略限制. --run-router参数，可透传源IP。
- 通过bgp实现路由策略.--run-firewall 参数
- 通过lvs实现代理策略，比kube-proxy的ipvs要高效很多。 --run-service-proxy

--run-firewall, --run-router, --run-service-proxy可以有选择地只启用kube-router所需的功能

- 只提供入口防火墙：--run-firewall=true --run-service-proxy=false --run-router=false
- 仅仅替换kube-proxy: --run-service-proxy=true --run-firewall=false --run-router=false

[网络功能介绍](https://cloudnativelabs.github.io/post/2017-05-22-kube-pod-networking/)

[代理功能介绍](https://cloudnativelabs.github.io/post/2017-05-10-kube-network-service-proxy/)

[网络策功能略介绍](https://cloudnativelabs.github.io/post/2017-05-1-kube-network-policies/)

#### 查看CIDR划分

```
kubectl get nodes -o json | jq '.items[] | .spec'
```
 
# 安装kube-router

#### 依赖

- 已有k8s集群
- kube-router 能够连接apiserver
- controller-manager必要配置参数 --allocate-node-cidrs=true --cluster-cidr=10.254.0.0/16,示例：
```
/usr/local/bin/kube-controller-manager --logtostderr=true --v=0 --master=http://127.0.0.1:8080 --address=127.0.0.1 --allocate-node-cidrs=true --cluster-cidr=10.254.0.0/16 --node-cidr-mask-size=24 --cluster-name=kubernetes --use-service-account-credentials  --cluster-signing-cert-file=/etc/kubernetes/ssl/ca.pem --cluster-signing-key-file=/etc/kubernetes/ssl/ca-key.pem --service-account-private-key-file=/etc/kubernetes/ssl/ca-key.pem --root-ca-file=/etc/kubernetes/ssl/ca.pem --leader-elect=true
```
- 直接在主机运行需要有ipset命令
- 以daemonseset 运行需要开启--allow-privileged=true
- 默认情况下pod并不能访问所属的svc，想要访问需要开启发夹模式,[介绍](http://www.bubuko.com/infodetail-1994270.html)
- 需要在kube-router守护进程清单中启用hostIPC：true和hostPID：true。并且必须将主路径/var/run/docker.sock设置为volumemount.
```
hairpin_mode 网络虚拟化技术中的概念，也即交换机端口的VEPA模式。这种技术借助物理交换机解决了虚拟机间流量转发问题。很显然，这种情况下，源和目标都在一个方向，所以就是从哪里进从哪里出的模式
```

#### 这里我们启用DR模式

```
kubectl --namespace=kube-system create configmap kube-proxy  --from-file=kubeconfig.conf=/root/.kube/config
kubectl create -f https://raw.githubusercontent.com/cloudnativelabs/kube-router/master/daemonset/kubeadm-kuberouter-all-features-dsr.yaml
```

#### 创建一个应用测试kube-router

```
yum install ipvsadm traceroute -y
kubectl run nginx --image=nginx --replicas=1
```

#### 暴露服务

- svc clusterip

```
$ kubectl expose nginx --target-port=80 --port=80
$ kubectl get svc nginx -o template --template='{{.spec.clusterIP}}'
10.254.116.179
```

在每台机器上查看lvs条目
```
ipvsadm -Ln
IP Virtual Server version 1.2.1 (size=4096)
Prot LocalAddress:Port Scheduler Flags
  -> RemoteAddress:Port           Forward Weight ActiveConn InActConn
TCP  10.254.0.1:443 rr persistent 10800
  -> 172.26.6.1:6443              Masq    1      0          0
TCP  10.254.116.179:80 rr 10800
  -> 10.254.11.2:80               Masq    1      0          0
```
发现本机SVCIP代理后端真实podip，使用rr算法，通过ip addr s可以看到每添加一个服务node节点上面的kube-dummy-if网卡就会增加一个虚IP

- svc session-affinity

```
kubectl delete svc nginx
kubectl expose deploy nginx --target-port=80 --port=80 --session-affinity=ClientIP
ipvsadm -Ln
IP Virtual Server version 1.2.1 (size=4096)
Prot LocalAddress:Port Scheduler Flags
  -> RemoteAddress:Port           Forward Weight ActiveConn InActConn
TCP  10.254.0.1:443 rr persistent 10800
  -> 172.26.6.1:6443              Masq    1      0          0
TCP  10.254.191.234:80 rr persistent 10800
  -> 10.254.11.2:80               Masq    1      0          0
我们可以看到 多个persistent，既lvs里面的持久链接
```

- svc NodePort

```
kubectl delete svc nginx
kubectl expose deploy nginx --target-port=80 --port=80 --type=NodePort
ipvsadm -Ln
IP Virtual Server version 1.2.1 (size=4096)
Prot LocalAddress:Port Scheduler Flags
  -> RemoteAddress:Port           Forward Weight ActiveConn InActConn
TCP  172.26.6.3:31117 rr
  -> 10.254.11.2:80               Masq    1      0          0
TCP  10.254.0.1:443 rr persistent 10800
  -> 172.26.6.1:6443              Masq    1      0          0
TCP  10.254.102.188:80 rr
  -> 10.254.11.2:80               Masq    1      0          0
可以看到不仅有虚拟IP条目，还多了对应主机的lvs条目
```

- 更改算法

```
kubectl annotate service nginx "kube-router.io/service.scheduler=dh"
```

#### network policy

kubectl annotate ns prod "net.beta.kubernetes.io/network-policy={\"ingress\":{\"isolation\":\"DefaultDeny\"}}"
测试可以看到其他命名空间ping不通该命名空间
 
查看路由表

 ```
 ip route s
 ```
 
查看bgp新信息

 ```
#  kubectl --namespace=kube-system  exec -it  kube-router-pk7fs /bin/bash 
#  gobgp neighbor -u 172.26.6.3 #从哪些IP获得更新
Peer          AS  Up/Down State       |#Received  Accepted
172.26.6.2 64512 01:03:03 Establ      |        1         1
#  gobgp global rib -u 172.26.6.3 #global rib相当于路由表
   Network              Next Hop             AS_PATH              Age        Attrs
*> 10.254.0.0/24        172.26.6.2                                01:03:24   [{Origin: i} {LocalPref: 100}]
*> 10.254.2.0/24        172.26.6.3                                00:00:32   [{Origin: i}]
```