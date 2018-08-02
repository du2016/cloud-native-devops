# istio 0.8 安装

> 0.8是第一个LTS版本

## 依赖

- k8s需要开启MutatingAdmissionWebhook,ValidatingAdmissionWebhook
- k8s 1.9+

## 下载依赖文件

```bash
curl -L https://git.io/getLatestIstio | sh -
cd istio-0.8.0
export PATH=$PWD/bin:$PATH
```

## 安装istio核心组件

- 不开启mtls
kubectl apply -f install/kubernetes/istio-demo.yaml
- 开启mtls
kubectl apply -f install/kubernetes/istio-demo-auth.yaml


## 验证安装

- 查看svc
```
kubectl get svc -n istio-system
#因默认使用loadbalance，需要手动改为nodeport
kubectl edit svc -n istio-system tracing
kubectl edit svc -n istio-system istio-ingressgateway
```
- 查看pod
```
kubectl get pods -n istio-system
```

## 部署服务

- 给ns添加label

```
kubectl label namespace default istio-injection=enabled
```

> 可以通过 `kubectl get MutatingWebhookConfiguration istio-sidecar-injector -o yaml`查看修改selector规则

