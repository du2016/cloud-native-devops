# 介绍

MCP是基于订阅的配置分发API。配置使用者（即sink）从配置生产者（即source）请求更新资源集合.添加,更新或删除资源时，source会将资源更新推送到sink.sink积极确认资源更新，如果sink接受，则返回ACK，如果被拒绝则返回NACK，例如: 因为资源无效。一旦对先前的更新进行了ACK/NACK，则源可以推送其他更新.该源一次只能运行一个未完成的更新(每个集合).

MCP是一对双向流gRPC API服务（ResourceSource和ResourceSink）。

- 当source是服务器而sink是客户端时，将使用ResourceSource服务.默认情况下，Galley实现ResourceSource服务，并且Pilot/Mixer连接作为客户端。

- 当source是客户端，而接收器是服务器时，将使用ResourceSink服务.可以将Galley配置为可选地`dial-out`到远程配置sink，例如 Pilot位于另一个群集中，在该群集中，它不能作为客户端启动与Galley的连接.在这种情况下，Pilot将实现ResourceSink服务，而Galley将作为客户端进行连接。

就消息交换而言，ResourceSource和ResourceSink在语义上是等效的.唯一有意义的区别是谁启动连接并打开grpc流。

# 数据模型

MCP是一种传输机制，可以通过管理器组件配置先导和混合器.MCP定义了每种资源的通用元数据格式，而资源特定的内容则在其他位置定义（例如https://github.com/istio/api/tree/master/networking/v1alpha3).

## Collections

相同类型的资源被组织到命名集合中.Istio API集合名称的格式为istio/<area>/<version>/<api>，其中<area>，<version>和<api由API样式准则定义.例如，VirtualService的collection名称为istio/networking/v1alpha3/virtualservices。

## 元数据


# 建立连接

ResourceSource服务-客户端是reource sink.客户端dail服务器并建立新的gRPC流.客户端发送RequestResources并接收Resources消息。

