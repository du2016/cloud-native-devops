# 生成证书

[官方文档](https://kubernetes.io/docs/concepts/cluster-administration/certificates/)
# cfssl方式

#### cfssl下载地址

```
VERSION=R1.2
for i in {cfssl,cfssljson,cfssl-certinfo}
do
wget https://pkg.cfssl.org/${VERSION}/${i}_linux-amd64 -O /usr/local/bin/${i}
done
```

#### 创建CA证书

- 生成CA配置文件
```
mkdir ssl && cd ssl
cfssl print-defaults config > config.json
cfssl print-defaults csr > csr.json
cat > ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "87600h"
    },
    "profiles": {
      "kubernetes": {
        "usages": [
            "signing",
            "key encipherment",
            "server auth",
            "client auth"
        ],
        "expiry": "87600h"
      }
    }
  }
}
EOF
```
- 生成CA签名配置文件

```
cat > ca-csr.json << EOF
{
  "CN": "kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names":[{
    "C": "CN",
    "ST": "Beijing",
    "L": "BeiJing",
    "O": "k8s",
    "OU": "System"
  }]
}
EOF

```

- 生成私钥和证书

```
cfssl gencert -initca ca-csr.json | cfssljson -bare ca
```
- 创建一个用于生成API Server的密钥和证书的JSON配置文件
```
cat > kubernetes-csr.json <<EOF
{
  "CN": "kubernetes",
  "hosts": [
    "127.0.0.1",
    "<MASTER_IP>",
    "<MASTER_CLUSTER_IP>",
    "kubernetes",
    "kubernetes.default",
    "kubernetes.default.svc",
    "kubernetes.default.svc.cluster",
    "kubernetes.default.svc.cluster.local"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [{
    "C": "CN",
    "ST": "Beijing",
    "L": "BeiJing",
    "O": "k8s",
    "OU": "System",
  }]
} 
EOF
#该文件需要包含所有使用该证书的ip和域名列表，包括etcd集群、kubernetes master集群、以及apiserver 集群内部cluster ip。
```
- 生成 kubernetes 证书和私钥

```
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kubernetes-csr.json | cfssljson -bare kubernetes
```

- 创建admin证书

```
cat > admin-csr.json <<EOF

{
  "CN": "admin",
  "hosts": [],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "BeiJing",
      "L": "BeiJing",
      "O": "system:masters",
      "OU": "System"
    }
  ]
}
EOF

cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes admin-csr.json | cfssljson -bare admin
# 证书O配置为system:masters 在集群内部cluster-admin的clusterrolebinding将system:masters组和cluster-admin clusterrole绑定在一起
```
- 创建kube-proxy证书

```
cat > kube-proxy-csr.json << EOF
{
  "CN": "system:kube-proxy",
  "hosts": [],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "BeiJing",
      "L": "BeiJing",
      "O": "k8s",
      "OU": "System"
    }
  ]
}
EOF
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes  kube-proxy-csr.json | cfssljson -bare kube-proxy
# 该证书用户名为system:kube-proxy，预定义的system:node-proxier clusterrolebindings将 system:kube-proxy用户和system:node-proxier角色绑定在一起
```

- 校验证书信息
```
cfssl-certinfo -cert kubernetes.pem
openssl x509  -noout -text -in  kubernetes.pem
```

- 拷贝证书
mkdir -p /etc/kubernetes/ssl/ && cp *.pem /etc/kubernetes/ssl/