# kubeadm 

#### 添加yum源

```
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
```

#### master初始化

```
初始化
#kubeadm init
指定参数
#kubeadm init --kubernetes-version=v1.9.0 --pod-network-cidr=10.244.0.0/16
```


### 取消污点

```
kubectl taint nodes --all node-role.kubernetes.io/master-
```

### 命令参数

https://k8smeetup.github.io/docs/reference/generated/kubeadm/
