[Traefik](https://traefik.io/) 是一款开源的反向代理与负载均衡工具，它监听后端的变化并自动更新服务配置。Traefik 最大的优点是能够与常见的微服务系统直接整合，可以实现自动化动态配置。目前支持 Docker、Swarm,、Mesos/Marathon、Mesos、Kubernetes、Consul、Etcd、Zookeeper、BoltDB 和 Rest API 等后端模型

![](https://docs.traefik.io/img/architecture.png)

[]traefik on k8s文档](https://docs.traefik.io/configuration/backends/kubernetes/)
- rbac 配置

```
https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-rbac.yaml

echo '
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
rules:
  - apiGroups:
      - ""
    resources:
      - services
      - endpoints
      - secrets
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: traefik-ingress-controller
subjects:
- kind: ServiceAccount
  name: traefik-ingress-controller
  namespace: kube-system' | kubectl apply -f -
```

#### deploy 安装
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-deployment.yaml

#### daemonset 安装
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-ds.yaml

#### 暴露自身

```
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/ui.yaml

apiVersion: v1
kind: Service
metadata:
  name: traefik-web-ui
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: traefik-web-ui
  namespace: kube-system
  annotations:
    kubernetes.io/ingress.class: traefik
spec:
  rules:
  - host: traefik-ui.minikube
    http:
      paths:
      - backend:
          serviceName: traefik-web-ui
          servicePort: 80
```

#### 认证

```
htpasswd -c ./auth myusername
kubectl create secret generic mysecret --from-file auth --namespace=monitoring

apiVersion: extensions/v1beta1
kind: Ingress
metadata:
 name: prometheus-dashboard
 namespace: monitoring
 annotations:
   kubernetes.io/ingress.class: traefik
   ingress.kubernetes.io/auth-type: "basic"
   ingress.kubernetes.io/auth-secret: "mysecret"
spec:
 rules:
 - host: dashboard.prometheus.example.com
   http:
     paths:
     - backend:
         serviceName: prometheus
         servicePort: 9090
```

#### traefik 支持基于名称路由基于path路由

#### 权重设置

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: wildcard-cheeses
  annotations:
    traefik.frontend.priority: "1"
spec:
  rules:
  - host: *.minikube
    http:
      paths:
      - path: /
        backend:
          serviceName: stilton
          servicePort: http

kind: Ingress
metadata:
  name: specific-cheeses
  annotations:
    traefik.frontend.priority: "2"
spec:
  rules:
  - host: specific.minikube
    http:
      paths:
      - path: /
        backend:
          serviceName: stilton
          servicePort: http
```

#### 禁止header传递
disablePassHostHeaders = true #全局
添加注释 traefik.frontend.passHostHeader: "false"