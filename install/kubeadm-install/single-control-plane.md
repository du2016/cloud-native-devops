# 配置 kubeadmc config

```
apiVersion: kubeadm.k8s.io/v1beta2
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token # token所属组
  token: abcdef.0123456789abcdef # 设置token
  ttl: 24h0m0s #token过期时间
  usages: # 签名信息
  - signing
  - authentication
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: 10.10.8.42 # 监听地址
  bindPort: 6443 #监听端口
nodeRegistration:
  criSocket: /var/run/dockershim.sock # cri socket
  name: 10.10.8.42 # 注册的名称
  taints: # 污点
  - effect: NoSchedule
    key: node-role.kubernetes.io/master
---
apiServer:
  timeoutForControlPlane: 4m0s
  certSANs: # 证书san
    - 127.0.0.1
    - 10.10.8.42
    - 200.0.0.1
    - kubernetes
    - kubernetes.default
    - kubernetes.default.svc
    - kubernetes.default.svc.cluster
    - kubernetes.default.svc.cluster.local
apiVersion: kubeadm.k8s.io/v1beta2
certificatesDir: /etc/kubernetes/pki # 证书目录
clusterName: kubernetes
controllerManager: {}
dns:
  type: CoreDNS
etcd:
  local:
    dataDir: /var/lib/etcd
imageRepository: registry.aliyuncs.com/google_containers # 设置使用阿里云镜像
kind: ClusterConfiguration
kubernetesVersion: v1.16.0
networking:
  dnsDomain: cluster.local
  serviceSubnet: 200.0.0.1/16 #svc cidr
  podSubnet: 10.201.0.0/16 # pod cidr
controlPlaneEndpoint: "10.10.8.200" # apiserver 负载均衡 IP 单点克不设置
scheduler: {}
```


# 安装

```
kubeadm init --config=kubeadm.conf
```

# 安装网络插件

```
kubectl apply -f https://docs.projectcalico.org/v3.8/manifests/canal.yaml
```

# 查看node状态

```
kubectl get nodes -w
```