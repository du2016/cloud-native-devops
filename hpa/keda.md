# 介绍
KEDA是基于Kubernetes的事件驱动自动缩放组件。
它为Kubernetes中运行的任何容器提供了事件驱动的扩容能力

KEDA可以对事件驱动的Kubernetes工作负载进行细粒度的自动缩放(包括从零到从零) 
KEDA充当Kubernetes Metrics Server，
并允许用户使用专用的Kubernetes自定义资源定义来定义自动缩放规则.
KEDA可以在云和边缘上运行，可以与Kubernetes组件(例如Horizo​​ntal Pod Autoscaler)进行本地集成，
并且没有外部依赖性。

# 安装

## 使用helm安装

1.添加helm repo

```
helm repo add kedacore https://kedacore.github.io/charts
```

2.更新helm仓库

```
helm repo update
```

3. 安装keda helmchart

```
helm install keda kedacore/keda --namespace keda
```

## 使用yaml部署

```
git clone https://github.com/kedacore/keda
cd keda
kubectl create namespace keda
kubectl apply -f deploy/crds/keda.k8s.io_scaledobjects_crd.yaml
kubectl apply -f deploy/crds/keda.k8s.io_triggerauthentications_crd.yaml
kubectl apply -f deploy/
```


# 使用rabbitmq作为事件源扩容应用服务

本示例将运行一个简单的docker容器，
它将接收来自RabbitMQ队列的消息并通过KEDA进行扩展。
接收者一次（每次实例）将收到一条消息，并睡眠1秒钟以模拟执行工作。
当添加大量队列消息时，KEDA将驱动容器根据事件源（RabbitMQ）进行扩展。