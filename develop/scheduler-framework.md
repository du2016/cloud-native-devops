# scheduling-framework

scheduling framework 是Kubernetes Scheduler的一种新的可插入架构，可简化调度程序的自定义,
它向现有的调度程序中添加了一组新的`plugin`API。插件被编译到调度程序中。
这些API允许大多数调度功能实现为插件，同时使调度`core`保持简单且可维护。有关该框架设计的更多技术信息，请参阅scheduling framework的
[设计建议](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/20180409-scheduling-framework.md)。

# 框架工作流程

scheduling framework定义了一系列扩展点，调度插件注册提交一个或多个扩展点，
一些可以改变调度结果，一些仅用于提供信息，
每次调度一个Pod的尝试都分为两个阶段，即scheduling cycle（调度周期）和binding cycle（绑定周期）。


# 调度周期和绑定周期

调度周期为Pod选择一个节点，并且绑定周期将该决策应用于集群。调度周期和绑定周期一起被称为`scheduling context`。

调度周期是串行运行的，而绑定周期可能是并行的。

如果确定Pod不可调度或存在内部错误，则可以中止调度周期或绑定周期。Pod将返回队列并重试。

# 扩展点


下图显示了Pod的调度上下文以及调度框架公开的扩展点。在此图片中，`Filter`等效于`预选`，`Scoring`等效于`优选`。

一个插件可以在多个扩展点注册以执行更复杂或有状态的任务

