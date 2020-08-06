# argo

## argo工作流是什么

Argo Workflows是一个开源的容器本机工作流引擎，用于在Kubernetes上协调并行作业。 
Argo Workflows通过Kubernetes CRD（自定义资源定义）实现。

- 定义工作流，其中工作流中的每个步骤都是一个容器。
- 将多步骤工作流建模为一系列任务，或者使用图形（DAG）捕获任务之间的依赖关系。
- 使用Kubernetes上的Argo Workflow，可以在短时间内轻松运行用于计算机学习或数据处理的计算密集型作业。
- 在Kubernetes上本地运行CI / CD管道而无需配置复杂的软件开发产品。

## 为什么选择Argo工作流？

- 从头开始设计容器，而没有传统VM和基于服务器的环境的开销和限制。
- 与云厂商无关，可以在任何Kubernetes集群上运行。
- 在Kubernetes上轻松编排高度并行的工作。
- Argo Workflows使一台云级超级计算机触手可及！

## 使用argo管理workflow

## 安装argo控制器

```
brew install argoproj/tap/argo
kubectl create namespace argo
kubectl apply -n argo -f https://raw.githubusercontent.com/argoproj/argo/stable/manifests/install.yaml
```

## 配置serviceaccount 以运行工作流

### 角色，角色绑定和服务帐户

为了使Argo支持artifacts，outputs，对secret的访问等功能，它需要使用Kubernetes API与Kubernetes资源进行通信。
为了与Kubernetes API通信，Argo使用ServiceAccount进行身份验证以向Kubernetes API进行身份验证。
您可以通过将a绑定到使用来指定Argo使用哪个Role（即哪些权限）ServiceAccountRoleServiceAccountRoleBinding

然后，在提交工作流时，您可以制定argo使用哪个ServiceAccount：

```
argo submit --serviceaccount <name>
```

如果ServiceAccount未提供，则Argo将使用default ServiceAccount运行它的名称空间中的from，默认情况下，该名称空间几乎总是没有足够的特权。

### 授予管理员权限

就本演示而言，我们将授予default ServiceAccountadmin特权（即，将admin Role绑定到当前命名空间的default ServiceAccount）：

请注意，这将向当前命名空间的default ServiceAccount授予管理员特权，因此您将只能在该名称空间中运行Workflows。

## 运行样本工作流程

```
argo submit --watch https://raw.githubusercontent.com/argoproj/argo/master/examples/hello-world.yaml
argo submit --watch https://raw.githubusercontent.com/argoproj/argo/master/examples/coinflip.yaml
argo submit --watch https://raw.githubusercontent.com/argoproj/argo/master/examples/loops-maps.yaml
argo list
argo get xxx-workflow-name-xxx
argo logs xxx-pod-name-xxx #from get command above
```

您也可以直接使用kubectl创建工作流。但是，Argo CLI提供了其他一些kubectl未提供的功能，
例如YAML验证，工作流可视化，参数传递，重试和重新提交，挂起和恢复等等。

```
kubectl create -f https://raw.githubusercontent.com/argoproj/argo/master/examples/hello-world.yaml
kubectl get wf
kubectl get wf hello-world-xxx
kubectl get po --selector=workflows.argoproj.io/workflow=hello-world-xxx --show-all
kubectl logs hello-world-yyy -c main
```

## 使用mino作为argo artifact存储库

安装minio

```
helm install argo-artifacts stable/minio \
  --set service.type=LoadBalancer \
  --set defaultBucket.enabled=true \
  --set defaultBucket.name=my-bucket \
  --set persistence.enabled=false \
  --set fullnameOverride=argo-artifacts
  --namespace=argo
```

暴露minioweb ui

```
minikube service --url argo-artifacts
```

编辑 workflow-controller ConfigMap以引用由Helm安装创建的服务名称
（argo-artifacts）和secret（argo-artifacts）

编辑workflow-controller ConfigMap:

```
kubectl edit cm -n argo workflow-controller-configmap
```

添加以下内容

官方文档有误 删除了`   config: |`
```
data:
    artifactRepository: |
      s3:
        bucket: my-bucket
        endpoint: argo-artifacts.default:9000
        insecure: true
        # accessKeySecret and secretKeySecret are secret selectors.
        # It references the k8s secret named 'argo-artifacts'
        # which was created during the minio helm install. The keys,
        # 'accesskey' and 'secretkey', inside that secret are where the
        # actual minio credentials are stored.
        accessKeySecret:
          name: argo-artifacts
          key: accesskey
        secretKeySecret:
          name: argo-artifacts
          key: secretkey
```


## 运行使用artifact的流程

```
argo submit https://raw.githubusercontent.com/argoproj/argo/master/examples/artifact-passing.yaml
```

## 访问Argo UI

minikube service -n argo --url argo-ui

![argo-ui](http://img.rocdu.top/20200528/argo-ui.png)


扫描关注我:

![微信](http://img.rocdu.top/20200528/qrcode_for_gh_7457c3b1bfab_258.jpg)