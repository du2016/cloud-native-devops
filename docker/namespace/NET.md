#!/bin/bash
#Create new namesapce
ip netns delete ns1
ip netns add ns1

ip link add veth0 type veth peer name veth1
ip link set veth0 netns ns1

echo "====Create New Network namespace and spcify a eth===="
ip netns list
ip netns ns exec if config -a

#Assign the IP and bring up
ip addr add 10.100.1.1/24 dev veth1
ip link set veth1 up

ip netns exec ns1 ip addr add 10.100.1.2/24 dev veth0
ip netns exec ns1 ip link set veth0 up
ip netns exec ns1 ip link set lo up

echo "====Bring up the veth0 and lo inside Namespace===="
ip netns exec ns1 ip addr show

#add route inside namespace
ip netns exec ns1 ip route add default via 10.100.1.1

echo "====Add new default rout inside the namespace===="
ip netns exec ns1 ip route show

echo "====Tryting to ping the veth1 on host===="
ip netns exec ns1 ping 10.100.1.1 -c 4

#Config the host to enable forwarding
echo 1 > /proc/sys/net/ipv4/ip_forward

iptables -P FORWARD DROP
iptables -F FORWARD

iptables -t nat -F

#enalbe masquerading of 10.100.1.0
iptables -t nat -A POSTROUTING -s 10.100.1.0/255.255.255.0 -o ens160 -j MASQUERADE

#Allow forwarding
iptables -A FORWARD -i ens160 -o veth1 -j ACCEPT
iptables -A FORWARD -o ens160 -i veth1 -j ACCEPT

echo "====Enable the forwarding of veth1 to ens160(NIC)on host ===="
echo "Show the iptables of filter"
iptables -L -n

echo "Show the iptables of nat"
iptables -t nat -L -n



ip link add veth0 type veth peer name veth1
ip link set veth0 netns 29756
nsenter -t 29756 -n ip link show