# kubectl(命令行工具)二进制文件获取

#### mac 

```
# 最新版本：
VERSION=`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`
# 指定版本：
VERSION=v1.9.0


curl -LO https://storage.googleapis.com/kubernetes-release/release/${VERSION}/bin/darwin/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl

```

#### linux

```
# 最新版本：
VERSION=`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`
# 指定版本：
VERSION=v1.9.0

curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```

#### win

```
curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.9.0/bin/windows/amd64/kubectl.exe
```

kubectl更多[获取方式](https://kubernetes.io/docs/tasks/tools/install-kubectl/)


# 各组件二进制获取

[官网文档](https://kubernetes.io/docs/getting-started-guides/binary_release/)

VERSION=`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`

#### 预构建版(github获取)
```
#版本列表 https://github.com/kubernetes/kubernetes/releases
wget https://github.com/kubernetes/kubernetes/releases/download/${VERSION}/kubernetes.tar.gz
tar xf kubernetes.tar.gz
./cluster/get-kube-binaries.sh
```

该脚本实际上下载了以下两个tar包

```
#推荐
VERSION=`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`
wget https://dl.k8s.io/${VERSION}/kubernetes-client-linux-amd64.tar.gz
wget https://dl.k8s.io/${VERSION}/kubernetes-server-linux-amd64.tar.gz
tar xf kubernetes-client-linux-amd64.tar.gz && tar xf kubernetes-server-linux-amd64.tar.gz
find ./kubernetes -perm  /111 -and -type f -exec  cp  {} /usr/local/bin/ \;
```

#### 源码编译

```
git clone https://github.com/kubernetes/kubernetes.git
cd kubernetes
make release
```

#### 下载并自动启动一个k8s集群
```
# wget version
export KUBERNETES_PROVIDER=YOUR_PROVIDER; wget -q -O - https://get.k8s.io | bash

# curl version
export KUBERNETES_PROVIDER=YOUR_PROVIDER; curl -sS https://get.k8s.io | bash
```