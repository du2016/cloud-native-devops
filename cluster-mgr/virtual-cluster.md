# k8s虚拟集群概览

# 前言

随着组织内部越来越多地使用Kubernetes，对应用程序和工程师进行Kubernetes访问的需求也在增长.由于始终使用整个物理Kubernetes集群既不可行也不具有成本效益，因此Kubernetes的虚拟化是显而易见的解决方案.

# VirtualCluster概述

VirtualCluster代表了一种新架构，可应对各种Kubernetes控制平面隔离挑战. 它通过为每个租户提供一个集群视图来扩展现有的基于命名空间的Kubernetes多租户模型. VirtualCluster完全利用了Kubernetes的可扩展性，并保留了完整的API兼容性. 话虽如此，核心Kubernetes组件并未在虚拟集群中进行修改.

使用VirtualCluster，每个租户都被分配了一个专用的租户主机，这是上游Kubernetes发行版. 租户可以在租户主机中创建群集作用域资源，例如名称空间和CRD，而不会影响其他资源. 结果，由于共享一个apiserver而导致的大多数隔离问题消失了. 管理实际物理节点的Kubernetes集群称为超级主节点，现在成为Pod资源提供者. VirtualCluster由以下组件组成：

- vc-manager：引入了新的CRD VirtualCluster对租户主机进行建模. vc-manager管理每个VirtualCluster自定义资源的生命周期. 根据规范，它可以在本地K8s集群中创建apiserver，etcd和controller-manager Pod，或者如果提供有效的kubeconfig则导入现有集群.

- syncer：一个集中式控制器，可将Pod设置所需的API对象从每个租户主机填充到超级主机，并双向同步对象状态. 它还定期扫描已同步的对象，以确保租户主机和超级主机之间的状态一致.

- vn-agent：一个节点守护程序，它将所有租户kubelet API请求代理到在节点中运行的kubelet进程. 它确保每个租户只能在节点中访问其自己的Pod.

综上所述，从租户的角度来看，每个租户主机的行为就像完整的Kubernetes，具有几乎完整的API功能.

# 功能及限制

VirtualCluster遵循无服务器设计模式.超级主节点拓扑未在租户主中完全公开.租户主机中仅显示正在运行的租户Pod的节点.结果，VirtualCluster在租户主服务器中不支持类似DaemonSet的工作负载.换句话说，如果规范中已设置其节点名，则同步器控制器将拒绝新创建的承租人Pod.

建议将租户主节点控制器--node-monitor-grace-period参数增加到更大的值(> 60秒，已在示例clusterversion yaml中完成).同步器控制器不会更新租户主机中的节点租用对象，因此默认宽限期太短.

Coredns不支持租户.因此，如果需要DNS，租户应在租户主机中安装coredns. DNS服务应使用名称kube-dns在kube-system命名空间中创建.然后，同步器控制器可以识别超级主服务器中的DNS服务群集IP，并将其注入到Pod spec dnsConfig中.

VirtualCluster使用自定义的coredns构建支持租户DNS服务.有关详细信息，请参见此文档.

VirtualCluster完全支持租户服务帐户.

VirtualCluster不支持租户PersistentVolume.所有PV和存储类均由超级主机提供.

建议租户主机和超级主机使用相同的Kubernetes版本，以避免API行为不兼容.同步器控制器和vn-agent使用Kubernetes 1.16 API构建，因此目前不支持更高版本的Kubernetes.

# 编译安装

编译 kubectl-vc

```
git clone https://github.com/kubernetes-sigs/multi-tenancy
make build WHAT=cmd/kubectl-vc GOOS=darwin
cp -f _output/bin/kubectl-vc /usr/local/bin
```

安装对应的CRD

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/multi-tenancy/master/incubator/virtualcluster/config/crds/tenancy.x-k8s.io_clusterversions.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/multi-tenancy/master/incubator/virtualcluster/config/crds/tenancy.x-k8s.io_virtualclusters.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/multi-tenancy/master/incubator/virtualcluster/config/setup/all_in_one.yaml
```

我使用的是kind安装的k8s集群没有使用推荐的1.6版本，所以需要修改，tenancy.x-k8s.io_clusterversions.yaml文件，手动将CRD protocol 设置为 required

创建clusterversion CR 
clusterversion CR指定一个租户主配置，vc-manager可以使用它来创建租户主组件. 以下cmd将创建一个cv-sample-np clusterversion CR，该CR为Kubernetes 1.15 apiserver，etcd和控制器管理器分别指定三个StatefulSet.

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/multi-tenancy/master/incubator/virtualcluster/config/sampleswithspec/clusterversion_v1_nodeport.yaml
```

# 创建虚拟集群

## 创建

```
kubectl vc create -f https://raw.githubusercontent.com/kubernetes-sigs/multi-tenancy/master/incubator/virtualcluster/config/sampleswithspec/virtualcluster_1_nodeport.yaml -o vc-1.kubeconfig
```

因为我使用的是mac上安装的kind,需要通过port-forward 转发端口

```
kubectl port-forward service/apiserver-svc 6443:6443 -n default-cf0191-vc-sample-1
```

修改vc-1.kubeconfig中的apiserver地址为https://127.0.0.1:6443

## 查看集群信息

```
kubectl cluster-info --kubeconfig vc-1.kubeconfig
Kubernetes master is running at https://127.0.0.1:6443

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

## 查看集群node

```
kubectl get node --kubeconfig vc-1.kubeconfig
No resources found in default namespace.
```

使用跟集群配置查看命名空间，可以看到我们的虚拟命名空间名称

```
 kubectl get ns
NAME                                         STATUS   AGE
default                                      Active   24h
default-cf0191-vc-sample-1                   Active   51m
default-cf0191-vc-sample-1-default           Active   49m
default-cf0191-vc-sample-1-kube-node-lease   Active   49m
default-cf0191-vc-sample-1-kube-public       Active   49m
default-cf0191-vc-sample-1-kube-system       Active   49m
kube-node-lease                              Active   24h
kube-public                                  Active   24h
kube-system                                  Active   24h
local-path-storage                           Active   24h
scas-c                                       Active   24h
vc-manager                                   Active   66m
```

## 在虚拟集群中创建服务

```
kubectl apply --kubeconfig vc-1.kubeconfig -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  labels:
    app: vc-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vc-test
  template:
    metadata:
      labels:
        app: vc-test
    spec:
      containers:
      - name: poc
        image: busybox
        command:
        - top
EOF
```

查看在虚拟集群中部署的服务

```
kubectl get pod --kubeconfig vc-1.kubeconfig
NAME                         READY   STATUS    RESTARTS   AGE
test-deploy-5f4bcd8c-cdw9h   1/1     Running   0          45m
```