# kubeflow

https://github.com/ksonnet 

## install ksonnet

macos
```
brew install ksonnet/tap/ks
```

linux 

```
wget https://github.com/ksonnet/ksonnet/releases/download/v0.8.0/ks-linux-amd64
```


## install kubeflow

```
ks init my-kubeflow
cd my-kubeflow
ks registry add kubeflow github.com/google/kubeflow/tree/master/kubeflow
ks pkg install kubeflow/core
ks pkg install kubeflow/tf-serving
ks pkg install kubeflow/tf-job
ks generate core kubeflow-core --name=kubeflow-core
```
## define env

minikube
```
kubectl config use-context minikube
ks env add minikube
ks apply minikube -c kubeflow-core
```

gke
```
kubectl config use-context gke
ks env add gke
ks apply gke -c kubeflow-core
```

access 

```
kubectl port-forward tf-hub-0 8100:8000
```

## training job 训练作业

ks generate tf-cnn cnn --name=cnn
ks apply minikube -c cnn

## 查看训练列表

ks prototype list tf-job
ks param set --env=gke cnn num_gpus 1
ks param set --env=gke cnn num_workers 1
ks apply gke -c cnn
