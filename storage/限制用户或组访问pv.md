## 通过GID限制PV访问控制

```
kind: PersistentVolume
apiVersion: v1
metadata:
  name: pv1
  annotations:
    pv.beta.kubernetes.io/gid: "1234"
```
挂在该PV需要pod有相同GID