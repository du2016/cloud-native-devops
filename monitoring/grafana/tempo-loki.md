
# Tempo简介
Grafana Tempo是一个开源、易于使用的大规模分布式跟踪后端.Tempo具有成本效益，仅需要对象存储即可运行，并且与Grafana，Prometheus和Loki深度集成.Tempo可以与任何开源跟踪协议一起使用,包括Jaeger、Zipkin和OpenTelemetry。它仅支持键/值查找，并且旨在与用于发现的日志和度量标准(示例性)协同工作.Tempo与Jaeger，Zipkin，OpenCensus和OpenTelemetry兼容.它以任何上述格式提取批处理，对其进行缓冲，然后将其写入GCS，S3或本地磁盘.因此，它强大、便宜且易于操作！

# 部署

## 部署grafana

```
kubectl create ns loki
helm install grafana stable/grafana -n loki
# 7.3版本支持Tempo，需要升级
kubectl set image deployment/grafana grafana=grafana/grafana:7.3.0 -n loki
```

获取grafana密码
```
kubectl get secret --namespace loki grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```

## 部署loki

使用helm进行部署

```
helm repo add loki https://grafana.github.io/loki/charts
helm repo update
```

为了后续能够使用loki产生traceid，需要设置为loki设置以下环境变量
```
        env:
        - name: JAEGER_AGENT_HOST
          value: tempo.default
        - name: JAEGER_ENDPOINT
          value: http://tempo.default:14268/api/traces
        - name: JAEGER_SAMPLER_TYPE
          value: const
        - name: JAEGER_SAMPLER_PARAM
          value: "1"
```

## 部署promtail

通过promtail进行日志收集，写入loki

