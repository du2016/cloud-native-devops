# edgehub 源码分析

edgehub是Edge上的通信接口模块，用于云边消息同步

# 结构定义及初始化

edgehub的结构定义

```
type EdgeHub struct {
	context       *beehiveContext.Context
	chClient      clients.Adapter
	config        *config.ControllerConfig
	reconnectChan chan struct{}
	cancel        context.CancelFunc
	syncKeeper    map[string]chan model.Message
	keeperLock    sync.RWMutex
}
```


在注册edgehub模块的时间 对edgehub进行了初始化

```
func Register() {
	core.Register(&EdgeHub{
		config:        &config.GetConfig().CtrConfig,
		reconnectChan: make(chan struct{}),
		syncKeeper:    make(map[string]chan model.Message),
	})
}
```



在模块启动时先将拿到beehiveContext，然后获取EdgehubConfig

在启动时间加载ControllerConfig 根据使用的ControllerConfig中的protocol加载对应的websocket config或者quicconfig

```
func (eh *EdgeHub) initial() (err error) {
	config.GetConfig().WSConfig.URL, err = bhconfig.CONFIG.GetValue("edgehub.websocket.url").ToString()
	if err != nil {
		klog.Warningf("failed to get cloud hub url, error:%+v", err)
		return err
	}

	cloudHubClient, err := clients.GetClient(eh.config.Protocol, config.GetConfig())
	if err != nil {
		return err
	}

	eh.chClient = cloudHubClient

	return nil
}
```

然后初始化EdgeHub.Controller.chClient 配置，然后初始化对应连接用于和cloudcore通信

# 给其他组件同步连接成功状态

然后通过pubConnectInfo给其他的组（resource,twin,func,user）发消息，告诉他们 云端连接成功了

```
func (eh *EdgeHub) pubConnectInfo(isConnected bool) {
	// var info model.Message
	content := connect.CloudConnected
	if !isConnected {
		content = connect.CloudDisconnected
	}

	for _, group := range groupMap {
		message := model.NewMessage("").BuildRouter(message.SourceNodeConnection, group,
			message.ResourceTypeNodeConnection, message.OperationNodeConnection).FillBody(content)
		eh.context.SendToGroup(group, *message)
	}
}
```

message结构定义如下

type Message struct {
	Header  MessageHeader `json:"header"`
	Router  MessageRoute  `json:"route,omitempty"`
	Content interface{}   `json:"content"`
}

header包含了

- 消息的ID
- 消息的父ID
- 时间戳
- 是否被同步


router 定义了以下对象

- 来源
- 广播到哪个组
- 动作
- 操作的资源


## 发消息

我们可以看到model.NewMessage("").BuildRouter接收四个参数，分别为：

- 来源
- 发给哪个组
- 资源类型
- 动作

这里的NewMessage parentID参数为空 证明这是消息的发起者

```
func NewMessage(parentID string) *Message {
	msg := &Message{}
	msg.Header.ID = uuid.NewV4().String()
	msg.Header.ParentID = parentID
	msg.Header.Timestamp = time.Now().UnixNano() / 1e6
	return msg
}
```


# 接下来启动了三个协程

- routeToEdge
- routeToCloud
- keepalive

## routeToEdge

routeToEdge接收信息 然后发送信息到对应的group,
判断group是否存在，判断是否是已有同步响应，如果没有发送给对应组。这里就使用到了beehive的messageContext.
接下来根据parentid将此条消息发送到syncKeeper channel里

```
func (ehc *Controller) routeToEdge() {
	for {
		message, err := ehc.chClient.Receive()
		if err != nil {
			klog.Errorf("websocket read error: %v", err)
			ehc.stopChan <- struct{}{}
			return
		}

		klog.Infof("received msg from cloud-hub:%+v", message)
		err = ehc.dispatch(message)
		if err != nil {
			klog.Errorf("failed to dispatch message, discard: %v", err)
		}
	}
}
```


## routeToCloud

将在channel收到的消息发送到云端，同时将消息保存在syncKeeper，这里创建了一个定时器，过期的话会自动删除

```
func (ehc *Controller) sendToCloud(message model.Message) error {
	ehc.keeperLock.Lock()
	err := ehc.chClient.Send(message)
	ehc.keeperLock.Unlock()
	if err != nil {
		klog.Errorf("failed to send message: %v", err)
		return fmt.Errorf("failed to send message, error: %v", err)
	}

	syncKeep := func(message model.Message) {
		tempChannel := ehc.addKeepChannel(message.GetID())
		sendTimer := time.NewTimer(ehc.config.HeartbeatPeriod)
		select {
		case response := <-tempChannel:
			sendTimer.Stop()
			ehc.context.SendResp(response)
			ehc.deleteKeepChannel(response.GetParentID())
		case <-sendTimer.C:
			klog.Warningf("timeout to receive response for message: %+v", message)
			ehc.deleteKeepChannel(message.GetID())
		}
	}

	if message.IsSync() {
		go syncKeep(message)
	}

	return nil
}
```


## keepalive

根据心跳时间向云端发送心跳

```
func (ehc *Controller) keepalive() {
	for {
		msg := model.NewMessage("").
			BuildRouter(ModuleNameEdgeHub, "resource", "node", "keepalive").
			FillBody("ping")

		// post message to cloud hub
		err := ehc.sendToCloud(*msg)
		if err != nil {
			klog.Errorf("websocket write error: %v", err)
			ehc.stopChan <- struct{}{}
			return
		}

		time.Sleep(ehc.config.HeartbeatPeriod)
	}
}
```


扫描关注我:

![微信](http://q08i5y6c2.bkt.clouddn.com/qrcode_for_gh_7457c3b1bfab_258.jpg)