# keycloak 介绍

[keycloak](https://github.com/keycloak/keycloak) 现代应用程序和服务的开源身份和访问管理

以最小的麻烦为应用程序和安全服务添加身份验证。无需处理存储用户或认证用户。开箱即用。您甚至可以获得高级功能，例如用户联合，身份代理和社交登录。 

# 以docker方式运行keycloak

和k8s交互要求必须启用https,我们使用docker启动没有配置证书，需要启动PROXY_ADDRESS_FORWARDING，然后通过NGINX配置证书，从而与apiserver交互

```
docker run -p 8080:8080 -e PROXY_ADDRESS_FORWARDING=true  -e KEYCLOAK_USER=admin -e KEYCLOAK_PASSWORD=admin quay.io/keycloak/keycloak:11.0.0
```
> 如果不开启PROXY_ADDRESS_FORWARDING，需要给keycloak配置证书，对于官方的docker镜像，需要将名为tls.crt和tls.key的文件挂载到/etc/x509/https，同时给api-server添加
> --oidc-ca-file=path/ca.pem
# 配置nginx代理keycloak

```
    server {
        listen       443 ssl;
        ssl_certificate      /tmp/rocdu-certs/fullchain.crt;
        ssl_certificate_key  /tmp/rocdu-certs/private.pem;

        ssl_session_cache    shared:SSL:1m;
        ssl_session_timeout  5m;

        ssl_ciphers  HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers  on;

        location / {
           proxy_set_header   X-Real-IP $remote_addr;
           proxy_set_header   Host      $http_host;
           proxy_set_header X-Forwarded-Proto  $scheme;
           proxy_pass http://127.0.0.1:8080;
        }
    }
```
# 配置keycloak

- 创建新的client

![](http://img.rocdu.top/20200812/1.png)

- 创建mapper

![](http://img.rocdu.top/20200812/2.png)

- 配置用户属性

![](http://img.rocdu.top/20200812/3.png)

# 配置k8s启动oidc认证


使用kubeadm安装k8s集群,kubeadmconfig配置如下

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
  advertiseAddress: 192.168.11.26 # 监听地址
  bindPort: 6443 #监听端口
nodeRegistration:
  criSocket: /var/run/dockershim.sock # cri socket
  name: 192.168.11.26 # 注册的名称
  taints: # 污点
  - effect: NoSchedule
    key: node-role.kubernetes.io/master
---
apiServer:
  timeoutForControlPlane: 4m0s
  certSANs: # 证书san
    - 127.0.0.1
    - 192.168.11.26
    - kubernetes
    - kubernetes.default
    - kubernetes.default.svc
    - kubernetes.default.svc.cluster
    - kubernetes.default.svc.cluster.local
  extraArgs:
    oidc-issuer-url: "https://keycloak.rocdu.top/auth/realms/master"
    oidc-client-id: "kubernetes"
    oidc-username-claim: "preferred_username"
    oidc-username-prefix: "-"
    oidc-groups-claim: "groups"
apiVersion: kubeadm.k8s.io/v1beta2
kubernetesVersion: v1.18.0
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
networking:
  dnsDomain: cluster.local
  serviceSubnet: 200.0.0.1/16 #svc cidr
  podSubnet: 10.201.0.0/16 # pod cidr
controlPlaneEndpoint: "192.168.11.26" # apiserver 负载均衡 IP 单点克不设置
scheduler: {}
```

# 配置kubeconfig及clusterrole

应用以下配置将cluster-admin角色赋予admin group

```
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: admin-group
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: Group
  name: admin
  apiGroup: rbac.authorization.k8s.io
```

设置oidc用户

```
kubectl config set-credentials aaa --auth-provider=oidc \
--auth-provider-arg=idp-issuer-url=https://keycloak.rocdu.top/auth/realms/master \
--auth-provider-arg=client-id=kubernetes \
--auth-provider-arg=client-secret=89aeaf34-d5c1-4e16-b025-42d58a957201 \
// --auth-provider-arg=idp-certificate-authority=/Users/dutianpeng/Desktop/ca.crt

# 切换为OIDC用户
kubectl config set-context --current --user=oidc
```

> 如果未开启PROXY_ADDRESS_FORWARDING 需要添加--auth-provider-arg=idp-certificate-authority=/Users/dutianpeng/Desktop/ca.crt 参数

对应的secret可以在对应clients的credentials中查到

![](http://img.rocdu.top/20200812/4.png)

# 通过kubelogin实现k8s oidc认证

```
brew install int128/kubelogin/kubelogin
kubelogin
```
执行上述命令后，kubelogin将打开浏览器，输入用户名密码认证成功后将显示以下信息，表明认证完成


![](http://img.rocdu.top/20200812/5.png)


此时查看kubeconfig发现oidc用户的refresh-token及id-token已经被配置

如果不适用kubelogin等工具也可以直接通过curl获取token信息
```
启用服务账户
curl -k 'https://keycloak.rocdu.top/auth/realms/master/protocol/openid-connect/token' -d "client_id=kubernetes" -d "client_secret=89aeaf34-d5c1-4e16-b025-42d58a957201" -d "response_type=code token" -d "grant_type=password" -d "username=admin" -d "password=admin" -d "scope=openid"

# 不启用服务账户
curl  http://192.168.8.10/auth/realms/master/protocol/openid-connect/token -d "client_id=istio" -d "client_secret=69ae93e2-4b41-4a20-a9de-1f472b0ca2a9"  -d "response_type=code token" -d "grant_type=password" -d "username=admin" -d "password=admin" -d "scope=openid" | jq -r ".access_token"
```


# 验证

此时执行kubectl get deploy可以看到能够正常列出信息，我们新创建一个test用于并设置groups属性为test,并赋予test组system:node角色

```
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-group
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:node
subjects:
- kind: Group
  name: test
  apiGroup: rbac.authorization.k8s.io
```

此时get deploy可以看到如下信息

```
Error from server (Forbidden): deployments.apps is forbidden: User "test" cannot list resource "deployments" in API group "apps" in the namespace "default"
```

可以看到我们已经可以通过keycloak实现k8s用户的统一认证及角色划分。

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