promtail.yaml 文件内容如下
```
client:
  backoff_config:
    maxbackoff: 5s
    maxretries: 20
    minbackoff: 100ms
  batchsize: 102400
  batchwait: 1s
  external_labels: {}
  timeout: 10s
positions:
  filename: /run/promtail/positions.yaml
server:
  http_listen_port: 3101
target_config:
  sync_period: 10s

scrape_configs:
- job_name: kubernetes-pods-name
  pipeline_stages:
    - docker: {}

  kubernetes_sd_configs:
  - role: pod
  relabel_configs:
  - source_labels:
    - __meta_kubernetes_pod_label_name
    target_label: __service__
  - source_labels:
    - __meta_kubernetes_pod_node_name
    target_label: __host__
  - action: drop
    regex: ''
    source_labels:
    - __service__
  - action: replace
    replacement: $1
    separator: /
    source_labels:
    - __meta_kubernetes_namespace
    - __service__
    target_label: namespace_service
  - action: replace
    source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod_name
  - replacement: /var/log/pods/*$1/*.log
    separator: /
    source_labels:
    - __meta_kubernetes_pod_uid
    - __meta_kubernetes_pod_container_name
    target_label: __path__
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_host_ip
    target_label: pod_host_ip
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container_name
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_site
    target_label: site
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_app
    target_label: app
- job_name: kubernetes-pods-app
  pipeline_stages:
    - docker: {}

  kubernetes_sd_configs:
  - role: pod
  relabel_configs:
  - action: drop
    regex: .+
    source_labels:
    - __meta_kubernetes_pod_label_name
  - source_labels:
    - __meta_kubernetes_pod_label_app
    target_label: __service__
  - source_labels:
    - __meta_kubernetes_pod_node_name
    target_label: __host__
  - action: drop
    regex: ''
    source_labels:
    - __service__
  - action: replace
    replacement: $1
    separator: /
    source_labels:
    - __meta_kubernetes_namespace
    - __service__
    target_label: namespace_service
  - action: replace
    source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod_name
  - replacement: /var/log/pods/*$1/*.log
    separator: /
    source_labels:
    - __meta_kubernetes_pod_uid
    - __meta_kubernetes_pod_container_name
    target_label: __path__
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_host_ip
    target_label: pod_host_ip
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container_name
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_site
    target_label: site
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_app
    target_label: app
- job_name: kubernetes-pods-direct-controllers
  pipeline_stages:
    - docker: {}

  kubernetes_sd_configs:
  - role: pod
  relabel_configs:
  - action: drop
    regex: .+
    separator: ''
    source_labels:
    - __meta_kubernetes_pod_label_name
    - __meta_kubernetes_pod_label_app
  - action: drop
    regex: '[0-9a-z-.]+-[0-9a-f]{8,10}'
    source_labels:
    - __meta_kubernetes_pod_controller_name
  - source_labels:
    - __meta_kubernetes_pod_controller_name
    target_label: __service__
  - source_labels:
    - __meta_kubernetes_pod_node_name
    target_label: __host__
  - action: drop
    regex: ''
    source_labels:
    - __service__
  - action: replace
    replacement: $1
    separator: /
    source_labels:
    - __meta_kubernetes_namespace
    - __service__
    target_label: namespace_service
  - action: replace
    source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod_name
  - replacement: /var/log/pods/*$1/*.log
    separator: /
    source_labels:
    - __meta_kubernetes_pod_uid
    - __meta_kubernetes_pod_container_name
    target_label: __path__
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_host_ip
    target_label: pod_host_ip
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container_name
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_site
    target_label: site
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_app
    target_label: app
- job_name: kubernetes-pods-indirect-controller
  pipeline_stages:
    - docker: {}

  kubernetes_sd_configs:
  - role: pod
  relabel_configs:
  - action: drop
    regex: .+
    separator: ''
    source_labels:
    - __meta_kubernetes_pod_label_name
    - __meta_kubernetes_pod_label_app
  - action: keep
    regex: '[0-9a-z-.]+-[0-9a-f]{8,10}'
    source_labels:
    - __meta_kubernetes_pod_controller_name
  - action: replace
    regex: '([0-9a-z-.]+)-[0-9a-f]{8,10}'
    source_labels:
    - __meta_kubernetes_pod_controller_name
    target_label: __service__
  - source_labels:
    - __meta_kubernetes_pod_node_name
    target_label: __host__
  - action: drop
    regex: ''
    source_labels:
    - __service__
  - action: replace
    replacement: $1
    separator: /
    source_labels:
    - __meta_kubernetes_namespace
    - __service__
    target_label: namespace_service
  - action: replace
    source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod_name
  - replacement: /var/log/pods/*$1/*.log
    separator: /
    source_labels:
    - __meta_kubernetes_pod_uid
    - __meta_kubernetes_pod_container_name
    target_label: __path__
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_host_ip
    target_label: pod_host_ip
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container_name
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_site
    target_label: site
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_app
    target_label: app
- job_name: kubernetes-pods-static
  pipeline_stages:
    - docker: {}

  kubernetes_sd_configs:
  - role: pod
  relabel_configs:
  - action: drop
    regex: ''
    source_labels:
    - __meta_kubernetes_pod_annotation_kubernetes_io_config_mirror
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_component
    target_label: __service__
  - source_labels:
    - __meta_kubernetes_pod_node_name
    target_label: __host__
  - action: drop
    regex: ''
    source_labels:
    - __service__
  - action: replace
    replacement: $1
    separator: /
    source_labels:
    - __meta_kubernetes_namespace
    - __service__
    target_label: namespace_service
  - action: replace
    source_labels:
    - __meta_kubernetes_namespace
    target_label: namespace
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_name
    target_label: pod_name
  - replacement: /var/log/pods/*$1/*.log
    separator: /
    source_labels:
    - __meta_kubernetes_pod_annotation_kubernetes_io_config_mirror
    - __meta_kubernetes_pod_container_name
    target_label: __path__
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_host_ip
    target_label: pod_host_ip
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_container_name
    target_label: container_name
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_site
    target_label: site
  - action: replace
    source_labels:
    - __meta_kubernetes_pod_label_app
    target_label: app
```

创建configmap

kubectl create configmap promtail-config --from-file=promtail.yaml=promtail-config -n loki

promtail-ds.yaml内容如下

