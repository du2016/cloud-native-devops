# 概述

Istio 在网格中部署的Pod中注入initContainer，istio-init。该istio-init容器设置荚的网络流量重定向到/从Istio三轮代理。这就要求将用户或服务帐户部署到网格上的Pod具有足够的Kubernetes RBAC权限才能部署具有NET_ADMIN和NET_RAW功能的容器。对于某些组织的安全合规性，要求Istio用户具有提升的Kubernetes RBAC权限是有问题的。Istio CNI插件是istio-init执行相同网络功能但不要求Istio用户启用提升的Kubernetes RBAC权限的容器的替代。

Istio CNI插件在Kubernetes Pod生命周期的网络设置阶段执行Istio Mesh Pod流量重定向，从而消除了 将Pod部署到Istio Mesh中的用户的需求NET_ADMIN和NET_RAW功能。Istio CNI插件取代了istio-init容器提供的功能。

# 配置istio启用istio-cni

```
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  components:
    cni:
      enabled: true
  values:
    cni:
      excludeNamespaces:
       - istio-system
       - kube-system
      logLevel: info

istioctl install -f istio-cni.yaml
```

此时查看启用自动注入的pod会发现init container的名称(istio-validation)有别于不启用istio-cni的名称(istio-init)，

且执行命令多了以下参数，

```
- --run-validation  # 校验流量是否劫持
- --skip-rule-apply # 不执行iptables策略
```

# install-cni

install-cni程序会根据获取配置通过模板对原始cni配置文件进行修改，配置模板存储在istio-cni-config configmap中，渲染后添加的配资如下

```
{
      "kubernetes": {
        "cni_bin_dir": "/opt/cni/bin",
        "exclude_namespaces": [
          "istio-system",
          "kube-system"
        ],
        "kubeconfig": "/etc/cni/net.d/ZZZ-istio-cni-kubeconfig"
      },
      "log_level": "info",
      "name": "istio-cni",
      "type": "istio-cni"
    }
```

分别执行以下操作：

- 拷贝所需要的相关二进制命令，包含istio-cni/istio-cni-repair/istio-iptables三个命令
- 创建所需要的kubeconfig，默认为/etc/cni/net.d/ZZZ-istio-cni-kubeconfig
- 生成cniconfig，即添加了一个kubernetes cni plugin，
- 永久等待


# istio-cni

当kubelet调用cni时会执行上述插件，
通过cni传过来的参数获取相关配置，如果pod包含istio-init则不进行相关操作

- 包含sidecar.istio.io/inject=true 注释
- 包含istio-init initcontainer
- 命名空间在Exclude列表中
- 包含 sidecar.istio.io/status 注释

通过NewRedirect生成Redirect信息

redirect结构如下，保存着后续iptables所需的信息

```
type Redirect struct {
	targetPort           string
	redirectMode         string
	noRedirectUID        string
	includeIPCidrs       string
	includePorts         string
	excludeIPCidrs       string
	excludeInboundPorts  string
	excludeOutboundPorts string
	kubevirtInterfaces   string
}
```

然后调用istio.Program方法进入命名空间执行istio-iptables命令

```
func (ipt *iptables) Program(netns string, rdrct *Redirect) error {
	netnsArg := fmt.Sprintf("--net=%s", netns)
	nsSetupExecutable := fmt.Sprintf("%s/%s", nsSetupBinDir, nsSetupProg)
	nsenterArgs := []string{
		netnsArg, //指定网络命名空间
		nsSetupExecutable, //istio-iptables命令路径
		"-p", rdrct.targetPort, //这些实际上就是istio-iptables的参数
		"-u", rdrct.noRedirectUID,
		"-m", rdrct.redirectMode,
		"-i", rdrct.includeIPCidrs,
		"-b", rdrct.includePorts,
		"-d", rdrct.excludeInboundPorts,
		"-o", rdrct.excludeOutboundPorts,
		"-x", rdrct.excludeIPCidrs,
		"-k", rdrct.kubevirtInterfaces,
	}
	log.Info("nsenter args",
		zap.Reflect("nsenterArgs", nsenterArgs))
	out, err := exec.Command("nsenter", nsenterArgs...).CombinedOutput() //执行命令
	if err != nil {
		log.Error("nsenter failed",
			zap.String("out", string(out)),
			zap.Error(err))
		log.Infof("nsenter out: %s", out)
	} else {
		log.Infof("nsenter done: %s", out)
	}
	return err
}
```

进入对应的网络命名空间执行istio-iptables命令进行端口劫持相关操作，实现了和istio-init init container相同的功能，同样执行的是istio-iptables命令


# istio-cni-repair

该程序相当于一个故障处理程序，主要解决以下两个问题

- 如果应用程序容器在istio-cni安装完成之前启动，则Kubelet不了解Istio CNI插件。结果是出现了没有Istio iptables规则的应用程序容器。该Pod可以访问网络，而其他Pod可以访问该网络，从而有效地绕过了所有Istio策略。这是一个安全问题，因为它无提示地绕过所有策略检查。
- 当kubernetes响应负载的突然增加时，应用程序pod和Istio CNI安装程序之间的竞争更有可能发生。节点启动后，Kubernetes将有许多可调度的Pod分配给该节点，这些Pod都与CNI安装程序竞争。如果某个节点突然终止，例如在GKE可抢占节点的情况下，则可能出现这种情况。

## 上述问题解决思路

- 通过init容器检测到iptables在pod中未正确配置，退出并返回错误（推荐方式）
- 物理安装，不通过daemonset安装
- 使用netd安装CNI

## istio-cni-repair执行逻辑

istio-cni-repair的目的就是为了检测返回错误的pod,主要执行以下逻辑.

- 初始化client-set
- 生成RepairController，根据标签检测istio-validation容器崩溃的pod
- 搜索istio-validation容器崩溃的pod
- 根据label-pods/delete-pods参数决定是否删除pod

# 总结

虽然istio-cni可以减少pod的授权，但是也带来了其他问题，增加了复杂性，如果对容器权限不敏感的情况下，不推荐使用istio-cni。


参考:

[缓解istio cni race设计草案](https://docs.google.com/document/d/1SQzrFxtcn3o_79OtJYbSHMPuqZNhR3EaEhbkpBVMXAw/edit#heading=h.7zgnj8bwqfld)

[istio源码](https://github.com/istio/istio)

[istio官方文档](https://istio.io/latest/docs/setup/additional-setup/cni/)