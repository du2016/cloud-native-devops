# 介绍

kube-dns基于skydns实现，SkyDNS2的作者创建了一个新的dns项目：coredns，采用模块化设计，拥有更高的性能，以下为社区做的简单性能对比测试：

```
Kube-DNS（缓存启用）：15k qps
CoreDNS（缓存启用）：8k qps
Kube-DNS（禁用缓存）：7k qps
CoreDNS（禁用缓存）：5k qps
```
kube-dns和coredns自由选择。

# 安装kube-dns

```
$ git clone https://github.com/kubernetes/kubernetes kubernetes-repo
```

#### 替换配置文件
```
cd kubernetes-repo/cluster/addons/dns
DNSDOMAIN=cluster.local
DNSSERVERIP=10.254.0.2
SERVICECLUSTERIPRANGE=10.254.0.0/16
sed -i "s#\$DNS_DOMAIN#$DNSDOMAIN#g" *.yaml.sed
sed -i "s#\$DNS_SERVER_IP#$DNSSERVERIP#g" *.yaml.sed
sed -i "s#\$SERVICE_CLUSTER_IP_RANGE#$SERVICECLUSTERIPRANGE#g" *.yaml.sed
```

#### 安装kube-dns

```
kubectl create -f kube-dns.yaml.sed
```
#### 配置kube-dns上游dns和子域

```
kubectl edit configmap -n kube-system kube-dns
data:
  upstreamNameservers: |
    ["10.10.3.201", "10.10.3.202"]
  stubDomains: |
    {"test.com": ["1.1.1.1"]}  
```


#### 安装coredns

```
kubectl create -f coredns.yaml.sed
```


#### 测试

```
kubectl run nginx --image=nginx && kubectl expose deploy nginx --target-port=80 --port=80
kubectl exec -it busybox-5788c675f7-8472h ping nginx
yum install *bin/dig
dig @10.254.0.2 nginx.default.svc.cluster.local
```

#### 使用coredns仓库安装

[coredns](https://github.com/coredns/deployment/tree/master/kubernetes)

```
后面可以有三个参数 svc-cidr pod-cidr domain
$ ./deploy.sh 10.3.0.0/12 172.17.0.0/16 | kubectl apply -f -
$ kubectl delete --namespace=kube-system deployment kube-dns
```