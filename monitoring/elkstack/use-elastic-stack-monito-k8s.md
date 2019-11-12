

## 启动elasticsearch+kibana

为了快速启动这里直接使用docker

```
docker run -d -v /etc/localtime:/etc/localtime -p 9200:9200 -p 9300:9300 --name=elasticsearch -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.4.1
docker run -d -v /etc/localtime:/etc/localtime  --link elasticsearch:elasticsearch -p 5601:5601 docker.elastic.co/kibana/kibana:7.4.1
```

## 集群安装


### 创建集群
```
cat >> kubeadm.config << EOF
apiVersion: kubeadm.k8s.io/v1beta2
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: abcdef.0123456789abcdef
  ttl: 24h0m0s
  usages:
  - signing
  - authentication
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: 10.10.8.42
  bindPort: 6443
nodeRegistration:
  criSocket: /var/run/dockershim.sock
  name: 10.10.8.42
  taints:
  - effect: NoSchedule
    key: node-role.kubernetes.io/master
---
apiServer:
  timeoutForControlPlane: 4m0s
apiVersion: kubeadm.k8s.io/v1beta2
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controllerManager: {}
dns:
  type: CoreDNS
etcd:
  local:
    dataDir: /var/lib/etcd
imageRepository: registry.aliyuncs.com/google_containers
kind: ClusterConfiguration
kubernetesVersion: v1.16.0
networking:
  dnsDomain: cluster.local
  serviceSubnet: 200.0.0.1/16
  podSubnet: 10.201.0.0/16
controlPlaneEndpoint: "10.10.8.200"
scheduler: {}
EOF

kubeadm init --config=kubeadm.config
```


### 设置网络

我们选用了canal插件
```
kubectl apply -f https://docs.projectcalico.org/v3.8/manifests/canal.yaml
```

## 安装kube-state-metrics

kube-state-metrics 用于通过apiserver获取k8s集群及创建对象的状态
```
git clone https://github.com/kubernetes/kube-state-metrics.git
cd kube-state-metrics/examples/standard/
kubectl apply -f .
```


## 安装metricbeat

```
git clone https://github.com/elastic/beats.git
cd beats/deploy/kubernetes/metricbeat/

# 修改镜像版本
sed -i "s/%VERSION%/7.4.1/g" *

# 修改对应es的host
TODO

kubectl apply -f .
```
## 安装filebeat

```
cd beats/deploy/kubernetes/metricbeat/
# 修改对应es的host
TODO

kubectl apply -f .
```

## 安装heartbeat


```
cat >> heartbeat.yaml << EOF
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: heartbeat-deployment-config
  namespace: kube-system
  labels:
    k8s-app: heartbeat
data:
  heartbeat.yml: |-
    heartbeat.autodiscover:
      providers:
        - type: kubernetes
          templates:
            - config:
                - type: icmp
                  hosts: ["${data.host}"]
                  schedule: '*/5 * * * * * *'

    cloud.auth: ${ELASTIC_CLOUD_AUTH}
    cloud.id: ${ELASTIC_CLOUD_ID}

    output.elasticsearch:
      hosts: ${ELASTICSEARCH_HOSTS}
      username: ${ELASTICSEARCH_USERNAME}
      password: ${ELASTICSEARCH_PASSWORD}
    setup.kibana:
      host: ${KIBANA_HOST}
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: heartbeat
  namespace: kube-system
  labels:
    k8s-app: heartbeat
spec:
  template:
    metadata:
      labels:
        k8s-app: heartbeat
    spec:
      serviceAccountName: heartbeat
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: heartbeat
        image: docker.elastic.co/beats/heartbeat:7.2.0
        args: [
          "-c", "/etc/heartbeat.yml",
          "-e",
        ]
        env:
        - name: ELASTIC_CLOUD_ID
        - name: ELASTIC_CLOUD_AUTH
        - name: ELASTICSEARCH_HOSTS
          value: "10.10.8.42"
        - name: KIBANA_HOST
          value: "10.10.8.42"
        - name: ELASTICSEARCH_USERNAME
          value: "admin"
        - name: ELASTICSEARCH_PASSWORD
          value: "admin"
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        securityContext:
          runAsUser: 0
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 100Mi
        volumeMounts:
        - name: config
          mountPath: /etc/heartbeat.yml
          readOnly: true
          subPath: heartbeat.yml
      volumes:
      - name: config
        configMap:
          defaultMode: 0600
          name: heartbeat-deployment-config
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: heartbeat
subjects:
- kind: ServiceAccount
  name: heartbeat
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: heartbeat
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: heartbeat
  labels:
    k8s-app: heartbeat
rules:
- apiGroups: [""]
  resources:
  - nodes
  - namespaces
  - events
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions"]
  resources:
  - replicasets
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources:
  - statefulsets
  - deployments
  verbs: ["get", "list", "watch"]
- apiGroups:
  - ""
  resources:
  - nodes/stats
  verbs:
  - get
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: heartbeat
  namespace: kube-system
  labels:
    k8s-app: heartbeat
---
EOF

kubectl apply -f heartbeat.yaml
```


## 效果展示

## pod列表
![aa](http://q08i5y6c2.bkt.clouddn.com/1.png)

## pod日志
![aa](http://q08i5y6c2.bkt.clouddn.com/2.png)

## pod监控
![aa](http://q08i5y6c2.bkt.clouddn.com/3.png)

## pod网络
![aa](http://q08i5y6c2.bkt.clouddn.com/4.png)


扫描关注我:

![微信](http://q08i5y6c2.bkt.clouddn.com/qrcode_for_gh_7457c3b1bfab_258.jpg)