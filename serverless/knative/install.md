# 在kubernetes集群上安装knative

本指南引导您使用预编译镜像安装最新版本的Knative

## 开始之前

Knative 需要 v1.10以上版本kubernetes 集群，kubectl v1.10也被需要，本指南假定您已经创建了一个kubernetes集群，您可以方便的安装Alpha软件

本指南假定您在Mac或Linux环境中使用BASH；在Windows环境中需要调整一些命令。

## 安装istio

Knative 依赖于istio,istio工作负载需要为init container开启特权模式

- 安装istio

```
kubectl apply -f https://raw.githubusercontent.com/knative/serving/v0.1.0/third_party/istio-0.8.0/istio.yaml
```

- 为default namespace 添加`istio-injection=enabled` 标签

```
kubectl label namespace default istio-injection=enabled
```

- 监控istio组件，直到所有组件显示`Running`或`Completed` `Status`:`bash kubectl get pods -n istio-system`。

```
所有组件运行和运行需要几分钟；您可以重新运行命令以查看当前状态。
```

## 安装Knative服务

+ 接下来，我们将安装Knative服务及其依赖关系。

```
kubectl apply -f https://github.com/knative/serving/releases/download/v0.1.0/release.yaml
```

+ 监控各部件，直到所有部件显示运行状态.

```
kubectl get pods -n knative-serving
```

就像ISTIO组件一样，这些组件可以运行和运行几秒钟；您可以重新运行命令以查看当前状态。

> 注意：不用重新运行命令，可以添加`--watch`到上面的命令，以实时查看组件的状态更新。使用Ctrl +C退出监视模式。

现在，您已经准备好将应用程序部署到新的Knative集群中。

## 构建应用程序

现在，你包含Knative的集群已经安装好了，你已经准备好部署一个应用程序了。

部署第一个应用程序有两个选项：
- 你可以循序渐进地[开始使用Knative应用程序部署指南](https://github.com/knative/docs/blob/master/install/getting-started-knative-app.md)。
- 您可以查看可用的[示例应用程序](https://github.com/knative/docs/blob/master/serving/samples/README.md)并部署您的选择之一。
