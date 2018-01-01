kubelet 的垃圾收集是非常有用的功能，它可以清除未使用的容器和镜像。kubelet 在每分钟和每五分钟分别回收容器和镜像。

不建议使用第三方的垃圾收集工具，因为这些工具可能会移除期望存在的容器进而破坏 kubelet 的行为。

https://k8smeetup.github.io/docs/concepts/cluster-administration/kubelet-garbage-collection/