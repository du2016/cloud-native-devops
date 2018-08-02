# 已官方bookinfo示例为基础

## 设置所有应用服务为v1版本

- 未启用mtls

```
istioctl create -f samples/bookinfo/routing/route-rule-all-v1.yaml
```

- 启用mtls

```
istioctl create -f samples/bookinfo/routing/route-rule-all-v1-mtls.yaml
```

- 查看路由规则

```
#当前都为v1版本
istioctl get virtualservices -o yaml
```

- 切换review服务到v2版本

```
istioctl replace -f samples/bookinfo/routing/route-rule-reviews-test-v2.yaml
```

- 查看路由规则

```
#当前为v1版本
istioctl get virtualservices  reviews -o yaml
```