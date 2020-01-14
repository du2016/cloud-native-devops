# scheduler-framework

本文将讲述如何使用scheduler-framework扩展原生调度器

目的： 在prefilter阶段检查pod是否添加有dely注释，如果未达到对应时间则不调度


# 分析需要实现的method


## 注册插件

WithPlugin返回一个注册选项，由此我们可以看出，我们的插件需要实现framework.PluginFactory 接口
```
func WithPlugin(name string, factory framework.PluginFactory) Option {
	return func(registry framework.Registry) error {
		return registry.Register(name, factory)
	}
}
```

该接口接收传入的附加参数和FrameworkHandle，关于FrameworkHandle的作用请查看之前文章

```
type PluginFactory = func(configuration *runtime.Unknown, f FrameworkHandle) (Plugin, error)
```

## PreFilterPlugin接口

```
// PreFilterPlugin is an interface that must be implemented by "prefilter" plugins.
// These plugins are called at the beginning of the scheduling cycle.
type PreFilterPlugin interface {
	Plugin
	// PreFilter is called at the beginning of the scheduling cycle. All PreFilter
	// plugins must return success or the pod will be rejected.
	PreFilter(ctx context.Context, state *CycleState, p *v1.Pod) *Status
	// PreFilterExtensions returns a PreFilterExtensions interface if the plugin implements one,
	// or nil if it does not. A Pre-filter plugin can provide extensions to incrementally
	// modify its pre-processed info. The framework guarantees that the extensions
	// AddPod/RemovePod will only be called after PreFilter, possibly on a cloned
	// CycleState, and may call those functions more than once before calling
	// Filter again on a specific node.
	PreFilterExtensions() PreFilterExtensions
}

// Plugin is the parent type for all the scheduling framework plugins.
type Plugin interface {
	Name() string
}
```

PreFilterExtensions()方法返回PreFilterExtensions接口

```
type PreFilterExtensions interface {
	// AddPod is called by the framework while trying to evaluate the impact
	// of adding podToAdd to the node while scheduling podToSchedule.
	AddPod(ctx context.Context, state *CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *schedulernodeinfo.NodeInfo) *Status
	// RemovePod is called by the framework while trying to evaluate the impact
	// of removing podToRemove from the node while scheduling podToSchedule.
	RemovePod(ctx context.Context, state *CycleState, podToSchedule *v1.Pod, podToRemove *v1.Pod, nodeInfo *schedulernodeinfo.NodeInfo) *Status
}
```

我们需要实现五个method:

- Name 返回插件名称
- PreFilter 对pod进行筛选
- PreFilterExtensions prefilter扩展功能，评估add/removepod的影响，如果不实现可返回nil
- AddPod 评估添加pod到node的影响
- RemovePod 评估删除pod到node的影响

# 代码实现

## 实现注册相关


```
const Name = "test"

var _ framework.PreFilterPlugin = &TestPlugin{}

type Args struct {
	KubeConfig string `json:"kubeconfig,omitempty"`
	Master     string `json:"master,omitempty"`
}

type TestPlugin struct {
	handle framework.FrameworkHandle
	Args   *Args
}

func New(rargs *runtime.Unknown, handle framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(rargs, args); err != nil {
		return nil, err
	}
	klog.Info(args)
	return &TestPlugin{
		handle: handle,
		Args:   args,
	}, nil
}
```


## 实现prefilter接口


```
# 返回名称，任何plugin都需要实现
func (self *TestPlugin) Name() string {
	return Name
}

# 实现PreFilter method
func (self *TestPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	klog.Error("into controller test")
	state.Write()
	var dtime int64
	var err error
    # 判断是否有延迟字段
	if v, ok := p.Annotations["delay"]; ok {
		if dtime, err = strconv.ParseInt(v, 10, 64); err != nil {
			return nil
		}
        # 距离当前大于延时间，则调度
		if time.Now().Unix()-p.CreationTimestamp.Unix() >= dtime {
			klog.Infof("scheduler: %s/%s", p.Namespace, p.Name)
			return nil
		}
        # 否则不调度
		klog.Infof("not reatch scheduler time: %s/%s", p.Namespace, p.Name)
		return framework.NewStatus(framework.Skip, "not reatch scheduler time")
	}
	return nil
}

# PreFilterExtensions AddPod method,返回nil即success
func (self *TestPlugin) AddPod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *schedulernodeinfo.NodeInfo) *framework.Status {
	return nil
}

# PreFilterExtensions RemovePod,返回nil即success
func (self *TestPlugin) RemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToRemove *v1.Pod, nodeInfo *schedulernodeinfo.NodeInfo) *framework.Status {
	return nil
}

# 这里也可以返回nil
func (self *TestPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	return self
}
```

# 测试

- 测试时使用kubeadm，停止原有kubelet

```
# 将static pod配置移走，kubelet会自动停止
mv /etc/kubernetes/manifests/kube-scheduler.yaml ./
```

- 编译运行

```
go build 
./test-scheduler-framework --config=config.yaml
```

- 测试pod

创建deploy 配置如下
```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: busybox
  name: busybox
  namespace: default
spec:
  replicas: 0
  selector:
    matchLabels:
      run: busybox
  template:
    metadata:
      annotations:
        delay: "15"
      labels:
        run: busybox
    spec:
      containers:
      - args:
        - sleep
        - "123456"
        image: busybox
        name: busybox
```

- 扩容观察pod状态

```
$ kubectl scale deploy busybox --replicas=1
$ kubectl get pods -l run=busybox -w
NAME                      READY   STATUS    RESTARTS   AGE
busybox-8d8554fc8-f9899   0/1     Pending   0          3s
busybox-8d8554fc8-f9899   0/1     Pending   0          85s
busybox-8d8554fc8-f9899   0/1     ContainerCreating   0          85s
busybox-8d8554fc8-f9899   0/1     ContainerCreating   0          86s
busybox-8d8554fc8-f9899   1/1     Running             0          90s
```

- 观察调度器日志
```
E1226 18:19:11.071899   92593 type.go:59] into controller test
I1226 18:19:11.071909   92593 type.go:70] not reatch scheduler time: default/busybox-8d8554fc8-f9899
E1226 18:19:11.071923   92593 framework.go:287] error while running "test" prefilter plugin for pod "busybox-8d8554fc8-f9899": not reatch scheduler time
E1226 18:19:11.071941   92593 factory.go:469] Error scheduling default/busybox-8d8554fc8-f9899: error while running "test" prefilter plugin for pod "busybox-8d8554fc8-f9899": not reatch scheduler time; retrying
E1226 18:19:11.071970   92593 scheduler.go:638] error selecting node for pod: error while running "test" prefilter plugin for pod "busybox-8d8554fc8-f9899": not reatch scheduler time
E1226 18:20:36.064233   92593 type.go:59] into controller test
I1226 18:20:36.064258   92593 type.go:67] scheduler: default/busybox-8d8554fc8-f9899
```

可以看到确实延时调度了，但是因为重新调度本身有时间间隔(30s)，所以并不是我们设置的值


扫描关注我:

![微信](http://img.rocdu.top/qrcode_for_gh_7457c3b1bfab_258.jpg)