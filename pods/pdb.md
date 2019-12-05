PodDisruptionBudget(pdb)限制应用程序的并发终端数，保证服务的可用性
（例如更新或其他动作时，可能造成pod进入notready状态，pdb保证服务所属的pod尽可能的处于ready状态）

使用minAvailable设置PDB
```
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: zk-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: zookeeper
```

使用maxUnavailable 设置PDB

```
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: zk-pdb
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: zookeeper
```