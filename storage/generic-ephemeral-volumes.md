# 通用临时卷

Kubernetes提供了其生命周期与pod绑定的卷插件，可用作临时空间（例如内置emptydir卷类型）或将某些数据加载到pod中（例如内置configmap和secret卷类型，或`CSI内联卷`）。新的alpha通用临时卷功能允许将任何支持动态预配置的现有存储驱动程序用作临时卷，并将该卷的生命周期绑定到Pod。它可用于提供与根磁盘不同的临时存储，例如永久性内存或该节点上的单独本地磁盘。支持所有用于卷配置的StorageClass参数。支持PersistentVolumeClaims支持的所有功能，例如存储容量跟踪，快照和还原以及卷大小调整。

# 使用

创建以下deployment
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - name: http
          containerPort: 80
        volumeMounts:
        - mountPath: "/scratch"
          name: scratch-volume
        command: [ "sleep", "1000000" ]
      volumes:
        - name: scratch-volume
          ephemeral:
            volumeClaimTemplate:
              metadata:
                labels:
                  type: my-frontend-volume
              spec:
                accessModes: [ "ReadWriteOnce" ]
                storageClassName: "standard"
                resources:
                  requests:
                    storage: 1Gi
```

查看资源状态：

```
kubectl get pods
NAME                     READY   STATUS    RESTARTS   AGE
nginx-67b97764d6-7tgvx   1/1     Running   0          4m30s

kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                           STORAGECLASS   REASON   AGE
pvc-8f7683f4-3472-412f-bbb3-b1e9d142a1f0   1Gi        RWO            Delete           Bound    default/nginx-67b97764d6-7tgvx-scratch-volume   standard                4m30s

kubectl get pvc
NAME                                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
nginx-67b97764d6-7tgvx-scratch-volume   Bound    pvc-8f7683f4-3472-412f-bbb3-b1e9d142a1f0   1Gi        RWO            standard       4m35s
```


实现通用临时卷后对于无状态服务可以实现动态挂盘，在机器学习/数据分析领域比较有使用场景,将数据单独存储在云盘内，降低本地IO压力，

# CSI临时内联卷

CSIInlineVolume 功能门  1.15引入 1.16beta默认开启

```
kind: Pod
apiVersion: v1
metadata:
  name: my-csi-app-inline-volume
spec:
  containers:
    - name: my-frontend
      image: busybox
      command: [ "sleep", "100000" ]
      volumeMounts:
      - mountPath: "/data"
        name: my-csi-volume
  volumes:
  - name: my-csi-volume
    csi:
      driver: pmem-csi.intel.com
      fsType: "xfs"
      volumeAttributes:
        size: "2Gi"
        nsmode: "fsdax"
```

https://kubernetes.io/blog/2020/01/21/csi-ephemeral-inline-volumes/