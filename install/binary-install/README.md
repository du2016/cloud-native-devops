# 前提 

#### centos 7 mini安装

```
yum install vim nfs-utils wget *bin/ifconfig mlocate lezsz epel* *bin/route *bin/traceroute nc -y

```

#### centos 7 默认使用firewall管理防火墙，我还是习惯使用iptables.

```
yum remove firewalld -y  && yum install iptables-services -y
```

#### 禁用selinux

```
setenforce 0
sed -i 's/enforcing/disabled/g' /etc/selinux/config
```

#### 利用nfs共享软件包

```
yum install nfs-utils -y
echo "/opt *(rw,no_root_squash)" > /etc/exports
#其他机器 mount
```

#### 安装高版本内核

链接可能失效，可以去[elrepo官网](http://elrepo.org/tiki/tiki-index.php)查看最新源。
```
rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org
rpm -Uvh http://www.elrepo.org/elrepo-release-7.0-3.el7.elrepo.noarch.rpm
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

#### 修改主机名

```
hostnamectl set-hostname node1
```