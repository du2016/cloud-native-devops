扩展到新领域-Istio中的智能DNS代理

使用本地工作负载DNS方案来简化VM集成，多集群等

DNS解析是Kubernetes上任何应用程序基础架构的重要组成部分.当您的应用程序代码尝试访问Kubernetes集群中的另一个服务甚至是Internet上的服务时，它必须先查找与该服务的主机名相对应的IP地址，然后再启动与该服务的连接.此名称查找过程通常称为服务发现。在Kubernetes中,server(无论是kube-dnsCoreDNS还是CoreDNS)将服务的主机名解析为唯一的不可路由的虚拟IP(VIP)，如果它是clusterIP类型的服务.在kube-proxy每个节点上这个VIP映射到该服务的一组pod，并随机选择一个pod进行转发。使用服务网格时，边车的工作原理就流量转发而言与kube-proxy相同。

下图描述了当今DNS的作用：

![](http://img.rocdu.top/20201117/role-of-dns-today.png)

# DNS带来的问题

尽管DNS在服务网格中的作用似乎微不足道，但它始终代表着将网格扩展到VM并实现无缝多集群访问的方式。

## 虚拟机访问Kubernetes服务

考虑到VM带有sidecar的情况。如下图所示，VM上的应用程序会查找Kubernetes群集内服务的IP地址，因为它们通常无法访问群集的DNS服务器。

![虚拟机访问Kubernetes服务时的DNS解析问题](http://img.rocdu.top/20201117/vm-dns-resolution-issues.png)

如果有人愿意参与一些涉及dnsmasq和使用NodePort服务对kube-dns进行外部暴露的复杂变通方法，从技术上讲，可以在虚拟机上使用kube-dns作为域名服务器：假设您设法说服集群管理员这样做。 即使这样，您仍在打开许多安全问题的大门。 归根结底，对于那些组织能力和领域专业知识有限的人来说，这些解决方案通常超出范围。

## 没有VIP的外部TCP服务

不仅网状网络中的VM遭受DNS问题。为了使Sidecar能够准确地区分网格外部的两个不同TCP服务之间的流量，这些服务必须位于不同的端口上，或者它们需要具有全局唯一的VIP，就像clusterIP分配给Kubernetes服务一样。但是，如果没有VIP，该怎么办？云托管服务(例如托管数据库)通常没有VIP。取而代之的是，提供者的DNS服务器返回实例IP之一，然后可由应用程序直接访问这些实例IP。例如，考虑以下两个服务条目，它们指向两个不同的AWS RDS服务：


```
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: db1
  namespace: ns1
spec:
  hosts:
  - mysql–instance1.us-east-1.rds.amazonaws.com
  ports:
  - name: mysql
    number: 3306
    protocol: TCP
  resolution: DNS
---
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: db2
  namespace: ns1
spec:
  hosts:
  - mysql–instance2.us-east-1.rds.amazonaws.com
  ports:
  - name: mysql
    number: 3306
    protocol: TCP
  resolution: DNS
```

边车上有一个侦听器 0.0.0.0:3306，该侦听器从公共DNS服务器查找mysql-instance1.us-east1.rds.amazonaws.com的IP地址并将流量转发给它。它无法将流量路由至db2因为它无法区分到达的流量 0.0.0.0:3306是绑定db1还是绑定db2。实现此目的的唯一方法是将解析设置为NONE，使Sidecar将端口上的所有流量盲目转发3306到应用程序请求的原始IP。这类似于在防火墙上打一个洞，使所有流量都可以3306传入端口，而与目标IP无关。为了使流量畅通，现在您不得不在系统的安全性上做出妥协。

### 为远程群集中的服务解析DNS

多群集网格的DNS限制是众所周知的。如果没有笨拙的解决方法(例如在调用方名称空间中创建存根服务)，则一个群集中的服务无法查找其他群集中服务的IP地址。

# 控制DNS

总而言之，DNS在Istio中一直是一个棘手的问题。现在是时候杀死那只野兽了。我们(Istio网络团队)决定以对您(最终用户)完全透明的方式彻底解决该问题。我们的首次尝试涉及利用Envoy的DNS代理。事实证明这是非常不可靠的，并且由于Envoy使用的c-ares DNS库普遍缺乏复杂性，因此总体上令人失望。为了解决这个问题，我们决定在Go语言编写的Istio sidecar代理中实现DNS代理。我们能够优化实现，以处理我们要解决的所有场景，而不会影响规模和稳定性。我们使用的Go DNS库与可扩展DNS实现(例如CoreDNS，Consul，Mesos等)使用的库相同。

从Istio 1.8开始，Sidecar上的Istio代理将附带由Istiod动态编程的缓存DNS代理。Istiod基于Kubernetes服务和集群中的服务条目，为应用程序可以访问的所有服务推送主机名到IP地址的映射。来自应用程序的DNS查找查询被Pod或VM中的Istio代理透明地拦截并提供服务。如果查询是针对网格中的服务，则无论该服务所在的群集是什么，代理都会直接对应用程序做出响应。如果不是，它将查询转发到/etc/resolv.conf中定义的上游域名服务器。下图描述了当应用程序尝试使用其主机名访问服务时发生的交互。

![](http://img.rocdu.top/20201117/dns-interception-in-istio.png)

正如您将在以下各节中看到的那样，DNS代理功能已在Istio的许多方面产生了巨大的影响。

## 降低DNS服务器的负载并提高解析度

群集中Kubernetes DNS server上的负载急剧下降，因为Istio在Pod内几乎解决了所有DNS查询。集群上的网格使用范围越大，DNS服务器上的负载就越小。在Istio代理中实现自己的DNS代理使我们能够实现出色的优化，例如CoreDNS auto-path，而不会出现CoreDNS当前面临的正确性问题。

要了解此优化的影响，让我们在标准Kubernetes集群中采用简单的DNS查找方案，而无需为Pod进行任何自定义DNS设置-即，默认/etc/resolv.conf中设置为ndots:5。当您的应用程序启动DNS查找 productpage.ns1.svc.cluster.local时，它会在按原查询主机之前将DNS搜索名称空间作为DNS查询的一部分附加在/etc/resolv.conf(例如ns1.svc.cluster.local)中。结果，实际上发出的第一个DNS查询看起来像 productpage.ns1.svc.cluster.local.ns1.svc.cluster.local，当不涉及Istio时，它将不可避免地使DNS解析失败。如果您 /etc/resolv.conf有5个搜索名称空间，则应用程序将为每个搜索名称空间发送两个DNS查询，一个用于IPv4 A记录，另一个用于IPv6 AAAA记录，然后是最后一对查询，其中包含代码中使用的确切主机名。在建立连接之前，该应用程序将为每个主机执行12个DNS查找查询！

使用Istio实现的CoreDNS样式自动路径技术，Sidecar代理将检测到在第一个查询中查询的真实主机名，并将cname记录 返回productpage.ns1.svc.cluster.local为该DNS响应的一部分以及的A/AAAA记录 productpage.ns1.svc.cluster.local。现在，收到此响应的应用程序可以立即提取IP地址，并继续建立与该IP的TCP连接。Istio代理中的智能DNS代理将DNS查询数量从12个大大减少到2个！

## 虚拟机到Kubernetes集成

由于Istio代理对网格内的服务执行了本地DNS解析，因此从VM进行的Kubernetes服务的DNS查找查询现在将成功完成，而无需笨拙的变通办法来暴露kube-dns 到群集外部。现在，无缝解析集群中内部服务的能力将简化您到微服务的旅程，因为VM现在可以访问Kubernetes上的微服务，而无需通过API网关进行其他级别的间接访问。

## 尽可能自动分配VIP

您可能会问，代理中的此DNS功能如何解决区分在同一端口上没有VIP的多个外部TCP服务的问题？

从Kubernetes获得启发，Istio现在将自动将不可路由的VIP(来自E类子网)分配给此类服务，只要它们不使用通配符主机即可。边车上的Istio代理将使用VIP作为来自应用程序的DNS查找查询的响应。现在，Envoy可以清楚地区分绑定到每个外部TCP服务的流量，并将其转发到正确的目标。通过引入DNS代理，您将不再需要`resolution: NONE`用于非通配TCP服务，从而改善了整体安全性。Istio在通配符外部服务(例如`*.us-east1.rds.amazonaws.com`)方面无济于事。您将不得不诉诸NONE解析模式来处理此类服务。

## 多集群DNS查找

对于喜欢冒险的人来说，尝试编织一个多集群网格，其中应用程序直接调用远程集群中名称空间的内部服务，DNS代理功能非常方便。您的应用程序可以解析任何名称空间中任何群集上的Kubernetes服务，而无需在每个群集中创建存根Kubernetes服务。

DNS代理的优势超出了Istio当前描述的多集群模型。在Tetrate，我们在客户的多群集部署中广泛使用此机制，以使Sidecar能够为网格中所有群集的入口网关处暴露的主机解析DNS，并通过相互的TLS访问它们。

## 结论思想

在跨多个群集，不同的环境编织网格以及集成外部服务时，由于对DNS缺乏控制而导致的问题通常经常被整体忽略和忽略。在Istio Sidecar代理中引入缓存DNS代理可以解决这些问题。通过对应用程序的DNS解析进行控制，Istio可以准确识别流量绑定到的目标服务，并增强Istio在群集内和群集之间的整体安全性，路由和遥测状态。

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
