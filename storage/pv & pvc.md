# 概念

PersistentVolume（PV）是由管理员设置的存储，它是群集的一部分。
就像节点是集群中的资源一样，PV 也是集群中的资源。
 PV 是 Volume 之类的卷插件，但具有独立于使用 PV 的 Pod 的生命周期。
 此 API 对象包含存储实现的细节，即 NFS、iSCSI 或特定于云供应商的存储系统。
 
PersistentVolumeClaim（PVC）是用户存储的请求。它与 Pod 相似。Pod 消耗节点资源，
PVC 消耗 PV 资源。Pod 可以请求特定级别的资源（CPU 和内存）。
声明可以请求特定的大小和访问模式（例如，可以以读/写一次或 只读多次模式挂载）

# 生命周期

创建PVC -- 查找PV -- 查找storageclass -- 失败

## 静态创建

首先手动创建pv,然后创建PVC，假如PVC能够找到符合自己的PV对象，则进行关联

## 动态创建

当PVC 找不到符合自己的PV时，集群会尝试通过storageclass动态创建PV

假如不想动态创建，可以设置PVC的storageClassName为""，如下所示

```
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: example-local-claim
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: ""
```

# PV回收策略

- Retain 不进行任何操作，需要人为干预
- Delete 直接删除
- Recycle 清空，供其他PVC使用

Recycle已经启用，可以通过定义kube-controller-manager Recycle来实现自定义回收

主要通过以下参数来进行定义
--pv-recycler-increment-timeout-nfs 删除每G空间添加到ActiveDeadlineSeconds的时间增量
--pv-recycler-timeout-increment-hostpath  hostpath ActiveDeadlineSeconds的时间增量 用于测试
--pv-recycler-minimum-timeout-hostpath hostpath最小的ActiveDeadlineSeconds 只用于测试
--pv-recycler-minimum-timeout-nfs nfs最小的ActiveDeadlineSeconds
--pv-recycler-pod-template-filepath-hostpath hostpath回收模板 用于测试
--pv-recycler-pod-template-filepath-nfs nfs回收模板

模板示例
```
apiVersion: v1
kind: Pod
metadata:
  name: pv-recycler
  namespace: default
spec:
  restartPolicy: Never
  volumes:
  - name: vol
    hostPath:
      path: /any/path/it/will/be/replaced
  containers:
  - name: pv-recycler
    image: "k8s.gcr.io/busybox"
    command: ["/bin/sh", "-c", "test -e /scrub && rm -rf /scrub/..?* /scrub/.[!.]* /scrub/*  && test -z \"$(ls -A /scrub)\" || exit 1"]
    volumeMounts:
    - name: vol
      mountPath: /scrub
```


# 扩展PVC

只有当storageclass allowVolumeExpansion字段为true时才能进行扩展

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gluster-vol-default
provisioner: kubernetes.io/glusterfs
parameters:
  resturl: "http://192.168.10.100:8080"
  restuser: ""
  secretNamespace: ""
  secretName: ""
allowVolumeExpansion: true
```

# 扩展包含文件系统的卷大小

只有Pod PersistentVolumeClaim在ReadWrite模式下使用时才调整文件系统的大小

# 调整使用中的PersistentVolumeClaim的大小

在这种情况下，您不需要删除并重新创建使用现有PVC的Pod或部署。
文件系统扩展后，所有使用中的PVC都将自动供其Pod使用。此功能对Pod或部署中未使用的PVC无效。
您必须创建一个使用PVC的Pod，然后才能完成扩展。

与其他卷类型类似-当由Pod使用时，FlexVolume卷也可以扩展。

# PV

apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv0003
spec:
  capacity:
    storage: 5Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Recycle
  storageClassName: slow
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    path: /tmp
    server: 172.17.0.2
    
    
- capacity： 容量
- volumeMode：
   - block 块设备
   - filesystem 文件系统
  
accessModes：
  - ReadWriteOnce –该卷可以通过单个节点以读写方式安装
  - ReadOnlyMany –该卷可以被许多节点只读安装
  - ReadWriteMany –该卷可以被许多节点读写安装
 
 同时只能设置一个，但是可以切换
 
storageClassName 指定storageClassName的PV只能被对应storageClassName名称的PVC


## 节点亲和性

显示设置PV需要再哪个节点才能访问
```
    nodeAffinity:
      required:
        nodeSelectorTerms:
        - matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
            - 10.10.8.42
```

## 阶段
- Available 尚未绑定到声明的资源
- Bound 绑定到PVC
- Released 已删除但是未回收
- Failed 无法回收

# PVC

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: myclaim
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 8Gi
  storageClassName: slow
  selector:
    matchLabels:
      release: "stable"
    matchExpressions:
      - {key: environment, operator: In, values: [dev]}
```
accessModes 访问模式与PV相同
volumeMode 卷模式与PV相同
resources 和pod 的resource一样，请求资源
selector 进一步选择


# 块设备
通过指定volumeMode了；哎指定块设备类型，当在pod内部映射块设备时，指定块设备路径而不是挂载点

# 从指定来源创建PVC

## 通过快照创建

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: restore-pvc
spec:
  storageClassName: csi-hostpath-sc
  dataSource:
    name: new-snapshot-test
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

## 通过现有PVC创建

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cloned-pvc
spec:
  storageClassName: my-csi-plugin
  dataSource:
    name: existing-src-pvc-name
    kind: PersistentVolumeClaim
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```