# Taint 和 toleration

Node 亲和性，根据之前的描述，是 pod 的一个属性，将其 吸引 到一个 node 集上（以倾向性或硬性要求的形式）。Taint 则刚好相反 – 它们允许一个 node 排斥 一些 pod。

Taint 和 toleration 一起工作，以确保 pod 不会被调度到不适当的 node 上。如果一个或多个 taint 应用于一个 node，这表示该节点不应该接受任何不能容忍 taint 的 pod。 Toleration 应用到 pod 上，则表示允许（但不要求）pod 调度到具有匹配 taint 的 node 上。

您可以通过使用 kubectl taint 命令来给 node 添加一个 taint。例如，

```
kubectl taint nodes node1 key=value:NoSchedule
```

# 为pod添加toleration

```
tolerations: 
- key: "key"
  operator: "Equal"
  value: "value"
  effect: "NoSchedule"
----------------------
tolerations: 
- key: key
  operator: Exists
  value: value
  effect: NoSchedule

```

# 最佳实践

- 专有node(指定pod运行)
  - 添加taint（NoSchedule）
  - 为pod添加toleration
- 具有特殊硬件的 node 
  - 添加taint（PreferNoSchedule）
  - 为pod添加toleration

符合taint则立即被驱逐
发生改变后tolerationSeconds 指定则在过期后驱逐，不添加则不驱逐

pod默认有以下toleration：
- node.alpha.kubernetes.io/notReady tolerationSeconds=300
- node.alpha.kubernetes.io/unreachable tolerationSeconds=300
- statefulset 会自动添加node.alpha.kubernetes.io/notReady node.alpha.kubernetes.io/unreachable