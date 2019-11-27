# edged

在边缘管理容器化的应用程序,edged实际上是kubelet的精简版本，复用了kubelet的主要功能，实现了和其他edgecore组件的通信同步功能

## 初始化

- 设置backoff的值
- 创建podmanager
- 创建image gc policy
- 创建事件记录器
- 初始化edged
- 创建pods目录
- 创建livenessmanager
- 定义noderef
- 创建StateProvider接口实现，用于在垃圾回收期间获取镜像信息
- 创建镜像GC策略
- 如果remote-image-endpoint为空则设置为remote-runtime-endpoint
- 根据配置创建一个docker shim grpc客户端
- 设置dns地址及配置
- 创建一个containerRefManager
- 创建image和运行时管理程序
- 创建容器声明周期管理器
- 创建容器运行时管理器
- 创建容器管理器
- 创建image GC管理器
- 创建容器GC管理器
- 创建卷插件管理器

## 启动

- 创建元数据管理客户端，用于和beehive交互同步消息
- 创建Clientset 包含kube client（k8s clientset）和MetaClient
- 创建pod状态管理器
- 创建、运行volume管理器 
- 创建运行probemanager 
- 运行pod add worker 
- 运行pod delete worker 
- 运行pod状态管理器
- 启动edge server,10255/pods
- 启动imageGCmanager
- 启动容器gc
- 创建CSI插件管理器
- sync pod


# syncPod


- 发送消息给metamanager获取register-node-namespace指定的命名空间下的pod列表
- 从beehive中读取发给自己的消息
- 获取资源的类型和id
- 当资源类型为pod
  - 当动作为response切ID为空且来源为metamanager,将pod加入运行队列
  - 当动作为response且ID为空来源为metamanager edgecontroller，将坡道加入队列
  - 其他情况
    - insert 加入队列
    - update 更新pod到指定队列
    - delete 加入删除队列
- configmap 对cachestore进行增删改
- secret 对secret进行增删改
- volume 对volume进行增删改

# consumePodAddition

重点看一下consumePodAddition

```
func (e *edged) consumePodAddition(namespacedName *types.NamespacedName) error {
	podName := namespacedName.Name
	klog.Infof("start to consume added pod [%s]", podName)
	pod, ok := e.podManager.GetPodByName(namespacedName.Namespace, podName)
	if !ok || pod.DeletionTimestamp != nil {
		return apis.ErrPodNotFound
	}

	if err := e.makePodDataDirs(pod); err != nil {
		klog.Errorf("Unable to make pod data directories for pod %q: %v", format.Pod(pod), err)
		return err
	}

	if err := e.volumeManager.WaitForAttachAndMount(pod); err != nil {
		klog.Errorf("Unable to mount volumes for pod %q: %v; skipping pod", format.Pod(pod), err)
		return err
	}

	secrets, err := e.getSecretsFromMetaManager(pod)
	if err != nil {
		return err
	}

	curPodStatus, err := e.podCache.Get(pod.GetUID())
	if err != nil {
		klog.Errorf("Pod status for %s from cache failed: %v", podName, err)
		return err
	}

	result := e.containerRuntime.SyncPod(pod, curPodStatus, secrets, e.podAdditionBackoff)
	if err := result.Error(); err != nil {
		// Do not return error if the only failures were pods in backoff
		for _, r := range result.SyncResults {
			if r.Error != kubecontainer.ErrCrashLoopBackOff && r.Error != images.ErrImagePullBackOff {
				// Do not record an event here, as we keep all event logging for sync pod failures
				// local to container runtime so we get better errors
				return err
			}
		}

		return nil
	}

	e.workQueue.Enqueue(pod.UID, utilwait.Jitter(time.Minute, workerResyncIntervalJitterFactor))
	klog.Infof("consume added pod [%s] successfully\n", podName)
	return nil
}
```

在edgecore启动的时间发送一个获取pod列表的消息给metamanager，当启动时间，从metamanager获取configmap和secret从而保证离线时间依旧能够运行