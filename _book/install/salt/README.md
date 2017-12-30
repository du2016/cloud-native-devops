[salt state](https://github.com/kubernetes/kubernetes/tree/master/cluster/saltbase/)

## 介绍

暂时只支持： gce, azure, aws, vagrant
saltstack是自动化运维工具相似工具：
- puppet
- ansible
- fabric



## 打开salt-master的auto_accept功能

```
[root@kubernetes-master] $ cat /etc/salt/master.d/auto-accept.conf
open_mode: True
auto_accept: True
```

## 给主机分配角色（通过定义grants实现）

```
$ cat /etc/salt/minion.d/grains.conf
grains:
  etcd_servers: $MASTER_IP
  cloud: vagrant
  roles:
    - kubernetes-master
```

## 测试grains

> salt 'tgt' grants.get etcd_servers

## 可自定义的值

Key | Value
-----------------------------------|----------------------------------------------------------------
`api_servers` | (Optional) The IP address / host name where a kubelet can get read-only access to kube-apiserver
`cbr-cidr` | (Optional) The minion IP address range used for the docker container bridge.
`cloud` | (Optional) Which IaaS platform is used to host Kubernetes, *gce*, *azure*, *aws*, *vagrant*
`etcd_servers` | (Optional) Comma-delimited list of IP addresses the kube-apiserver and kubelet use to reach etcd.  Uses the IP of the first machine in the kubernetes_master role, or 127.0.0.1 on GCE.
`hostnamef` | (Optional) The full host name of the machine, i.e. uname -n
`node_ip` | (Optional) The IP address to use to address this node
`hostname_override` | (Optional) Mapped to the kubelet hostname-override
`network_mode` | (Optional) Networking model to use among nodes: *openvswitch*
`networkInterfaceName` | (Optional) Networking interface to use to bind addresses, default value *eth0*
`publicAddressOverride` | (Optional) The IP address the kube-apiserver should use to bind against for external read-only access
`roles` | (Required) 1. `kubernetes-master` means this machine is the master in the Kubernetes cluster.  2. `kubernetes-pool` means this machine is a kubernetes-node.  Depending on the role, the Salt scripts will provision different resources on the machine.
