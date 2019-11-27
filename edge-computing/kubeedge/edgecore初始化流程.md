# edcore功能

由官方文档我们知道，kubeedge核心为cloudcore和edgecore，edgecore主要分为以下几个组件

- Edged：在边缘管理容器化的应用程序。
- EdgeHub：Edge上的通信接口模块。
- EventBus：使用MQTT处理内部边缘通信。
- DeviceTwin：它是用于处理设备元数据的设备的软件镜像。
- MetaManager：它管理边缘节点上的元数据。


# 启动代码
```
Run: func(cmd *cobra.Command, args []string) {
    verflag.PrintAndExitIfRequested()
    flag.PrintFlags(cmd.Flags())

    // To help debugging, immediately log version
    klog.Infof("Version: %+v", version.Get())

    registerModules()
    // start all modules
    core.Run()
},
```

# 配置初始化

在初始化时，会初始化配置，配置使用了华为CSE github.com/go-chassis/go-archaius的微服务框架
go-archaius依次从配置中心，命令行，环境变量，文件，外部配置源读取配置，

在beehive中的InitializeConfig中以下代码为本地文件配置的代码实现，在不进行配置configpath的情况下遍历当前的conf目录读取yml或yaml结尾的文件：

```
confLocation := getConfigDirectory() + "/conf"
_, err = os.Stat(confLocation)
if !os.IsExist(err) {
	os.Mkdir(confLocation, os.ModePerm)
}
err = filepath.Walk(confLocation, func(location string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() {
		return nil
	}
	ext := strings.ToLower(path.Ext(location))
	if ext == ".yml" || ext == ".yaml" {
		archaius.AddFile(location)
	}
	return nil
})
```


# registerModules

edgecore通过registerModules函数注册对应的模块

```
func registerModules() {
	devicetwin.Register()
	edged.Register()
	edgehub.Register()
	eventbus.Register()
	edgemesh.Register()
	metamanager.Register()
	servicebus.Register()
	test.Register()
	dbm.InitDBManager()
}
```

在模块注册时会判断在配置中是否启用了该模块，从而选择性加载

```
func Register(m Module) {
	if isModuleEnabled(m.Name()) {
		modules[m.Name()] = m
		klog.Infof("Module %v registered", m.Name())
	} else {
		disabledModules[m.Name()] = m
		klog.Warningf("Module %v is not register, please check modules.yaml", m.Name())
	}
}
```

另外registerModules中的initdbmanager比较简单，根据orm初始化数据库，不再赘述

# StartModules


初始化一个上下文，当前只有一种类型，也就是go channel,可能以后会有更多的上下文实现

初始化全局上下文context，context包含两个对象moduleContext，messageContext，

moduleContext用于模块管理

```
type ModuleContext interface {
	AddModule(module string)
	AddModuleGroup(module, group string)
	Cleanup(module string)
}
```


messageContext 用于消息同步

```
type MessageContext interface {
	// async mode
	Send(module string, message model.Message)
	Receive(module string) (model.Message, error)
	// sync mode
	SendSync(module string, message model.Message, timeout time.Duration) (model.Message, error)
	SendResp(message model.Message)
	// group broadcast
	SendToGroup(moduleType string, message model.Message)
	SendToGroupSync(moduleType string, message model.Message, timeout time.Duration) error
}
```

> 这部分消息同步代码即为kubeedge中的消息同步框架[beehive](https://github.com/kubeedge/beehive)的实现

遍历要加载的models，通过AddModule将模块添加到模块上下文的channel上下文里面，
AddModuleGroup 根据组模块添加channel到typeChannels

有三个channel  channels,typeChannels,anonChannels：

```
type ChannelContext struct {
	channels,     map[string]chan model.Message
	chsLock      sync.RWMutex
	typeChannels map[string]map[string]chan model.Message
	typeChsLock  sync.RWMutex
	anonChannels map[string]chan model.Message
	anonChsLock  sync.RWMutex
}
```

分别有以下group

```
const (
	BusGroup = "bus"   -- servicebus,eventbus
	HubGroup = "hub"  -- edgehub
	TwinGroup = "twin" -- devicetwin
	MetaGroup = "meta"   -- testmanager
	EdgedGroup = "edged"   -- edged 
	UserGroup = "user"  
	MeshGroup = "mesh"  -- edgemesh
)
```

然后每个模块通过根据Context启动自身

```
go module.Start(coreContext)
```

# 优雅停止

捕捉信号量，然后对模块进行清理
```
func GracefulShutdown() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	select {
	case s := <-c:
		klog.Infof("Get os signal %v", s.String())
		//Cleanup each modules
		modules := GetModules()
		for name, module := range modules {
			klog.Infof("Cleanup module %v", name)
			module.Cleanup()
		}
	}
}
```

扫描关注我:

![微信](http://q08i5y6c2.bkt.clouddn.com/qrcode_for_gh_7457c3b1bfab_258.jpg)