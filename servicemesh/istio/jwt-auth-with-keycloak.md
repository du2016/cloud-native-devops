# envoy rbac介绍

基于角色的访问控制(RBAC)为服务提供服务级别和方法级别的访问控制。RBAC政策是附加的。依次检查策略。根据操作以及是否找到匹配的策略，允许或拒绝请求。

策略配置主要包括两个部分。

- permissions

由AuthorizationPolicy中to转换过来

```
定义角色的权限集。 每个权限都与OR语义匹配。 为了匹配此策略的所有操作，应使用any字段设置为true的单个Permission。
```

- principals

由AuthorizationPolicy中to和when字段转换过来

```
根据操作分配/拒绝角色的主体集。 每个主体都与OR语义匹配。 为了匹配此策略的所有下游，应使用any字段设置为true的单个Principal。
```

本文将基于istio和keyclock应用envoy的rbac策略，实现基于jwt的权限控制。

# 启动keycloak

```
docker run -d  -p 8443:8443 -p 80:8080 -e PROXY_ADDRESS_FORWARDING=true  -e KEYCLOAK_USER=admin -e KEYCLOAK_PASSWORD=admin quay.io/keycloak/keycloak:11.0.0
```

# 配置keycloak

创建istioclient

![](http://img.rocdu.top/20200909/create-client.png)

创建clientrole
![](http://img.rocdu.top/20200909/role.png)

分配role
![](http://img.rocdu.top/20200909/assaign-role.png)

创建rolemapper,如果不创建信息会保存在resource_access.istio.roles，但是istio的jwt auth无法获取子路径下的信息，需要将信息映射出来

![](http://img.rocdu.top/20200909/role-mapper.png)

# 安装服务

```
kubectl create ns foo
kubectl apply -f <(istioctl kube-inject -f samples/httpbin/httpbin.yaml) -n foo
kubectl apply -f <(istioctl kube-inject -f samples/sleep/sleep.yaml) -n foo
```

# 为ingressgate配置认证策略

为服务httpbin创建Gateway

```
$ kubectl apply -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: httpbin-gateway
  namespace: foo
spec:
  selector:
    istio: ingressgateway # use Istio default gateway implementation
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
EOF
```

创建vs

```
kubectl apply -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: httpbin
  namespace: foo
spec:
  hosts:
  - "*"
  gateways:
  - httpbin-gateway
  http:
  - route:
    - destination:
        port:
          number: 8000
        host: httpbin.foo.svc.cluster.local
EOF
```

应用授权策略，只有通过认证的服务才能访问

```
kubectl apply -f - <<EOF
apiVersion: "security.istio.io/v1beta1"
kind: "RequestAuthentication"
metadata:
  name: "jwt-example"
  namespace: istio-system
spec:
  selector:
    matchLabels:
      istio: ingressgateway
  jwtRules:
  - issuer: "http://192.168.8.10/auth/realms/master"
    jwksUri: "http://192.168.8.10/auth/realms/master/protocol/openid-connect/certs"
EOF
```
## 测试访问

获取token

```
export TOKEN=`curl -s -d "audience=master" -d "client_secret=69ae93e2-4b41-4a20-a9de-1f472b0ca2a9" -d "client_id=istio" -d "grant_type=client_credentials" http://127.0.0.1:8080/auth/realms/master/protocol/openid-connect/token | jq -r ".access_token"`

curl "http://ingress-gateway-ip:8080/headers"  -H "Authorization: Bearer $TOKEN"
```

http://127.0.0.1:5556/dex/token -d "client_id=istio" -d "response_type=code token" -d "grant_type=password" -d "username=admin" -d "password=admin" -d "scope=openid"

可以正常访问

# 使用jwt对特定路径进行认证授权

应用以下策略在GET/POST时判断headers时验证客户端是否具有fuckistio角色，

```
kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: require-jwt
  namespace: foo
spec:
  selector:
    matchLabels:
      app: httpbin
  action: ALLOW
  rules:
  - from:
    - source:
        requestPrincipals: ["*"]
    to:
    - operation:
        paths: ["/headers"]
        methods: ["POST"]
    when:
    - key: request.auth.claims[roles]
      values: ["fuckistio"]
  - from:
    - source:
        notRequestPrincipals: ["*"]
    to:
    - operation:
        paths: ["/headers"]
        methods: ["GET"]
  - from: # 由于envoy中当有allow条件时，如果无法匹配默认会拒绝所以需要应用以下策略在访问非headers时不验证客户端信息
    - source:
        notRequestPrincipals: ["*"]
    to:
    - operation:
        notPaths: ["/headers"]
EOF
```

## 验证

尝试请求/headers POST method,可以访问，但是需要添加token
```
[root@centos /]# curl -XPOST "http://httpbin.foo:8000/headers"
RBAC: access denied
[root@centos /]# curl -XPOST "http://httpbin.foo:8000/headers"  -H "Authorization: Bearer $TOKEN"
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<title>405 Method Not Allowed</title>
<h1>Method Not Allowed</h1>
<p>The method is not allowed for the requested URL.</p>
```

get /headers 无需认证即可访问
```
[root@centos /]#  curl  "http://httpbin.foo:8000/headers"
{
  "headers": {
    "Accept": "*/*",
    "Content-Length": "0",
    "Host": "httpbin.foo:8000",
    "User-Agent": "curl/7.29.0",
    "X-B3-Sampled": "1",
    "X-B3-Spanid": "81d7413f45dd1e9e",
    "X-B3-Traceid": "18a755df138a124a81d7413f45dd1e9e"
  }
}
```

ip 接口无需认证即可使用访问任意方法访问
```
[root@centos /]# curl -XPOST "http://httpbin.foo:8000/ip"
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<title>405 Method Not Allowed</title>
<h1>Method Not Allowed</h1>
<p>The method is not allowed for the requested URL.</p>
[root@centos /]# curl "http://httpbin.foo:8000/ip"
{
  "origin": "127.0.0.1"
}
```

# 总结

使用keycloak结合istio可以实现细粒度的认证授权策略，客户端只需要到认证授权中心获取token,服务端无需关心任何认证授权细节，专注以业务实现，实现业务逻辑与基础设施的解耦

参考
https://www.doag.org/formes/pubfiles/11143470/2019-NN-Sebastien_Blanc-Easily_Secure_your_Microservices_with_Keycloak-Praesentation.pdf
https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/jwt_authn/v3/config.proto
https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter#config-http-filters-jwt-authn

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
