# 开始使用Knative部署应用程序

本指南向您展示如何使用Knative部署应用程序，然后使用CURL发送请求与应用程序进行交互。

## 开始之前

你需要：
- 已[部署Knative](/serverless/knative/install.md)的kubernetes集群
- 一个你想要部署程序的的容器镜像已上传到镜像仓库，本指南中使用的示例应用程序的图像可在谷歌容器镜像仓库上找到。

## 示例应用

本指南使用Hello World示例应用程序来演示部署应用程序的基本工作流程，
这些步骤可以适用于您自己的应用程序，假如如果您可以在docker hub、谷歌容器镜像仓库或其它容器镜像仓库上提供镜像。


Hello World示例应用程序从`.yaml`配置文件中读取`env`变量，然后打印"Hello World：${Task}！"，如果`TARGET`没有定义，它将打印"NOT SPECIFIED"。

## 配置你的部署

要使用Knative部署应用程序，需要配置一个定义服务的`.yaml`文件。有关服务对象的更多信息,查看
[资源类型文档](https://github.com/knative/serving/blob/master/docs/spec/overview.md#service)

此配置文件指定有关应用程序的元数据，指向部署的应用程序的宿主映像。
并允许`deployment`可配置。有关哪些配置选项可用的，更多信息查看[服务规格文档](https://github.com/knative/serving/blob/master/docs/spec/spec.md)


创建一个名为`service.yaml`的新文件，然后复制并粘贴下面的内容:

```
apiVersion: serving.knative.dev/v1alpha1 # Current version of Knative
kind: Service
metadata:
  name: helloworld-go # The name of the app
  namespace: default # The namespace the app will use
spec:
  runLatest:
    configuration:
      revisionTemplate:
        spec:
          container:
            image: gcr.io/knative-samples/helloworld-go # The URL to the image of the app
            env:
            - name: TARGET # The environment variable printed out by the sample app
              value: "Go Sample v1"
```

如果要部署示例应用程序，请将配置文件按原样保留。
如果您正在部署自己的应用程序的镜像，则相应地更新应用程序的名称和镜像的URL。

## 部署应用

在创建`service.yaml`文件的目录中，应用配置：

```
kubectl apply -f service.yaml
```

既然你的服务被创建了，那么 Knative 将执行以下步骤：

- 为这个版本的应用程序创建一个新的不可更改的版本
- 执行网络编程，为应用程序创建路由、入口、服务和负载均衡器
- 根据流量自动缩放你的pods，包括零活pods

### 与应用程序交互

要查看您的应用程序是否已成功部署，您需要由Knative创建的host URL和IP地址.

1. 要找到您的服务的IP地址，请输入`kubectl get svc knative-ingressgateway -n istio-system`。
如果您的群集是新的，则该服务可能需要一段时间来获得外部IP地址。

```
export IP_ADDRESS=$(kubectl get svc knative-ingressgateway -n istio-system -o 'jsonpath={.status.loadBalancer.ingress[0].ip}')
```

> 注意：如果使用minikube或没有外部负载均衡器的裸集群，则`EXTERNAL-IP`字段显示为`<pending>`。
你需要使用`NodeIP`和`NodePort`来代替你的应用程序。要获得应用程序的`NoDEIP`和`NodePort`，请输入以下命令：

```
export IP_ADDRESS=$(kubectl get node  -o 'jsonpath={.items[0].status.addresses[0].address}'):$(kubectl get svc knative-ingressgateway -n istio-system   -o 'jsonpath={.spec.ports[?(@.port==80)].nodePort}')
```

2. 要查找您的服务的主机URL，请输入

```
export HOST_URL=$(kubectl get services.serving.knative.dev helloworld-go  -o jsonpath='{.status.domain}')
```

如果在创建`.yaml`文件时将名称从`helloworld-go`更改为其他内容，请在上面的命令中用你输入的名称替换`helloworld-go`。

3. 现在你可以向你的应用程序请求查看结果。用您写下的`EXTERNAL-IP`替换`IP_ADDRESS`，
并用前一步返回的域替换`helloworld-go.default.example.com`。

如果部署了自己的应用程序，您可能希望自定义curl请求，以便与应用程序交互.

```
curl -H "Host: ${HOST_URL}" http://${IP_ADDRESS}
Hello World: Go Sample v1!
```

它可以花几秒钟的时间来扩展你的应用程序并返回一个响应。

## 清理

若要从群集中移除示例应用程序，请删除service记录：

```
kubectl delete -f service.yaml
```
