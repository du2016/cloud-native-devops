# Readiness Gates

kubernetes从1.11版本开始引入了Pod Ready++特性对Readiness探测机制进行扩展，在1.14版本时达到GA稳定版本，称其为Pod Readiness Gates。
通过Pod Readiness Gates机制，用户可以将自定义的ReadinessProbe探测方式设置在Pod上，辅助kubernetes设置Pod何时达到服务可用状态Ready，为了使自定义的ReadinessProbe生效，用户需要提供一个外部的控制器Controller来设置相应的Condition状态。Pod的Readiness Gates在pod定义中的ReadinessGates字段进行设置，

# 应用场景

其根本目的是为了让用户能够随心所欲的控制pod的状态，可以想到的场景如下：

- 阶段更新（例如更新deployment，我们想先更新部分比例查看效果）
- 原地升级（保证原地升级的优雅性）
- 依赖于其他组件提供服务（保证关联资源状态完成后再接收流量）

关于原地升级，需要了解pod spec的校验机制，
https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/validation/validation.go#L3927。

更新pod的image后kubelet会删除container进行重建，但是直接更新显然会造成服务异常，由此我们可以借助readnessgate 在condition为false时动态的更新image实现原地升级

# 用法示例

如下设置了一个类型为www.example.com/feature-1的新Readiness Gates：

```
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: centos
  name: centos
  namespace: default
spec:
  containers:
  - args:
    - sleep
    - 10d
    image: centos
    imagePullPolicy: Always
    name: centos
  readinessGates:
    - conditionType: "www.example.com/feature-1"
```

apply后查看Ready condition是false,如果设置有endpoint，也不会出现在endpoint列表里。如果要想容器正常提供服务，就需要将对应的conditionType设置为true.
通俗的来讲就是设置readinessGates字段，然后将对应的condition通过patch操作设置为true

注意kubectl 是无法通过patch更改status里面的字段的。

## 调用方式

- curl 直接调用

```
kubectl proxy
curl http://localhost:8001/api/v1/namespaces/default/pods/centos/status -X PATCH -H "Content-Type: application/json-patch+json" -d '[{"op": "add", "path": "/status/conditions/-", "value": {"type": "www.example.com/feature-1", "status": "True", "lastProbeTime": null}}]'
 kubectl get pods centos -o json | jq .status.conditions
[
  {
    "lastProbeTime": null,
    "lastTransitionTime": null,
    "status": "True",
    "type": "www.example.com/feature-1"
  },
{
  "lastProbeTime": null,
  "lastTransitionTime": "2020-07-11T03:19:55Z",
  "status": "True",
  "type": "Ready"
},
  ...
```

可以看到此时容器状态已经正常了

- 通过clientgo对容器进行patch


# 实现readnessgate controller

接下来我们将通过clientgo实现一个简单的readnessgate controller，用于通过标签控制容器的 readnessgate.

实现以下功能：

- 当example标签为true时设置www.example.com/feature-1 condition为true
- 当example标签为false时设置www.example.com/feature-1 condition为false

## condition patch


通过以下patch来对pod添加定影的condition
```
const addTruePatch = "[{\"op\": \"add\", \"path\": \"/status/conditions/-\", \"value\": {\"type\": \"www.example.com/feature-1\", \"status\": \"True\", \"lastProbeTime\": null}}]"
const addFalsePatch = "[{\"op\": \"add\", \"path\": \"/status/conditions/-\", \"value\": {\"type\": \"www.example.com/feature-1\", \"status\": \"False\", \"lastProbeTime\": null}}]"
```


## 定义如下informer

通过informer机制来监听pod的add和update事件，作出相应操作。

```
	clientset, err = kubernetes.NewForConfig(config)
	podLW := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"pods",
		v1.NamespaceDefault,
		fields.Everything(),
	)
	_, podinformer := cache.NewInformer(
		podLW,
		&v1.Pod{},
		time.Second*30,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handlepodsAdd,
			UpdateFunc: handlepodsupdate,
		},
	)
```

## addpatch

如果没有标签则通过clientset 进行patch操作

```
clientset.
CoreV1().
Pods(pod.Namespace).
Patch(context.TODO(),
    pod.Name,
    types.JSONPatchType,
    []byte(addTruePatch),
    metav1.PatchOptions{},
    "status")
```

## replace patch

如果已有但是状态不符合预期则进行更新

```
patch,err:=getchangePatch(pod, true)
if err!=nil {
    log.Println("get patch error:",err)
}
patchbytes, err := json.Marshal(patch)
if err != nil {
    log.Println("Marshal patch error:",err)
}
log.Println("patch pod: ",string(patchbytes))
_, err = clientset.
    CoreV1().
    Pods(pod.Namespace).
    Patch(context.TODO(),
        pod.Name,
        types.JSONPatchType,
        patchbytes,
        metav1.PatchOptions{},
        "status")
```


扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
