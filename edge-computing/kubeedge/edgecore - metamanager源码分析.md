# metamanager

# Start

定期发送MetaSync操作消息，以同步在边缘节点上运行的Pod的状态。同步间隔可在conf/edge.yaml中配置

```
func (m *metaManager) Start() {
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	InitMetaManagerConfig()

	go func() {
		period := getSyncInterval()
		timer := time.NewTimer(period)
		for {
			select {
			case <-ctx.Done():
				klog.Warning("MetaManager stop")
				return
			case <-timer.C:
				timer.Reset(period)
				msg := model.NewMessage("").BuildRouter(MetaManagerModuleName, GroupResource, model.ResourceTypePodStatus, OperationMetaSync)
				beehiveContext.Send(MetaManagerModuleName, *msg)
			}
		}
	}()

	m.runMetaManager(ctx)
}
```

# runMetaManager

runMetaManager从beehive中读取metamanager的消息，交给process处理

```
func (m *metaManager) runMetaManager(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				klog.Warning("MetaManager mainloop stop")
				return
			default:

			}
			if msg, err := beehiveContext.Receive(m.Name()); err == nil {
				klog.Infof("get a message %+v", msg)
				m.process(msg)
			} else {
				klog.Errorf("get a message %+v: %v", msg, err)
			}
		}
	}()
}
```


# process

判断动作，然后执行相对应的操作
```
func (m *metaManager) process(message model.Message) {
	operation := message.GetOperation()
	switch operation {
	case model.InsertOperation:
		m.processInsert(message)
	case model.UpdateOperation:
		m.processUpdate(message)
	case model.DeleteOperation:
		m.processDelete(message)
	case model.QueryOperation:
		m.processQuery(message)
	case model.ResponseOperation:
		m.processResponse(message)
	case messagepkg.OperationNodeConnection:
		m.processNodeConnection(message)
	case OperationMetaSync:
		m.processSync(message)
	case OperationFunctionAction:
		m.processFunctionAction(message)
	case OperationFunctionActionResult:
		m.processFunctionActionResult(message)
	case constants.CSIOperationTypeCreateVolume,
		constants.CSIOperationTypeDeleteVolume,
		constants.CSIOperationTypeControllerPublishVolume,
		constants.CSIOperationTypeControllerUnpublishVolume:
		m.processVolume(message)
	}
}
```


# insert

- insert获取了消息的内容
- 使用parseResource，根据resource返回资源类型，资源ID，

```
// Resource format: <namespace>/<restype>[/resid]
// return <reskey, restype, resid>
func parseResource(resource string) (string, string, string) {
	tokens := strings.Split(resource, constants.ResourceSep)
	resType := ""
	resID := ""
	switch len(tokens) {
	case 2:
		resType = tokens[len(tokens)-1]
	case 3:
		resType = tokens[len(tokens)-2]
		resID = tokens[len(tokens)-1]
	default:
	}
	return resource, resType, resID
}
```
- 将数据存入数据库，
- 判断资源类型是service和endpoint的消息发送给edgemesh，其他的发送给Edged
- 最后将所有的insert消息发送给edgehub



# update

- 当资源类型为 servicelist  endpointslist podlist时需要更新数据库，发送消息给edgemesh,发送消息到edgehub

- 当资源类型为其他类型时首先判断资源是否变化（实际上只判断了pod的状态），代码如下

```
func resourceUnchanged(resType string, resKey string, content []byte) bool {
	if resType == model.ResourceTypePodStatus {
		dbRecord, err := dao.QueryMeta("key", resKey)
		if err == nil && len(*dbRecord) > 0 && string(content) == (*dbRecord)[0] {
			return true
		}
	}

	return false
}
```

- 更新数据库
- 当来源是edged时消息发送给cloud和edged
- 当消息来源是edgecontroller时，判断资源类型是否为service或endpoint类型。如果是发送消息给edgemesh,不是则发送消息给edged
- 当来源是funcmgr时，发送消息给edgefunction
- 当来源是edgefunction时消息发送给edgehub

# delete

- 删除数据库内容
- 转发消息给edged
- 返回响应到edgehub

# query

- 判断资源类型是否依赖远端查询，切远端为已连接状态且连接到云端，
  - 当资源在本地查询失败，或者资源类型为node或类型为volume attachment
    - 则向云端发送查询请求，将返回信息存储，并根据类型（service,endpoints）发送消息给edgemesh还是edged
  - 其他资源直接由metamanager返回，并通知edged同步状态
- 当资源不需要远端查询，则在本地查询进行返回

# response

- 当来源是云端，判断类型是否为service和endpoint，则发送消息给edgemesh,否则发给edged
- 不是云端则发给云端

# node/connection

- 设置云端连接状态

# meta-internal-sync

用于同步pod状态

- 当pod在db中无记录则跳过
- 当有pod状态记录无pod记录时，删除pod状态记录
- 将所有的pod状态记录发送到云端


# action

本地保存后将消息发送给，edgefunction

# action_result

本地保存函数执行结果，返回给云端

# createvolume、deletevolume、controllerpublishvolume、controllerunpublishvolume

发消息给edged,返回结果传给云端