# 使用prom替换heapster实现基于mem/cpu的hpa

# 依赖

- k8s 1.8+

# 安装prometheus监控

使用  prometheus-operator安装prometheus
```
git clone https://github.com/coreos/prometheus-operator
cd prometheus-operator
kubectl apply -f ./contrib/kube-prometheus/manifests/
```

访问prom ui查看指标

# 证书生成

## 生成验证请求客户端身份的根证书
```
cat <<EOF > front-proxy-ca-csr.json
{
    "CN": "kubernetes",
    "key": {
        "algo": "rsa",
        "size": 2048
    }
}
EOF
cfssl gencert -initca front-proxy-ca-csr.json | cfssljson -bare front-proxy-ca
```

## 生成证明apiserver身份的客户端证书（或者其它聚合器）

```
cat <<EOF > front-proxy-client-csr.json
{
    "CN": "front-proxy-client",
    "key": {
        "algo": "rsa",
        "size": 2048
    }
}
EOF

cfssl gencert \
  -ca=front-proxy-ca.pem \
  -ca-key=front-proxy-ca-key.pem \
  -config=ca-config.json \
  -profile=kubernetes \
  front-proxy-client-csr.json | cfssljson -bare front-proxy-client
```

将根证书拷贝到其它节点

```
scp /etc/kubernetes/ssl/front-proxy-ca.pem xxx:/etc/kubernetes/ssl/front-proxy-ca.pem
```

修改metrics-server-deployment.yaml 指定根证书
```
        volumeMounts:
        - mountPath: /etc/kubernetes/ssl
          name: ca
      volumes:
      - name: ca
        hostPath:
          path: /etc/kubernetes/ssl
```

# 配置

## kube-apiserver添加配置

```
--requestheader-client-ca-file=/etc/kubernetes/ssl/front-proxy-ca.pem --requestheader-allowed-names=aggregator --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-group-headers=X-Remote-Group --requestheader-username-headers=X-Remote-User --proxy-client-cert-file=/etc/kubernetes/ssl/front-proxy-client.pem --proxy-client-key-file=/etc/kubernetes/ssl/front-proxy-client-key.pem
```

## controller-manager添加配置

```
--horizontal-pod-autoscaler-use-rest-clients=true
```


# metrics-server

## 安装

```
git clone https://github.com/kubernetes-incubator/metrics-server
cd metrics-server/deploy
kubectl apply -f 1.8+/
```

## 验证配置
```
#查看node指标
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/nodes" | jq .
#查看pod指标
kubectl get --raw "/apis/metrics.k8s.io/v1beta1/pods" | jq .
```

# 验证hpa

## 创建podinfo deployment

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: podinfo
spec:
  replicas: 2
  template:
    metadata:
      labels:
        app: podinfo
      annotations:
        prometheus.io/scrape: 'true'
    spec:
      containers:
      - name: podinfod
        image: stefanprodan/podinfo:0.0.1
        imagePullPolicy: Always
        command:
          - ./podinfo
          - -port=9898
          - -logtostderr=true
          - -v=2
        volumeMounts:
          - name: metadata
            mountPath: /etc/podinfod/metadata
            readOnly: true
        ports:
        - containerPort: 9898
          protocol: TCP
        readinessProbe:
          httpGet:
            path: /readyz
            port: 9898
          initialDelaySeconds: 1
          periodSeconds: 2
          failureThreshold: 1
        livenessProbe:
          httpGet:
            path: /healthz
            port: 9898
          initialDelaySeconds: 1
          periodSeconds: 3
          failureThreshold: 2
        resources:
          requests:
            memory: "32Mi"
            cpu: "1m"
          limits:
            memory: "256Mi"
            cpu: "100m"
      volumes:
        - name: metadata
          downwardAPI:
            items:
              - path: "labels"
                fieldRef:
                  fieldPath: metadata.labels
              - path: "annotations"
                fieldRef:
                  fieldPath: metadata.annotations
```

## 创建podinfo hpa

```
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: podinfo
spec:
  scaleTargetRef:
    apiVersion: extensions/v1beta1
    kind: Deployment
    name: podinfo
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      targetAverageUtilization: 80
  - type: Resource
    resource:
      name: memory
      targetAverageValue: 200Mi
```

## 验证

查看是否获取到指标数据
```
kubectl get hpa
```

压力测试

```
ab -c 1000 -n 100000000000 http://podinfosvc:port/index.html
```

可以看到已经扩容


# 有关apiserver的认证机制参考 https://github.com/kubernetes-incubator/apiserver-builder/blob/master/docs/concepts/auth.md