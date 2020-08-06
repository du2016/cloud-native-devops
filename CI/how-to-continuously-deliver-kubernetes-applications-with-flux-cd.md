# 如何使用Flux CD持续交付Kubernetes应用程序

> 适用于Kubernetes工作负载的GitOps

Flux CD是一个连续交付工具，正在迅速普及。Weaveworks最初开发了该项目，然后将其开源到CNCF.

它成功的原因是它可以感知Kubernetes变化并且易于设置。它提供的最亮点的功能是，它允许团队以声明方式管理其Kubernetes部署。
Flux CD通过定期轮询存储库来将存储在源代码存储库中的Kubernetes manifests文件与Kubernetes集群同步，
因此团队无需担心运行kubectl命令和监视环境以查看他们是否部署了正确的工作负载。
相反，Flux CD确保Kubernetes集群始终与源代码存储库中定义的配置保持同步。


它使团队可以实现GitOps，它具有以下原则：
- Git是的唯一的真实来源。
- Git是操作所有环境的唯一场所，所有配置都是代码。
- 所有更改都是可观察/可验证的。


# 为什么使用 FLUX CD？

使用Kubernetes的传统CI/CD部署遵循以下模式：

![](http://img.rocdu.top/20200601/0*hb9tlQkodgmRCaT5.png)


- 开发人员创建代码并编写Dockerfile。他们还为应用程序创建Kubernetes manifests和Helm Charts。
- 他们将代码推送到源代码存储库。
- 源代码存储库使用提交后的钩子触发Jenkins构建。
- Jenkins CI流程将构建Docker映像和Helm软件包，并将其推送到依赖仓库。
- 然后，Jenkins CD程序部署helm charts到k8s cluster。

这个过程听起来合理，或多或少是行业标准。但是，有一些限制：
- 您需要将Kubernetes 凭据存储在Jenkins服务器中。由于服务器是共享的，这是折中的做法。
- 尽管您可以使用Jenkins创建和更改配置，但无法使用它删除现有资源。例如，如果您从存储库中删除清单文件，则kubectl不会将其从服务器中删除。这是自动化GitOps的最大障碍。


# Flux CD如何工作

Flux CD允许团队以声明方式使用YAML清单指定所有必需的Kubernetes配置。
- 团队编写Kubernetes manifests并将其推送到源代码存储库。
- memcached pod存储当前配置。
- Flux定期（默认为五分钟）使用Kubernetes operator轮询存储库以进行更改。Flux容器将其与memcached中的现有配置进行比较。
如果检测到更改，它将通过运行一系列kubectl apply/delete命令将配置与集群同步。然后，它将最新的元数据再次存储在memcached存储中。

![](http://img.rocdu.top/20200601/1*HauufXN9jQkfKLCIH78KNg.png)

另外，如果要自动升级工作负载，Flux CD允许您轮询docker registry,并使用最新镜像更新Git存储库中的Kubernetes manifests。
由于Flux CD以Kubernetes operator的身份运行，因此设置非常简单，而且启动也很轻松。
让我们看一下动手演示，以便我们更好地理解它。

# 先决条件

确保您具有一个正在运行的Kubernetes集群，并具有cluster-admin的角色以部署Flux CD 。

# 安装fluxctl
Flux CD提供了一个fluxctl二进制文件，可以帮助您在Kubernetes集群中部署和管理Flux CD。
下载的最新版本fluxctl并将其移动到/usr/bin目录中。

```
$ wget https://github.com/fluxcd/flux/releases/download/1.19.0/fluxctl_linux_amd64
$ mv fluxctl_linux_amd64 /usr/bin/fluxctl
$ sudo chmod +x /usr/bin/fluxctl
```

在此示例中，让我们使用GitHub作为源代码存储库。forck [bharatmicrosystems/nginx-kubernetes](https://github.com/bharatmicrosystems/nginx-kubernetes)存储库到您的GitHub的帐户中。

该存储库包含目录中的清单nginx-deployment和nginx-service清单以及workloads目录中的web名称空间定义namespaces。

```
├─ namespaces
│  └─ web-ns.yaml
├─ workloads
│  ├─ nginx-deployment.yaml
│  └─ nginx-service.yaml
├─ .gitignore
├─ LICENSE
└─ README.md
```

在GHUSER环境变量中提供GitHub用户的名称，在环境变量中提供GitHub存储库GHREPO，如下所示。创建一个名为的新名称空间，flux并在Kubernetes集群中安装Flux CD操作符。

该fluxctl install命令根据以下选项生成所需的Kubernetes清单：

- git-user— Git用户。在这种情况下，GitHub用户名
- git-email — Git用户电子邮件。在这种情况下，默认的GitHub电子邮件
- git-url — Git存储库的URL
- git-path — Git存储库中用于同步更改的目录
- namespace —部署flux运算符的名称空间

```
$ export GHUSER="<YOUR_GITHUB_USER>"
$ export GHREPO="<YOUR_GITHUB_REPO>"
$ kubectl create ns flux
namespace/flux created
$ fluxctl install \
--git-user=${GHUSER} \
--git-email=${GHUSER}@users.noreply.github.com \
--git-url=git@github.com:${GHUSER}/${GHREPO} \
--git-path=namespaces,workloads \
--namespace=flux | kubectl apply -f -
service/memcached created
serviceaccount/flux created
clusterrole.rbac.authorization.k8s.io/flux created
clusterrolebinding.rbac.authorization.k8s.io/flux created
deployment.apps/flux created
secret/flux-git-deploy created
deployment.apps/memcached created
```

检查Flux部署是否成功。

```
$ kubectl -n flux rollout status deployment/flux
deployment "flux" successfully rolled out
```

让我们获取flux名称空间中的所有资源，以查看对象的当前状态。
如您所见，有一个flux吊舱和一个memcached吊舱。吊舱需要与之交互时，还有一项memcached服务flux。

```
$ kubectl get all -n flux
NAME                             READY   STATUS    RESTARTS   AGE
pod/flux-86d86b868-lndhn         1/1     Running   0          2m
pod/memcached-86869f57fd-qwnts   1/1     Running   0          2m
NAME                TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
service/memcached   ClusterIP   10.8.11.199   <none>        11211/TCP   2m
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/flux        1/1     1            1           2m
deployment.apps/memcached   1/1     1            1           2m
NAME                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/flux-86d86b868       1         1         1       2m
replicaset.apps/memcached-86869f57fd 1         1         1       2m
```

# 授权Flux CD连接到您的Git存储库

现在，我们需要允许Flux CD操作员与Git存储库进行交互，因此，我们需要将其公共SSH密钥添加到存储库中。
使用获取公共SSH密钥fluxctl。

```
$ fluxctl identity --k8s-fwd-ns flux
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCryxSADyA+GIxtyCwpO3R9EuRcjZCqScKbYO246LZknyeluxKz0SlHYZHrlqxvla+k5GpPqnbImLLhuAD+YLzn0DbI58hUZLsrvxPWKiku--REDACTED--MKoPyEtQ+JiR3ZiADx6Iq8tYRRR+WBs1k5Hc8KNpg+FSRP8I8+CJRkCG4JQacPwK8FESP4qr1dxVv1tE8ZXyb8CdiToKpK7Mkc= root@flux-b9b4cc4f9-p9w88
```


将SSH密钥添加到您的存储库中，以便Flux CD可以访问它。
转到https://github.com/<YOUR_GITHUB_USER>/nginx-kubernetes/settings/keys
在标题部分的密钥中添加一个名称。
将SSH密钥粘贴到“密钥”部分。
选中“允许写访问权限”。

![](http://img.rocdu.top/20200601/1*yGovA5t5l7TE2WVJ0bpZhg.png)

Flux CD每五分钟自动与配置的Git存储库同步一次。但是，如果要立即将Flux与Git存储库同步，则可以使用fluxctl sync，如下所示。

```
$ fluxctl sync --k8s-fwd-ns flux
Synchronizing with ssh://git@github.com/bharatmicrosystems/nginx-kubernetes
Revision of master to apply is 8db9163
Waiting for 8db9163 to be applied ...
Done.
```

现在让我们看一下Pod，看看是否有两个Nginx副本。

```
$ kubectl get pod -n web
NAME                                READY   STATUS    RESTARTS   AGE
nginx-deployment-7fd6966748-lj8zd   1/1     Running   0          20s
nginx-deployment-7fd6966748-rbxqs   1/1     Running   0          20s
```

获得该服务，您应该会看到一个暴露80端口的nginx负载均衡器服务。如果您的Kubernetes集群可以动态加载负载平衡器，那么您应该会看到一个外部IP。

```
$ kubectl get svc -n web
NAME            TYPE           CLUSTER-IP   EXTERNAL-IP      PORT(S)        AGE
nginx-service   LoadBalancer   10.8.10.33   35.222.174.212   80:30609/TCP   94s
```

使用外部IP测试服务。如果您的群集不支持负载均衡器，则可以使用NodeIP:NodePort组合。

```
$ curl http://35.222.174.212/
This is version 1
```

从workloads/nginx-deployment.yaml上更新镜像bharamicrosystems/nginx:v2

```
$ sed -i "s/nginx:v1/nginx:v2/g" workloads/nginx-deployment.yaml
$ git add --all
$ git commit -m 'Updated version to v2'
$ git push origin master
```

现在，让我们等待五分钟以进行自动同步，同时，查看pod的更新。

```
$ watch -n 30 'kubectl get pod -n web'
NAME                                READY   STATUS        RESTARTS   AGE
nginx-deployment-5db4d6cb84-8lbsk   1/1     Running       0          11s
nginx-deployment-5db4d6cb84-qc6jp   1/1     Running       0          10s
nginx-deployment-6784c95fc7-zqptk   0/1     Terminating   0          6m43s
```

如您所见，旧的pod正在终止，新pod正在滚动更新。检查pod状态，以确保所有pod都在运行。

```
$ kubectl get pod -n web
NAME                                READY   STATUS    RESTARTS   AGE
nginx-deployment-5db4d6cb84-8lbsk   1/1     Running   0          1m
nginx-deployment-5db4d6cb84-qc6jp   1/1     Running   0          1m
```


现在，让我们再次调用该服务。

```
$ curl http://35.222.174.212/
This is version 2
```


如您所见，该版本现已更新为v2。
恭喜你！您已经在Kubernetes集群上成功设置了Flux CD。


# 结论

Flux是声明式地将Git存储库中的Kubernetes配置与集群进行同步的最轻量的方法之一，尤其是从GitOps着手时。

原文：https://medium.com/better-programming/how-to-continuously-deliver-kubernetes-applications-with-flux-cd-502e4fb8ccfe


扫描关注我:

![微信](http://img.rocdu.top/20200601/qrcode_for_gh_7457c3b1bfab_258.jpg)