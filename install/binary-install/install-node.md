# node节点安装

#### docker安装

- 安装
```
yum install epel* flannel conntrack-tools docker  -y
```

- 配置

除centos外，overlay需要3.18以上内核，overlay2需要4.0以上内核，关于[overlay存储说明](https://docs.docker.com/engine/userguide/storagedriver/overlayfs-driver/)，为了更好的兼容新属性，最好升级内核。
```
cat > /etc/sysconfig/docker <<EOF
# /etc/sysconfig/docker

# Modify these options if you want to change the way the docker daemon runs
OPTIONS='--selinux-enabled --log-driver=json-file --ip-masq=false --signature-verification=false -s overlay2'
if [ -z "${DOCKER_CERT_PATH}" ]; then
    DOCKER_CERT_PATH=/etc/docker
fi

# Do not add registries in this file anymore. Use /etc/containers/registries.conf
# from the atomic-registries package.
#

# docker-latest daemon can be used by starting the docker-latest unitfile.
# To use docker-latest client, uncomment below lines
#DOCKERBINARY=/usr/bin/docker-latest
#DOCKERDBINARY=/usr/bin/dockerd-latest
#DOCKER_CONTAINERD_BINARY=/usr/bin/docker-containerd-latest
#DOCKER_CONTAINERD_SHIM_BINARY=/usr/bin/docker-containerd-shim-latest
EOF
```
- 修改镜像源
```
cat > /etc/docker/daemon.json << EOF
{
  "registry-mirrors": ["https://registry.docker-cn.com"]
}
EOF
```
#### 配置flannel

- 参数配置

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

- etcd种指定网络类型

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

- 启动flannel

```
systemctl start flanneld
systemctl status flanneld
```

- 启动docker

```
service docker start
#ifconfig 查看docker0是否启用flannel网段
#多个node可以route -n查看静态路由
```

#### 通用配置文件

```
cat > /etc/kubernetes/config <<EOF
###
# kubernetes system config
#
# The following values are used to configure various aspects of all
# kubernetes services, including
#
#   kube-apiserver.service
#   kube-controller-manager.service
#   kube-scheduler.service
#   kubelet.service
#   kube-proxy.service
# logging to stderr means we get it in the systemd journal
KUBE_LOGTOSTDERR="--logtostderr=true"

# journal message level, 0 is debug
KUBE_LOG_LEVEL="--v=0"

# Should this cluster be allowed to run privileged docker containers
KUBE_ALLOW_PRIV="--allow-privileged=true"

# How the controller-manager, scheduler, and proxy find the apiserver
KUBE_MASTER="--master=https://172.26.6.131:6443"
EOF
```

#### kubelet配置

mkdir /var/lib/kubelet

- service文件配置

```
/usr/lib/systemd/system/kubelet.service
[Unit]
Description=Kubernetes Kubelet Server
Documentation=https://github.com/GoogleCloudPlatform/kubernetes
After=docker.service
Requires=docker.service

[Service]
WorkingDirectory=/var/lib/kubelet
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/kubelet
ExecStart=/usr/local/bin/kubelet \
            $KUBE_LOGTOSTDERR \
            $KUBE_LOG_LEVEL \
            $KUBELET_API_SERVER \
            $KUBELET_ADDRESS \
            $KUBELET_PORT \
            $KUBELET_HOSTNAME \
            $KUBE_ALLOW_PRIV \
            $KUBELET_POD_INFRA_CONTAINER \
            $KUBELET_ARGS
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

- 参数配置文件

fail-swap-on参数在启用swap时候需要添加，不然就需要卸载swap

```
cat > /etc/kubernetes/kubelet <<EOF
###
## kubernetes kubelet (minion) config
#
## The address for the info server to serve on (set to 0.0.0.0 or "" for all interfaces)
KUBELET_ADDRESS="--address=0.0.0.0"
#
## The port for the info server to serve on
#KUBELET_PORT="--port=10250"
#
## You may leave this blank to use the actual hostname
KUBELET_HOSTNAME="--hostname-override=172.26.6.2"
#
## location of the api-server
#
## pod infrastructure container
#
## Add your own!
KUBELET_ARGS="--cgroup-driver=systemd --cluster-dns=10.254.0.2 --experimental-bootstrap-kubeconfig=/etc/kubernetes/bootstrap.kubeconfig --kubeconfig=/etc/kubernetes/kubeconfig --cluster-domain=cluster.local --fail-swap-on=false --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
EOF
```

- 绑定kubelet-bootstrap用户到system:node-bootstrapper角色

```
#system:node-bootstrapper集群预定义角色对于证书有相关操作权限
kubectl create clusterrolebinding kubelet-bootstrap --clusterrole=system:node-bootstrapper --user=kubelet-bootstrap
```

- 启动

```
systemctl daemon-reload
systemctl enable kubelet
systemctl start kubelet
systemctl status kubelet
```

- master节点接受请求

```
kubectl get csr | awk '/Pending/ {print $1}' | xargs kubectl certificate approve
```

```
kubectl get nodes
#想要使用集群启动docker需要下载沙箱容器镜像
docker pull gcr.io/google_containers/pause-amd64:3.0
```