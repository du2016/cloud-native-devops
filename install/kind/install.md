```
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
featureGates:
  GenericEphemeralVolume: true
networking:
  podSubnet: "10.244.0.0/16"
  serviceSubnet: "10.96.0.0/12"
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
      apiServerExtraArgs:
#      feature-gates: GenericEphemeralVolume=true
#      service-account-issuer: kubernetes.default.svc
#      service-account-signing-key-file: /etc/kubernetes/pki/sa.key
#      service-account-api-audiences: kubernetes.default.svc
- role: worker
  image: kindest/node:v1.19.0
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      imageRepository: registry.aliyuncs.com/google_containers
      kubeletExtraArgs:
        pod-infra-container-image: registry.aliyuncs.com/google_containers/pause:3.1
```
  
  
查看可用镜像版本https://hub.docker.com/r/kindest/node/tags

```
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: weight-shifting
  namespace: default
spec:
  virtualhost:
    fqdn: weights.bar.com
  routes:
    - services:
        - name: s1
          port: 80
          weight: 10
        - name: s2
          port: 80
          weight: 90
```