![scheduling framework extension points](http://img.rocdu.top/20191224/scheduling-framework-extensions.png)

### Queue sort

这些插件用于在调度队列中对Pod进行排序。队列排序插件本质上将提供`less(Pod1, Pod2)`功能。一次只能启用一个队列排序插件。

### Pre-filter

这些插件用于预处理有关Pod的信息，或检查集群或Pod必须满足的某些条件。如果预过滤器插件返回错误，则调度周期将中止。

### Filter

这些插件用于过滤无法运行Pod的节点。对于每个节点，调度程序将按其配置顺序调用过滤器插件。
如果有任何过滤器插件将节点标记为不可行，则不会为该节点调用其余插件。可以同时评估节点。

### Post-filter

这是一个信息扩展点。将使用通过过滤阶段的节点列表来调用插件。插件可以使用这些数据来更新内部状态或生成日志/指标。

> 注意: 希望执行“预评分”工作的插件应使用后过滤器扩展点。

### Scoring

这些插件用于对已通过过滤阶段的节点进行排名。调度程序将为每个节点调用每个计分插件。
将有一个定义明确的整数范围，代表最小和最大分数。
在 Normalize scoring 阶段之后，调度程序将根据配置的插件权重合并所有插件的节点分数

### Normalize scoring

这些插件用于在调度程序计算节点的最终排名之前修改分数。
注册此扩展点的插件将被调用，并带有同一插件的评分结果。每个插件每个调度周期调用一次。

例如，假设一个插件BlinkingLightScorer根据节点闪烁的指示灯数量进行排序。

```
func ScoreNode(_ *v1.pod, n *v1.Node) (int, error) {
   return getBlinkingLightCount(n)
}
```

但是，与相比，闪烁灯的最大数量可能会比 NodeScoreMax小。要解决此问题，BlinkingLightScorer也应该注册该扩展点。

```
func NormalizeScores(scores map[string]int) {
   highest := 0
   for _, score := range scores {
      highest = max(highest, score)
   }
   for node, score := range scores {
      scores[node] = score*NodeScoreMax/highest
   }
}
```

如果任何规范化评分插件返回错误，则调度周期将中止。

> 注意：希望执行`pre-reserve`工作的插件应使用规范化评分扩展点。



### Reserve

这是一个信息扩展点。当为给定Pod保留节点上的资源时，维护运行时状态的插件（也称为`stateful plugins`）应使用此扩展点由调度程序通知。
这是在调度程序实际将Pod绑定到节点之前发生的，并且它存在是为了防止竞争条件，同时调度程序会等待绑定成功。

这是调度周期的最后一步。一旦Pod处于保留状态，它将在绑定周期结束时触发Unreserve插件（失败）或 Post-bind插件（成功）。

注意：此概念以前称为`assume`。

### Permit

这些插件用于防止或延迟Pod的绑定。许可插件可以做以下三件事之一。

1.  **approve** \
    所有许可插件都批准Pod后，将其发送以进行绑定。

1.  **deny** \
    如果任何许可插件拒绝Pod，则将其返回到调度队列。这将触发Unreserve插件。

1.  **wait** (with a timeout) \
    如果许可插件返回`wait`，则Pod将保持在许可阶段，直到插件批准它为止。
    如果发生超时，wait将变为拒绝，并且Pod将返回到调度队列，从而触发 Unreserve插件。

批准Pod绑定:

尽管任何插件都可以从缓存中访问`waiting`的Pod列表并批准它们（请参阅参考资料FrameworkHandle），
但我们希望只有allow插件才能批准处于`waiting`状态的预留Pod的绑定。批准Pod后，将其发送到预绑定阶段。

### Pre-bind

这些插件用于执行绑定Pod之前所需的任何工作。例如，预绑定插件可以在允许Pod在此处运行之前预配置网络卷并将其安装在目标节点上。

如果任何预绑定插件返回错误，则Pod被拒绝并返回到调度队列。

### Bind

这些插件用于将Pod绑定到节点。在所有预绑定插件完成之前，不会调用绑定插件。每个绑定插件均按配置顺序调用。
绑定插件可以选择是否处理给定的Pod。如果绑定插件选择处理Pod，则会跳过其余的绑定插件。

### Post-bind

这是一个信息扩展点。成功绑定Pod后，将调用后绑定插件。这是绑定周期的结束，可用于清理关联的资源。




### Unreserve

这是一个信息扩展点。如果Pod被保留，然后在以后的阶段中被拒绝，则将通知取消保留的插件。取消保留的插件应清除与保留的Pod相关联的状态。

使用此扩展点的插件通常也应使用 Reserve。

## Plugin API

插件API分为两个步骤。首先，插件必须注册并配置，然后才能使用扩展点接口。扩展点接口具有以下形式。

```
type Plugin interface {
   Name() string
}

type QueueSortPlugin interface {
   Plugin
   Less(*v1.pod, *v1.pod) bool
}

type PreFilterPlugin interface {
   Plugin
   PreFilter(PluginContext, *v1.pod) error
}

// ...
```

# CycleState

大多数*插件函数将使用CycleState参数调用。 CycleState表示当前的调度上下文。 
CycleState将提供API，用于访问范围为当前调度上下文的数据。
因为绑定周期可以同时执行，所以插件可以使用CycleState来确保它们正在处理正确的请求。

CycleState还提供类似于context.WithValue的API，可用于在不同扩展点的插件之间传递数据。
多个插件可以共享状态或通过此机制进行通信。仅在单个调度上下文中保留状态。值得注意的是，假定插件是受信任的。
调度程序不会阻止一个插件访问或修改另一个插件的状态。

唯一的例外是队列排序插件。 

> 警告：在调度上下文结束后，通过CycleState获得的数据无效，并且插件保存该数据的引用的时间不应超过必要的时间。

# FrameworkHandle

虽然CycleState提供与单个调度上下文有关的API，但是FrameworkHandle提供与插件的生存期有关的API。
这就是插件如何获取客户端（kubernetes.Interface）和SharedInformerFactory或从调度程序的群集状态缓存读取数据的方式。
该句柄还将提供API以列出和批准或拒绝等待的Pod。

警告：FrameworkHandle提供对kubernetes API服务器和调度程序的内部缓存的访问。不能保证两者都是同步的，编写使用这两个数据的插件时应格外小心。
 
要实现有用的功能，必须为插件提供对API服务器的访问权限，特别是当这些功能使用了调度程序通常不考虑的对象类型时，
尤其如此。提供SharedInformerFactory可使插件安全共享缓存。

# Plugin Registration

每个插件必须定义一个构造函数并将其添加到硬编码的registry中。有关构造函数args的更多信息，请参见可选Args。

```
type PluginFactory = func(runtime.Unknown, FrameworkHandle) (Plugin, error)

type Registry map[string]PluginFactory

func NewRegistry() Registry {
   return Registry{
      fooplugin.Name: fooplugin.New,
      barplugin.Name: barplugin.New,
      // New plugins are registered here.
   }
}
```

也可以将插件添加到Registry对象，然后将其注入调度程序中。请参阅自定义调度程序插件

# 插件生命周期

## Initialization

插件初始化有两个步骤。首先，注册插件。其次，调度程序使用其配置来确定要实例化的插件。如果插件注册了多个扩展点，则仅实例化一次。 

实例化插件时，将向其传递config args和FrameworkHandle。

## Concurrency

插件编写者应考虑两种并发类型。在评估多个节点时，一个插件可能会被同时调用几次，而一个插件可能会从不同的调度上下文中被并发调用。

注意：在一个调度上下文中，将对每个扩展点进行串行评估。

在调度程序的主线程中，一次仅处理一个调度周期。在下一个调度周期开始之前，直至并包括预留空间的任何扩展点都将完成。
在保留阶段之后，绑定周期将异步执行。这意味着可以从两个不同的调度上下文中同时调用一个插件，前提是至少有一个调用要在保留后到达扩展点。
有状态的插件应谨慎处理这些情况。

最后，根据拒绝Pod的方式，可以从Permit线程或Bind线程调用取消un-reserve插件。

队列排序扩展点是一种特殊情况。它不是调度上下文的一部分，可以为许多吊舱对同时调用。

![scheduling framework extension points](http://img.rocdu.top/20191224/20180409-scheduling-framework-threads.png)

# Plugin Configuration

可以在调度程序配置中启用插件。另外，可以在配置中禁用默认插件(这里好像没实现)。在1.15中，调度框架没有默认插件。

调度程序配置也可以包括插件的配置。这样的配置将在调度程序初始化插件时传递给插件。该配置是任意值。接收插件应解码并处理配置。

插件分为两个部分：

- 每个扩展点已启用插件的列表（及其运行顺序）。如果省略了这些列表之一，则将使用默认列表。
- 每个插件的一组可选的自定义插件参数。省略插件的配置参数等效于使用该插件的默认配置。

插件配置由扩展点组织。每个列表中都必须包含一个注册有多个要点的插件。

```
type KubeSchedulerConfiguration struct {
    // ... other fields
    Plugins      Plugins
    PluginConfig []PluginConfig
}

type Plugins struct {
    QueueSort      []Plugin
    PreFilter      []Plugin
    Filter         []Plugin
    PostFilter     []Plugin
    Score          []Plugin
    Reserve        []Plugin
    Permit         []Plugin
    PreBind        []Plugin
    Bind           []Plugin
    PostBind       []Plugin
    UnReserve      []Plugin
}

type Plugin struct {
    Name   string
    Weight int // Only valid for Score plugins
}

type PluginConfig struct {
    Name string
    Args runtime.Unknown
}
```


下面的示例示出了调度器的配置，它使一些插件在reserve和preBind扩展点和禁用一个插件。它还提供了plugin的配置foo。



```yaml
apiVersion: kubescheduler.config.k8s.io/v1alpha1
kind: KubeSchedulerConfiguration

...

plugins:
  reserve:
    enabled:
    - name: foo
    - name: bar
    disabled:
    - name: baz
  preBind:
    enabled:
    - name: foo
    disabled:
    - name: baz

pluginConfig:
- name: foo
  args: >
    Arbitrary set of args to plugin foo
```

当配置中省略扩展点时，将使用该扩展点的默认插件。当存在扩展名并enabled提供扩展名时enabled，
除默认插件外，还将调用插件。首先调用默认插件，然后以配置中指定的相同顺序调用其他已启用的插件。
如果希望以不同的顺序调用默认插件，则默认插件必须为，disabled且 enabled顺序为所需。

假设有一个插件叫做默认foo的reserve，我们要添加插件bar，我们想要被调用之前foo，我们应该顺序禁用foo 和启用bar和foo。
以下示例显示实现此目的的配置：

```yaml
apiVersion: kubescheduler.config.k8s.io/v1alpha1
kind: KubeSchedulerConfiguration

...

plugins:
  reserve:
    enabled:
    - name: bar
    - name: foo
    disabled:
    - name: foo
```

## 启用/禁用

指定后，将仅启用特定扩展点的插件列表。如果配置中省略了扩展点，则默认插件集将用于该扩展点。

## 更改评估顺序

关联时，插件评估顺序由插件在配置中出现的顺序指定。注册多个扩展点的插件在每个扩展点的顺序可以不同。

## 可选的Args

插件可以从其配置中以任意结构接收参数。因为一个插件可能出现在多个扩展点中，所以配置位于PluginConfig的单独列表中。

配置参数：
```
{
   "name": "ServiceAffinity",
   "args": {
      "LabelName": "app",
      "LabelValue": "mysql"
   }
}
```

解析参数：
```
func NewServiceAffinity(args *runtime.Unknown, h FrameworkHandle) (Plugin, error) {
    if args == nil {
        return nil, errors.Errorf("cannot find service affinity plugin config")
    }
    if args.ContentType != "application/json" {
        return nil, errors.Errorf("cannot parse content type: %v", args.ContentType)
    }
    var config struct {
        LabelName, LabelValue string
    }
    if err := json.Unmarshal(args.Raw, &config); err != nil {
        return nil, errors.Wrap(err, "could not parse args")
    }
    //...
}
```