```
# Source: promtail/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: promtail
    release: calling-quail
  name: promtail
  namespace: loki
---
# Source: promtail/templates/clusterrole.yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  labels:
    app: promtail
    release: calling-quail
  name: promtail-clusterrole
  namespace: loki
rules:
- apiGroups: [""] # "" indicates the core API group
  resources:
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "watch", "list"]
---
# Source: promtail/templates/clusterrolebinding.yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: promtail-clusterrolebinding
  labels:
    app: promtail
    release: calling-quail
subjects:
  - kind: ServiceAccount
    name: promtail
    namespace: loki
roleRef:
  kind: ClusterRole
  name: promtail-clusterrole
  apiGroup: rbac.authorization.k8s.io
---
# Source: promtail/templates/role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: promtail
  namespace: loki
  labels:
    app: promtail
    release: calling-quail
rules:
- apiGroups:      ['extensions']
  resources:      ['podsecuritypolicies']
  verbs:          ['use']
  resourceNames:  [promtail]
---
# Source: promtail/templates/rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: promtail
  namespace: loki
  labels:
    app: promtail
    release: calling-quail
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: promtail
subjects:
- kind: ServiceAccount
  name: promtail
---
# Source: promtail/templates/daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: promtail
  namespace: loki
  labels:
    app: promtail
    release: calling-quail
spec:
  selector:
    matchLabels:
      app: promtail
      release: calling-quail
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: promtail
        release: calling-quail
      annotations:
        checksum/config: 674c34bb782c907c837d96c00242c2c953b4b482d2c00475b03145696d4f301f
        prometheus.io/port: http-metrics
        prometheus.io/scrape: "true"
    spec:
      serviceAccountName: promtail
      containers:
        - name: promtail
          image: grafana/promtail:v1.3.0
          imagePullPolicy: IfNotPresent
          args:
            - "-config.file=/etc/promtail/promtail.yaml"
            - "-client.url=http://loki:3100/loki/api/v1/push"
          volumeMounts:
            - name: config
              mountPath: /etc/promtail
            - name: run
              mountPath: /run/promtail
            - mountPath: /var/lib/docker/containers
              name: docker
              readOnly: true
            - mountPath: /var/log/pods
              name: pods
              readOnly: true
          env:
          - name: HOSTNAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          ports:
            - containerPort: 3101
              name: http-metrics
          securityContext:
            readOnlyRootFilesystem: true
            runAsGroup: 0
            runAsUser: 0
          readinessProbe:
            failureThreshold: 5
            httpGet:
              path: /ready
              port: http-metrics
            initialDelaySeconds: 10
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
      tolerations:
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: Exists
      volumes:
        - name: config
          configMap:
            name: promtail-config
        - name: run
          hostPath:
            path: /run/promtail
        - hostPath:
            path: /var/lib/docker/containers
          name: docker
        - hostPath:
            path: /var/log/pods
          name: pods
```

部署 promtail daemonset

```
kubectl apply -f promtail-ds.yaml
```

## 部署tempo

这里为了快速安装选择了single-binary的安装方式，先上建议使用微服务部署模式，Tempo的微服务部署具有容错能力，高容量，可独立扩展。

```
git clone https://github.com/grafana/tempo
cd tempo/operations/helm
helm install tempo-single-binary/ --generate-name -n loki
```

# grafana配置

因为我使用的是kind,无法直接访问服务，需要使用port-forward进行端口转发

kubectl port-forward --namespace loki service/grafana 3000:80

## 配置loki数据源

添加一个loki数据源

![](http://img.rocdu.top/20201029/1.png)

创建派生字段，派生字段可用于从日志消息中提取新字段并根据其值创建链接。
![](http://img.rocdu.top/20201029/2.png)

## 配置tempo数据源

![](http://img.rocdu.top/20201029/3.png)


## 数据查询

先随便查询几次以让loki产生数据，输入以下内容查询loki服务产生的包含traceID的数据。

```
{namespace="loki",pod_name="loki-0"} |= "traceID"
```
可以看到TraceID字段后面的Tempo按钮，点击可以看到对应的trace信息

![](http://img.rocdu.top/20201029/4.png)


# 总结

grafana tempo的诞生完善了grafana traceing体系，实现了grafana apm logging、tracing、metrics最后的一环，相对来说较便捷的集成到了grafana ui中，有点对标elk stack的意思，不过elk stack的链路追踪好像是付费功能。tempo其键值存储实现也决定了其功能的局限性，还不支持链路的完整展示，在查询时必须要先获得traceid才能进行查询，所以只能通过日志打印traceid,然后再根据traceid进行查询从而进行展示。

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
