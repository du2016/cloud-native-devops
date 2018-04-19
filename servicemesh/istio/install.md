# 安装文档

> 当前版本为0.7.1

## 前提条件

- k8s集群

- 启用以下admissioncontrol

```
--admission-control=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,Initializers,NamespaceExists
```

- 启用dynamic admission controller API

```
--runtime-config=rbac.authorization.k8s.io/v1beta1=true,admissionregistration.k8s.io/v1alpha1=true
```

## 准备安装包

```
curl -L https://git.io/getLatestIstio | sh -
cd istio-0.7.1
export PATH=$PWD/bin:$PATH
```

## 不启用sidecar之间的tls认证的安装方式

```
kubectl apply -f install/kubernetes/istio.yaml
```

## 启用认证的安装方式

```
kubectl apply -f install/kubernetes/istio-auth.yaml
```

## 启用自动注入

- 生成证书
```
./install/kubernetes/webhook-create-signed-cert.sh \
    --service istio-sidecar-injector \
    --namespace istio-system \
    --secret sidecar-injector-certs
```
- 添加configmap

```
kubectl apply -f install/kubernetes/istio-sidecar-injector-configmap-release.yaml
```

- 生成最终配置文件

```
cat install/kubernetes/istio-sidecar-injector.yaml | \
       ./install/kubernetes/webhook-patch-ca-bundle.sh > \
       install/kubernetes/istio-sidecar-injector-with-ca-bundle.yaml
```

- 添加自动注入配置

```
kubectl apply -f install/kubernetes/istio-sidecar-injector-with-ca-bundle.yaml
```

- 如何卸载

```
kubectl delete -f install/kubernetes/istio-sidecar-injector-with-ca-bundle.yaml
```

## 验证安装

> 以下为安装了自动注入的效果

```
# 默认 istio-ingress 使用loadbalance,需要云平台支持，可以修改为nodeport方式
kubectl get svc -n istio-system
NAME                     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                                                             AGE
istio-ingress            NodePort    10.254.83.223    <none>        80:32013/TCP,443:30784/TCP                                          1h
istio-mixer              ClusterIP   10.254.246.253   <none>        9091/TCP,15004/TCP,9093/TCP,9094/TCP,9102/TCP,9125/UDP,42422/TCP    1h
istio-pilot              ClusterIP   10.254.198.74    <none>        15003/TCP,15005/TCP,15007/TCP,15010/TCP,8080/TCP,9093/TCP,443/TCP   1h
istio-sidecar-injector   ClusterIP   10.254.113.47    <none>        443/TCP

# 查看各个组件的运行状态
kubectl get pods -n istio-system
NAME                                      READY     STATUS    RESTARTS   AGE
istio-ca-75fb7dc8d5-8674c                 1/1       Running   0          1h
istio-ingress-577d7b7fc7-hlztf            1/1       Running   0          1h
istio-mixer-859796c6bf-nv8gg              3/3       Running   0          1h
istio-pilot-65648c94fb-m2tml              2/2       Running   0          1h
istio-sidecar-injector-844b9d4f86-5bns5   1/1       Running   0          46m
```

## 测试

```
# 添加测试服务
kubectl apply -f samples/sleep/sleep.yaml

# 查看 deploy的状态
kubectl get deployment -o wide

# 查看pod状态
kubectl get pod

# 给ns添加label 这里是因为使用的MutatingWebhookConfiguration功能通过 kubectl get MutatingWebhookConfiguration istio-sidecar-injector 查看对应配置
kubectl label namespace default istio-injection=enabled

# 删除pod
kubectl delete pods -l app=sleep

# 重新查看状态，可以看到deploy中虽然只启动了一个pod，但是MutatingWebhookConfiguration注入了一个sidecarpod
kubectl get pods -l app=sleep
NAME                     READY     STATUS        RESTARTS   AGE
sleep-86f6b99f94-qqhzl   2/2       Running       0          35s
```