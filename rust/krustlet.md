# 介绍

Krustlet 是基于k8s运行wasm程序的负载，通过亲和性来运行wasm程序，其实现了kubelet api，且兼容了 `kubectl logs` 和 `kubectl delete` 命令。

接下来将一步步安装、运行krustlet

# kind安装

## kind config配置
```
kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha3
kubeadmConfigPatches:
- |
  apiVersion: kubeadm.k8s.io/v1beta1
  kind: ClusterConfiguration
  metadata:
    name: config
  networking:
    serviceSubnet: 10.0.0.0/16
  imageRepository: registry.aliyuncs.com/google_containers
  nodeRegistration:
    kubeletExtraArgs:
      pod-infra-container-image: registry.aliyuncs.com/google_containers/pause:3.1
- |
  apiVersion: kubeadm.k8s.io/v1beta1
  kind: InitConfiguration
  metadata:
    name: config
  networking:
    serviceSubnet: 10.0.0.0/16
  imageRepository: registry.aliyuncs.com/google_containers
nodes:
- role: control-plane
- role: worker
- role: worker
```

## 创建集群

```
kind create --config=kind-config.yaml
```
# 运行krustlet

## 生成kubeconfig

通过脚本生成bootstrap config
```
bash <(curl https://raw.githubusercontent.com/deislabs/krustlet/master/docs/howto/assets/bootstrap.sh)
```

## krustlet安装

```
wget https://krustlet.blob.core.windows.net/releases/krustlet-v0.3.0-macos-amd64.tar.gz
tar xf krustlet-v0.3.0-macos-amd64.tar.gz
chmod +x krustlet-*
mv krustlet-* /usr/local/bin/
```

## 启动krustlet

```
/usr/local/bin/krustlet-wasi  --node-ip 192.168.11.5 --bootstrap-file=~/.krustlet/config/bootstrap.conf
```


## approve 证书

像kubelet一样，通过bootstrap config启动后会发起一个csr,我们需要对其approve

```
kubectl certificate approve  Mbp.local-tls
```

## 查看状态


可以看到已经处于ready状态

```
kubectl get nodes
NAME                 STATUS   ROLES    AGE   VERSION
kind-control-plane   Ready    master   16m   v1.18.2
kind-worker          Ready    <none>   15m   v1.18.2
kind-worker2         Ready    <none>   15m   v1.18.2
mbp.local            Ready    agent    9s    0.3.0
```

查看污点

```
kubectl get nodes mbp.local -o json | jq .spec.taints
[
  {
    "effect": "NoExecute",
    "key": "krustlet/arch",
    "value": "wasm32-wasi"
  }
]
```


# krustlet运行wasi

## 下载wasm-to-oci

我们需要下载wasm-to-oci工具，使用该工具可以将wasm程序推送到docker仓库。

```
wget https://github.com/engineerd/wasm-to-oci/releases/download/v0.1.1/darwin-amd64-wasm-to-oci
```

## 生成wasi程序

创建项目
```
cargo new krust-test
```

代码如下
```
#[no_mangle]
fn _start() {
    let mut i:i32 = 0;
    loop {
        println!("hello {:?}",i);
        i+=1;
        sleep(Duration::new(1, 0))
    }
}
```
编译

```
cargo build --target=wasm32-wasi
```

## 准备一个oci v2registry

支持以下几种registry


- Distribution (open source, version 2.7+)
- Azure Container Registry
- Google Container Registry
- Harbor Container Registry v2.0

这里我自行搭建了一个，证书通过letsencrypt生成
```
docker run   -v /media:/certs \
>   -e REGISTRY_HTTP_ADDR=0.0.0.0:443 \
>   -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/fullchain.crt \
>   -e REGISTRY_HTTP_TLS_KEY=private.pem \
>   -p 443:443 \
>   registry:2
```

将编译好的wasm文件推送到仓库
```
wasm-to-oci push .//target/wasm32-wasi/release/krustlet_test.wasm reg.rocdu.top/test/krustlet:v1
```

## 运行wasi wavm程序

我们将运行一个wasm pod,通过容忍策略让该pod可以运行在我们的krustlet节点上面

```
apiVersion: v1
kind: Pod
metadata:
  name: krustlet-tutorial
spec:
  containers:
    - name: krustlet-tutorial
      image: reg.rocdu.top/test/krustlet:v1
  tolerations:
    - key: "krustlet/arch"
      operator: "Equal"
      value: "wasm32-wasi"
      effect: "NoExecute"
```

查看运行状态，可以看到已经处于运行状态，让我们看一下日志的输出，可以看到是我们预期的输出。

```
kubectl get pods -w
NAME                READY   STATUS    RESTARTS   AGE
krustlet-tutorial   0/1     Pending   0          14s
krustlet-tutorial   0/1     Running   0          14s


kubectl logs krustlet-tutorial -f
hello 1
hello 2
hello 3
```


# 总结

krustlet以及wasi都是很不成熟的项目，仅限用于测试。
 

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
