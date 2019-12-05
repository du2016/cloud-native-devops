# 介绍

一个准入控制插件是一段代码，它会在请求通过认证和授权之后、对象被持久化之前拦截到达
API server 的请求。插件代码运行在 API server 进程中，必须将其编译为二进制文件，
以便在此时使用。

在每个请求被集群接受之前，准入控制插件依次执行。
如果插件序列中任何一个拒绝了该请求，则整个请求将立即被拒绝并且返回一个错误给终端用户。

准入控制插件可能会在某些情况下改变传入的对象，从而应用系统配置的默认值。
另外，作为请求处理的一部分，准入控制插件可能会对相关的资源进行变更，
以实现类似增加配额使用量这样的功能


# 用途

- 对对象进行修改
- 设置对象默认值

# 插件列表

- AlwaysAdmit 通过所有请求
- AlwaysPullImages 总是重新拉去镜像，多租户环境通过这个插件实现镜像检查，因为重新拉取会重新验证imagepullsecret
- AlwaysDeny 拒绝所有请求，用于测试
- DenyEscalatingExec 不允许在特权容器中执行exec和attach
- ImagePolicyWebhook 通过webhook实现镜像的访问权限
  ```
    开启方式,api server配置以下参数
    --admission-control=ImagePolicyWebhook
    --admission-control-config-file=ac.json
  
    ac.json格式如下：
  
    {
      "imagePolicy": {
         "kubeConfigFile": "path/to/kubeconfig/for/backend",
         "allowTTL": 50,           // 缓存通过的秒数
         "denyTTL": 50,            // 缓存拒绝的秒数
         "retryBackoff": 500,      // 重试的间隔 毫秒
         "defaultAllow": true      // 请求webhook失败是否通过
      }
    }
     backend的kubeconfig格式如下
  
     # 指定webhook服务.
     clusters:
     - name: name-of-remote-imagepolicy-service
       cluster:
         certificate-authority: /path/to/ca.pem    # 指定CA.
         server: https://images.example.com/policy # 查询的接口，协议必须是 'https'.
     
     # 指定webhook的认证信息.
     users:
     - name: name-of-api-server
       user:
         client-certificate: /path/to/cert.pem # cert for the webhook plugin to use
         client-key: /path/to/key.pem          # key matching the cert
  ```
- ServiceAccount 自动创建serviceaccount
- SecurityContextDeny 禁止给pod配置安全上下文
- ResourceQuota 用来执行ResourceQuota资源配置策略，限制namespace中的资源总使用量
- LimitRanger 用来执行limitrange资源配置策略，限制namespace中单个pod资源的使用量
- InitialResources 自动给pod设置request和limit
- NamespaceLifecycle 不允许在已删除的命名空间中创建资源
- DefaultStorageClass 给没有设置StorageClass的pvc关联StorageClass
- DefaultTolerationSeconds 设置pod在不符合容忍策略的情况下 5分钟后被驱逐
- PodNodeSelector 通过命名空间注解或者全局配置限制命名空间中的容器可以在哪些节点上运行
- EventRateLimit  可以在命名空间和用户级别限制时间的产生速度
- ExtendedResourceToleration 对具有扩展资源的节点自动添加污点
- LimitPodHardAntiAffinityTopology 不允许pod定义kubernetes.io/hostname之外的requiredDuringSchedulingRequiredDuringExecution类型的反亲和性
- MutatingAdmissionWebhook 修改资源的admisionwebgook
- NamespaceAutoProvision 查看资源使用的命名空间是否存在，没有自动创建
- NamespaceExists 查看资源使用的命名空间是否存在，没有报错
- NodeRestriction 限制node可以修改的node和pod对象
  ```
  禁止操作node-restriction.kubernetes.io/前缀标签，
  允许操作以下标签
  kubernetes.io/hostname
  kubernetes.io/arch
  kubernetes.io/os
  beta.kubernetes.io/instance-type
  failure-domain.beta.kubernetes.io/region
  failure-domain.beta.kubernetes.io/zone
  kubelet.kubernetes.io/前缀标签
  node.kubernetes.io/前缀标签
  ```
- OwnerReferencesPermissionEnforcement 只有对metadata.ownerReferences具有“删除”权限的用户才能更改它
- PersistentVolumeClaimResize禁止显示声明allowVolumeExpansion之外的storageclass调整PVC大小
- PodPreset 该准入控制器使用匹配的PodPreset中指定的字段注入一个pod
- PodSecurityPolicy 根据安全上下文判断是否允许创建pod
- PodToleranceRestriction 检查pod的容忍策略和命名空间的容忍策略是否冲突，合并容忍策略到pod
- Priority 查看priorityClassName的数值，没有则拒绝
- RuntimeClass 设置pod的运行时，并且设置对应的pod.Spec.Overhead(运行pod时占用的基础资源开销)
- StorageObjectInUseProtection 确保删除PV或者PVC完成时再删除关联的PV或者PVC
- TaintNodesByCondition  将新创建的Node标记为NotReady和NoSchedule
- ValidatingAdmissionWebhook 验证资源的webook


查看准入控制列表：

```
kube-apiserver -h | grep enable-admission-plugins
```