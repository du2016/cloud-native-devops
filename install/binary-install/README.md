# 前提 

#### centos 7 mini安装
```
yum install wget lrzsz mlocate vim -y
```

#### centos 7 默认使用firewall管理防火墙，我还是习惯使用iptables.
```
yum remove firewalld -y 
```
#### 禁用selinux
```
setenforce 0
sed -i 's/enforcing/disabled/g' /etc/selinux/config
```
### 卸载firewalld
```
yum remove firewalld -y 
```

#### 安装高版本内核
```
rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org
rpm -Uvh http://www.elrepo.org/elrepo-release-7.0-2.el7.elrepo.noarch.rpm
yum --enablerepo=elrepo-kernel install kernel-ml-devel kernel-ml -y
#查看已安装内核版本
awk -F\' '$1=="menuentry " {print $2}' /etc/grub2.cfg
#设置为新版本内核
grub2-set-default 0
#重启
reboot
#验证是否安装成功
uname -r
# 因为该源是会更新内核版本的，为了保证内核版本统一，最好下载rpm包到yum源
yum  --downloadonly  --enablerepo=elrepo-kernel install kernel-ml-devel kernel-ml
#更新yum源
createrepo
```

#### 安装docker等依赖
```
yum install epel* 
yum install flanneld conntrack-tools docker iptables-services -y
```

#### 修改docker的日志引擎、默认镜像源、存储驱动
overlay需要3.18以上内核，overlay2需要4.0以上内核，关于[overlay存储说明](https://docs.docker.com/engine/userguide/storagedriver/overlayfs-driver/)
```
cat > /etc/sysconfig/docker <<EOF
# /etc/sysconfig/docker

# Modify these options if you want to change the way the docker daemon runs
OPTIONS='--selinux-enabled --log-driver=json-file --signature-verification=false -s overlay2'
if [ -z "${DOCKER_CERT_PATH}" ]; then
    DOCKER_CERT_PATH=/etc/docker
fi

# Do not add registries in this file anymore. Use /etc/containers/registries.conf
# from the atomic-registries package.
#

# docker-latest daemon can be used by starting the docker-latest unitfile.
# To use docker-latest client, uncomment below lines
#DOCKERBINARY=/usr/bin/docker-latest
#DOCKERDBINARY=/usr/bin/dockerd-latest
#DOCKER_CONTAINERD_BINARY=/usr/bin/docker-containerd-latest
#DOCKER_CONTAINERD_SHIM_BINARY=/usr/bin/docker-containerd-shim-latest
```
修改镜像源
```
cat > /etc/docker/daemon.json << EOF
{
  "registry-mirrors": ["https://registry.docker-cn.com"]
}
EOF
```
#### 修改主机名
```
hostnamectl set-hostname node1
```