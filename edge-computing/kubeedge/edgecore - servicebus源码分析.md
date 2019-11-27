# service bus 

ServiceBus是一个运行在边缘的HTTP客户端，接受来自云上服务的请求，
与运行在边缘端的HTTP服务器交互，提供了云上服务通过HTTP协议访问边缘端HTTP服务器的能力。


# 代码逻辑

servicebus的功能比较简单

```
func (sb *servicebus) Start(c *beehiveContext.Context) {
	// no need to call TopicInit now, we have fixed topic
	var ctx context.Context
	sb.context = c
	ctx, sb.cancel = context.WithCancel(context.Background())
	var htc = new(http.Client)
	htc.Timeout = time.Second * 10

	var uc = new(util.URLClient)
	uc.Client = htc

	//Get message from channel
	for {
		select {
		case <-ctx.Done():
			klog.Warning("ServiceBus stop")
			return
		default:

		}
		msg, err := sb.context.Receive("servicebus")
		if err != nil {
			klog.Warningf("servicebus receive msg error %v", err)
			continue
		}
		go func() {
			klog.Infof("ServiceBus receive msg")
			source := msg.GetSource()
			if source != sourceType {
				return
			}
			resource := msg.GetResource()
			r := strings.Split(resource, ":")
			if len(r) != 2 {
				m := "the format of resource " + resource + " is incorrect"
				klog.Warningf(m)
				code := http.StatusBadRequest
				if response, err := buildErrorResponse(msg.GetID(), m, code); err == nil {
					sb.context.SendToGroup(modules.HubGroup, response)
				}
				return
			}
			content, err := json.Marshal(msg.GetContent())
			if err != nil {
				klog.Errorf("marshall message content failed %v", err)
				m := "error to marshal request msg content"
				code := http.StatusBadRequest
				if response, err := buildErrorResponse(msg.GetID(), m, code); err == nil {
					sb.context.SendToGroup(modules.HubGroup, response)
				}
				return
			}
			var httpRequest util.HTTPRequest
			if err := json.Unmarshal(content, &httpRequest); err != nil {
				m := "error to parse http request"
				code := http.StatusBadRequest
				klog.Errorf(m, err)
				if response, err := buildErrorResponse(msg.GetID(), m, code); err == nil {
					sb.context.SendToGroup(modules.HubGroup, response)
				}
				return
			}
			operation := msg.GetOperation()
			targetURL := "http://127.0.0.1:" + r[0] + "/" + r[1]
			resp, err := uc.HTTPDo(operation, targetURL, httpRequest.Header, httpRequest.Body)
			if err != nil {
				m := "error to call service"
				code := http.StatusNotFound
				klog.Errorf(m, err)
				if response, err := buildErrorResponse(msg.GetID(), m, code); err == nil {
					sb.context.SendToGroup(modules.HubGroup, response)
				}
				return
			}
			resp.Body = http.MaxBytesReader(nil, resp.Body, maxBodySize)
			resBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				if err.Error() == "http: request body too large" {
					err = fmt.Errorf("response body too large")
				}
				m := "error to receive response, err: " + err.Error()
				code := http.StatusInternalServerError
				klog.Errorf(m, err)
				if response, err := buildErrorResponse(msg.GetID(), m, code); err == nil {
					sb.context.SendToGroup(modules.HubGroup, response)
				}
				return
			}

			response := util.HTTPResponse{Header: resp.Header, StatusCode: resp.StatusCode, Body: resBody}
			responseMsg := model.NewMessage(msg.GetID())
			responseMsg.Content = response
			responseMsg.SetRoute("servicebus", modules.UserGroup)
			sb.context.SendToGroup(modules.HubGroup, *responseMsg)
		}()
	}
}
```

根据代码分为以下步骤：

- 拿到回传的beehiveContext，初始化http连接，设置超时十秒
- 初始化URLClient
- 从beehiveContext 里面接收接收servicebus的消息，获取消息来源，来源必须是router_rest
- 然后获取消息包含的resource，根据代码可以看出resource是一个以冒号分割的字符串，根据后续代码可以知道，实际上就是 port:url
- 获取对应的消息内容，该内容是一个json，最终被反序列化为

```
type HTTPRequest struct {
    Header http.Header `json:"header"`
    Body   []byte      `json:"body"`
}
```
  
- 获取动作 - 即http的 get/post/put...
- 发出请求接收返回，判断返回数据量大小，最大为 5*1e6个字节，超过就报错
- 以接收到的消息ID作为父ID通过返回数据给edgehub

扫描关注我:

![微信](http://q08i5y6c2.bkt.clouddn.com/qrcode_for_gh_7457c3b1bfab_258.jpg)
