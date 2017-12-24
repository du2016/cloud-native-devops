# 安装kube-proxy

#### 服务文件
```
/usr/lib/systemd/system/kube-proxy.service
[Unit]
Description=Kubernetes Kube-Proxy Server
Documentation=https://github.com/GoogleCloudPlatform/kubernetes
After=network.target

[Service]
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/proxy
ExecStart=/usr/local/bin/kube-proxy \
        $KUBE_LOGTOSTDERR \
        $KUBE_LOG_LEVEL \
        $KUBE_MASTER \
        $KUBE_PROXY_ARGS
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

#### 参数配置文件

```
cat > /etc/kubernetes/proxy << EOF
###
# kubernetes proxy config

# default config should be adequate

# Add your own!
KUBE_PROXY_ARGS="--hostname-override=172.26.6.1 --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig --cluster-cidr=10.254.0.0/16"
#enable ipvs 需要安装ipset命令
#KUBE_PROXY_ARGS="--hostname-override=172.26.6.1 --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig --cluster-cidr=10.254.0.0/16 --masquerade-all --feature-gates=SupportIPVSProxyMode=true --proxy-mode=ipvs --ipvs-min-sync-period=5s --ipvs-sync-period=5s  --ipvs-scheduler=rr"

EOF
```

#### 启动
```
systemctl daemon-reload
systemctl enable kube-proxy
systemctl start kube-proxy
systemctl status kube-proxy
```