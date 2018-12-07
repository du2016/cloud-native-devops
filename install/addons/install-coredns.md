#### 安装coredns

```
sed -i 's/$DNS_DOMAIN/cluster.local/g' coredns.yaml.sed
sed -i 's/$DNS_SERVER_IP/10.254.0.2/g' coredns.yaml.sed
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


#### coredns corefile

默认coredns配置

```
# . 代表根区
.:53 {
    #插件列表
    errors 
    health 
    kubernetes cluster.local in-addr.arpa ip6.arpa { 
        pods insecure 
        upstream # cname配置 
        fallthrough in-addr.arpa ip6.arpa
    }
    prometheus :9153
    proxy . /etc/resolv.conf
    cache 30
    loop
    reload
    loadbalance
}
```

##### 多区域选择

```
多个区包含同一域名，则使用最具体的配置
example.org {
    whoami
}
org {
    whoami
}
将使用example.org配置
```

##### 插件

- whoami

```
返回解析器本地的IP
```

- [chaos](https://coredns.io/plugins/chaos/)

```
返回CH类中的dns查询，通常用于返回服务器版本和作者信息
```

- [proxy](https://coredns.io/plugins/proxy/)

```
反带功能，可以实现kube-dns中指定子域和上游dns的功能
```

- [kubernetes](https://coredns.io/plugins/kubernetes/)


```
从k8s读取区域信息
```

- [prometheus](https://coredns.io/plugins/metrics/)

```
prometheus 指标
```

- [cache](https://coredns.io/plugins/cache/)

```
缓存
```

- [loop](https://coredns.io/plugins/loop/)


```
防止自身循环查询
```

- [reload](https://coredns.io/plugins/reload/)

```
重新加载corefile
```

- [loadbalance](https://coredns.io/plugins/loadbalance/)

```
forward策略 只有一种算法：轮训
```

- [forward](https://coredns.io/plugins/forward/)

```
新版本的proxy
```