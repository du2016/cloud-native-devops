# 介绍

OAM是构建云原生应用程序的规范
专注于分离开发和运营需求，Open Application Model将模块化，
可扩展和可移植的设计引入到Kubernetes等平台上，以构建和交付应用程序。


rudr是开放应用模型规范(oam)的Kubernetes实现,允许用户轻松地在任何Kubernetes集群上部署和管理应用程序，
而无需担心应用程序开发人员和运营商的问题

> Rudr目前处于Alpha状态。它可能反映了我们纳入Open App Model规范之前正在审查的API或功能

# 创建云原生应用程序并不难

![](http://127.0.0.1:8000/rudr-1.png)
用户希望专注于轻松地描述和构建应用程序，
但是使用Kubernetes直接实现这一点很复杂。
从本质上讲，容器编排平台将应用程序原语与基础结构原语密不可分。
开发人员和操作人员等不同角色必须彼此关注彼此域中的问题，以便了解底层基础结构的整体情况。
深入了解容器基础架构的要求为应用程序部署和管理引入了以下问题

- 没有针对云原生应用程序的标准定义，这使用户难以寻找更简便的现代化方法。
- 有许多工具和方法可以完成任务。一方面，这是积极的，因为它使用户可以自由选择自己的路径。
但是，对于正在寻找自以为是的方式的用户而言，这是一个机会
- 在基础设施运营商，应用程序运营商和开发人员之间很难明确区分角色。
用户接触到其域外的结构，他们必须学习这些结构才能完成日常任务

# 方法：让我们一次迈出一步

![](http://127.0.0.1:8000/rudr-2.png)

- 这使应用程序开发人员可以专注于构建OAM组件，应用程序运营商可以通过OAM应用程序配置来专注于运营功能，而基础架构运营商可以专注于Kubernetes
- 通过利用开放应用程序模型，用户现在拥有一个框架，可以在其Kubernetes集群上定义其应用程序
- 目前，Rudr将利用已定义的特征来完成任务。这样就可以自由使用用户想要的任何基础工具，同时提供着重于功能而不是技术的特征。
将来，Rudr可能会提供一组默认技术来提供特征所需的功能。

# 从头开始创建应用

在本教程中，我们将构建一个用Python编写的简单Web应用程序组件，
您可以将其用于测试。它读取一个环境变量TARGET并显示"Hello $ {TARGET}!"。
如果未指定TARGET，它将使用"world"作为TARGET

## 先决条件

- 现有的k8s集群，当前支持1.15以上版本

### 安装rudr

####  安装rudr,kubectl,helm

```
git clone https://github.com/oam-dev/rudr.git
curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/darwin/amd64/kubectl"
wget https://get.helm.sh/helm-v3.0.0-linux-amd64.tar.gz
tar xf helm-v3.0.0-linux-amd64.tar.gz
cp helm /usr/local/bin/helm
helm install rudr ./charts/rudr --wait --set image.tag=v1.0.0-alpha.1
```

#### 验证安装

```
kubectl get crds -l app.kubernetes.io/part-of=core.oam.dev
kubectl get deployment rudr
```


### 升级rudr

```
helm upgrade rudr charts/rudr
```

### 卸载rudr

```
helm delete rudr
```

这样删除将保留CRD，可以通过以下命令删除CRD

```
kubectl delete crd -l app.kubernetes.io/part-of=core.oam.dev
```

## 安装具体特性的实现

Rudr提供了多个特征，包括入口和自动缩放器。但是，它不会安装其中一些的默认实现。这是因为它们映射到可由不同控制器实现的原始Kubernetes功能。
查找符合您的特征的实现的最佳位置是Helm Hub。

### 手动缩放

手动缩放没有外部依赖性

### ingress

要成功使用ingress特性，您将需要安装Kubernetes入口控制器之一。
我们建议使用nginx-ingress。

- 首先，将稳定版本库添加到您的Helm安装中。

```
helm repo add stable https://kubernetes-charts.storage.googleapis.com/
```

- 使用Helm 3安装NGINx ingress

```
helm install nginx-ingress stable/nginx-ingress
```

> 您仍然还必须管理DNS配置。如果您也无法控制example.com的域映射，则无法将入口映射到example.com。

# 使用rudr

一旦安装了Rudr，就可以开始创建和部署应用程序。
部署应用程序的第一步是部署其组成组件。在部署组件的父应用程序之前，
该组件实际上不会运行。但是，必须先部署它，然后再部署应用程序

首先，安装示例组件:

```
$ kubectl apply -f examples/helloworld-python-component.yaml
```

该组件声明了一个用Python编写的简单Web应用程序。您可以阅读Scratch文档中的[创建组件](https://github.com/oam-dev/rudr/blob/master/docs/how-to/create_component_from_scratch.md)以了解我们如何构建它。
之后，您可以使用kubectl列出所有可用的组件：

```
$ kubectl get componentschematics
NAME              AGE
helloworld-python-v1   14s
```


您可以查看单个组件

```
$ kubectl get componentschematic helloworld-python-v1 -o yaml
apiVersion: core.oam.dev/v1alpha1
kind: ComponentSchematic
metadata:
  creationTimestamp: "2019-10-08T13:02:23Z"
  generation: 1
  name: helloworld-python-v1
  namespace: default
  resourceVersion: "1989944"
  ...
spec:
  containers:
  - env:
    - fromParam: target
      name: TARGET
# ... more YAML
```

## 查看Trait

Rudr提供了一种在安装时附加操作功能的方法。这使应用程序操作有机会在安装时提供自动缩放，缓存或入口控制等功能，而无需开发人员更改组件中的任何内容。
您还可以列出Rudr上可用的特征:

```
$ kubectl get traits
NAME            AGE
autoscaler      19m
ingress         19m
manual-scaler   19m
volume-mounter  19m
```

您可以像研究组件一样查看单个特征:

```
$ kubectl get trait ingress -o yaml
apiVersion: core.oam.dev/v1alpha1
kind: Trait
metadata:
  creationTimestamp: "2019-10-02T19:57:37Z"
  generation: 1
  name: ingress
  namespace: default
  resourceVersion: "117813"
  selfLink: /apis/core.oam.dev/v1alpha1/namespaces/default/traits/ingress
  uid: 9f82c346-c8c6-4780-9949-3ecfd47879f9
spec:
  appliesTo:
  - core.oam.dev/v1alpha1.Server
  - core.oam.dev/v1alpha1.SingletonServer
  properties:
  - description: Host name for the ingress
    name: hostname
    required: true
    type: string
  - description: Port number on the service
    name: service_port
    required: true
    type: int
  - description: Path to expose. Default is '/'
    name: path
    required: false
    type: string
```

上面描述了一种Trait，该Trait将入口附加到组件上，处理到该应用的流量路由

## 安装应用程序配置

当您准备尝试安装某些产品时，请查看examples/first-app-config.yaml，
它显示了应用了单个trait的基本应用程序配置：

```
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  name: first-app
spec:
  components:
    - componentName: helloworld-python-v1
      instanceName: first-app-helloworld-python-v1
      parameterValues:
        - name: target
          value: Rudr
        - name: port
          value: '9999'
      traits:
        - name: ingress
          parameterValues:
            - name: hostname
              value: example.com
            - name: path
              value: /
            - name: service_port
              value: 9999
```

这是一个应用程序的示例，该应用程序由单个组件组成，该组件的入口特征为example.com，服务端口为9999。

要安装此应用程序配置，请使用kubectl：

```
$ kubectl apply -f examples/first-app-config.yaml
configuration.core.oam.dev/first-app created
```

您需要等待一两分钟才能完全部署它。
在幕后，Rudr正在创建所有必要的对象。 完全部署后，您可以看到您的配置：

```
$ kubectl get configurations
NAME        AGE
first-app   4m23s
$ kubectl get configuration first-app -o yaml
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  annotations:
     ...
  creationTimestamp: "2019-10-08T12:39:07Z"
  generation: 6
  name: first-app
  namespace: default
  resourceVersion: "2020150"
  selfLink: /apis/core.oam.dev/v1alpha1/namespaces/default/applicationconfigurations/first-app
  uid: 2ea9f384-993c-42b0-803a-43a1c273d291
spec:
  components:
  - instanceName: first-app-helloworld-python-v1
    componentName: helloworld-python-v1
    parameterValues:
    - name: target
      value: Rudr
    - name: port
      value: "9999"
    traits:
    - name: ingress
      parameterValues:
      - name: hostname
        value: example.com
      - name: path
        value: /
      - name: service_port
        value: 9999
status:
  components:
    helloworld-python-v1:
      deployment/first-app-helloworld-python-v1: running
      ingress/first-app-helloworld-python-v1-trait-ingress: Created
      service/first-app-helloworld-python-v1: created
  phase: synced
```

## 访问web服务

在不同平台上，访问Web应用程序的方式可能有所不同
让我们使用端口转发通过运行以下命令来帮助我们获取应用程序URL

```
export POD_NAME=$(kubectl get pods -l "oam.dev/instance-name=first-app-helloworld-python-v1,app.kubernetes.io/name=first-app" -o jsonpath="{.items[0].metadata.name}")
echo "Visit http://127.0.0.1:9999 to use your application"
kubectl port-forward $POD_NAME 9999:9999
```

kubectl port-forward 命令将阻塞并处理您的请求。

您将获得以下输出：

```
Hello Rudr!
```

## 升级应用程序配置文件

现在，我们已经成功安装了Web应用程序并检查了结果，该应用程序运行良好。但是总有一天，操作员可能需要更改某些内容。例如：

- hostname：可能是因为与其他应用程序发生冲突，假设我们将主机名更改为oamexample.com。
- env(target): 假设我们将目标的值更改为World，这可能代表一些正常的更新情况

### 更改应用程序配置文件

因此，您可以如下更改first-app-config.yaml：

```
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  name: first-app
spec:
  components:
    - componentName: helloworld-python-v1
      instanceName: first-app-helloworld-python-v1
      parameterValues:
        - name: target
-         value: Rudr
+         value: World
        - name: port
          value: '9999'
      traits:
        - name: ingress
          parameterValues:
            - name: hostname
-             value: example.com
+             value: oamexample.com
            - name: path
              value: /
            - name: service_port
              value: 9999
```

### 应用更改的文件

再次，我们应用这个yaml：

```
$ kubectl apply -f examples/first-app-config.yaml
applicationconfiguration.core.oam.dev/first-app configured
```

### 检查更新的应用

然后先检查应用的Yaml：

```
$ kubectl get configuration first-app -o yaml
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  annotations:
    ...
  creationTimestamp: "2019-10-08T12:39:07Z"
  generation: 9
  name: first-app
  namespace: default
  resourceVersion: "2022598"
  selfLink: /apis/core.oam.dev/v1alpha1/namespaces/default/applicationconfigurations/first-app
  uid: 2ea9f384-993c-42b0-803a-43a1c273d291
spec:
  components:
  - instanceName: first-app-helloworld-python-v1
    componentName: helloworld-python-v1
    parameterValues:
    - name: target
      value: World
    - name: port
      value: "9999"
    traits:
    - name: ingress
      parameterValues:
      - name: hostname
        value: oamexample.com
      - name: path
        value: /
      - name: service_port
        value: 9999
status:
  components:
    helloworld-python-v1:
      deployment/first-app-helloworld-python-v1: running
      ingress/first-app-helloworld-python-v1-trait-ingress: Created
      service/first-app-helloworld-python-v1: created
  phase: synced
```

您可以看到字段已更改。

再次，通过运行以下命令获取应用程序URL：

```
export POD_NAME=$(kubectl get pods -l "oam.dev/instance-name=first-app-helloworld-python-v1,app.kubernetes.io/name=first-app" -o jsonpath="{.items[0].metadata.name}")
echo "Visit http://127.0.0.1:9999 to use your application"
kubectl port-forward $POD_NAME 9999:9999
```

让我们再次访问该Web应用程序并找到以下结果：

```
Hello World!
```

响应表明我们的环境更改成功。


## 更改升级后的组件

假设已经过去了几天，并且开发人员已经开发了Web应用程序的新版本

例如，我们将响应的前缀从Hello更改为Goodbye，
然后制作一个名为helloworld-python-v2的新组件。
您可以在升级组件中找到有关我们如何创建它的更多详细信息。

### 更改并应用应用程序配置文件

我们需要更改并应用配置文件以使组件升级工作。

```
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  name: first-app
spec:
  components:
-   - componentName: helloworld-python-v1
+   - componentName: helloworld-python-v2
-     instanceName: first-app-helloworld-python-v1
+     instanceName: first-app-helloworld-python-v2
      parameterValues:
        - name: target
          value: World
        - name: port
          value: '9999'
      traits:
        - name: ingress
          parameterValues:
            - name: hostname
              value: oamexample.com
            - name: path
              value: /
            - name: service_port
              value: 9999
```

应用它：

```
$ kubectl apply -f examples/first-app-config.yaml
applicationconfiguration.core.oam.dev/first-app configured
```

### 检查升级结果

您可以自己再次检查应用的yaml。您应该找到组件名称已更改。 让我们直接访问该网站：

```
$ curl oamexample.com
Goodbye World!
```

更新的Web应用程序运行良好！

现在，我们已经成功地使我们的新组件正常工作。
这可能更容易，因为开发人员只需要关心组件更新，而操作员只需要关心应用程序配置。

## 卸载应用程序

您可以使用kubectl轻松删除配置

```
$ kubectl delete configuration first-app
configuration.core.oam.dev "first-app" deleted
```

这将删除您的应用程序和所有相关资源。 

它不会删除特征和组件，它们很高兴在下一个应用程序配置中等待您的使用

```
$ kubectl get traits,components
NAME                                AGE
trait.core.oam.dev/autoscaler      31m
trait.core.oam.dev/empty           31m
trait.core.oam.dev/ingress         31m
trait.core.oam.dev/manual-scaler   31m

NAME                                             AGE
component.core.oam.dev/alpine-replicable-task   19h
component.core.oam.dev/alpine-task              19h
component.core.oam.dev/hpa-example-replicated   19h
component.core.oam.dev/nginx-replicated         19h
component.core.oam.dev/nginx-singleton          19h
```


rudr基于OAM集成了云原生应用程序所需要的ingress,scale,volume等周边的管理功能，从而更加快捷的进行定义