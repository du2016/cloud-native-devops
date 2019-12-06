
# 创建网络命名空间

```
# 删除命名空间
ip netns delete ns1
# 添加命名空间
ip netns add ns1

# 创建虚拟网络设备对
ip link add veth0 type veth peer name veth1
# 将虚拟设备移动到命名空间
ip link set veth0 netns ns1

# 将虚拟网络设备移动到PID网络命名空间
ip link set veth0 netns $PID

# 查看命名空间列表
ip netns list
# 在网络命名空间执行命令
ip netns ns exec if config -a

# 为虚拟网络设备设置IP
ip addr add 10.100.1.1/24 dev veth1
# 更改状态
ip link set veth1 up
# 为命名空间内网络设备设置IP
ip netns exec ns1 ip addr add 10.100.1.2/24 dev veth0
# 命名空间内网络设备更改状态
ip netns exec ns1 ip link set veth0 up
ip netns exec ns1 ip link set lo up

# 命名空间内网络设备查看状态
ip netns exec ns1 ip addr show

# 命名空间内添加路由
ip netns exec ns1 ip route add default via 10.100.1.1


# 开启路由转发
echo 1 > /proc/sys/net/ipv4/ip_forward

```

# 为网络配置SNAT

```
iptables -P FORWARD DROP
iptables -F FORWARD
iptables -t nat -F
iptables -t nat -A POSTROUTING -s 10.100.1.0/255.255.255.0 -o ens160 -j MASQUERADE
iptables -A FORWARD -i ens160 -o veth1 -j ACCEPT
iptables -A FORWARD -o ens160 -i veth1 -j ACCEPT
```

nsenter -t 29756 -n ip link show