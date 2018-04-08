#  CNI-Genie

# 介绍

CNI-Genie 使 Kubernetes 无缝连接到一种 CNI 插件，例如：Flannel、Calico、Canal、Romana 或者 Weave。

# 安装

## daemonset 
- 1.7-版本

```bash
$ kubectl apply -f https://raw.githubusercontent.com/Huawei-PaaS/CNI-Genie/master/conf/1.5/genie.yaml
```

- 1.8+

```
$ kubectl apply -f https://raw.githubusercontent.com/Huawei-PaaS/CNI-Genie/master/conf/1.8/genie.yaml
```

## 源码安装

请注意，在更改源代码之前，应先安装genie。这确保了genie conf文件被成功生成。

在对源代码进行更改后，运行以下命令构建gen​​ie二进制文件

```
$ make all
```

将genie的二进制文件放在/opt/cni/bin/ 目录

```bash
$ cp dist/genie /opt/cni/bin/genie
```