# K8S deployment故障可视化排查可视化指南

![](https://learnk8s.io/a/d0e338e90d78e0c46beafb168ee12e32.png)

这是一个示意图，可帮助您调试Kubernetes中的deployemnt,

![](https://learnk8s.io/a/f65ffe9f61de0f4a417f7a05306edd4c.png)

当您希望在Kubernetes中部署应用程序时，通常定义三个组件：

- 一个deployment - 这是创建名为Pods的应用程序副本的秘诀

- 一个service - 内部负载平衡器路由流量到pod

- 一个ingress - 从外部访问集群服务的网络流向的描述

以下是快速视觉回顾。

在Kubernetes中，您的应用程序通过两层负载均衡器公开：内部和外部。

![](https://learnk8s.io/a/5c4ee7aa22c0d5d7b61746cae882aedc.svg)

内部的负载均衡器称为Service，而外部的负载均衡器称为Ingress。

![](https://learnk8s.io/a/616def92e99bcaa69bb2e7bb7768a595.svg)

pod未直接部署。相反，deploymeny会在其上创建和watchPod。

![](https://learnk8s.io/a/aa7dc4e26be246133054a6603aa07a77.svg)

假设您希望部署一个简单的Hello World应用程序，则该应用程序的YAML应该类似于以下内容：

```

apiVersion: apps/v1

kind: Deployment

metadata:

  name: my-deployment

  labels:

    track: canary

spec:

  selector:

    matchLabels:

      any-name: my-app

  template:

    metadata:

      labels:

        any-name: my-app

    spec:

      containers:

      - name: cont1

        image: learnk8s/app:1.0.0

        ports:

        - containerPort: 8080

---

apiVersion: v1

kind: Service

metadata:

  name: my-service

spec:

  ports:

  - port: 80

    targetPort: 8080

  selector:

    name: app

---

apiVersion: networking.k8s.io/v1beta1

kind: Ingress

metadata:

  name: my-ingress

spec:

  rules:

  - http:

    paths:

    - backend:

        serviceName: app

        servicePort: 80

      path: /

```

定义很长，很容易忽略组件之间的相互关系

例如： 

- 什么时候应使用端口80，何时应使用端口8080？

- 您是否应该为每个服务创建一个新端口，以免它们冲突？

- 标签名称重要吗？所有的都应该一样吗？

在进行调试之前，让我们回顾一下这三个组件如何相互链接。 

让我们从Deployment和Service开始。

# 连接Deployment和Service

令人惊讶的消息是，Deployment和Service根本没有连接。

而是，该服务直接指向Pod，并完全跳过部署。

因此，您应该注意的是Pod和Service之间的相互关系。

您应该记住三件事：

- 服务选择器应至少与Pod的一个标签匹配

- 服务targetPort应与containerPortPod中容器的匹配

- 服务port可以是任何号码。多个服务可以使用同一端口，因为它们分配了不同的IP地址。

下图总结了如何连接端口：

考虑Service暴露的以下Pod。

![](https://learnk8s.io/a/244048d9100f0468f0fb2cc5d24908b9.svg)

创建Pod时，应为Pod containerPort中的每个容器定义端口。

![](https://learnk8s.io/a/ba0a5f40770add57a39f48fc1fa68ea9.svg)

创建服务时，可以定义port和targetPort。但是您应该连接哪一个容器？

![](https://learnk8s.io/a/198396281daf41d2b64bf92e39afe5d5.svg)

targetPort并且containerPort应该始终匹配

![](https://learnk8s.io/a/56f461d209901a168ba9b6aabd56f8a7.svg)

如果您的容器暴露了端口3000，则targetPort应当与该端口号匹配。

![](https://learnk8s.io/a/3e6865f66e7f366feeffaf79a636b397.svg)

如果您查看YAML，则标签和ports/ targetPort应该匹配：

```

apiVersion: apps/v1

kind: Deployment

metadata:

  name: my-deployment

  labels:

    track: canary

spec:

  selector:

    matchLabels:

      any-name: my-app

  template:

    metadata:

      labels:

        any-name: my-app

    spec:

      containers:

      - name: cont1

        image: learnk8s/app:1.0.0

        ports:

        - containerPort: 8080

---

apiVersion: v1

kind: Service

metadata:

  name: my-service

spec:

  ports:

  - port: 80

    targetPort: 8080

  selector:

    any-name: my-app

```

Deployment 头部的`track: canary`是什么？

那也应该匹配吗？

该标签属于Deployment，Service的选择器未使用它来路由流量。

换句话说，您可以安全地删除它或为其分配其他值。

那matchLabels选择器呢？

它始终必须与Pod标签匹配，并且由Deployment用来跟踪Pod。

假设您进行了正确的更改，如何测试它？

您可以使用以下命令检查Pod是否具有正确的标签：

```

kubectl get pods --show-labels

```

或者，如果您具有属于多个应用程序的Pod：

```

kubectl get pods --selector any-name=my-app --show-labels

```

`any-name=my-app`标签在哪里`any-name: my-app`。

还有问题吗？

您也可以连接到Pod！

您可以使用kubectl中的`port-forward`命令连接到服务并测试连接。

```

kubectl port-forward service/<service name> 3000:80

```

如果：

- service/<service name> 是服务的名称-在当前的YAML中是`my-service`

- 3000是您希望在计算机上打开的端口

- 80是服务在port现场暴露的端口

如果可以连接，则说明设置正确。

如果不行，则很可能是您放错了标签或端口不匹配。

# 连接Service和ingress

暴露您的应用的下一步是配置Ingress。

  

Ingress必须知道如何检索服务，然后检索Pod并将流量路由到它们。

  

Ingress按名称和公开的端口检索正确的服务。

  

在Ingress和Service中应该匹配两件事：

  

- 在Ingress中该servicePort应该匹配port的服务

- 在Ingress中该serviceName应该匹配name的服务

下图总结了如何连接端口：

您已经知道该服务公开了一个端口。

![](https://learnk8s.io/a/3d0bb4d60860ca4bf934cd8e0f6296bb.svg)

Ingress有一个名为servicePort的字段

![](https://learnk8s.io/a/c11fe067be3b340235842f9e43b021b0.svg)

Service端口和ingress servicePort应始终匹配。

![](https://learnk8s.io/a/bc44ba06abef481a6dd6db4562092e52.svg)

如果决定为服务分配端口80，则也应将servicePort更改为80。

![](https://learnk8s.io/a/f198bbe14b7a3f9faed51fa2834557df.svg)

在实践中，您应该查看以下几行：

```

apiVersion: v1

kind: Service

metadata:

  name: my-service

spec:

  ports:

  - port: 80

    targetPort: 8080

  selector:

    any-name: my-app

---

apiVersion: networking.k8s.io/v1beta1

kind: Ingress

metadata:

  name: my-ingress

spec:

  rules:

  - http:

    paths:

    - backend:

        serviceName: my-service

        servicePort: 80

      path: /

```

您如何测试Ingress的功能？

您可以使用与以前相同的策略kubectl port-forward，但是应该连接到Ingress控制器，而不是连接到服务。

首先，使用以下命令检索Ingress控制器的Pod名称：

```

kubectl get pods --all-namespaces

NAMESPACE   NAME                              READY STATUS

kube-system coredns-5644d7b6d9-jn7cq          1/1   Running

kube-system etcd-minikube                     1/1   Running

kube-system kube-apiserver-minikube           1/1   Running

kube-system kube-controller-manager-minikube  1/1   Running

kube-system kube-proxy-zvf2h                  1/1   Running

kube-system kube-scheduler-minikube           1/1   Running

kube-system nginx-ingress-controller-6fc5bcc  1/1   Running

```

确定Ingress Pod（可能在不同的命名空间中）并描述它以检索端口

```

kubectl describe pod nginx-ingress-controller-6fc5bcc \

 --namespace kube-system \

 | grep Ports

Ports:         80/TCP, 443/TCP, 18080/TCP

```

最后，连接到Pod：

```

kubectl port-forward nginx-ingress-controller-6fc5bcc 3000:80 --namespace kube-system

```

此时，每次您访问计算机上的端口3000时，请求都会转发到Ingress控制器Pod上的端口80。

如果访问http://localhost:3000，则应该找到提供网页的应用程序。

# 回顾端口

快速回顾一下哪些端口和标签应该匹配：

- 服务选择器应与Pod的标签匹配

- 服务targetPort应与containerPortPod中容器的匹配

- 服务端口可以是任何数字。多个服务可以使用同一端口，因为它们分配了不同的IP地址。

- 在servicePort该入口的应该匹配port在服务

- 服务名称应与serviceNameIngress 中的字段匹配

- 知道如何构造YAML定义只是故事的一部分。

出问题了怎么办？

Pod可能无法启动，或者正在崩溃。

# 解决Kubernetes Deployment问题的3个步骤

在深入研究异常的Deployment之前，必须有一个明确定义的Kubernetes工作方式的思维模型。

由于每个部署中都有三个组件，因此您应该从底部开始依次调试所有组件。

- 您应该确保Pods正在运行，然后

- 专注于让服务将流量路由到Pod，然后

- 检查是否正确配置了Ingress

您应该从底部开始对Deployment进行故障排除。首先，检查Pod是否已就绪并正在运行。

![](https://learnk8s.io/a/f1bcc6166f088371a58d7b0b04661908.svg)

如果Pod已就绪，则应调查服务是否可以将流量分配给Pod。

![](https://learnk8s.io/a/ef0791df2a69439059143c7f54c2a249.svg)

最后，您应该检查服务与入口之间的连接。

![](https://learnk8s.io/a/39a656b37862d1bbd310e175a8a5de47.svg)

## Pod故障排除

在大多数情况下，问题出在Pod本身。

您应该确保Pod正在运行并准备就绪。

您如何检查？

```

kubectl get pods

NAME                    READY STATUS            RESTARTS  AGE

app1                    0/1   ImagePullBackOff  0         47h

app2                    0/1   Error             0         47h

app3-76f9fcd46b-xbv4k   1/1   Running           1         47h

```

在上述会话中，最后一个Pod为Running and Ready - 但是，前两个Pod 既不是Running也不为Ready。

您如何调查出了什么问题？

有四个有用的命令可以对Pod进行故障排除：

- kubectl logs <pod name> 有助于检索Pod容器的日志

- kubectl describe pod <pod name> 检索与Pod相关的事件列表很有用

- kubectl get pod <pod name> 用于提取存储在Kubernetes中的Pod的YAML定义

- kubectl exec -ti <pod name> bash 在Pod的一个容器中运行交互式命令很有用

您应该使用哪一个？

没有一种万能的。

相反，您应该结合使用它们。

## 常见pod错误

Pod可能会出现启动和运行时错误。

启动错误包括：

- ImagePullBackoff

- ImageInspectError

- ErrImagePull

- ErrImageNeverPull

- registry不可用

- InvalidImageName

运行时错误包括：

- CrashLoopBackOff

- RunContainerError

- KillContainerError

- VerifyNonRootError

- RunInitContainerError

- CreatePodSandboxError

- ConfigPodSandboxError

- KillPodSandboxError

- SetupNetworkError

- TeardownNetworkError

有些错误比其他错误更常见。

以下是最常见的错误以及如何修复它们的列表。

### ImagePullBackOff

当Kubernetes无法检索Pod容器之一的registry时，将出现此错误。

共有三个罪魁祸首：

- image名称无效-例如，您拼错了名称，或者image不存在

- 您为image指定了不存在的标签

- 您尝试检索的image属于一个私有registry，而Kubernetes没有凭据可以访问它

前两种情况可以通过更正image名称和标记来解决。

最后，您应该将凭证添加到`secret`中的私人resistry中，并在Pod中引用它。

### CrashLoopBackOff

如果容器无法启动，则Kubernetes将CrashLoopBackOff消息显示为状态。

通常，在以下情况下容器无法启动：

- 应用程序中存在错误，导致无法启动

- 您未[正确配置容器](https://stackoverflow.com/questions/41604499/my-kubernetes-pods-keep-crashing-with-crashloopbackoff-but-i-cant-find-any-lo)

- Liveness探针失败太多次

您应该尝试从该容器中检索日志，以调查其失败的原因。

如果由于容器重新启动太快而看不到日志，则可以使用以下命令：

```

kubectl logs <pod-name> --previous

```

将打印前一个容器的错误信息

### RunContainerError

当容器无法启动时出现错误。

甚至在容器内的应用程序启动之前。

该问题通常是由于配置错误，例如：

- 挂载不存在的卷，例如ConfigMap或Secrets

- 将只读卷安装为可读写

您应该使用`kubectl describe pod <pod-name>`收集和分析错误。

### Pods处于Pending状态

当您创建Pod时，该Pod保持Pending状态。

为什么？

假设您的调度程序组件运行良好，原因如下：

- 群集没有足够的资源（例如CPU和内存）来运行Pod

- 当前的命名空间具有ResourceQuota对象，创建Pod将使命名空间超过配额

- 该Pod绑定到一个待处理的 PersistentVolumeClaim

检查`event`部分最好的办法是运行`kubectl describe`命令：

```

kubectl describe pod <pod name>

```

对于因ResourceQuotas而导致的错误，可以使用以下方法检查群集的日志：

```

kubectl get events --sort-by=.metadata.creationTimestamp

```

### Pods处于 not Ready状态

如果Pod正在运行但not Ready，则表明`readiness`探针失败。

当`readiness`探针失败时，Pod未连接到服务，并且没有流量转发到该实例。

准备就绪探针失败是特定于应用程序的错误，因此您应通过`kubectl describe`检查其中的`event`部分以识别错误。

## Service故障排除

如果您的Pod正在运行并处于就绪状态，但仍无法收到应用程序的响应，则应检查服务的配置是否正确。

服务旨在根据流量的标签将流量路由到Pod。

因此，您应该检查的第一件事是服务定位了多少个Pod。

您可以通过检查Service中的endpoint来做到这一点：

```

kubectl describe service <service-name> | grep Endpoints

```

端点是一对<ip address:port>，并且在服务以Pod为目标时，应该至少有一个。

如果"Endpoints"部分为空，则有两种解释：

- 您没有运行带有正确标签的Pod（提示：您应检查自己是否在正确的命名空间中）

您selector在服务标签上有错字

- 如果您看到端点列表，但仍然无法访问您的应用程序，则targetPort可能是您服务中的罪魁祸首。

您如何测试服务？

无论服务类型如何，您都可以使用kubectl port-forward它来连接：

```

kubectl port-forward service/<service-name> 3000:80

```

即：

- <service-name> 是服务的名称

- 3000 是您希望在计算机上打开的端口

- 80 是服务公开的端口

## 对Ingress进行故障排除

如果您已到达本节，则：

- pod正在运行并准备就绪

- 服务会将流量分配到Pod

但是您仍然看不到应用程序的响应。

这意味着最有可能Ingress配置错误。

由于正在使用的Ingress控制器是集群中的第三方组件，因此有不同的调试技术，具体取决于Ingress控制器的类型。

但是在深入研究Ingress专用工具之前，您可以检查一些简单的方法。

入口使用serviceName和servicePort连接到服务。

您应该检查这些配置是否正确。

您可以检查是否已使用以下命令正确配置了Ingress：

```

kubectl describe ingress <ingress-name>

```

如果`Backend`列为空，则配置中一定有一个错误。

如果您可以在`Backend`列中看到端点，但仍然无法访问该应用程序，则可能是以下问题：

- 您如何将Ingress暴露于公共互联网

- 您如何将群集暴露于公共互联网

您可以通过直接连接到Ingress Pod来将基础结构问题与Ingress隔离开。

首先，为您的Ingress控制器（可以位于其他名称空间中）检索Pod：

```

kubectl get pods --all-namespaces

NAMESPACE   NAME                              READY STATUS

kube-system coredns-5644d7b6d9-jn7cq          1/1   Running

kube-system etcd-minikube                     1/1   Running

kube-system kube-apiserver-minikube           1/1   Running

kube-system kube-controller-manager-minikube  1/1   Running

kube-system kube-proxy-zvf2h                  1/1   Running

kube-system kube-scheduler-minikube           1/1   Running

kube-system nginx-ingress-controller-6fc5bcc  1/1   Running

```

describe以检索端口：

```

kubectl describe pod nginx-ingress-controller-6fc5bcc --namespace kube-system \

 | grep Ports

```

最后，连接到Pod：

```

kubectl port-forward nginx-ingress-controller-6fc5bcc 3000:80 --namespace kube-system

```

此时，每次您访问计算机上的端口3000时，请求都会转发到Pod上的端口80。

现在可以用吗？

- 如果可行，则问题出在基础架构中。您应该调查流量如何路由到您的群集。

- 如果不起作用，则问题出在Ingress控制器中。您应该调试Ingress。

如果仍然无法使Ingress控制器正常工作，则应开始对其进行调试。

有许多不同版本的Ingress控制器。

热门选项包括Nginx，HAProxy，Traefik等。

您应该查阅Ingress控制器的文档以查找故障排除指南。

由于[Ingress Nginx](https://github.com/kubernetes/ingress-nginx)是最受欢迎的Ingress控制器，因此在下一部分中我们将介绍一些技巧。

### 调试Ingress Nginx

Ingress-nginx项目有一个[Kubectl](https://kubernetes.github.io/ingress-nginx/kubectl-plugin/) [官方插件](https://kubernetes.github.io/ingress-nginx/kubectl-plugin/)。

您可以kubectl ingress-nginx用来：

- 检查日志，后端，证书等。

- 连接到入口

- 检查当前配置

您应该尝试的三个命令是：

- kubectl ingress-nginx lint，它会检查 nginx.conf

- kubectl ingress-nginx backend，以检查后端（类似于kubectl describe ingress <ingress-name>）

- kubectl ingress-nginx logs，查看日志

> 请注意，您可能需要使用来为Ingress控制器指定正确的名称空间--namespace <name>。

# 摘要

如果您不知道从哪里开始，在Kubernetes中进行故障排除可能是一项艰巨的任务。

您应该始终牢记从下至上解决问题：从Pod开始，然后通过Service and Ingress向上移动堆栈。

您在本文中了解到的相同调试技术可以应用于其他对象，例如：

失败的Jobs和CronJobs

StatefulSets 和 DaemonSets

