# 环境准备

使用kind进行集群安装，通过静态路由打通两个集群的容器网络。

## cluster1 初始化

cluster1 kind配置如下

```
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
featureGates:
  GenericEphemeralVolume: true
networking:
  podSubnet: "10.241.0.0/16"
  serviceSubnet: "10.95.0.0/16"
nodes:
- role: control-plane
  extraMounts:
  - hostPath: /Users/dutianpeng/working/rocdu-certs
    containerPath: /files
  image: kindest/node:v1.19.0
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      imageRepository: registry.aliyuncs.com/google_containers
      kubeletExtraArgs:
        pod-infra-container-image: registry.aliyuncs.com/google_containers/pause:3.1
        cluster-domain: cluster.local
```

安装cluster1

```
kind create cluster --name=cluster1 --config=cluster1.yaml
```

初始化istio ca

```
kubectl create namespace istio-system
kubectl create secret generic cacerts -n istio-system \
    --from-file=samples/certs/ca-cert.pem \
    --from-file=samples/certs/ca-key.pem \
    --from-file=samples/certs/root-cert.pem \
    --from-file=samples/certs/cert-chain.pem
```

## cluster2 初始化

cluster2 kind配置如下

```
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
featureGates:
  GenericEphemeralVolume: true
networking:
  podSubnet: "10.244.0.0/16"
  serviceSubnet: "10.96.0.0/16"
nodes:
- role: control-plane
  extraMounts:
  - hostPath: /Users/dutianpeng/working/rocdu-certs
    containerPath: /files
  image: kindest/node:v1.19.0
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      imageRepository: registry.aliyuncs.com/google_containers
      kubeletExtraArgs:
        pod-infra-container-image: registry.aliyuncs.com/google_containers/pause:3.1
        cluster-domain: cluster.local
```

安装cluster2

```
kind create cluster --name=cluster1 --config=cluster1.yaml
```

初始化istio ca

```
kubectl create namespace istio-system
kubectl create secret generic cacerts -n istio-system \
    --from-file=samples/certs/ca-cert.pem \
    --from-file=samples/certs/ca-key.pem \
    --from-file=samples/certs/root-cert.pem \
    --from-file=samples/certs/cert-chain.pem
```

# 添加静态路由

因为是kind安装，所以每个节点都是我们对应的容器，进入容器查看ip如下

```
cluster1: 172.18.0.2
cluster2： 172.18.0.3
```

添加静态路由

```
docker exec cluster1-control-plane ip route add 10.244.0.0/16 via 172.18.0.3
docker exec cluster2-control-plane ip route add 10.241.0.0/16 via 172.18.0.2
```

## 初始化环境变量


```
export MAIN_CLUSTER_CTX=kind-cluster1
export REMOTE_CLUSTER_CTX=kind-cluster2
export MAIN_CLUSTER_NAME=main0
export REMOTE_CLUSTER_NAME=remote0
export MAIN_CLUSTER_NETWORK=network1
export REMOTE_CLUSTER_NETWORK=network1
```

## 生成cluster1 IstioOperator配置

```
cat <<EOF> istio-main-cluster.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  values:
    global:
      multiCluster:
        clusterName: ${MAIN_CLUSTER_NAME}
      network: ${MAIN_CLUSTER_NETWORK}

      # Mesh network configuration. This is optional and may be omitted if
      # all clusters are on the same network.
      meshNetworks:
        ${MAIN_CLUSTER_NETWORK}:
          endpoints:
          - fromRegistry:  ${MAIN_CLUSTER_NAME}
          gateways:
          - registry_service_name: istio-ingressgateway.istio-system.svc.cluster.local
            port: 443

        ${REMOTE_CLUSTER_NETWORK}:
          endpoints:
          - fromRegistry: ${REMOTE_CLUSTER_NAME}
          gateways:
          - registry_service_name: istio-ingressgateway.istio-system.svc.cluster.local
            port: 443

      # Use the existing istio-ingressgateway.
      meshExpansion:
        enabled: true
EOF
```

通过istioctl进行安装

```
istioctl install -f istio-main-cluster.yaml --context=${MAIN_CLUSTER_CTX}
# 因为是本地安装需要将istiod改为NodePort
```

## 远程集群安装

查看cluster1中istiod的IP，设置环境变量

```
export ISTIOD_REMOTE_EP=10.241.0.5
```

cluster2 IstioOperator config 配置如下

