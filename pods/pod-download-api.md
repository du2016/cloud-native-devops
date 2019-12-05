# 暴露 Pod 和 Container 字段给运行的容器

# 环境变量

容器创建时运行的所有服务的列表都会作为环境变量提供给容器。

- 默认环境变量

查看环境变量：
```bash
env
```

- fieldRef
  - 使用pod字段
    ```
          env:
            - name: MY_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: MY_POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: MY_POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: MY_POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
            - name: MY_POD_SERVICE_ACCOUNT
              valueFrom:
                fieldRef:
                  fieldPath: spec.serviceAccountName
    ```
  - 使用pod字段
    ```
          env:
            - name: MY_CPU_REQUEST
              valueFrom:
                resourceFieldRef:
                  containerName: test-container
                  resource: requests.cpu
            - name: MY_CPU_LIMIT
              valueFrom:
                resourceFieldRef:
                  containerName: test-container
                  resource: limits.cpu
            - name: MY_MEM_REQUEST
              valueFrom:
                resourceFieldRef:
                  containerName: test-container
                  resource: requests.memory
            - name: MY_MEM_LIMIT
              valueFrom:
                resourceFieldRef:
                  containerName: test-container
                  resource: limits.memory
    ```

# downloadapi

- 获取pod字段
```
  volumes:
    - name: podinfo
      downwardAPI:
        items:
          - path: "labels"
            fieldRef:
              fieldPath: metadata.labels
          - path: "annotations"
            fieldRef:
              fieldPath: metadata.annotations

```

- 获取container字段
```
  volumes:
    - name: podinfo
      downwardAPI:
        items:
          - path: "cpu_limit"
            resourceFieldRef:
              containerName: client-container
              resource: limits.cpu
          - path: "cpu_request"
            resourceFieldRef:
              containerName: client-container
              resource: requests.cpu
          - path: "mem_limit"
            resourceFieldRef:
              containerName: client-container
              resource: limits.memory
          - path: "mem_request"
            resourceFieldRef:
              containerName: client-container
              resource: requests.memory

```
