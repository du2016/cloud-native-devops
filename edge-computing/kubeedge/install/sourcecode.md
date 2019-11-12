# 依赖

- [安装docker](../../../docker/install.md) 
> 注意docker的cgroup确定和集群kubelet的要一致
- [安装kubeadm/kubectl](../../../install/kubeadm-install/README.md)
- [创建k8s集群](../../../install/kubeadm-install/single-control-plane.md)
- [安装golang](../../../golang/install.md)
- 在每个edge节点安装mosquitto

> 如果边缘节点为centos可以只直接yum安装,如果为其他系统参见[官方文档](https://mosquitto.org/download/), centos执行以下命令:
> `yum install epel* -y && yum install -y mosquitto && mosquitto -d -p 1883`

# 运行kubeedge

## 初始化云端

### 克隆kubeedge

```
git clone https://github.com/kubeedge/kubeedge.git $GOPATH/src/github.com/kubeedge/kubeedge
cd $GOPATH/src/github.com/kubeedge/kubeedge
```

### 生成证书

要为KubeEdge进行设置，需要RootCA证书和一个证书/密钥对。云和边缘都可以使用相同的证书/密钥对。

```
$GOPATH/src/github.com/kubeedge/kubeedge/build/tools/certgen.sh genCertAndKey edge
```

### 二进制运行

- 查看gcc是否安装
```
gcc --version
```

- 编译cloudcore

```
cd $GOPATH/src/github.com/kubeedge/kubeedge/
make all WHAT=cloudcore
```
> 编译时因为墙的问题，可能拉不下来，可以使用vendor，根据屏幕输出的go build命令添加-mod vendor参数即可

- 创建设备模块和设备CRD
```
cd $GOPATH/src/github.com/kubeedge/kubeedge/build/crds/devices
kubectl create -f devices_v1alpha1_devicemodel.yaml
kubectl create -f devices_v1alpha1_device.yaml
```

- 复制cloudcore二进制文件和配置文件

```
cd $GOPATH/src/github.com/kubeedge/kubeedge/cloud
# run edge controller
# `conf/` should be in the same directory as the cloned KubeEdge repository
# verify the configurations before running cloud(cloudcore)
mkdir -p ~/cmd/conf
cp cloudcore ~/cmd/
cp -rf conf/* ~/cmd/conf/
```


> ~/cmd/dir是一个示例，在以下示例中，我们继续使用~/cmd/作为二进制启动目录。您
可以将cloudcore或edgecore二进制文件移动到任何地方，但需要在与二进制文件相同的目录中创建conf目录。

> 如果cloud和edge是同一台机器请注意和edgecore分开目录

- 配置

```
cd ~/cmd/conf
vim controller.yaml
```

大多默认配置即可，具体内容不再展示，在实验过程中发现 "~/"不能识别改成绝对路径就好了

- 运行

```
cd ~/cmd/
nohup ./cloudcore &
```

- 使用systemd启动cloudcore

也可以使用systemd启动cloudcore。如果需要，可以使用示例systemd-unit-file。以下命令将向您展示如何进行设置:

```
sudo ln build/tools/cloudcore.service /etc/systemd/system/cloudcore.service
sudo systemctl daemon-reload
sudo systemctl start cloudcore
```

> 请在cloudcore.service中修改ExecStart路径。不要使用相对路径，而要使用绝对路径

### 部署edge节点

我们提供了一个示例node.json来在kubernetes中添加一个节点。请确保在Kubernetes中添加了Edge节点。运行以下步骤以添加边缘节点

- 复制`$GOPATH/src/github.com/kubeedge/kubeedge/build/node.json`并且更改 metadata.name 为自己的边缘节点名称

```
mkdir ~/cmd/yaml
cp $GOPATH/src/github.com/kubeedge/kubeedge/build/node.json ~/cmd/yaml
```

- 确保该节点的角色设置为edge。为此，形式为`node-role.kubernetes.io/edge`的标签必须设置

- 请确保将标签`node-role.kubernetes.io/edge`添加到`build/node.json`文件中
```
{
  "kind": "Node",
  "apiVersion": "v1",
  "metadata": {
    "name": "edge-node",
    "labels": {
      "name": "edge-node",
      "node-role.kubernetes.io/edge": ""
    }
  }
}
```

> name需要用到，edgecore上报的名称需要与这里一致

- 如果未为节点设置角色，则无法在云中创建/更新的pod，configmap和secret与它们所针对的节点同步
- 部署edge node,只是创建了节点，并没有状态，状态依赖于edgecore向cloudcore上报，cloudcore将更改节点状态

```
kubectl apply -f ~/cmd/yaml/node.json
```

## 初始化边缘侧

### 克隆KubeEdge

```
git clone https://github.com/kubeedge/kubeedge.git $GOPATH/src/github.com/kubeedge/kubeedge
cd $GOPATH/src/github.com/kubeedge/kubeedge
```

### 运行edge

#### 二进制运行

- 构建edge

```
cd $GOPATH/src/github.com/kubeedge/kubeedge
make all WHAT=edgecore
```

- 配置edgecore

```
mkdir ~/cmd/conf
cp $GOPATH/src/github.com/kubeedge/kubeedge/edge/conf/* ~/cmd/conf
vim ~/cmd/conf/edge.yaml
```

> edgecore和cloudcore在一台，默认配置即可，如果不在一台请修改edgehub对应配置

- 运行 

```
cp $GOPATH/src/github.com/kubeedge/kubeedge/edge/edgecore ~/cmd/
cd ~/cmd
./edgecore
# or
nohup ./edgecore > edgecore.log 2>&1 &
```

- systemd 启动

```
sudo ln build/tools/edgecore.service /etc/systemd/system/edgecore.service
sudo systemctl daemon-reload
sudo systemctl start edgecore
sudo systemctl enable edgecore
```

- 验证 

```
kubectl get nodes
```

# 在云端部署应用

请按照以下步骤尝试部署示例应用程序。

```
kubectl apply -f $GOPATH/src/github.com/kubeedge/kubeedge/build/deployment.yaml
```

- 验证服务是否运行在边缘节点及状态是否正常

```
kubectl get pods
```