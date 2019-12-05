除了在主机上运行的Kubernetes核心组件（如api-server, scheduler, controller-manager ）外，
还有许多附加组件，由于各种原因，这些附加组件必须在常规群集节点
（而不是Kubernetes master）上运行。其中一些附加组件对于功能齐全的群集至关重要，
例如metrics-server，DNS和UI。
如果紧急附加组件被驱逐（手动或作为其他操作（如升级）的副作用）并变为挂起状态（
例如，当该群集被高度利用且有其他挂起的Pod计划进入该群集时，该群集可能会停止正常工作）
被驱逐的关键附加组件腾出的空间或节点上可用的资源量由于其他原因而发生了变化）。

请注意，将pod标记为关键并不意味着完全防止驱逐。它只能防止pod永久不可用。
对于静态pod，这意味着无法将其逐出，但对于非静态pod，
这仅意味着它们将始终被重新调度。

# 设置critical pod

在v1.11之前，关键Pod必须在kube-system命名空间中运行，在v1.11之后，此限制已被删除，
并且可以通过以下两种方式将任何命名空间中的pod配置为关键Pod：

- 确保已启用PodPriority 功能门。将priorityClassName设置为
`system-cluster-critical`或`system-node-critical`，
后者是整个集群中最高的，这是自v1.10 +起可用的两个优先级名称

- 或者，确保同时启用PodPriority和ExperimentCriticalPodAnnotation功能门，
您可以将`scheduler.alpha.kubernetes.io/critical-pod`作为键添加注释，
并将空字符串作为值添加到您的pod，但是从1.13版开始不推荐使用此注释，
并且在将来的版本中将删除该注释。


# 原理分析

当资源节点的资源不足时，新的pod就会尝试抢占已有pod,kubelet源码中会根据一些列条件进行判断是否可以被抢占
https://github.com/kubernetes/kubernetes/blob/0939f9010381fb5ce56ee0543109c2554681dacb/pkg/kubelet/types/pod_update.go#L168

```
func Preemptable(preemptor, preemptee *v1.Pod) bool {
	if IsCriticalPod(preemptor) && !IsCriticalPod(preemptee) {
		return true
	}
	if (preemptor != nil && preemptor.Spec.Priority != nil) &&
		(preemptee != nil && preemptee.Spec.Priority != nil) {
		return *(preemptor.Spec.Priority) > *(preemptee.Spec.Priority)
	}

	return false
}
```

- 首先判断抢占者为关键pod,被抢占pod非关键pod,则抢占成功
- 如果都设置的有Priority，则抢占者大于被抢占pod的优先级时，抢占成功

# 关键pod判定

```
func IsCriticalPod(pod *v1.Pod) bool {
	if IsStaticPod(pod) {
		return true
	}
	if IsMirrorPod(pod) {
		return true
	}
	if pod.Spec.Priority != nil && IsCriticalPodBasedOnPriority(*pod.Spec.Priority) {
		return true
	}
	return false
}
```

- 如果为静态pod则为关键pod
- 如果为mirrorpod则为关键pod 即带有`kubernetes.io/config.mirror`注释的pod,实际上只要是static pod,都会加上这个注释，和上面的有重复
- 通过IsCriticalPodBasedOnPriority判断 大于2000000000的pod

# 为什么大于2000000000的pod即判定为关键pod

```
kubectl get priorityClass
NAME                      VALUE        GLOBAL-DEFAULT   AGE
system-cluster-critical   2000000000   false            6d1h
system-node-critical      2000001000   false            6d1h
```

kubeadm生成的所有master组件都为system-cluster-critical

kube-proxy为system-node-critical，由此看出在一个机器上会先保证kube-proxy的可用性





