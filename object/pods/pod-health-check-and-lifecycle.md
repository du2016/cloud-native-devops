# 健康检查与生命周期

健康检查作为服务存存活的依据，kubelet将根据健康检查结果判定何时重启容器，决定了pod的生命周期。

# 类型

- liveness probe（存活探针）
- readiness probe（就绪探针）

# 定义

在源码中liveness probe和readiness probe结构体完全一致

通用参数：
- initialDelaySeconds 启动后的初始化时间
- TimeoutSeconds 超时时间
- PeriodSeconds 检查频率
- SuccessThreshold 几次判断为成功
- FailureThreshold 几次判断为失败

## 健康检查方式

- Exec
    参数：
      - Command
```
    livenessProbe:
      exec:
        command:
        - cat
        - /tmp/healthy
      initialDelaySeconds: 5
      periodSeconds: 5
```

- HTTPGet
    参数：
      - Path
      - Port
      - Host
      - Scheme
      - HTTPHeaders
```
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080 #可以使用定义的port name
        httpHeaders:
        - name: X-Custom-Header
          value: Awesome
      initialDelaySeconds: 3
      periodSeconds: 3
```

- TCPSocket
- HTTPGet
    参数：
      - Port
      - Host
```
    livenessProbe:
      tcpSocket:
        port: 8080
      initialDelaySeconds: 15
      periodSeconds: 20
```


# 生命周期

- 挂起（Pending）：Pod 已被 Kubernetes 系统接受，但有一个或者多个容器镜像尚未创建。等待时间包括调度 Pod 的时间和通过网络下载镜像的时间，这可能需要花点时间。
- 运行中（Running）：该 Pod 已经绑定到了一个节点上，Pod 中所有的容器都已被创建。至少有一个容器正在运行，或者正处于启动或重启状态。
- 成功（Succeeded）：Pod 中的所有容器都被成功终止，并且不会再重启。
- 失败（Failed）：Pod 中的所有容器都已终止了，并且至少有一个容器是因为失败终止。也就是说，容器以非0状态退出或者被系统终止。
- 未知（Unknown）：因为某些原因无法取得 Pod 的状态，通常是因为与 Pod 所在主机通信失败。

## restartpolicy

- Always
- OnFailure
- Never

### 特殊情况

- Job 仅适用于重启策略为 OnFailure 或 Never 的 Pod
- ReplicationController 仅适用于具有 restartPolicy 为 Always 的 Pod
- DaemonSet 为每台机器运行一个 Pod 


## 定义启动结束操作

```
apiVersion: v1
kind: Pod
metadata:
  name: lifecycle-demo
spec:
  containers:
  - name: lifecycle-demo-container
    image: nginx
    lifecycle:
      postStart:
        exec:
          command: ["/bin/sh", "-c", "echo Hello from the postStart handler > /usr/share/message"]
      preStop:
        exec:
          command: ["/usr/sbin/nginx","-s","quit"]
```

### 说明

postStart和容器启动为异步执行，postStart执行完后pod才可能变为running状态

Kubernetes在容器创建之后就会马上发送postStart事件，但是并没法保证一定会 这么做，
它会在容器入口被调用之前调用postStart操作，因为postStart的操作跟容器的操作是异步的，
而且Kubernetes控制台会锁住容器直至postStart完成，
因此容器只有在 postStart操作完成之后才会被设置成为RUNNING状态。


# pod readinessgate

1.11 引入 ，1.14稳定

定义时添加spec.readinessGates
```
Kind: Pod
...
spec:
  readinessGates:
    - conditionType: "www.example.com/feature-1"
```
然后通过api更新即可控制pod的状态