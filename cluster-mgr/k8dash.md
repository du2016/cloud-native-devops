# K8Dash - Kubernetes Dashboard

- K8Dash是管理Kubernetes集群的最简单方法。为什么？ 
- 全面的群集管理：命名空间，节点，窗格，副本集，部署，存储，RBAC等 
- 快速且始终如一的即时更新：无需刷新页面即可查看最新信息
- 一目了然地快速可视化集群运行状况：实时图表可帮助快速跟踪性能不佳的资源
- 易于CRUD和扩展：加上内联API文档，可以轻松了解每个字段的作用
- 简单的OpenID集成：无需特殊代理
- 安装简单：使用提供的yaml资源在不到1分钟的时间内启动K8Dash并运行（不严重）

# 依赖

- 运行中的k8s集群
- 安装metric-server(可以查看历史文章)
- k8s集群为OpenId配置连接认证

# 安装

- 部署

```
# 很久没更新了高版本需要改一下deployment的版本 apps/v1,端口改为nodeport
kubectl apply -f https://raw.githubusercontent.com/herbrandson/k8dash/master/kubernetes-k8dash.yaml
```

- 确保pod和svc状态正常

```
kubectl get  -n kube-system deploy/k8dash svc/k8dash
NAME                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/k8dash   1/1     1            1           2m55s

NAME             TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
service/k8dash   NodePort   200.0.160.93   <none>        80:30354/TCP   4m17s
```

- 生成token

```
kubectl create serviceaccount k8dash -n kube-system
kubectl create clusterrolebinding k8dash --clusterrole=cluster-admin --serviceaccount=kube-system:k8dash
kubectl get secret k8dash-token-kpt25 -n kube-system -o yaml | grep 'token:' | awk '{print $2}' | base64 -d
eyJhbGciOiJSUzI1NiIsImtpZCI6ImZ6UWpVcGVfUktkc0tfU0FLOFFlRnQ4QTJGR1JwRmZZNzJFWEZCUi1xTlUifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Ims4ZGFzaC10b2tlbi1rcHQyNSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJrOGRhc2giLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJkNjgxNDBlNi0zMWE2LTRhZDgtYmRlYy1jZGMwMDI0ZTFiY2IiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDprOGRhc2gifQ.sqYyMQPWeHwbaKEp-GahWJiPWSGETGMD-12sHIS08l2dXZEsv1zr8r_mWK56u7LHAnpEKeW8HtVZ-8VMpbYAyQdYBn_rqOpa81E0Gi7JsGTKCKuHJ4UB8fx6zGS4O397Pcn9iKxtQKjEo0JhnIfhDuZUC4yl0Fren60csBpHsGbUs6uSTH1n7BFL1Xk_Slzym9hZVnrrdyWlBXnHPo8xt7GvvbL7hMKJZ23Fk9HqNejjxcEUQMliMi25-rVkh8muO-n6uYoTdupMMwTpk34d8vTgq_XfuM95elCEMc2VWjGXYrRVkViIyomIzRHn_taQ-udRraWS-9_q6khjjWOd2g
```

- 使用token访问k8dash

![](http://img.rocdu.top/20191224/k8dash-2.png)
参考：

https://github.com/herbrandson/k8dashdash