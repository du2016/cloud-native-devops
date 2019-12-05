# 分配 Pod 到 Node 上

# 介绍

在默认调度算法的基础上，我们可以通过以下两种方式指定pod的运行位置：
- nodeSelector
- Affinity & anti-affinity


# nodeSelector

## 为node添加标签

```
kubectl label nodes kubernetes-foo-node-1.c.a-robinson.internal disktype=ssd
# 验证
kubectl get nodes -l disktype=ssd
```

## 创建pod

```
cat << EOF | kubectl create -f -
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
  nodeSelector:
    disktype: ssd
EOF
```

## 验证

```
kubectl get pods nginx -o wide
```

# 亲和性（Affinity）和反亲和性（anti-affinity）

相对于nodeSelector Affinity&anti-affinity有以下优势

- 表达语言更丰富（不仅仅是 "AND 精确匹配"）
- 你可以指定规则是 "软"/"偏好" 的，而不是一个硬性要求，所以即使 scheduler 不能满足它的规则，pod 仍然会被调度
- 您可以针对 node 上正在运行的其它 pod （或者其它的拓扑域）制定标签，而不仅仅是 node 自身，这就能指定哪些 pod 能够（或者不能够）落在同一节点。

有以下类型

- requiredDuringSchedulingIgnoredDuringExecution（硬性限制，node变化不会被驱逐）
- preferredDuringSchedulingIgnoredDuringExecution（硬性限制，node变化不会被驱逐）
- requiredDuringSchedulingRequiredDuringExecution （硬性限制，node变化会被驱逐，未实现）
- requiredDuringSchedulingIgnoredDuringExecution（硬性限制，node变化会被驱逐，未实现）

- nodeAffinity

    ```
    apiVersion: v1
    kind: Pod
    metadata:
      name: with-node-affinity
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: kubernetes.io/e2e-az-name
                operator: In
                values:
                - e2e-az1
                - e2e-az2
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 1
            preference:
              matchExpressions:
              - key: another-node-label-key
                operator: In
                values:
                - another-node-label-value
      containers:
      - name: with-node-affinity
        image: k8s.gcr.io/pause:2.0
    ```

- podAffinity

    ```
    apiVersion: v1
    kind: Pod
    metadata:
      name: with-pod-affinity
    spec:
      affinity:
        podAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: security
                operator: In
                values:
                - S1
            topologyKey: failure-domain.beta.kubernetes.io/zone
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: security
                  operator: In
                  values:
                  - S2
              topologyKey: kubernetes.io/hostname
      containers:
      - name: with-pod-affinity
        image: k8s.gcr.io/pause:2.0
    ```
    
    
# 参考

https://kubernetes.io/docs/concepts/configuration/assign-pod-node/