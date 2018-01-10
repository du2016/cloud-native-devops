Flannel 是一个可以用于 Kubernetes 的 overlay 网络提供者

> 需要先安装etcd


# 配置flannel

#### 参数配置

```
cat > /etc/sysconfig/flanneld <<EOF
# Flanneld configuration options

# etcd url location.  Point this to the server where etcd runs
FLANNEL_ETCD_ENDPOINTS="https://172.26.6.1:4001,https://172.26.6.2:4001,https://172.26.6.3:4001"

# etcd config key.  This is the configuration key that flannel queries
# For address range assignment
FLANNEL_ETCD_PREFIX="/atomic.io/network"

# Any additional options that you want to pass
#FLANNEL_OPTIONS=""
FLANNEL_OPTIONS="-etcd-cafile=/etc/kubernetes/ssl/ca.pem -etcd-certfile=/etc/kubernetes/ssl/kubernetes.pem -etcd-keyfile=/etc/kubernetes/ssl/kubernetes-key.pem"
EOF
```

#### etcd种指定网络类型

```
cat > flannel.json << EOF
{
"Network": "10.254.0.0/16",
"SubnetLen": 26,
"SubnetMin": "10.254.0.64",
"SubnetMax": "10.254.250.192",
"Backend":
  {
    "Type": "host-gw"
  }
}
EOF

etcdctl --endpoints=https://172.26.6.1:4001,https://172.26.6.2:4001,https://172.26.6.3:4001   --ca-file=/etc/kubernetes/ssl/ca.pem   --cert-file=/etc/kubernetes/ssl/kubernetes.pem   --key-file=/etc/kubernetes/ssl/kubernetes-key.pem set  /atomic.io/network/config < flannel.json
etcdctl --endpoints=https://172.26.6.1:4001,https://172.26.6.2:4001,https://172.26.6.3:4001   --ca-file=/etc/kubernetes/ssl/ca.pem   --cert-file=/etc/kubernetes/ssl/kubernetes.pem   --key-file=/etc/kubernetes/ssl/kubernetes-key.pem get  /atomic.io/network/config
```

#### 启动flannel

```
systemctl start flanneld
systemctl status flanneld
```

#### 启动docker

```
service docker start
#ifconfig 查看docker0是否启用flannel网段
#多个node可以route -n查看静态路由
```