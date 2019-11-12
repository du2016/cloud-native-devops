# ingress 
ingress 是k8s集群内部的对象名称，ingress-controller是具体的运行程序，根据ingress配置来产生最终的代理规则

```
ingress 是除了 hostport  nodeport  clusterIP以及云环境专有的负载均衡器外的访问方式,官方提供了Nginx ingress controller。通常情况下，service和pod的IP可以被集群网络访问。外部访问的所有流量被丢弃或转发到别处。ingress是允许入站连接到达群集服务的规则集合.可以为外部提供可访问服务的URL，流量负载均衡，可被终止的ssl连接，以及基于配置的虚拟主机。配置服务器或负载平衡器是比想象中要难。大多数Web服务器的配置文件非常相似。虽然一些应有有一些奇怪的特点，但是我们可以用相似的逻辑去实现期望的结果。ingress体现了这一理念，ingress controller是用来处理所有共同特性的。ingress controller通过监听/ingresses接口从而更新ingress 从而达到ingress的预期，这里我们先讲述怎么搭建ingress service

api: /apis/extensions/v1beta1/ingresses
```

官方的部署方式
https://github.com/kubernetes/ingress-nginx

#### 部署Default backend server

https://github.com/kubernetes/ingress-nginx/blob/master/deploy/README.md

```
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/namespace.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/default-backend.yaml \
    | kubectl apply -f -
```

#### ingress cntroller

```
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/configmap.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/tcp-services-configmap.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/udp-services-configmap.yaml \
    | kubectl apply -f -
#rbac

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/rbac.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/with-rbac.yaml \
    | kubectl apply -f -
```


#### 验证

```
curl -H "Host:foo.bar.com"  127.0.0.1/foo  
```

#### 配置https

```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /tmp/tls.key -out /tmp/tls.crt -subj "/CN=foo.bar.com"  
kubectl create secret tls foo-secret --key /tmp/tls.key --cert /tmp/tls.crt

echo "  
apiVersion: extensions/v1beta1  
kind: Ingress  
metadata:  
  name: foo  
  namespace: default  
spec:  
  tls:  
  - hosts:  
    - foo.bar.com  
    secretName: foo-secret  
  rules:  
  - host: foo.bar.com  
    http:  
      paths:  
      - backend:  
          serviceName: echoheaders-x  
          servicePort: 80  
        path: /  
" | kubectl create -f -  
kubectl get ing 
``` 
可添加默认证书`        - --default-ssl-certificate=default/foo-secret`

#### http跳转配置
 在configmap 中ssl-redirect: "false" 关闭跳转
 configmap 通过--nginx-configmap指定
 
#### HSTS

```
关闭HSTS(强制客户端（如浏览器）使用HTTPS与服务器创建连接)
hsts=false
```


#### 指定ingress-controller

```
kubernetes.io/ingress.class: "nginx"
```

#### 查看ingress controller版本

```
POD_NAMESPACE=ingress-nginx
POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app=ingress-nginx -o jsonpath={.items[0].metadata.name})
kubectl exec -it $POD_NAME -n $POD_NAMESPACE -- /nginx-ingress-controller --version
```


#### helm 安装

```
helm install stable/nginx-ingress --name my-nginx
```

#### 自定义错误

```
实际就是通过更改官方例子中的custom-default-backend角色实现，通过X-Code 和 X-Format两个维度控制
```
#### 启用状态页

```
enable-vts-status=true
#启用 ngx_http_stub_status_module 查看/nginx_status 18080端口
```

#### 自动加密Kube-Lego（使用 Let's Encrypt证书）

```
https://github.com/jetstack/kube-lego
```

#### tcp端口暴露

```
kind: ConfigMap  
metadata:  
  name: tcp-configmap-example  
data:  
  9000: "default/example-go:8080"
```
#### udp端口暴露

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: udp-configmap-example
data:
  53: "kube-system/kube-dns:53"
```

#### 服务追踪
[服务追踪]https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/opentracing.md

#### 文档

[注释]https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/annotations.md

[参数]https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/cli-arguments.md

[主配置文件]https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/configmap.md

[服务追踪]https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/opentracing.md