# fluentd-elasticsearch


cd kubernetes-repo/cluster/addons/fluentd-elasticsearch

####
- es镜像

```
cd /opt/kubernetes-repo/cluster/addons/fluentd-elasticsearch/es-image 
make binary && make build
```

- fluentd镜像

```
cd /opt/kubernetes-repo/cluster/addons/fluentd-elasticsearch/fluentd-es-image && make build
```


#### 创建 fluentd-elasticsearch

- 给node添加label

因为默认的配置添加了nodeselector
```
      nodeSelector:
        beta.kubernetes.io/fluentd-ds-ready: "true"
```
所系需要给需要收集日志的node添加label

```
kubectl label nodes 172.26.6.2 beta.kubernetes.io/fluentd-ds-ready=true
```

- 创建
```
kubectl create -f es-statefulset.yaml
kubectl create -f es-service.yaml
kubectl create -f kibana-service.yaml
kubectl create -f kibana-deployment.yaml
kubectl create -f fluentd-es-configmap.yaml
kubectl create -f fluentd-es-ds.yaml
```