![](http://img.rocdu.top/20200818/1.png)

ResourceSink服务-客户端是资源源.客户端拨打服务器并建立新的gRPC流.服务器发送RequestResources并接收Resources消息。

![](http://img.rocdu.top/20200818/2.png)


# 配置升级

以下概述适用于ResourceSink和ResourceSource服务，无论客户端/服务器角色如何。

资源更新协议由增量xDS协议派生。除了资源提示已被删除之外，协议交换几乎是相同的。下面的大多数文本和图表都是从增量xDS文档中复制并进行相应调整的。

在MCP中，资源首先按collection进行组织。在每个collection中，资源可以通过元数据名称唯一地标识。对各个资源进行版本控制，以区分同一命名资源的较新版本。

可以在两种情况下发送RequestResource消息：

- MCP双向更改流中的初始消息

- 作为对先前资源消息的ACK或NACK响应。在这种情况下，response_nonce设置为资源消息中的现时值.ACK/NACK由后续请求中是否存在error_detail决定。

初始的RequestResources消息包括与所订阅的资源集（例如VirtualService）相对应的集合，节点接收器标识符和nonce字段以及initial_resource_version（稍后会详细介绍）。当请求的资源可用时，source将发送资源消息。处理资源消息后，sink在流上发送新的RequestResources消息，指定成功应用的最后一个版本以及源提供的随机数。

随机数字段用于将每个集合的RequestResources和Resources消息配对。源一次只能发送一个未完成的资源消息（每个collection），并等待接收器进行ACK/NACK。接收到更新后，接收器将在解码，验证更新并将更新持久保存到其内部配置存储后，期望相对较快地发送ACK/NACK。

source应忽略具有过期和未知随机数的请求，这些请求与最近发送的Resource消息中的随机数不匹配。

## 成功示例

以下示例显示接收器接收到已成功ACK的一系列更改。


![](http://img.rocdu.top/20200818/3.png)

以下示例显示了与增量更新一起交付的所需资源.此示例假定source支持增量.当source不支持增量更新时，考虑到接收器是否请求增量更新，推送的资源将始终将增量设置为false.在任何时候，源都可以决定推送完整状态更新，而不必考虑接收器的请求.双方必须协商（即同意）在每个请求/响应的基础上使用增量，以增量发送更新。

![](http://img.rocdu.top/20200818/4.png)

## 错误示例

以下示例显示了无法应用更改时发生的情况

![](http://img.rocdu.top/20200818/5.png)

接收器仅在特殊情况下应为NACK。例如，如果一组资源无效，格式错误或无法解码。NACK的更新应发出警报，以供人类随后进行调查.源不应该重新发送先前NACK相同的资源集.在将金丝雀推送到更大数量的资源接收器之前，也可以将金丝雀推送到专用接收器，以验证正确性（非NACK）。

MCP中的随机数用于匹配RequestResources和Resource。在重新连接时，接收器可以通过为每个集合指定带有initial_resource_version的已知资源版本来尝试恢复与同一源的会话。

# mcp实现探

接下来以官方的测试用例分析官方的mcp实现，代码地址：https://github.com/istio/istio/blob/master/pkg/mcp/testing/server.go
对于mcpsource通过

```
mcp.RegisterResourceSourceServer(gs, srcServer) 
```
进行注册

# Server.EstablishResourceStream

获取客户端信息，并进行鉴权，交给ProcessStream进行处理

```
func (s *Server) EstablishResourceStream(stream mcp.ResourceSource_EstablishResourceStreamServer) error {
	if s.rateLimiter != nil {
		if err := s.rateLimiter.Wait(stream.Context()); err != nil {
			return err
		}

	}
	var authInfo credentials.AuthInfo
	if peerInfo, ok := peer.FromContext(stream.Context()); ok { //获取客户端信息
		authInfo = peerInfo.AuthInfo
	} else {
		scope.Warnf("No peer info found on the incoming stream.")
	}

	if err := s.authCheck.Check(authInfo); err != nil { // 认证
		return status.Errorf(codes.Unauthenticated, "Authentication failure: %v", err)
	}

	if err := stream.SendHeader(s.metadata); err != nil {
		return err
	}
	err := s.src.ProcessStream(stream) 
	code := status.Code(err)
	if code == codes.OK || code == codes.Canceled || err == io.EOF {
		return nil
	}
	return err
}
```

# Source.ProcessStream

- s.newConnection(stream)

初始化connection，

```
con := &connection{
    stream:   stream,
    peerAddr: peerAddr,
    requestC: make(chan *mcp.RequestResources),
    watches:  make(map[string]*watch),
    watcher:  s.watcher,
    id:       atomic.AddInt64(&s.nextStreamID, 1),
    reporter: s.reporter,
    limiter:  s.requestLimiter.Create(),
    queue:    internal.NewUniqueScheduledQueue(len(s.collections)),
}
```

watcher即为Server.src的watcher

```
func NewServer(srcOptions *Options, serverOptions *ServerOptions) *Server {
	s := &Server{
		src:         New(srcOptions),
		authCheck:   serverOptions.AuthChecker,
		rateLimiter: serverOptions.RateLimiter,
		metadata:    serverOptions.Metadata,
	}
	return s
}
```

- con.receive()

通过receive不断将数据写入requestC channel

- 从requestC 读取请求

```
case req, more := <-con.requestC:
    if !more {
        return con.reqError
    }
    if con.limiter != nil {
        if err := con.limiter.Wait(stream.Context()); err != nil {
            return err
        }

    }
    if err := con.processClientRequest(req); err != nil {
        return err
    }
```


- 响应入队

```
func (con *connection) queueResponse(resp *WatchResponse) {
	if resp == nil {
		con.queue.Close()
	} else {
		con.queue.Enqueue(resp.Collection, resp) //响应入队
	}
}
```
			
- 响应出队

从queue中获取需要返回的response，返回给sink
```
collection, item, ok := con.queue.Dequeue() // 从queue读取处理后的返回结果
if !ok {
    break
}

resp := item.(*WatchResponse)

w, ok := con.watches[collection]
if !ok {
    scope.Errorf("unknown collection in dequeued watch response: %v", collection)
    break // bug?
}

// the response may have been cleared before we got to it
if resp != nil {
    if err := con.pushServerResponse(w, resp); err != nil { //通过pushServerResponse 将响应发送给sink
        return err
    }
}
```
# connection.processClientRequest

通过调用Watch方法处理请求

```
sr := &Request{
			SinkNode:    req.SinkNode,
			Collection:  collection,
			VersionInfo: versionInfo,
			incremental: req.Incremental,
		}
w.cancel = con.watcher.Watch(sr, con.queueResponse, con.peerAddr)
```

由此我们可以看出我们实现一个mcp的核心在于实现一个watcher，传递给mcp source server

```
type Watcher interface {
	Watch(*Request, PushResponseFunc, string) CancelWatchFunc
}
```

# 官方mcp watcher 实现

pkg/mcp/testing/server.go


cache := snapshot.New(groups.DefaultIndexFn)


重点关注cache.Watch

```
func (c *Cache) Watch(
	request *source.Request,
	pushResponse source.PushResponseFunc,
	peerAddr string) source.CancelWatchFunc {
	group := c.groupIndex(request.Collection, request.SinkNode) // 获取sink要获取的group

	c.mu.Lock()
	defer c.mu.Unlock()

	info := c.fillStatus(group, request, peerAddr) // 初始化对应group peer的同步状态信息

	collection := request.Collection

	// return an immediate response if a snapshot is available and the
	// requested version doesn't match.
	if snapshot, ok := c.snapshots[group]; ok {  // 获取对应组的snapshot

		version := snapshot.Version(request.Collection) // 计算版本信息
		scope.Debugf("Found snapshot for group: %q for %v @ version: %q",
			group, request.Collection, version)

		if version != request.VersionInfo {        // 如果sink的当前版本和source的版本不一致则推送response
			scope.Debugf("Responding to group %q snapshot:\n%v\n", group, snapshot)
			response := &source.WatchResponse{
				Collection: request.Collection,
				Version:    version,
				Resources:  snapshot.Resources(request.Collection),
				Request:    request,
			}
			pushResponse(response)
			return nil
		}
		info.synced[request.Collection][peerAddr] = true
	}

	// 如果版本一致则.返回一个cancel，同时记录对应的watchs，当SetSnapshot时，也就是有更新时会进行调用
	c.watchCount++
	watchID := c.watchCount

	scope.Infof("Watch(): created watch %d for %s from group %q, version %q",
		watchID, collection, group, request.VersionInfo)

	info.mu.Lock()
	info.watches[watchID] = &responseWatch{request: request, pushResponse: pushResponse}
	info.mu.Unlock()

	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if s, ok := c.status[group]; ok {
			s.mu.Lock()
			delete(s.watches, watchID)
			s.mu.Unlock()
		}
	}
	return cancel
}
```

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
