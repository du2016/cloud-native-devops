从1.8版开始，Kubernetes Storage SIG停止接受树内卷插件，并建议所有存储提供商实施树外插件。目前有两种推荐的实现方式：容器存储接口（CSI）和Flexvolume。

# Flexvolume

## 介绍

flexvolume使用户能够编写自己的驱动程序并在Kubernetes中添加对卷的支持。如果--enable-controller-attach-detach启用Kubelet选项，则供应商驱动程序应安装在每个Kubelet节点和主节点上的卷插件路径中。

Flexvolume是Kubernetes 1.8版本以后的GA特性。

## 先决条件

在插件路径中的所有节点上安装供应商驱动程序，--enable-controller-attach-detach设置为true，
安装插件的路径：\<plugindir\>/\<vendor~driver\>/\<driver\>。
默认的插件目录是/usr/libexec/kubernetes/kubelet-plugins/volume/exec/。
可以通过--volume-plugin-dir标志在Kubelet中进行更改，
并通过标志在控制器管理器中进行更改--flex-volume-plugin-dir。

例如，要添加cifs驱动程序，供应商foo将驱动程序安装在：
/usr/libexec/kubernetes/kubelet-plugins/volume/exec/foo~cifs/cifs

供应商和驱动程序名称必须与卷规格中的flexVolume.driver匹配，'〜'替换为'/'。
例如，如果flexVolume.driver设置为foo/cifs，那么供应商是foo，而驱动程序是cifs

## 动态插件发现

Flexvolume从v1.8开始支持动态检测驱动程序的能力。
系统初始化时不需要存在驱动程序，或者需要重新启动kubelet或控制器管理器，
则可以在系统运行时安装，升级/降级和卸载驱动程序。有关更多信息，请参[阅设计文档](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/flexvolume-deployment.md)

## 自动插件安装/升级

