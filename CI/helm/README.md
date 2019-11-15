# 介绍

Helm为团队提供了在Kubernetes内部创建，安装和管理应用程序时需要协作的工具。

有了Helm，可以：

- 查找要安装和使用的预打包软件（chart）
- 轻松创建和托管自己的软件包
- 将软件包安装到任何Kubernetes集群中
- 查询集群以查看已安装和正在运行哪些程序包
- 更新，删除，回滚或查看已安装软件包的历史记录
- 通过Helm，可以轻松在Kubernetes中运行应用程序。

# helm3
helm3于 2019。11.13发布,和helm2有以下区别

- 主要是移除了tiller
- oci支持
- go sdk

未来功能实现

- 增强helm test
- 对Helm OCI集成的改进
- Go客户端库的增强功能

# 示例

假设您有一个Kubernetes集群正在运行并且配置正确kubectl，使用Helm就是小菜一碟。

通过添加社区托管的存储库，Helm可以轻松地搜索新chart。

```
$ helm repo add nginx https://helm.nginx.com/stable
```

添加一些存储库后，您可以搜索chart：

```
$ helm search repo nginx-ingress
NAME                    CHART VERSION   APP VERSION     DESCRIPTION
nginx/nginx-ingress     0.3.7           1.5.7           NGINX Ingress Controller
```

Helm为您提供了一种使用以下方法安装该chart的快速方法helm install：

```
$ helm install my-ingress-controller nginx/nginx-ingress
```

如果我们使用以下命令检查集群kubectl：

```
$ kubectl get deployments
```

我们正在运行一个入口控制器！我们可以使用轻松删除它`helm uninstall my-ingress-controller`。


好的。您已经尝试了一些chart。您已经自定义了一些。现在您已经准备好构建自己的了。helm也使这一部分变得容易。

```
$ helm create diy
Creating diy
```