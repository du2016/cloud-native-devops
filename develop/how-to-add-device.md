# 设备插件

从1.8版本开始，Kubernetes 提供一种 设备插件框架，
该框架能够让第三方设备资源开发者在不修改 Kubernetes 核心代码
的前提下将设备接入 Kubernetes，从而使得 Kubernetes 能够使用第三方设备资源。
通过该框架，第三方开发者可以实现一种通过手工或者 DaemonSet 部署的设备插件。
这些设备一般包括如 GPU、高性能NIC、FPGA 和 InfiniBand 等需要第三方开发
者进行特定初始化和设置的计算资源

[不透明资源管理](/object/pods/pod-opaque-integer-resource.md)
# 插件注册

- 开启DevicePlugins特性
- 向grpc注册

```
service Registration {
	rpc Register(RegisterRequest) returns (Empty) {}
}
```

- 注册信息
  - 设备的 Unix Socket 文件名。
  - 设备插件开发时对应的 Kubernetes API 版本。
  - 将要注册的 ResourceName。这里的 ResourceName 应该遵循 可扩展资源命名规则，一般采用 第三方域名/资源名 的方式。如 Nvidia GPU 的命名为 nvidia.com/gpu。
  
当成功注册资源后，设备插件还需要向 kubelet 发送其可用设备的列表，
之后 kubelet 会将这些资源作为节点状态的一部分上报给 API server。
例如，有一个名为 vendor-domain/foo 的设备插件向 kubelet 进行注册，并发送两个可用设备的列表，这时候查看 node status 就能看到有两个可用的 vendor-domain/foo 设备了。

然后，开发人员可以在申请 容器资源 时使用和 不透明整数资源 相同的过程来请求第三方设备资源。在1.8版本中，可扩展资源只支持按整数分配，并且在容器规格定义中的 limit 必须和 request 相等。

# 设备插件的实现

- 初始化。在这个阶段，设备插件将执行自身特定的初始化和配置操作，以确保设备正常运行。
- 插件将启动一个gRPC服务，并将其监听的 Unix Socket 文件放在宿主机的 /var/lib/kubelet/device-plugins/ 目录，该服务需要实现以下接口：

  ```
  service DevicePlugin {
        // ListAndWatch returns a stream of List of Devices
        // Whenever a Device state change or a Device disapears, ListAndWatch
        // returns the new list
        rpc ListAndWatch(Empty) returns (stream ListAndWatchResponse) {}
  
        // Allocate is called during container creation so that the Device
        // Plugin can run device specific operations and instruct Kubelet
        // of the steps to make the Device available in the container
        rpc Allocate(AllocateRequest) returns (AllocateResponse) {}
  }
  ```
  
- 插件通过宿主机的 Unix Socket 文件向 kubelet 注册，该文件位于：/var/lib/kubelet/device-plugins/kubelet.sock
- 当注册成功后，设备插件需要处于服务模式，在此期间需要监控设备的健康状态并且将设备状态的变动上报给 kubelet。同时，设备插件还负责处理 Allocate gRPC 请求。在 Allocate 过程中，设备插件可能会进行一些特定的准备工作，例如 GPU 的清理或者 QRNG 初始化等等。当 Allocate 操作成功后，第三方资源返回一个 AllocateResponse 给 kubelet，该返回包括能够访问设备的容器运行平台配置，然后 kubelet 将这些信息发送给容器运行平台

当 kubelet 重启时，设备插件必须能够感知并且重新发起注册。在1.8版本中，kubelet 启动时将会清空 /var/lib/kubelet/device-plugins 目录下的所有 Unix Socket 文件，设备插件可以通过监控自己的 Unix Socket 文件是否被删除从而重新发起注册。

# 部署

设备插件能够通过手工或者 DaemonSet 部署。通过 DaemonSet 部署的优势在于，当设备插件自身运行出错时，Kubernetes 将会自动重启该设备插件。否则，设备插件需要实现额外的机制来确保其自身的错误恢复。同时，对于指定的 /var/lib/kubelet/device-plugins 目录需要特定的访问权限，所以设备插件需要确保拥有这些访问权限。如果以 DaemonSet 的模式部署，/var/lib/kubelet/device-plugins 目录必须要在 PodSpec 中以 Volume 方式进行挂载。

# 样例

https://github.com/GoogleCloudPlatform/container-engine-accelerators