安装和升级Flexvolume驱动程序的一种可能方式是使用DaemonSet。见[推荐驱动程序部署方法](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/flexvolume-deployment.md#recommended-driver-deployment-method)的详细信息。

## 插件详细信息

该插件希望为后端驱动程序实现以下调用。有些标注是可选的。
调用是从Kubelet和Controller管理器节点调用的。
只有当启用了“--enable-controller-attach-detach”Kubelet选项时，
才会从Controller-manager调用调用。

### 驱动程序调用模型

#### init

初始化驱动程序。在Kubelet＆Controller manager初始化期间调用。
成功时，该函数返回一个功能映射，显示驱动程序是否支持每个Flexvolume功能。当前功能:

- attach - 指示驱动是否需要附加和分离操作的布尔字段。
该字段是必需的，但为了向后兼容，默认值设置为true，即需要附加和分离。
有关功能图格式，请参阅[驱动程序输出](https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md#driver-output)。

```
<driver executable> init
```

#### attach

在给定主机上附加给定规范指定的卷。成功时，返回设备连接到节点的设备路径。
如果启用了“--enable-controller-attach-detach”Kubelet选项，
则Nodename参数才是有效/相关的。来自Kubelet＆Controllermanager。

此调出不会传递Flexvolume规范中指定的“secrets”。如果您的驱动程序需要secrets，
请不要执行此调出，而是使用“mount”调出并在该调出中执行attach和调用。

```
<driver executable> attach <json options> <node name>
```

#### Detach

从Kubelet节点分离卷。只有在启用启用了“--enable-controller-attach-detach”Kubelet选项时
Nodename参数才是有效/相关的。Kubelet & Controller manager进行调用

```
<driver executable> detach <mount device> <node name>
```

#### Wait for attach

等待卷连接到远程节点上。成功后，返回设备的路径。从Kubelet & Controller manager进行调用，超时时间为10毫秒([代码](https://git.k8s.io/kubernetes/pkg/kubelet/volumemanager/volume_manager.go#L88)),


```
<driver executable> waitforattach <mount device> <json options>
```

#### Volume is Attached
检查节点上是否连接了卷。从Kubelet & Controller manager进行调用.

```
<driver executable> isattached <json options> <node name>
```

#### Mount device

挂载设备将设备挂载到全局路径，然后各个容器可以动态绑定，只能从kubelet调用。

此调出不会传递Flexvolume规范中指定的“secrets”。如果您的驱动程序需要secrets，
请不要执行此调出，而是使用“mount”调出并在该调出中执行attach和调用。
```
<driver executable> mountdevice <mount dir> <mount device> <json options>
```

#### Unmount device

取消所有挂载，一旦所有绑定挂载已被卸载，就会调用它。只能从Kubelet中调用。
```
<driver executable> unmountdevice <mount device>
```

#### Mount

将卷挂载到挂载目录。此调出默认为绑定挂载实现attach和mount-device调出的驱动程序。只能从Kubelet中调用。

```
<driver executable> mount <mount dir> <json options>
```


#### Unmount

卸载卷，此调出默认为绑定挂载实现附加和挂载设备调出的驱动程序。只能从Kubelet中调用。

```
<driver executable> unmount <mount dir>
```

有关如何编写简单的flexvolume驱动程序的简单示例，请参阅[lvm](https://git.k8s.io/kubernetes/examples/volumes/flexvolume/lvm)＆[nfs](https://git.k8s.io/kubernetes/examples/volumes/flexvolume/nfs)。

### 驱动输出

Flexvolume希望驱动程序以下列格式返回操作状态。
```
{
	"status": "<Success/Failure/Not supported>",
	"message": "<Reason for success/failure>",
	"device": "<Path to the device attached. This field is valid only for attach & waitforattach call-outs>"
	"volumeName": "<Cluster wide unique name of the volume. Valid only for getvolumename call-out>"
	"attached": <True/False (Return true if volume is attached on the node. Valid only for isattached call-out)>
    "capabilities": <Only included as part of the Init response>
    {
        "attach": <True/False (Return true if the driver implements attach and detach)>
    }
}
```

### 默认json选项

除了用户在FlexVolumeSource的Options字段中指定的标志之外，还将以下标志传递给可执行文件。注意：秘密只传递给“mount/umount”调出

### Flexvolume的示例

有关如何在pod中使用Flexvolume的快速示例，请参见[nginx.yaml](https://git.k8s.io/kubernetes/examples/volumes/flexvolume/nginx.yaml)＆[nginx-nfs.yaml](https://git.k8s.io/kubernetes/examples/volumes/flexvolume/nginx-nfs.yaml)。



https://github.com/sigma/cifs_k8s_plugin
https://github.com/kubernetes/kubernetes/blob/master/examples/volumes/flexvolume/lvm
https://github.com/kubernetes/kubernetes/blob/master/examples/volumes/flexvolume/nfs

# CSI

## 介绍

CSI提供了一个单一的界面，存储供应商可以实现它们的存储解决方案，以跨多个不同的容器编排器工作，并且卷插件被设计为out-of-tree。这是一项巨大的努力，CSI的全面实施需要几个季度的时间，并且需要立即为存储供应商提供解决方案，以继续添加卷插件。
它使许多不同类型的存储系统能够：
- 在需要时自动创建存储。
- 使存储在任何计划的地方都可用。
- 不再需要时自动删除存储。

## 创建csi的原因
Kubernetes卷插件目前是“in-tree”，意味着它们与核心kubernetes二进制文件进行链接，编译，构建和发布。
为Kubernetes（卷插件）添加对新存储系统的支持需要将代码检入核心Kubernetes存储库。
但是与Kubernetes发布流程对许多插件开发者来说是痛苦的。

现有的Flex Volume插件试图通过暴露外部卷插件的基于exec的API来解决这个问题。
尽管它使第三方存储供应商能够在树外编写驱动程序，但为了部署第三方驱动程序文件，
它需要访问节点和主机的根文件系统。

除了难以部署之外，Flex并没有解决插件依赖的痛苦：插件往往有很多外部需求（例如，在挂载和文件系统工具上）。
假定这些依赖关系在底层主机操作系统上可用，而这往往不是这种情况（并且安装它们需要访问节点机器的根文件系统）。

CSI解决了所有这些问题，使存储插件能够通过标准的Kubernetes基元进行树外，容器化，部署，
并通过用户所熟悉并喜爱的Kubernetes存储原语（PersistentVolumeClaims，PersistentVolumes，StorageClasses）进行使用。

## csi驱动

https://kubernetes-csi.github.io/docs/Drivers.html

## 用法

### 启用csi

csi在1.9是alpha版本，要想使用它，设置以下参数：

- API server binary:
  - --feature-gates=CSIPersistentVolume=true
  - --runtime-config=storage.k8s.io/v1alpha1=true

- API server binary and kubelet binaries:

  - --feature-gates=MountPropagation=true
  - --allow-privileged=true

### 预配置volume
预先配置的驱动程序的工作方式与之前一样，管理员将创建一个PersistentVolume规范，该规范将描述要使用的卷。PersistentVolume规范需要根据你的驱动程序进行设置，这里的区别在于有一个叫做csi的新部分需要相应地设置。请参阅Kubernetes关于CSI卷的文档（LINK TBD）。

以下是由CSI驱动程序管理的预配置卷的PersistentVolume规范示例：

```
apiVersion: v1
kind: PersistentVolume
metadata:
  name: manually-created-pv
spec:
  capacity:
    storage: 5Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  csi:
    driver: com.example.team/csi-driver
    volumeHandle: existingVolumeName
    readOnly: false
```

### 动态预配
为了设置系统进行动态配置，管理员需要设置StorageClass指向CSI驱动程序的外部配置器并指定驱动程序所需的任何参数。这是一个StorageClass的例子：

```
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: fast-storage
provisioner: com.example.team/csi-driver
parameters:
  type: pd-ssd
```

提供者：必须设置为CSI驱动程序的名称
参数：必须包含特定于CSI驱动程序的任何参数。
然后用户可以使用这个StorageClass 创建一个PersistentVolumeClaim，如下所示：

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: request-for-storage
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: fast-storage
```

## 在k8s中使用
本节介绍如何部署csi驱动程序到k8s 1.9集群

在Kubernetes 1.9中，有三个新的组件加上kubelet使CSI驱动为Kubernetes提供存储。
新组件是边车容器，负责与Kubernetes和CSI驱动沟通，监控events从而适时调用csi接口。

### external-attacher

external-attacher是一个边车容器，用于监视Kubernetes VolumeAttachment对象并触发针对驱动程序端点的CSI ControllerPublish和ControllerUnpublish操作。在撰写本文时，外部助理不支持领导者选举，因此每个CSI车手只能运行一次。欲了解更多信息，请阅读附加和分离。

请注意，即使这称为外部附件，它的功能是调用CSI API调用ControllerPublish和ControllerUnpublish。这些调用很可能发生在不是将要安装卷的节点中。因此，许多CSI驱动程序不支持这些调用，而是在要安装的节点上的kubelet所完成的CSI NodePublish和NodeUnpublish调用中执行attach / detach和mount / unmount 。

### external-provisioner

external-provisioner是一个Sidecar容器，用于监视Kubernetes PersistentVolumeClaim对象并触发针对驱动程序端点的CSI CreateVolume和DeleteVolume操作。有关更多信息，请阅读供应和删除。

### driver-registrar

driver-registrar是一个边车容器，它用kubelet注册CSI驱动程序，并将驱动程序自定义NodeId添加到Kubernetes Node API对象上的标签。它通过与CSI驱动程序上的身份服务进行通信并调用CSI GetNodeId操作来完成此操作。驱动程序注册器必须具有通过环境变量设置的节点的Kubernetes名称，KUBE_NODE_NAME如下所示：

```
        - name: csi-driver-registrar
          imagePullPolicy: Always
          image: docker.io/k8scsi/driver-registrar
          args:
            - "--v=5"
            - "--csi-address=$(ADDRESS)"
          env:
            - name: ADDRESS
              value: /csi/csi.sock
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
```

### pod配置

```
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
            - name: mountpoint-dir
              mountPath: /var/lib/kubelet/pods
              mountPropagation: "Bidirectional"
      volumes:
        - name: socket-dir
          hostPath:
            path: /var/lib/kubelet/plugins/csi-hostpath
            type: DirectoryOrCreate
        - name: mountpoint-dir
          hostPath:
            path: /var/lib/kubelet/pods
            type: Directory
```

### rbac配置

```
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-hostpath-role
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["create", "delete", "get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
```


# 参考
https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md
https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/flexvolume-deployment.md
https://github.com/container-storage-interface/spec/blob/master/spec.md
https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface.md
http://blog.kubernetes.io/2018/01/introducing-container-storage-interface.html
https://github.com/container-storage-interface
https://kubernetes-csi.github.io/docs/


欢迎加入QQ群：k8s开发与实践（482956822）一起交流k8s技术