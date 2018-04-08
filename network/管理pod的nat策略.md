
# 关闭docker及flannel的snat策略

## 关闭dockersnat

docker默认开启masq,可以通过 `--ip-masq=false`参数关闭masq

## 关闭flannel snat策略

flannel默认通过参数注入的方式开启masq：

- 使用daemonset方式启动可以通过删除--ip-masq参数实现
- 在系统直接部署的可以修改/usr/libexec/flannel/mk-docker-opts.sh设置ipmasq=false


# 通过ip-masq-agent实现

## 介绍

ip-masq-agent 配置iptables规则为MASQUERADE（除link-local外），
并且可以附加任意IP地址范围

它会创建一个iptables名为IP-MASQ-AGENT的链，包含link local（169.254.0.0/16）
和用户指定的IP地址段对应的规则，
它还创建了一条规则POSTROUTING，以保证任何未绑定到LOCAL目的地的流量会跳转到这条链上。

匹配到IP-MASQ-AGENT中对应规则的IP（最后一条规则除外），
通过IP-MASQ-AGENT将不受MASQUERADE管理（他们提前从链上返回），
ip-masq-agent链最后一条规则将伪装任何非本地流量。

## 安装

此仓库包含一个示例yaml文件，
可用于启动ip-masq-agent作为Kubernetes集群中的DaemonSet。

```
kubectl create -f ip-masq-agent.yaml
```
ip-masq-agent.yaml中的规则指定kube-system为对应DaemonSet Pods 运行的名称空间。

## 配置代理

提示：您不应尝试在Kubelet也配置有非伪装CIDR的集群中运行此代理。
您可以传递--non-masquerade-cidr=0.0.0.0/0给Kubelet以取消其规则，
这将防止Kubelet干扰此代理。

默认情况下，代理配置为将RFC 1918指定的三个私有IP范围视为非伪装CIDR。
这些范围是10.0.0.0/8，172.16.0.0/12和192.168.0.0/16。
该代理默认将link-local（169.254.0.0/16）视为非伪装CIDR。

默认情况下，代理配置为每60秒从其容器中的/etc/config/ip-masq-agent文件重新加载其配置，

代理配置文件应该以yaml或json语法编写，并且可能包含三个可选项：

- nonMasqueradeCIDRs []string：CIDR表示法中的列表字符串，用于指定非伪装范围。
- masqLinkLocal bool：是否伪装流量169.254.0.0/16。默认为False。
- resyncInterval string：代理尝试从磁盘重新加载配置的时间间隔。语法是Go的time.ParseDuration函数接受的任何格式。

该代理将在其容器中查找配置文件/etc/config/ip-masq-agent。这个文件可以通过一个configmap提供，通过一个ConfigMap管道进入容器ConfigMapVolumeSource。
因此，该客户端可以通过创建或编辑ConfigMap实现实时群集中重新配置代理程序。

这个仓库包括一个ConfigMap，可以用来配置代理（agent-config目录）的目录表示。
使用此目录在急群众创建ConfigMap：

```
kubectl create configmap ip-masq-agent --from-file=agent-config --namespace=kube-system
```

请注意，我们在与DaemonSet Pods相同的命名空间中创建了configmap，
并切该ConfigMap的名称与ip-masq-agent.yaml中的配置一致。
这对于让ConfigMap对于出现在Pods的文件系统中是必需的。

## 理论基础

该代理解决了为集群中的非伪装配置CIDR范围的问题（通过iptables规则）。
现在这以功能是通过--non-masquerade-cidr向Kubelet 传递一个标志来实现的，
该标志只允许一个CIDR被配置为非伪装。RFC 1918定义了三个范围（10/8，172.16/12，192.168/16为私有IP地址空间）。

有些用户会希望在不伪装这些范围之间进行通信-例如，
如果企业现有的网络使用10/8范围，
他们可能希望运行群集以及PodS 在 192.168/16以避免IP冲突。
他们也希望这些Pods能够有效地（不伪装）与他人和他们现有的网络资源进行沟通10/8。
这要求集群中的每个节点都跳过两个范围的伪装。

我们正在尝试从Kubelet中删除网络代码，
因此，而不是将Kubelet扩展为接受多个CIDR，
ip-masq-agent允许您运行一个将CIDR列表配置为非伪装的DaemonSet。


# 使用ospf网络

- 关闭 docker及flannel的snat

- 路由需要配置开启ospf划分区域

- 安装quagga

      ```
      yum install quagga -y
      ```

- 配置zebra 及ospfd的账号密码等信息

这样就直接和网络打通，acl可以在路由上设置，从而实现外部网络直接访问pod，

# 总结

在生产环境中，打通pod与集群外外部服务的网络很有必要，使用nat策略，当多个相同pod在一个node上，
一旦出现错误对应服务无法获取哪个pod在连接该服务，取消nat策略可直接拿到pod IP，方便debug。


欢迎加入QQ群：k8s开发与实践（482956822）一起交流k8s技术