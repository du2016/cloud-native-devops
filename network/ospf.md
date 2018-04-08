# OSPF

# 介绍

OSPF(Open Shortest Path First开放式最短路径优先）是一个内部网关协议(Interior Gateway Protocol，简称IGP），用于在单一自治系统（autonomous system,AS）内决策路由。是对链路状态路由协议的一种实现，隶属内部网关协议（IGP），故运作于自治系统内部。著名的迪克斯加算法(Dijkstra)被用来计算最短路径树。OSPF分为OSPFv2和OSPFv3两个版本,其中OSPFv2用在IPv4网络，OSPFv3用在IPv6网络。OSPFv2是由RFC 2328定义的，OSPFv3是由RFC 5340定义的。与RIP相比，OSPF是链路状态协议，而RIP是距离矢量协议。

# quagga

Quagga是一款功能比较强大的开源路由软件，支持ip,ripng,ospfv2,ospfv3，bgp等协议。安装Quagga的目的是使装有linux系统的电脑变成一台路由器;其主要的功能支持动态+静态路由的配置功能

# 实现方式

容器桥接到docker0，为docker0分配一个C段，通过ospf发布docker0所在网段，实现内网全互联

# 配置

```
yum install -y quagga

cat > /etc/quagga/ospfd.conf << EOF
! -*- ospf -*-
!
! OSPFd sample configuration file
!
!
hostname ospfd
password zebra
enable password zebra
!
router ospf
  network 10.9.0.0/16 area 172.16.0.1
  network 10.12.0.0/24 area 172.16.0.1

!log /var/log/quagga/ospfd.log
EOF


cat > /etc/quagga/zebra.conf << EOF
! -*- zebra -*-
!
! zebra sample configuration file
!
! $Id: zebra.conf.sample,v 1.1 2002/12/13 20:15:30 paul Exp $
!
hostname Router
password zebra
enable password zebra
!log file zebra.log
EOF

service ospfd restart && service zebra restart
```