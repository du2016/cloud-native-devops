
一个pod安全策略是一个群集级别的资源，该pod规范的控制安全敏感的方面。
这些PodSecurityPolicy对象定义了Pod必须运行的一组条件才能被系统接受，
以及相关字段的默认值。它们允许管理员控制以下各项:

控制方面 | 字段名称
--------|--------
运行特权容器 | privileged
主机名称空间的用法 | hostPID, hostIPC
主机网络和端口的使用 | hostNetwork, hostPorts
卷类型的用法 | volumes
主机文件系统的用法 | allowedHostPaths
FlexVolume驱动程序白名单 | allowedFlexVolumes
分配拥有Pod的卷的FSGroup | fsGroup
要求使用只读根文件系统 | readOnlyRootFilesystem
容器的用户和组ID | runAsUser, runAsGroup, supplementalGroups
将升级限制为root特权 | allowPrivilegeEscalation, defaultAllowPrivilegeEscalation
Linux功能 | defaultAddCapabilities, requiredDropCapabilities, allowedCapabilities
容器的SELinux上下文 | seLinux
容器的允许的Proc Mount类型 | allowedProcMountTypes
容器使用的AppArmor配置文件 | 注解实现
容器使用的seccomp配置文件 | 注解实现
容器使用的sysctl配置文件 | forbiddenSysctls,allowedUnsafeSysctls

# 创建策略

```
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: example
spec:
  privileged: false
  # The rest fills in some required fields.
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  runAsUser:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
  - '*'
```
kubectl-admin create -f example-psp.yaml

# 授权

kubectl-admin create role psp:unprivileged \
    --verb=use \
    --resource=podsecuritypolicy \
    --resource-name=example

kubectl-admin create rolebinding fake-user:psp:unprivileged \
    --role=psp:unprivileged \
    --serviceaccount=psp-example:fake-user
    
kubectl --as=system:serviceaccount:psp-example:fake-user -n psp-example auth can-i use podsecuritypolicy/example

# 创建pod

```
kubectl-user create -f- <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: pause
spec:
  containers:
    - name:  pause
      image: k8s.gcr.io/pause
EOF
```

如果添加      securityContext.privileged: true 则依旧创建失败，因为当前策略不允许创建特权容器

# 默认的安全策略

```
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: privileged
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: '*'
spec:
  privileged: true
  allowPrivilegeEscalation: true
  allowedCapabilities:
  - '*'
  volumes:
  - '*'
  hostNetwork: true
  hostPorts:
  - min: 0
    max: 65535
  hostIPC: true
  hostPID: true
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```
    
# 设置默认PodSecurityPolicy
当我们启用PodSecurityPolicy时pod必须要有PodSecurityPolicy才能创建，为了创建一般的普通容器，我们可以创建普通PodSecurityPolicy后
关联到命名空间的serviceaccount，这样就不需要每次指定serviceaccount了

```
kubectl -n psp-example create rolebinding default:psp:unprivileged \
    --role=psp:unprivileged \
    --serviceaccount=psp-example:default
```


[apparmor百度百科](https://baike.baidu.com/item/apparmor/8178991?fr=aladdin)

[Seccompwiki](https://en.wikipedia.org/wiki/Seccomp)