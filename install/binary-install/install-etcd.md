# 获取二进制文件
```
wget https://github.com/coreos/etcd/releases/download/v3.2.12/etcd-v3.2.12-linux-amd64.tar.gz
tar xf etcd-v3.2.12-linux-arm64.tar.gz
cp  etcd-v3.2.12-linux-arm64/{etcd,etcdctl} /usr/local/bin/
```


#### etcd v2

```
etcdctl  --ca-file=/etc/kubernetes/ssl/ca.pem   --cert-file=/etc/kubernetes/ssl/kubernetes.pem   --key-file=/etc/kubernetes/ssl/kubernetes-key.pem ls /
```

#### etcd v3
ETCDCTL_API=3
```
etcdctl  --cacert=/etc/kubernetes/ssl/ca.pem   --cert=/etc/kubernetes/ssl/kubernetes.pem   --key=/etc/kubernetes/ssl/kubernetes-key.pem
```
# 配置

#### 服务文件配置
```
cat > /usr/lib/systemd/system/etcd.service <<EOF
[Unit]
Description=Etcd Server
After=network.target
After=network-online.target
Wants=network-online.target
Documentation=https://github.com/coreos

[Service]
Type=notify
WorkingDirectory=/var/lib/etcd/
EnvironmentFile=-/etc/etcd/etcd.conf
ExecStart=/usr/local/bin/etcd \
  --name \${ETCD_NAME} \
  --cert-file=/etc/kubernetes/ssl/kubernetes.pem \
  --key-file=/etc/kubernetes/ssl/kubernetes-key.pem \
  --peer-cert-file=/etc/kubernetes/ssl/kubernetes.pem \
  --peer-key-file=/etc/kubernetes/ssl/kubernetes-key.pem \
  --trusted-ca-file=/etc/kubernetes/ssl/ca.pem \
  --peer-trusted-ca-file=/etc/kubernetes/ssl/ca.pem \
  --initial-advertise-peer-urls \${ETCD_INITIAL_ADVERTISE_PEER_URLS} \
  --listen-peer-urls \${ETCD_LISTEN_PEER_URLS} \
  --listen-client-urls \${ETCD_LISTEN_CLIENT_URLS},https://127.0.0.1:4001 \
  --advertise-client-urls \${ETCD_ADVERTISE_CLIENT_URLS} \
  --initial-cluster-token \${ETCD_INITIAL_CLUSTER_TOKEN} \
  --initial-cluster infra1=https://\${ETCD_NODE1}:7001,infra2=https://\${ETCD_NODE2}:7001,infra3=https://\${ETCD_NODE3}:7001 \
  --initial-cluster-state new \
  --data-dir=\${ETCD_DATA_DIR}
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
```

#### 配置文件

```
mkdir /etc/etcd
cat > /etc/etcd/etcd.conf <<EOF
NODE_IP=172.26.6.2
ETCD_NAME=infra2
ETCD_DATA_DIR="/var/lib/etcd"
ETCD_LISTEN_PEER_URLS="https://${NODE_IP}:7001"
ETCD_LISTEN_CLIENT_URLS="https://${NODE_IP}:4001"
ETCD_INITIAL_ADVERTISE_PEER_URLS="https://${NODE_IP}:7001"
ETCD_INITIAL_CLUSTER_TOKEN="etcd-cluster"
ETCD_ADVERTISE_CLIENT_URLS="https://${NODE_IP}:4001"


ETCD_NODE1=172.26.6.1
ETCD_NODE2=172.26.6.2
ETCD_NODE3=172.26.6.3
EOF
```

#### 启动
```
mkdir /var/lib/etcd/
systemctl daemon-reload
systemctl enable etcd
systemctl start etcd
systemctl status etcd
```

#### 查看状态

```
etcdctl --endpoints=https://172.26.6.1:4001,https://172.26.6.2:4001,https://172.26.6.3:4001   --ca-file=/etc/kubernetes/ssl/ca.pem   --cert-file=/etc/kubernetes/ssl/kubernetes.pem   --key-file=/etc/kubernetes/ssl/kubernetes-key.pem cluster-health
```