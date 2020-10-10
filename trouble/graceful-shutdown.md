![](http://img.rocdu.top/20200813/80d5b78be240372287a428b1be4cf47e.png)

在本文中,您将学习如何在Pod启动或关闭时防止断开的连接.您还将学习如何正常关闭长时间运行的任务.

![](http://img.rocdu.top/20200813/55d21503055aaf9ef8a04d5e595ed505.png)

在Kubernetes中,创建和删除Pod是最常见的任务之一.

当您执行滚动更新,扩展部署,每个新版本,每个作业和cron作业等时,都会创建Pod.

但是在驱逐后,Pods也会被删除并重新创建-例如,当您将节点标记为不可调度时.

如果这些Pod的性质如此短暂,那么当Pod在响应请求时却被告知关闭时会发生什么呢？

请求在关闭之前是否已完成？

接下来的请求又如何呢？

在讨论删除Pod时会发生什么之前,有必要讨论一下创建Pod时会发生什么.

假设您要在集群中创建以下Pod:


```
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
    - name: web
      image: nginx
      ports:
        - name: web
          containerPort: 80
```

您可以使用以下方式将YAML定义提交给集群:

```
kubectl apply -f pod.yaml
```

输入命令后,kubectl便将Pod定义提交给Kubernetes API.

这是旅程的起点.

# 在数据库中保存集群状态

API接收并检查Pod定义,然后将其存储在数据库etcd中.

Pod也将添加到调度程序的队列中.

调度程序:
- 检查定义
- 收集有关工作负载的详细信息,例如CPU和内存请求,然后
- 确定哪个节点最适合运行它(通过称为过滤器和谓词的过程).

在过程结束时:

- 在etcd中将Pod标记为Scheduled.
- 为Pod分配了一个节点.
- Pod的状态存储在etcd中.

但是Pod仍然不存在.

当您使用kubectl apply -fYAML 提交Pod时,会将其发送到Kubernetes API.
![](http://img.rocdu.top/20200813/893bc3afd34f208bcbb4292e7b604d46.png)

API将Pod保存在数据库-etcd中.
![](http://img.rocdu.top/20200813/9a125d7981e0a171a966d329b726d314.png)

调度程序为该Pod分配最佳节点,并且Pod的状态更改为Pending.Pod仅存在于etcd中.
![](http://img.rocdu.top/20200813/2ad73046afa1f93533f3e2f0bbf997c5.png)

# Kubelet-Kubernetes agent

`kubelet的工作是轮询控制平面以获取更新.`

您可以想象kubelet不断地向主节点询问:`我关注工作节点1,是否对我有任何新的Pod？`.

当有Pod时,kubelet会创建它.

kubelet不会自行创建Pod.而是将工作委托给其他三个组件:

- 容器运行时接口(CRI) -为Pod创建容器的组件.
- 容器网络接口(CNI) -将容器连接到群集网络并分配IP地址的组件.
- 容器存储接口(CSI) -在容器中装载卷的组件.

在大多数情况下,容器运行时接口(CRI)的工作类似于:

```
docker run -d <my-container-image>
```

容器网络接口(CNI)有点有趣,因为它负责:

- 为Pod生成有效的IP地址.
- 将容器连接到网络的其余部分.

可以想象,有几种方法可以将容器连接到网络并分配有效的IP地址(您可以在IPv4或IPv6之间进行选择,也可以分配多个IP地址).

例如,Docker创建虚拟以太网对并将其连接到网桥,而AWS-CNI将Pods直接连接到虚拟私有云(VPC)的其余部分.

当容器网络接口完成其工作时,Pod已连接到网络的其余部分,并分配了有效的IP地址.

只有一个问题.

`Kubelet知道IP地址(因为它调用了容器网络接口),但是控制平面却不知道.`

没有人告诉主节点,该Pod已分配了IP地址,并准备接收流量.

就控制平面而言,仍在创建Pod.


`kubelet的工作是收集Pod的所有详细信息(例如IP地址)并将其报告回控制平面.`


您可以想象检查etcd不仅可以显示Pod的运行位置,还可以显示其IP地址.


Kubelet轮询控制平面以获取更新.

![](http://img.rocdu.top/20200813/c4fe175187b09a6c73548583b707987f.png)

将新的Pod分配给其节点后,kubelet将检索详细信息
![](http://img.rocdu.top/20200813/631d60eb75372856349ffadccd034bc9.png)

kubelet不会自行创建Pod.它依赖于三个组件:容器运行时接口,容器网络接口和容器存储接口.
![](http://img.rocdu.top/20200813/b91daa3399ef822a51a63ea13230f402.png)

一旦所有三个组件都成功完成,该Pod便在您的节点中运行,并分配了IP地址.
![](http://img.rocdu.top/20200813/80fe8245ca1e8387335964dff537d3f3.png)

kubelet将IP地址报告回控制平面
![](http://img.rocdu.top/20200813/077d7fd0deb0949f81b028c9bcde48a5.png)

如果Pod不是任何服务的一部分,那么这就是旅程的终点​​.

Pod已创建并可以使用.

如果Pod是服务的一部分,则还需要执行几个步骤.

# pod和service

创建service时,通常需要注意两点信息:

- 选择器,用于指定将接收流量的Pod.
- 本targetPort-通过舱体使用的端口接收的流量.

服务的典型YAML定义如下所示:

```
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  ports:
  - port: 80
    targetPort: 3000
  selector:
    name: app
```

当使用kubectl apply将Service提交给集群时,Kubernetes会找到所有具有与选择器(name: app)相同标签的Pod,并收集其IP地址-但前提是它们已通过Readiness探针.

然后,对于每个IP地址,它将IP地址和端口连接在一起.

如果IP地址是10.0.0.3和,targetPort则3000Kubernetes将两个值连接起来并称为endpoint.

```
IP address + port = endpoint
---------------------------------
10.0.0.3   + 3000 = 10.0.0.3:3000
```

endpoint存储在etcd的另一个名为Endpoint的对象中.


困惑？

Kubernetes 参考:

- endpoint(在本文和Learnk8s资料中,这称为小写e endpoint)是IP地址+端口对(10.0.0.3:3000).
- endpoint(在本文和Learnk8s材料中,被称为大写E endpoint)是endpoint的集合.

endpoint对象是Kubernetes中的真实对象,对于每个服务Kubernetes都会自动创建一个endpoint对象.

您可以使用以下方法进行验证:

```
kubectl get services,endpoints
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)
service/my-service-1   ClusterIP   10.105.17.65   <none>        80/TCP
service/my-service-2   ClusterIP   10.96.0.1      <none>        443/TCP

NAME                     ENDPOINTS
endpoints/my-service-1   172.17.0.6:80,172.17.0.7:80
endpoints/my-service-2   192.168.99.100:8443
```


endpoint从Pod收集所有IP地址和端口.

但不仅仅是一次.

在以下情况下,将使用新的endpoint列表刷新Endpoint对象:

- 创建一个Pod.
- Pod已删除.
- 在Pod上修改了标签.

因此,您可以想象,每次创建Pod并在kubelet将其IP地址发布到主节点后,Kubernetes都会更新所有endpoint以反映更改:

```
kubectl get services,endpoints
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)
service/my-service-1   ClusterIP   10.105.17.65   <none>        80/TCP
service/my-service-2   ClusterIP   10.96.0.1      <none>        443/TCP

NAME                     ENDPOINTS
endpoints/my-service-1   172.17.0.6:80,172.17.0.7:80,172.17.0.8:80
endpoints/my-service-2   192.168.99.100:8443
```

很好,endpoint存储在控制平面中,并且endpoint对象已更新.


在此图中,集群中部署了一个Pod.Pod属于服务.如果您要检查etcd,则可以找到Pod的详细信息以及服务.
![](http://img.rocdu.top/20200813/97d10d05baaf6d7bc3d0e8689b8a5be7.png)
![](http://img.rocdu.top/20200813/97d10d05baaf6d7bc3d0e8689b8a5be7.png)

部署新的Pod会怎样？
![](http://img.rocdu.top/20200813/f1cef889d7aa6bcec038ece4a1885fb6.png)

Kubernetes必须跟踪Pod及其IP地址.服务应将流量路由到新endpoint,因此应传播IP地址和端口.
![](http://img.rocdu.top/20200813/0f31b21a662f18e3849952d2201d3f83.png)

当另一个pod部署时会发生什么情况？
![](http://img.rocdu.top/20200813/63cd7aeac7a44b8ac5ebaf7771564a2e.png)

完全相同的过程.在数据库中为Pod创建新的`row`,并传播endpoint.
![](http://img.rocdu.top/20200813/d98ed5e6463333b72e9dc579450d8465.png)

但是,删除Pod会发生什么？
![](http://img.rocdu.top/20200813/397ef6c048e31a6b6ecc72117b89443c.png)

该服务会立即删除endpoint,最终,Pod也将从数据库中删除.
![](http://img.rocdu.top/20200813/8744537c6af4d50c764226bf37e7994b.png)

Kubernetes对集群中的每一个小变化都会做出反应.
![](http://img.rocdu.top/20200813/c1db737b604ce41c103fcdf50954e05c.png)

您准备好开始使用Pod了吗？

# 在Kubernetes中使用endpoint

endpoint由Kubernetes中的多个组件使用.

Kube-proxy使用endpoint在节点上设置iptables规则.

因此,每次对endpoint(对象)进行更改时,kube-proxy都会检索IP地址和端口的新列表,并编写新的iptables规则.


让我们考虑具有两个Pod且不包含Service的三节点群集.Pod的状态存储在etcd中.
![](http://img.rocdu.top/20200813/2738f23286b5d1c126c35260e18e012f.png)

创建服务时会发生什么？
![](http://img.rocdu.top/20200813/db75e32821da5102751f62ffe921d9c3.png)


Kubernetes创建了一个Endpoint对象,并从Pod收集了所有endpoint(IP地址和端口对).
![](http://img.rocdu.top/20200813/f9ebfc9b8eb6baad2025a658ce27cd9b.png)

Kube-proxy守护程序已订阅对endpoint的更改.
![](http://img.rocdu.top/20200813/195b367931c3aa09a67eeb2409f88b1c.png)

添加,删除或更新endpoint时,kube-proxy会检索新的endpoint列表.
![](http://img.rocdu.top/20200813/a3511e3fe706155e43f371c6232869d1.png)


Kube-proxy使用endpoint在群集的每个节点上创建iptables规则.
![](http://img.rocdu.top/20200813/6653390e884d68dc2293e46aeb12d905.png)

Ingress控制器使用相同的endpoint列表.

Ingress控制器是群集中将外部流量路由到群集中的那个组件.

设置Ingress清单时,通常将Service指定为destination:

```
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

实际上,流量不会路由到服务.

取而代之的是,Ingress控制器设置了一个订阅,每次该服务的endpoint更改时都将收到通知.

`Ingress会将流量直接路由到Pod,从而跳过服务.`


可以想象,每次对endpoint(对象)进行更改时,Ingress都会检索IP地址和端口的新列表,并将控制器重新配置为包括新的Pod.

在这张照片中,有一个Ingress控制器,它带有两个副本和一个Service的Deployment.
![](http://img.rocdu.top/20200813/6b6e3675d3bc9a79aaca6a9499888569.png)


如果您要通过Ingress将外部流量路由到Pod,则应创建一个Ingress清单(YAML文件).
![](http://img.rocdu.top/20200813/f5a5659d33dbbafd674185e850feb168.png)

一旦您执行`kubectl apply -f ingress.yaml`,Ingress控制器就会从控制平面检索文件.
![](http://img.rocdu.top/20200813/1a5270ad70fea564a2735748739b8de3.png)


Ingress YAML具有serviceName描述其应使用的服务的属性.
![](http://img.rocdu.top/20200813/4080d2b2e23bf222ac56d48ae4c2cb19.png)


Ingress控制器从服务中检索endpoint列表,并跳过它.流量直接流向endpoint(Pods)
![](http://img.rocdu.top/20200813/ebba24ed3e2cb2c39a74a123c2fab16a.png)

创建新的Pod会怎样？
![](http://img.rocdu.top/20200813/fbc368817a83bde0ee51fb1236156512.png)

您已经知道Kubernetes如何创建Pod并传播endpoint.
![](http://img.rocdu.top/20200813/86a8e45fd552946c2a968c653bde1ee5.png)

入口控制器正在订阅对endpoint的更改.由于存在传入更改,因此它将检索新的endpoint列表.
![](http://img.rocdu.top/20200813/cfd09b9a7eb6f4180b7b43923b3a682a.png)

入口控制器将流量路由到新的Pod.
![](http://img.rocdu.top/20200813/00d34c8164e9eb46eecd7ee54a9f16f7.png)

还有更多的Kubernetes组件示例可以订阅对endpoint的更改.

集群中的DNS组件CoreDNS是另一个示例.

如果您使用Headless类型的服务,则每次添加或删除endpoint时,CoreDNS都必须订阅对endpoint的更改并重新配置自身.

相同的endpoint被Istio或Linkerd之类的服务网格所使用,云提供商也创建了type:LoadBalancer无数运营商的服务.

您必须记住,有几个组件订阅了对endpoint的更改,它们可能会在不同时间收到有关endpoint更新的通知.

够了吗,还是在创建Pod之后有什么事发生？


`这次您完成了！`


快速回顾一下创建Pod时发生的情况:

- Pod存储在etcd中.
- 调度程序分配一个节点.它将节点写入etcd.
- 向kubelet通知新的和预定的Pod.
- kubelet将创建容器的委托委派给容器运行时接口(CRI).
- kubelet代表将容器附加到容器网络接口(CNI).
- Kubelet将容器中的安装卷委托给容器存储接口(CSI).
- 容器网络接口分配IP地址.
- Kubelet将IP地址报告给控制平面.
- IP地址存储在etcd中.

如果您的Pod属于服务:

- Kubelet等待成功的Readiness探针.
- 通知所有相关的endpoint(对象)更改.
- endpoint将新endpoint(IP地址+端口对)添加到其列表中.
- 通知Kube-proxyendpoint更改.Kube-proxy更新每个节点上的iptables规则.
- 通知endpoint变化的入口控制器.控制器将流量路由到新的IP地址.
- CoreDNS通知endpoint更改.如果服务的类型为Headless,则更新DNS条目.
- 向云提供商通知endpoint更改.如果服务为type: LoadBalancer,则将新endpoint配置为负载平衡器池的一部分.
- endpoint更改将通知群集中安装的所有服务网格.
- 订阅endpoint更改的任何其他操作员也会收到通知.

如此长的列表令人惊讶地是一项常见任务-创建一个Pod.

Pod正在运行.现在是时候讨论删除它时会发生什么.

# 删除POD

您可能已经猜到了,但是删除Pod时,必须遵循相同的步骤,但要相反.

首先,应从endpoint(对象)中删除endpoint.

这次,“就绪”探针将被忽略,并且将endpoint立即从控制平面移除.

依次触发所有事件到kube-proxy,Ingress控制器,DNS,服务网格等.

这些组件将更新其内部状态,并停止将流量路由到IP地址.

由于组件可能忙于执行其他操作,因此无法保证从其内部状态中删除IP地址需要花费多长时间.

对于某些来说,可能不到一秒钟.对于其他,可能需要更多时间.

如果您要使用删除Pod kubectl delete pod,则该命令首先会到达Kubernetes API.
![](http://img.rocdu.top/20200813/52bfaf15e838a12cec28ed656479d515.png)


该消息被控制平面中的特定控制器(endpoint控制器)截获
![](http://img.rocdu.top/20200813/e7e784358791597480b1e90efa5a180c.png)

endpoint控制器向API发出命令,以从endpoint对象中删除IP地址和端口.
![](http://img.rocdu.top/20200813/ee1385cf9cdf974d741733be04f7d5ab.png)

谁在听endpoint更改？更改将通知Kube-proxy,Ingress控制器,CoreDNS等.
![](http://img.rocdu.top/20200813/ebc4293a1a6713162af536f40ecfac6b.png)

诸如kube-proxy之类的一些组件可能需要一些额外的时间才能进一步传播更改.
![](http://img.rocdu.top/20200813/d48e4005dc26411e5aaf9fd58dbf0f00.png)

同时,etcd中Pod的状态更改为Termination.

将通知kubelet更改并委托:

- 将任何卷从容器卸载到容器存储接口(CSI).
- 从网络上分离容器并将IP地址释放到容器网络接口(CNI).
- 将容器销毁到容器运行时接口(CRI).

换句话说,Kubernetes遵循与Pod完全相同的步骤来创建Pod.


如果您要使用删除Pod kubectl delete pod,则该命令首先会到达Kubernetes API.
![](http://img.rocdu.top/20200813/6b4c471a91b57589260f8be6eccf5c3d.png)

当kubelet轮询控制平面以获取更新时,它会注意到Pod已删除.
![](http://img.rocdu.top/20200813/88ff9f8f44b0530226430da3ca8c892e.png)

kubelet代表将Pod销毁到容器运行时接口,容器网络接口和容器存储接口.
![](http://img.rocdu.top/20200813/1.png)


但是,存在细微但必不可少的差异.

当您终止Pod时,将同时删除endpoint和发送到kubelet的信号.

首次创建Pod时,Kubernetes等待kubelet报告IP地址,然后启动endpoint传播.

但是,当您删除Pod时,事件将并行开始.

这可能会导致很多比赛情况.

如果在传播endpoint之前删除Pod怎么办？

删除endpoint和删除Pod会同时发生.

![](http://img.rocdu.top/20200813/4e92253a939a7e8bb97c6eb37f7a3d7e.png)

因此,您可能最终会在kube-proxy更新iptables规则之前删除endpoint.
![](http://img.rocdu.top/20200813/900616530908d5f4d5e883cc5812404e.png)

或者,您可能会更幸运,并且只有在endpoint完全传播之后才能删除Pod.
![](http://img.rocdu.top/20200813/72fe21f86a36f4d3cafa2c9827fdcc64.png)

# 优雅停机

当Pod在终结点从kube-proxy或Ingress控制器中删除之前终止时,您可能会遇到停机时间.

而且,如果您考虑一下,这是有道理的.

Kubernetes仍将流量路由到IP地址,但Pod不再存在.

Ingress控制器,kube-proxy,CoreDNS等没有足够的时间从其内部状态中删除IP地址.

理想情况下,在删除Pod之前,Kubernetes应该等待集群中的所有组件具有更新的endpoint列表.

但是Kubernetes不能那样工作.

Kubernetes提供了强大的原语来分发endpoint(即Endpoint对象和更高级的抽象,例如Endpoint Slices).

但是,Kubernetes不会验证订阅endpoint更改的组件是否是集群状态的最新信息.

那么,如何避免这种竞争情况并确保在传播endpoint之后删除Pod？


你应该等一下

当Pod即将被删除时,它会收到SIGTERM信号.

您的应用程序可以捕获该信号并开始关闭.

由于endpoint不太可能立即从Kubernetes中的所有组件中删除,因此您可以:

- 请稍等片刻,然后退出.
- 尽管有SIGTERM,仍然可以处理传入流量.
- 最后,关闭现有的长期连接(也许是数据库连接或WebSocket).
- 关闭该过程.

你应该等多久？

默认情况下,Kubernetes将发送SIGTERM信号并等待30秒,然后强制终止该进程.

因此,您可以在最初的15秒内继续操作,因为什么都没有发生.

希望该间隔应足以将endpoint删除传播到kube-proxy,Ingress控制器,CoreDNS等.

因此,越来越少的流量将到达您的Pod,直到停止为止.

15秒后,可以安全地关闭与数据库的连接(或任何持久连接)并终止该过程.

如果您认为需要更多时间,则可以在20或25秒时停止该过程.

但是,您应该记住,Kubernetes将在30秒后强行终止该进程(除非您更改terminationGracePeriodSecondsPod定义中的).

如果您无法更改代码以等待更长的时间怎么办？

您可以调用脚本以等待固定的时间,然后退出应用程序.

在调用SIGTERM之前,Kubernetes preStop在Pod中公开一个钩子.

您可以将preStop钩子设置为等待15秒.

让我们看一个例子:

```
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
    - name: web
      image: nginx
      ports:
        - name: web
          containerPort: 80
      lifecycle:
        preStop:
          exec:
            command: ["sleep", "15"]
```

该preStop hook是Pod LifeCycle hook之一.

建议延迟15秒吗？

这要视情况而定,但这可能是开始测试的明智方法.

以下是您可以选择的选项的概述:


您已经知道,当删除Pod时,将通知kubelet更改.
![](http://img.rocdu.top/20200813/2ddb2f96d586a7eb508f652728535922.png)

如果Pod具有preStop hook,则会首先调用它.
![](http://img.rocdu.top/20200813/50e256378e2a73a86a43eae7c1cf143d.png)

当preStop完成时,kubelet发送SIGTERM信号到容器上.从那时起,容器应关闭所有长期连接并准备终止.
![](http://img.rocdu.top/20200813/c537202f15000748135f98fdfcd6c31f.png)

默认情况下,该过程将有30秒退出,其中包括该preStop挂钩.如果到那时还没有退出该进程,则kubelet发送SIGKILL信号并强制终止该进程.
![](http://img.rocdu.top/20200813/f7985c9e68bd8cf65dca13edefd96f81.png)

Kubelet通知控制平面Pod已成功删除.
![](http://img.rocdu.top/20200813/2281bf7553ed051cf4ead780661c6b42.png)

# 宽限期和滚动更新

优雅关闭适用于要删除的Pod.

但是,如果不删除Pod,该怎么办？

即使您不这样做,Kubernetes也会始终删除Pod.

尤其是,每次部署较新版本的应用程序时,Kubernetes都会创建和删除Pod.

在部署中更改映像时,Kubernetes会逐步推出更改.

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  replicas: 3
  selector:
    matchLabels:
      name: app
  template:
    metadata:
      labels:
        name: app
    spec:
      containers:
      - name: app
        # image: nginx:1.18 OLD
        image: nginx:1.19
        ports:
          - containerPort: 3000
```

如果您有三个副本,并且一旦提交新的YAML资源Kubernetes,则:

- 用新的容器图像创建一个Pod.
- 销毁现有的Pod.
- 等待Pod准备就绪.

并重复上述步骤,直到所有Pod都迁移到较新的版本.

Kubernetes仅在新的Pod准备好接收流量(换句话说,它通过就绪检查)之后才重复每个周期.

Kubernetes是否在移到下一个Pod之前等待Pod被删除？

没有.

如果您有10个Pod,并且Pod需要2秒钟的准备时间和20个关闭的时间,则会发生以下情况:

创建第一个Pod,并终止前一个Pod.
Kubernetes创建一个新的Pod之后,需要2秒钟的准备时间.
同时,被终止的Pod会终止20秒
20秒后,所有新Pod 均已启用(10 Pod ,在2秒后就绪),并且所有之前的10 Pod 都将终止(第一个Terminated Pod将要退出).

总共,您在短时间内将Pod的数量增加了一倍(运行 10次​​,终止 10次).
![](http://img.rocdu.top/20200813/ed8e17bf72f5d5424d1f58c9602e2e81.png)

与就绪探针相比,宽限期越长,您同时具有`Running`(和`Terminating`)的Pod越多.

不好吗

不一定,因为您要小心不要断开连接.

# 终止长时间运行的任务

那长期工作呢？

- 如果您要对大型视频进行转码,是否有任何方法可以延迟停止Pod？

假设您有一个包含三个副本的Deployment.
每个副本都分配了一个视频进行转码,该任务可能需要几个小时才能完成.

当您触发滚动更新时,Pod会在30秒内完成任务,然后将其杀死.

- 如何避免延迟关闭Pod？

您可以将其terminationGracePeriodSeconds增加到几个小时.

`但是,此时Pod的endpoint不可达.`

![](http://img.rocdu.top/20200813/3ef3d002d72dfad05f538ea00f5cbd01.png)

如果公开指标以监视Pod,则您的设备将无法访问Pod.

为什么？

`诸如Prometheus之类的工具依赖于Endpoints来在群集中刮取Pod指标.`

但是,一旦删除Pod,endpoint删除就会在群集中传播,甚至传播到Prometheus！

`您应该考虑为每个新版本创建一个新的部署,而不是增加宽限期.`

当您创建全新的部署时,现有的部署将保持不变.

长时间运行的作业可以照常继续处理视频.

完成后,您可以手动删除它们.

如果希望自动删除它们,则可能需要设置一个自动缩放器,当它们用尽任务时,可以将部署扩展到零个副本.

这种Pod自动定标器的一个示例是Osiris,它是Kubernetes的通用,从零缩放的组件.

该技术有时被称为Rainbow部署,并且在每次您必须使以前的Pod 运行超过宽限期的时间时很有用.

- 另一个很好的例子是WebSockets.

如果您正在向用户流式传输实时更新,则可能不希望在每次发布时都终止WebSocket.

如果您白天经常出游,则可能会导致实时Feed多次中断.

为每个版本创建一个新的部署是一个不太明显但更好的选择.

现有用户可以继续流更新,而最新的Deployment服务于新用户.

当用户断开与旧Pod的连接时,您可以逐渐减少副本并退出过去的Deployment.

# 摘要

您应该注意将Pod从群集中删除,因为Pod的IP地址可能仍用于路由流量.

与其立即关闭Pods,不如考虑在应用程序中等待更长的时间或设置一个preStop钩子.

仅在将集群中的所有endpoint传播并从kube-proxy,Ingress控制器,CoreDNS等中删除后,才应删除Pod.

如果您的Pod运行诸如视频转码或使用WebSockets进行实时更新之类的长期任务,则应考虑使用Rainbow部署.

在Rainbow部署中,您为每个发行版创建一个新的Deployment,并在耗尽连接(或任务)后删除上一个发行版.

长时间运行的任务完成后,您可以手动删除较旧的部署.

或者,您可以自动将部署扩展到零个副本以自动化该过程


原文：https://learnk8s.io/graceful-shutdown

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
