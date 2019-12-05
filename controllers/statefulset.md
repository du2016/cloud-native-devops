StatefulSets旨在与有状态应用程序和分布式系统一起使用


# 创建一个STS

 ```
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 80
    name: web
  clusterIP: None
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "nginx"
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
```

sts特点

- 序号为0-n-1
- 前一个ready创建下一个
- 主机名以服务名-序号
- sts的service需要时headlessservice
- dns自动解析各个pod的条目
- 以DNS保证在POD重启时服务的稳定性

# 更新

## RollingUpdate

更新时倒序更新

kubectl patch statefulset web -p '{"spec":{"updateStrategy":{"type":"RollingUpdate"}}}'

### 分段更新

通过RollingUpdate partition字段来指定索引值

```
kubectl patch statefulset web -p '{"spec":{"updateStrategy":{"type":"RollingUpdate","rollingUpdate":{"partition":3}}}}'
statefulset.apps/web patched
```

3及以上的更新，以下的不更新
```
kubectl patch statefulset web --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"k8s.gcr.io/nginx-slim:0.7"}]'
statefulset.apps/web patched
```

退出分段更新，即指定partition为0

```
kubectl patch statefulset web -p '{"spec":{"updateStrategy":{"type":"RollingUpdate","rollingUpdate":{"partition":0}}}}'
statefulset.apps/web patched
```

## OnDelete

历史遗留更新策略，只有在被删除时才会主动更新pod,不会主动获取template内容的更新


# 删除策略

## 非级联删除

kubectl delete statefulset web --cascade=false


即使删除了sts,pod依旧存在，
此时可以通过指定partirion来指定分区重新创建sts来实现部分更新

## 级联删除

kubectl delete statefulset web # 删除sts同时删除pod

# pod管理策略

- OrderedReady 顺序就绪策略，必须等上一个就绪，启动下一个
- Parallel 并行启动终止，不需要保证顺序

