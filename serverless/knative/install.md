

本指南将引导您完成最新版本Knative的安装。

Knative有两个组件，可以独立安装或一起使用。为了帮助您挑选适合自己的作品，以下是每个组件的简要说明：

- Serving 为基于无状态请求的服务提供了一种零扩展抽象。
- Eventing提供了抽象来启用绑定事件源（例如Github Webhooks，Kafka）和使用者（例如Kubernetes或Knative Services）的绑定。


Knative还具有一个Observability插件，该插件提供了标准工具，可用于查看Knative上运行的软件的运行状况

# 在你开始之前
本指南假定您要在Kubernetes群集上安装上游Knative版本。 越来越多的供应商已经管理Knative产品。 有关完整列表，请参见Knative产品页面。

Knative v0.15.0需要Kubernetes集群v1.15或更高版本，以及兼容的kubectl。 本指南假定您已经创建了Kubernetes集群，并且在Mac或Linux环境中使用bash。 在Windows环境中需要调整一些命令

# 安装Serving组件

1.使用以下命令安装crd

```
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.15.0/serving-crds.yaml
```

2.serving的安装核心组件

```
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.15.0/serving-core.yaml
```

3.安装网络层
    

    - 安装contour
    
    kubectl apply --filename https://github.com/knative/net-contour/releases/download/v0.15.0/contour.yaml
    
    - 安装knative contour controller
    
    kubectl apply --filename https://github.com/knative/net-contour/releases/download/v0.15.0/net-contour.yaml
    
    - 配置knativeserving使用Contour
    
    kubectl patch configmap/config-network \
      --namespace knative-serving \
      --type merge \
      --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
      
    - 获取ip
    
    kubectl --namespace contour-external get service envoy
    
4. 配置DNS
  
  因为我们使用kind安装此步骤跳过
  
# 安装Eventing组件

1.安装crd

```
kubectl apply  --selector knative.dev/crd-install=true \
--filename https://github.com/knative/eventing/releases/download/v0.15.0/eventing.yaml
```

2.安装Eventing组件

```
kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.15.0/eventing.yaml
```

3.安装默认channel

这里选用kafka

- 创建kafka命名空间

```  
  kubectl create namespace kafka
```

  - 安装Strimzi operator

```  
  curl -L "https://github.com/strimzi/strimzi-kafka-operator/releases/download/0.16.2/strimzi-cluster-operator-0.16.2.yaml" \
    | sed 's/namespace: .*/namespace: kafka/' \
    | kubectl -n kafka apply -f -
```

  - 查看kafka的yaml

```
  apiVersion: kafka.strimzi.io/v1beta1
  kind: Kafka
  metadata:
    name: my-cluster
  spec:
    kafka:
      version: 2.4.0
      replicas: 1
      listeners:
        plain: {}
        tls: {}
      config:
        offsets.topic.replication.factor: 1
        transaction.state.log.replication.factor: 1
        transaction.state.log.min.isr: 1
        log.message.format.version: "2.4"
      storage:
        type: ephemeral
    zookeeper:
      replicas: 3
      storage:
        type: ephemeral
    entityOperator:
      topicOperator: {}
      userOperator: {}
```
  - 部署
```
    kubectl apply -n kafka -f kafka.yaml
```
  - 检查kafka集群状态
```
    $ kubectl get pods -n kafka
    NAME                                          READY   STATUS    RESTARTS   AGE
    my-cluster-entity-operator-65995cf856-ld2zp   3/3     Running   0          102s
    my-cluster-kafka-0                            2/2     Running   0          2m8s
    my-cluster-zookeeper-0                        2/2     Running   0          2m39s
    my-cluster-zookeeper-1                        2/2     Running   0          2m49s
    my-cluster-zookeeper-2                        2/2     Running   0          2m59s
    strimzi-cluster-operator-77555d4b69-sbrt4     1/1     Running   0          3m14s
```
  - 安装kafkachannel
```
  curl -L "https://github.com/knative/eventing-contrib/releases/download/v0.15.0/kafka-channel.yaml" \
   | sed 's/REPLACE_WITH_CLUSTER_URL/my-cluster-kafka-bootstrap.kafka:9092/' \
   | kubectl apply --filename -
```
  - 安装broker
```
  kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.15.0/channel-broker.yaml
```
  - 配置使用的broker
```
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: config-br-defaults
    namespace: knative-eventing
  data:
    default-br-config: |
      # This is the cluster-wide default broker channel.
      clusterDefault:
        brokerClass: ChannelBasedBroker
        apiVersion: v1
        kind: ConfigMap
        name: kafka-channel
        namespace: knative-eventing
```

  - broker具体配置

```
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: imc-channel
    namespace: knative-eventing
  data:
    channelTemplateSpec: |
      apiVersion: messaging.knative.dev/v1beta1
      kind: InMemoryChannel
  ---
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: kafka-channel
    namespace: knative-eventing
  data:
    channelTemplateSpec: |
      apiVersion: messaging.knative.dev/v1alpha1
      kind: KafkaChannel
      spec:
        numPartitions: 3
        replicationFactor: 1
```


查看 eventing组件状态

```
kubectl get pods --namespace knative-eventing
```