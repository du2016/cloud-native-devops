# EventBus
 
EventBus 是一个MQTT客户端

# 初始化

在初始化eventbus时获取mqtt模式 external/internal

#  启动

根据配置初始化Mqttclient,创建Internal Mqtt client或者external Mqtt client,设置qs,retain策略和队列的大小



# external mqtt broker


## InitSubClient

设置连接参数启动连接

```
func (mq *Client) InitSubClient() {
	timeStr := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	right := len(timeStr)
	if right > 10 {
		right = 10
	}
	subID := fmt.Sprintf("hub-client-sub-%s", timeStr[0:right])
	subOpts := util.HubClientInit(mq.MQTTUrl, subID, "", "")
	subOpts.OnConnect = onSubConnect
	subOpts.AutoReconnect = false
	subOpts.OnConnectionLost = onSubConnectionLost
	mq.SubCli = MQTT.NewClient(subOpts)
	util.LoopConnect(subID, mq.SubCli)
	klog.Info("finish hub-client sub")
}
```

## 以下两个函数定义了当失联和连接时的处理逻辑
```
func onSubConnectionLost(client MQTT.Client, err error) {
	klog.Errorf("onSubConnectionLost with error: %v", err)
	go MQTTHub.InitSubClient()
}

func onSubConnect(client MQTT.Client) {
	for _, t := range SubTopics {
		token := client.Subscribe(t, 1, OnSubMessageReceived)
		if rs, err := util.CheckClientToken(token); !rs {
			klog.Errorf("edge-hub-cli subscribe topic: %s, %v", t, err)
			return
		}
		klog.Infof("edge-hub-cli subscribe topic to %s", t)
	}
}
```
token用于确定连接状态
可以看到 它订阅了以下topic

```
	SubTopics = []string{
		"$hw/events/upload/#",
		"$hw/events/device/+/state/update",
		"$hw/events/device/+/twin/+",
		"$hw/events/node/+/membership/get",
		"SYS/dis/upload_records",
	}
```

当在这些topic中获得消息时，通过mqtt的Subscribe方法回调OnSubMessageReceived

## OnSubMessageReceived
```
func OnSubMessageReceived(client MQTT.Client, message MQTT.Message) {
	klog.Infof("OnSubMessageReceived receive msg from topic: %s", message.Topic())
	// for "$hw/events/device/+/twin/+", "$hw/events/node/+/membership/get", send to twin
	// for other, send to hub
	// for "SYS/dis/upload_records", no need to base64 topic
	var target string
	resource := base64.URLEncoding.EncodeToString([]byte(message.Topic()))
	if strings.HasPrefix(message.Topic(), "$hw/events/device") || strings.HasPrefix(message.Topic(), "$hw/events/node") {
		target = modules.TwinGroup
	} else {
		target = modules.HubGroup
		if message.Topic() == "SYS/dis/upload_records" {
			resource = "SYS/dis/upload_records"
		}
	}
	// routing key will be $hw.<project_id>.events.user.bus.response.cluster.<cluster_id>.node.<node_id>.<base64_topic>
	msg := model.NewMessage("").BuildRouter(modules.BusGroup, "user",
		resource, "response").FillBody(string(message.Payload()))
	klog.Info(fmt.Sprintf("received msg from mqttserver, deliver to %s with resource %s", target, resource))
	ModuleContext.SendToGroup(target, *msg)
}
```

该函数判断topic，"$hw/events/device"和"$hw/events/node"开头发送给twingroup也就是devicetwin，其他信息发送给edgehub
然后通过SendToGroup发送到devicetwin

## InitPubClient

```
func (mq *Client) InitPubClient() {
	timeStr := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	right := len(timeStr)
	if right > 10 {
		right = 10
	}
	pubID := fmt.Sprintf("hub-client-pub-%s", timeStr[0:right])
	pubOpts := util.HubClientInit(mq.MQTTUrl, pubID, "", "")
	pubOpts.OnConnectionLost = onPubConnectionLost
	pubOpts.AutoReconnect = false
	mq.PubCli = MQTT.NewClient(pubOpts)
	util.LoopConnect(pubID, mq.PubCli)
	klog.Info("finish hub-client pub")
}
```
InitPubClient只是创建了一个MQTTclient,然后每五秒钟连接一次mqtt server，当失败是通过，重新初始化


## Internal mqtt broker

启动一个内置的qttserver

```
mqttServer = mqttBus.NewMqttServer(sessionQueueSize.(int), internalMqttURL.(string), retain.(bool), qos.(int))
mqttServer.InitInternalTopics()
err := mqttServer.Run()
```

# pubCloudMsgToEdge


在启动/连接完MQTTserver后，调用了pubCloudMsgToEdge方法

pubCloudMsgToEdge执行以下操作

- 从beehive获取消息
- 获取消息的动作和资源
- 当动作为 subscribe 时从MQTT订阅消息
- 当动作为 message 时，将消息的message发送给MQTT broker，消息类型是一个map,
- 当动作为 publish 时，将消息的message发送给MQTT broker, 消息为一个字符串，topic和resource一致
- 当动作为 get_result时，resource必须为auth_info，
然后发送消息到`fmt.Sprintf("$hw/events/node/%s/authInfo/get/result", mqttBus.NodeID)`topic

```
func (eb *eventbus) pubCloudMsgToEdge(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			klog.Warning("EventBus PubCloudMsg To Edge stop")
			return
		default:
		}
		accessInfo, err := eb.context.Receive(eb.Name())
		if err != nil {
			klog.Errorf("Fail to get a message from channel: %v", err)
			continue
		}
		operation := accessInfo.GetOperation()
		resource := accessInfo.GetResource()
		switch operation {
		case "subscribe":
			eb.subscribe(resource)
			klog.Infof("Edge-hub-cli subscribe topic to %s", resource)
		case "message":
			body, ok := accessInfo.GetContent().(map[string]interface{})
			if !ok {
				klog.Errorf("Message is not map type")
				return
			}
			message := body["message"].(map[string]interface{})
			topic := message["topic"].(string)
			payload, _ := json.Marshal(&message)
			eb.publish(topic, payload)
		case "publish":
			topic := resource
			var ok bool
			// cloud and edge will send different type of content, need to check
			payload, ok := accessInfo.GetContent().([]byte)
			if !ok {
				content := accessInfo.GetContent().(string)
				payload = []byte(content)
			}
			eb.publish(topic, payload)
		case "get_result":
			if resource != "auth_info" {
				klog.Info("Skip none auth_info get_result message")
				return
			}
			topic := fmt.Sprintf("$hw/events/node/%s/authInfo/get/result", mqttBus.NodeID)
			payload, _ := json.Marshal(accessInfo.GetContent())
			eb.publish(topic, payload)
		default:
			klog.Warningf("Action not found")
		}
	}
}
```

扫描关注我:

![微信](http://q08i5y6c2.bkt.clouddn.com/qrcode_for_gh_7457c3b1bfab_258.jpg)