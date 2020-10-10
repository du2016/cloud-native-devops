# 优化

net.netfilter.nf_conntrack_tcp_loose = 1
net.netfilter.nf_conntrack_tcp_be_liberal = 1

net.nf_conntrack_max = 1048576
net.netfilter.nf_conntrack_tcp_timeout_close_wait = 60
net.netfilter.nf_conntrack_tcp_timeout_established = 7200
net.netfilter.nf_conntrack_tcp_timeout_fin_wait = 120
net.netfilter.nf_conntrack_tcp_timeout_last_ack = 30
net.netfilter.nf_conntrack_tcp_timeout_syn_recv = 60
net.netfilter.nf_conntrack_tcp_timeout_syn_sent = 30
net.netfilter.nf_conntrack_tcp_timeout_time_wait = 30


[k8s 丢包排查](https://kubernetes.io/blog/2019/03/29/kube-proxy-subtleties-debugging-an-intermittent-connection-reset/)

# iptables=false

Iptables must be enable when docker start.
Iptables can be disable after docker started.