# 安装dashboard

```
cd kubernetes-repo/cluster/addons/dashboard
kubectl create -f dashboard-secret.yaml
kubectl create -f dashboard-configmap.yaml
kubectl create -f dashboard-rbac.yaml
kubectl create -f dashboard-controller.yaml
kubectl create -f dashboard-service.yaml
```