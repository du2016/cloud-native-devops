# edgecontroller


# initConfig

```
func initConfig() {
	config.InitBufferConfig() // 初始化云边通信相关channel的管道大小
	config.InitContextConfig() //设置edgecontroller默认发/收/响应模块的名称
	config.InitKubeConfig()  //初始化kube相关的名称
	config.InitLoadConfig() // 设置相关携程数量
	config.InitMessageLayerConfig()  // 设置上下文类型
}
```


# NewDownstreamController

用于watch k8s apiserver  同步消息到edge

- 创建kubeclient

## NewPodManager

通过配置创建了pod 管理器

```
func NewPodManager(kubeClient *kubernetes.Clientset, namespace, nodeName string) (*PodManager, error) {
	var lw *cache.ListWatch
	if "" == nodeName {
		lw = cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", namespace, fields.Everything())
	} else {
		selector := fields.OneTermEqualSelector("spec.nodeName", nodeName)
		lw = cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", namespace, selector)
	}
	realEvents := make(chan watch.Event, config.PodEventBuffer)
	mergedEvents := make(chan watch.Event, config.PodEventBuffer)
	rh := NewCommonResourceEventHandler(realEvents)
	si := cache.NewSharedInformer(lw, &v1.Pod{}, 0)
	si.AddEventHandler(rh)

	pm := &PodManager{realEvents: realEvents, mergedEvents: mergedEvents}

	stopNever := make(chan struct{})
	go si.Run(stopNever)
	go pm.merge()

	return pm, nil
}
```


NewCommonResourceEventHandler注册了一系列的回调函数，当有对应的事件时，将事件写入realEvents
然后启动一个shardinformer 用于同步apiserver到cache

接着调用了podmanager的merge方法，
从realEvents中读取事件，根据动作操作pods这个字典，并将信息发送给mergedEvents


## NewConfigMapManager

创建一个map chan,watch apiserver,将事件写入，其他manager也一样

## initLocating

用于判断configmap和secret发送到哪个节点

- 根据标签过滤 "node-role.kubernetes.io/edge"： ""
- 将节点的状态缓存到本地cache
- 获取所有pods
- 如果pod的节点是边缘节点则执行AddOrUpdatePod

## AddOrUpdatePod

- 获取pod的configmap和secret
- 根据configmap获取节点，如果原本有节点则添加，没有则新建
- srecret同上，最终维护了一个configmap/secret 到节点的映射

## Start

- syncPod
    - 从mergedEvents获取pod
    - 判断是否为边缘节点
    - 根据不同的动作构建消息发送resource消息
    - 发送消息到beehive

- syncConfigMap
- syncSecret
    - 不同动作决定动作
    - 构建消息
    - 发送消息
- syncEdgeNodes
- syncService
- syncEndpoints



# NewUpstreamController

用于从边缘同步消息给k8s apiserver

- updateNodeStatus
- updatePodStatus 
- queryConfigMap
- queryEndpoints
- queryPersistentVolume
- queryPersistentVolumeClaim
- queryVolumeAttachment
- queryNode
- updateNode

