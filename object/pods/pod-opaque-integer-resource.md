# 不透明资源

# 创建不透明资源

```
curl --header "Content-Type: application/json-patch+json" \
--request PATCH \
--data '[{"op": "add", "path": "/status/capacity/pod.alpha.kubernetes.io~1opaque-int-resource-dongle", "value": "4"}]' \
http://localhost:8001/api/v1/nodes/<your-node-name>/status
```

# 删除不透明资源

```
curl --header "Content-Type: application/json-patch+json" \
--request PATCH \
--data '[{"op": "remove", "path": "/status/capacity/pod.alpha.kubernetes.io~1opaque-int-resource-dongle"}]' \
http://localhost:8001/api/v1/nodes/<your-node-name>/status
```

# 为容器分配不透明资源

若想请求不透明整数资源，请在您的容器 manifest 中包含 resources:requests 字段。
不透明整数资源拥有前缀 pod.alpha.kubernetes.io/opaque-int-resource-

```
apiVersion: v1
kind: Pod
metadata:
  name: oir-demo
spec:
  containers:
  - name: oir-demo-ctr
    image: nginx
    resources:
      requests:
        pod.alpha.kubernetes.io/opaque-int-resource-dongle: 3
```

# 参考
https://kubernetes.io/docs/tasks/administer-cluster/opaque-integer-resource-node/
https://kubernetes.io/docs/tasks/configure-pod-container/opaque-integer-resource/