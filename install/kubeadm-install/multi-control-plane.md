## 初始化控制面板

### 创建前段代理

通过nginx+keepalived实现 vip为10.10.8.200

### 创建kubeadm配置

个人喜欢直接使用IP，不同节点请修改nodename
```
apiVersion: kubeadm.k8s.io/v1beta2
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: abcdef.0123456789abcdef
  ttl: 24h0m0s
  usages:
  - signing
  - authentication
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: 10.10.8.42
  bindPort: 6443
nodeRegistration:
  criSocket: /var/run/dockershim.sock
  name: 10.10.8.42
  taints:
  - effect: NoSchedule
    key: node-role.kubernetes.io/master
---
apiServer:
  timeoutForControlPlane: 4m0s
  certSANs:
    - 10.10.8.42
    - 10.10.8.43
    - 10.10.8.44
    - 10.10.8.200
    - 200.0.0.1
    - kubernetes
    - kubernetes.default
    - kubernetes.default.svc
    - kubernetes.default.svc.cluster
    - kubernetes.default.svc.cluster.local
apiVersion: kubeadm.k8s.io/v1beta2
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controllerManager: {}
dns:
  type: CoreDNS
etcd:
    external:
        endpoints:
        - https://10.10.8.42:2379
        - https://10.10.8.43:2379
        - https://10.10.8.44:2379
        caFile: /etc/kubernetes/pki/etcd/ca.crt
        certFile: /etc/kubernetes/pki/apiserver-etcd-client.crt
        keyFile: /etc/kubernetes/pki/apiserver-etcd-client.key
imageRepository: registry.aliyuncs.com/google_containers
kind: ClusterConfiguration
kubernetesVersion: v1.16.0
networking:
  dnsDomain: cluster.local
  serviceSubnet: 200.0.0.1/16
  podSubnet: 10.201.0.0/16
controlPlaneEndpoint: "10.10.8.200"
scheduler: {}
```

# 初始化

```
kubeadm init --config=kubeadm.conf --upload-certs
```

# 加入新节点

```
kubeadm join 10.10.8.200:6443 --token abcdef.0123456789abcdef --discovery-token-ca-cert-hash sha256:xxx --control-plane --certificate-key xxxx
```

# 如果没有添加--upload-certs

请复制/etc/kubernetes/到其他节点

