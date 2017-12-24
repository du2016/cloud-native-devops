# minikube 快速搭建k8s

#### 下载对应的kubelet和kubectl 添加PATH
```
wget https://storage.googleapis.com/minikube/releases/v0.22.3/minikube-darwin-amd64 && mv minikube-darwin-amd64 /usr/local/bin
curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/darwin/amd64/kubectl
chmod +x kubectl && mv kubectl  /usr/local/bin
```

#### install fusion/vbox....

#### start

```
minikube start --vm-driver vmwarefusion --docker-env HTTP_PROXY=http://xxxx --docker-env HTTPS_PROXY=https://xxxx  -v 10 --docker-opt bip=10.0.0.1/24 
route add -net 10.0.0.1/24 ${MINIKUBE-IP}
```