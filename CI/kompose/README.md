kompose 是一个将docker-compose迁移到kubernetes的工具，kompose会把Docker Compose文件翻译成Kubernetes资源文件

官方网站[http://kompose.io](http://kompose.io)。

kompose 是从本地Docker管理到使用Kubernetes管理您的应用程序的便利工具。 Docker的转换撰写格式到Kubernetes资源清单可能不是精确的，但会起到参考作用，尤其是初次在Kubernetes上部署应用程序


# 安装

## Linux and macOS

```
# Linux
curl -L https://github.com/kubernetes/kompose/releases/download/v1.19.0/kompose-linux-amd64 -o kompose

# macOS
curl -L https://github.com/kubernetes/kompose/releases/download/v1.19.0/kompose-darwin-amd64 -o kompose

chmod +x kompose
sudo mv ./kompose /usr/local/bin/kompose
```

## shell自动补全

```
# Bash (add to .bashrc for persistence)
source <(kompose completion bash)

# Zsh (add to .zshrc for persistence)
source <(kompose completion zsh)
```

# 用例说明

如果您有一个Docker Compose的docker-compose.yml文件或者一个Docker分布式应用捆绑包的docker-compose-bundle.dab文件，可以通过kompose命令将它们生成为Kubernetes的deplyment、service的资源文件，如下所示：

## 转换文件
```
$ kompose -f docker-compose.yml convert
WARN: Unsupported key networks - ignoring
file "redis-svc.yaml" created
file "web-svc.yaml" created
file "web-deployment.yaml" created
file "redis-deployment.yaml" created
```

## 直接启动

```
$ kompose up
We are going to create Kubernetes Deployments, Services and PersistentVolumeClaims for your Dockerized application. 
If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead. 

INFO Successfully created Service: redis          
INFO Successfully created Service: web            
INFO Successfully created Deployment: redis       
INFO Successfully created Deployment: web         

Your application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods,pvc' for details.
```