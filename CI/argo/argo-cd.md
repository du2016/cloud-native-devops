# Argo CD-基于Kubernetes的声明式持续交付工具

## 什么是Argo CD？

Argo CD是用于Kubernetes的声明性GitOps连续交付工具。

![](http://img.rocdu.top/20200528/argocd-ui.gif)

## 为什么选择Argo CD？
应用程序定义，配置和环境应为声明性的，并受版本控制。应用程序部署和生命周期管理应该是自动化的，可审核的且易于理解的。

## argo cd 架构

Argo CD被实现为kubernetes控制器，该控制器连续监视正在运行的应用程序，
并将当前的活动状态与所需的目标状态（在Git存储库中指定）进行比较。
其活动状态偏离目标状态的已部署应用程序被标记为OutOfSync。
Argo CD报告并可视化差异，同时提供了自动或手动将实时状态同步回所需目标状态的功能。
在Git存储库中对所需目标状态所做的任何修改都可以自动应用并反映在指定的目标环境中。

![](http://img.rocdu.top/20200528/argocd_architecture.png)

## 支持的部署方式

- kustomize应用程序
- helm chat
- ksonnet应用
- jsonnet文件
- YAML / json清单的普通目录
- 任何配置为配置管理插件的自定义配置管理工具

# 使用argocd进行持续部署

## 安装argo cd

```
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

## 下载argocd cli

从github 下载最新版本的argo cli ` https://github.com/argoproj/argo-cd/releases/latest`

mac 环境可以这样安装：

```
brew tap argoproj/tap
brew install argoproj/tap/argocd
```

## 暴露argocd api/ui

```
kubectl port-forward svc/argocd-server -n argocd 8080:443
argo login 127.0.0.1:8080
```
默认密码为argocd-server pod的名称，如果重新生成，可以这样修改密码为`password`
```
kubectl -n argocd patch secret argocd-secret \\n  -p '{"stringData": {\n    "admin.password": "$2a$10$rRyBsGSHK6.uc8fntPwVIuLVHgsAhAX7TcdrqW/RADU0uh7CaChLa",\n    "admin.passwordMtime": "'$(date +%FT%T%Z)'"\n  }}'
```

## 注册集群到argocd

```
argocd cluster add #列出当前配置的上下文列表
argocd cluster add kubernetes-admin@kubernetes
```

## 从Git存储库创建应用程序并进行同步

```
argocd app create guestbook --repo https://github.com/argoproj/argocd-example-apps.git --path guestbook --dest-server https://kubernetes.default.svc --dest-namespace default
argocd app get guestbook
argocd app sync guestbook
```


## 通过argo server ui访问

在页面上可以看到各个资源的状态、配置、以及关联关系

![](http://img.rocdu.top/20200528/argocd-ui.pbg.png)

扫码关注我:

![微信](http://img.rocdu.top/20200528/qrcode_for_gh_7457c3b1bfab_258.jpg)