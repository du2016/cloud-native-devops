# 介绍

在使用k8s的过程中在特定场景可能需要控制pod的执行顺序，接下来我们将学习各个开源组件的实现方式


# istio中的实现

今天在测试istio新功能时注意到istio中添加了`values.global.proxy.holdApplicationUntilProxyStarts`，使sidecar注入器在pod容器列表的开始处注入sidecar，并将其配置为阻止所有其他容器的开始，直到代理就绪为止。

在查看代码后发现对istio-proxy容器注入了以下内容。

```
        lifecycle:
          postStart:
            exec:
              command:
              - pilot-agent
              - wait
```

熟悉k8s人可能会记得，poststart 不能保证在调用Container的入口点之前先调用postStart处理程序，那这样怎么通过postStart保证业务容器的延迟启动。


这里就来到了一个误区，大家可能都认为pod的初始化容器完成后，将并行启动pod的常规容器，事实上并不是。

[容器启动代码](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kuberuntime/kuberuntime_manager.go#L835)

可以看到pod中的容器时顺序启动的，按照pod spec.containers 中容器的顺序进行启动。

虽然是顺序启动，但是并不能保证当一个容器依赖于另外一个容器时，在依赖的容器启动完成后再进行启动，istio proxy sidecar 就是一个常见问题，经常出现503问题。


1. 需要将Proxy指定为中的第一个容器spec.containers，但这只是解决方案的一部分，因为它只能确保首先启动代理容器，而不必等待它准备就绪。其他容器立即启动，从而导致容器之间的竞争状态。我们需要防止Kubelet在代理准备好之前启动其他容器。

2. 为第一个容器注入PostStart 生命周期钩子


这样就实现了，如果sidecar容器提供了一个等待该sidecar就绪的可执行文件，则可以在容器的启动后挂钩中调用该文件，以阻止pod中其余容器的启动。

![](http://img.rocdu.top/20200827/1*doJhrU_cgrh8jq2jNrQNFA.png)

以下方式通过/bin/wait-until-ready.sh保证sidecar container早于application容器启动。

```
apiVersion: v1
kind: Pod
metadata:
  name: sidecar-starts-first
spec:
  containers:
  - name: sidecar
    image: my-sidecar
    lifecycle:
      postStart:
        exec:
          command:
          - /bin/wait-until-ready.sh
  - name: application
    image: my-application
```

# k8s自有的Sidecar container

从Kubernetes 1.18可以将容器标记为sidecar，以便它们在正常容器之前启动，而在所有其他容器终止后关闭。因此它们仍然像普通容器一样工作，唯一的区别在于它们的生命周期。目前istio并未使用该方式保证istio-proxy容器的启动顺序，可能是基于版本考虑，并且Sidecar container。

```
apiVersion: v1
kind: Pod
metadata:
  name: bookings-v1-b54bc7c9c-v42f6
  labels:
    app: demoapp
spec:
  containers:
  - name: bookings
    image: banzaicloud/allspark:0.1.1
    ...
  - name: istio-proxy
    image: docker.io/istio/proxyv2:1.4.3
    lifecycle:
      type: Sidecar
```

但是sidecar 容器只能保证sidecar早于业务容器启动，不能保证业务容器的启动先后顺序。有什么方式保证么？


# tekton中的实现

1.tekton中依赖于entrypoint初始化容器初始化脚本，生成各个容器需要执行的entrypoint，通过挂载目录共享到各个容器，共享entrypoint命令，
2.当所有容器ready时，通过downward-api将ready信息反馈给初始化容器
3.初始化容器开始进行初始化操作
4.初始完成后在共享目录完成后，创建一个文件
5.task容器在执行时会监听文件变化，当需要的文件创建完成，开始执行具体的逻辑

代码： https://github.com/tektoncd/pipeline/tree/master/cmd/entrypoint


扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
