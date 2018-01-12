Contiv 为多种用例提供可配置网络（使用 BGP 的原生 L3，使用 vxlan 的 overlay，经典 L2 和 Cisco-SDN/ACI）和丰富的策略框架。
Contiv 项目完全开源。安装工具同时提供基于和不基于 kubeadm 的安装选项。


VERSION=1.1.7
CONTIV_MASTER=172.26.6.1
curl -L -O https://github.com/contiv/install/releases/download/$VERSION/contiv-$VERSION.tgz
tar oxf contiv-$VERSION.tgz
cd contiv-$VERSION

# VXLAN安装
./install/k8s/install.sh -n $CONTIV_MASTER

# 指定vlan
./install/k8s/install.sh -n $CONTIV_MASTER -v \<data plane interface like eth1\>


# ACL
./install/k8s/install.sh -n $CONTIV_MASTER -a <APIC URL> -u <APIC User> -p <APIC Password> -l 
\<Leaf Nodes\> -d \<Physical Domain> -e \<EPG Bridge domain> -m \<APIC contracts unrestricted mode> 