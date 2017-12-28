# 生成证书

三种方式，实际上都是openssl

* cfssl
* easyrsa
* openssl

[官方文档](https://kubernetes.io/docs/concepts/cluster-administration/certificates/)
## cfssl方式

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
    "OU": "System"
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

## easyrsa

```
curl -L -O https://storage.googleapis.com/kubernetes-release/easy-rsa/easy-rsa.tar.gz
tar xzf easy-rsa.tar.gz
cd easy-rsa-master/easyrsa3
./easyrsa init-pki
./easyrsa --batch "--req-cn=172.26.6.1@`date +%s`" build-ca nopass
```

生成 kubernetes 证书和私钥

```
./easyrsa --subject-alt-name="IP:172.26.6.1,IP:10.254.0.1,DNS:kubernetes.default" build-server-full kubernetes-master nopass
```

签发admin证书

```
./easyrsa --dn-mode=org --req-cn=admin --req-org=system:masters --req-c= --req-st= --req-city= --req-email= --req-ou= build-client-full admin nopass
```

## openssl

- 生成ca

```
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=172.26.6.1" -days 10000 -out ca.crt
```

- kubernetes 证书和私钥

```
openssl genrsa -out server.key 2048
cat >csr.conf <<EOF
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn
    
[ dn ]
C = <country>
ST = <state>
L = <city>
O = <organization>
OU = <organization unit>
CN = 172.26.6.1
    
[ req_ext ]
subjectAltName = @alt_names
    
[ alt_names ]
DNS.1 = kubernetes
DNS.2 = kubernetes.default
DNS.3 = kubernetes.default.svc
DNS.4 = kubernetes.default.svc.cluster
DNS.5 = kubernetes.default.svc.cluster.local
IP.1 = 172.26.6.1
IP.2 = 10.254.0.1
    
[ v3_ext ]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment
extendedKeyUsage=serverAuth,clientAuth
subjectAltName=@alt_names
EOF

openssl req -new -key server.key -out server.csr -config csr.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 10000 -extensions v3_ext -extfile csr.conf
openssl x509  -noout -text -in ./server.crt
```

- admin证书

```
openssl genrsa -out admin.key 2048
openssl req -new -key admin.key -out admin.csr -subj "/O=system:masters/CN=dmin"
openssl x509 -req -set_serial $(date +%s%N) -in admin.csr -CA ca.crt -CAkey ca.key -out admin.crt -days 365 -extensions v3_req -extfile req.conf
```