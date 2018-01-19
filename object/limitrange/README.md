```
apiVersion: v1
kind: LimitRange
metadata:
  name: limitrange
spec:
  limits:
  - default:
      cpu: "8"
      memory: 16Gi
    defaultRequest:
      cpu: 200m
      memory: 1Gi
    max:
      cpu: "8"
      memory: 16Gi
    min:
      cpu: 200m
      memory: 1Gi
    type: Container
```