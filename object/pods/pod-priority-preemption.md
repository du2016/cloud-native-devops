# 优先抢占

# 介绍

Kubernetes 1.8 及其以后的版本中可以指定 Pod 的优先级。
优先级表明了一个 Pod 相对于其它 Pod 的重要性。当 Pod 无法被调度时，
scheduler 会尝试抢占（驱逐）低优先级的 Pod，使得这些挂起的 pod 可以被调度。
在 Kubernetes 未来的发布版本中，优先级也会影响节点上资源回收的排序。

# 配置

## 启用
- api开启 --feature-gates=PodPriority=true
- 增加一个或者多个 PriorityClass
- 创建拥有字段 PriorityClassName 的 Pod
- api和scheduler配置 --runtime-config=scheduling.k8s.io/v1alpha1=true

## 创建PriorityClass

```
apiVersion: v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000000
globalDefault: false
description: "This priority class should be used for XYZ service pods only."
```

## pod使用PriorityClass

```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  priorityClassName: high-priority
```

# 抢占

```
Pod 生成后，会进入一个队列等待调度。scheduler 从队列中选择一个 Pod，然后尝试将其调度到某个节点上。如果没有任何节点能够满足 Pod 指定的所有要求，对于这个挂起的 Pod，抢占逻辑就会被触发。当前假设我们把挂起的 Pod 称之为 P。抢占逻辑会尝试查找一个节点，在该节点上移除一个或多个比 P 优先级低的 Pod 后， P 能够调度到这个节点上。如果节点找到了，部分优先级低的 Pod 就会从该节点删除。Pod 消失后，P 就能被调度到这个节点上了。
```

在此阶段PDB不被支持

# 参考

https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/