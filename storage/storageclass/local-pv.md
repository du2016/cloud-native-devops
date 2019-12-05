# 直接绑定
## 创建local-pv

```
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv
spec:
    capacity:
      storage: 1Gi
    accessModes:
    - ReadWriteOnce
    persistentVolumeReclaimPolicy: Delete
    storageClassName: local-storage
    local:
      path: /media/aaa
    nodeAffinity:
      required:
        nodeSelectorTerms:
        - matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
            - 10.10.8.42
```

## 创建pvc

```
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: example-local-claim
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: local-storage
```


# Topology-Aware Volume

## DelayBinding

创建storageclass 设置绑定模式为等待使用时绑定
```
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
# 等待第一次被使用时绑定
volumeBindingMode: WaitForFirstConsumer
```


```
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nginx
spec:
  replicas: 1
  serviceName: nginx-headless
  selector:
    matchLabels:
      name: nginx
  template:
    metadata:
      name: nginx
      labels:
        name: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        resources:
          limits:
            cpu: 5300m
            memory: 5Gi
        volumeMounts:
        - mountPath: /var/lib/mysql
          name: data
  volumeClaimTemplates:
  - metadata:
      annotations:
        volume.beta.kubernetes.io/storage-class: local-storage
      name: data
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi

---

apiVersion: v1
kind: Service
metadata:
  name: nginx-headless
  labels:
    name: nginx
spec:
  ports:
  - port: 80
    name: web
  clusterIP: None
  selector:
    name: nginx
```