```
cat <<EOF> istio-remote0-cluster.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  values:
    global:
      # The remote cluster's name and network name must match the values specified in the
      # mesh network configuration of the primary cluster.
      multiCluster:
        clusterName: ${REMOTE_CLUSTER_NAME}
      network: ${REMOTE_CLUSTER_NETWORK}

      # Replace ISTIOD_REMOTE_EP with the the value of ISTIOD_REMOTE_EP set earlier.
      remotePilotAddress: ${ISTIOD_REMOTE_EP}

  ## The istio-ingressgateway is not required in the remote cluster if both clusters are on
  ## the same network. To disable the istio-ingressgateway component, uncomment the lines below.
  #
  # components:
  #  ingressGateways:
  #  - name: istio-ingressgateway
  #    enabled: false
EOF
```

istioctl进行安装

```
istioctl install -f istio-remote0-cluster.yaml --context ${REMOTE_CLUSTER_CTX}
# 因为是本地安装需要改为NodePort
```

# 配置secret使cluster1能够访问cluster2

创建cluster1访问cluster2的kubeconfig secret

```
istioctl x create-remote-secret --name ${REMOTE_CLUSTER_NAME} --context=${REMOTE_CLUSTER_CTX} > remote-kubeconfig.yaml

# 因为本地的kubeconfig是通过端口映射的，所以需要修改为 DOCKER_IP:6443

cat remote-kubeconfig.yaml | kubectl apply -f - --context=${MAIN_CLUSTER_CTX}
```

这里实际上创建了istio-remote-secret-remote0 secret,该secret具有

```
istio/multiCluster: "true"
```

istio中secretcontorller 会watch具有该label的secret加入registry

# 测试     

## 在远程集群安装helloworld v2服务


```
kubectl create namespace sample --context=${REMOTE_CLUSTER_CTX}
kubectl label namespace sample istio-injection=enabled --context=${REMOTE_CLUSTER_CTX}

kubectl create -f samples/helloworld/helloworld.yaml -l app=helloworld -n sample --context=${REMOTE_CLUSTER_CTX}
kubectl create -f samples/helloworld/helloworld.yaml -l version=v2 -n sample --context=${REMOTE_CLUSTER_CTX}
```

## 在主集群安装helloworld v1 服务

```
kubectl create namespace sample --context=${MAIN_CLUSTER_CTX}
kubectl label namespace sample istio-injection=enabled --context=${MAIN_CLUSTER_CTX}

kubectl create -f samples/helloworld/helloworld.yaml -l app=helloworld -n sample --context=${MAIN_CLUSTER_CTX}
kubectl create -f samples/helloworld/helloworld.yaml -l version=v1 -n sample --context=${MAIN_CLUSTER_CTX}
```


## 安装sleep服务

```
kubectl apply -f samples/sleep/sleep.yaml -n sample --context=${MAIN_CLUSTER_CTX}
kubectl apply -f samples/sleep/sleep.yaml -n sample --context=${REMOTE_CLUSTER_CTX}
```

## 测试连通性

```
kubectl exec -it -n sample -c sleep --context=${MAIN_CLUSTER_CTX} $(kubectl get pod -n sample -l app=sleep --context=${MAIN_CLUSTER_CTX} -o jsonpath='{.items[0].metadata.name}') -- curl helloworld.sample:5000/hello

kubectl exec -it -n sample -c sleep --context=${REMOTE_CLUSTER_CTX} $(kubectl get pod -n sample -l app=sleep --context=${REMOTE_CLUSTER_CTX} -o jsonpath='{.items[0].metadata.name}') -- curl helloworld.sample:5000/hello
```
这里可以看到我们不管在本地还是远程集群的sleep服务访问helloworld时返回结果随机为v1或者v2,证明已经实现跨集群的流量控制

查看cluster ep列表

```
kubectl get pod -n sample -l app=sleep --context=${MAIN_CLUSTER_CTX} -o name | cut -f2 -d'/' | \
    xargs -I{} istioctl -n sample --context=${MAIN_CLUSTER_CTX} proxy-config endpoints {} --cluster "outbound|5000||helloworld.sample.svc.cluster.local"
    
 kubectl get pod -n sample -l app=sleep --context=${REMOTE_CLUSTER_CTX} -o name | cut -f2 -d'/' | \
     xargs -I{} istioctl -n sample --context=${REMOTE_CLUSTER_CTX} proxy-config endpoints {} --cluster "outbound|5000||helloworld.sample.svc.cluster.local" 
```

可以看到都可以获取到分别在两个集群上的helloworld服务

# 总结

本文在本地以kind模拟了多k8s集群共享控制面(单网络平面)的部署方式，