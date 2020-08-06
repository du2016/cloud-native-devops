
# authenticate

开启XDS_AUTH的情况下，将添加下面的证书链

authenticate.ClientCertAuthenticator{}

通过证书链对证书进行验证

```
func (cca *ClientCertAuthenticator) Authenticate(ctx context.Context) (*Caller, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok || peer.AuthInfo == nil {
		return nil, fmt.Errorf("no client certificate is presented")
	}

	if authType := peer.AuthInfo.AuthType(); authType != "tls" {
		return nil, fmt.Errorf("unsupported auth type: %q", authType)
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	chains := tlsInfo.State.VerifiedChains
	if len(chains) == 0 || len(chains[0]) == 0 {
		return nil, fmt.Errorf("no verified chain is found")
	}

	ids, err := util.ExtractIDs(chains[0][0].Extensions)
	if err != nil {
		return nil, err
	}

	return &Caller{
		AuthSource: AuthSourceClientCertificate,
		Identities: ids,
	}, nil
}
```


s.globalPushContext().InitContext 对pushcontext再次初始化


创建一个channel，用于处理请求


# processRequest

通过url判断请求类型

```
	switch discReq.TypeUrl {
	case v2.ClusterType, v3.ClusterType:
		if err := s.handleTypeURL(discReq.TypeUrl, &con.node.RequestedTypes.CDS); err != nil {
			return err
		}
		if err := s.handleCds(con, discReq); err != nil {
			return err
		}
	case v2.ListenerType, v3.ListenerType:
		if err := s.handleTypeURL(discReq.TypeUrl, &con.node.RequestedTypes.LDS); err != nil {
			return err
		}
		if err := s.handleLds(con, discReq); err != nil {
			return err
		}
	case v2.RouteType, v3.RouteType:
		if err := s.handleTypeURL(discReq.TypeUrl, &con.node.RequestedTypes.RDS); err != nil {
			return err
		}
		if err := s.handleRds(con, discReq); err != nil {
			return err
		}
	case v2.EndpointType, v3.EndpointType:
		if err := s.handleTypeURL(discReq.TypeUrl, &con.node.RequestedTypes.EDS); err != nil {
			return err
		}
		if err := s.handleEds(con, discReq); err != nil {
			return err
		}
```

处理CDS请求

```
func (s *DiscoveryServer) handleCds(con *Connection, discReq *discovery.DiscoveryRequest) error {
	if con.Watching(v3.ClusterShortType) {
		if !s.shouldRespond(con, cdsReject, discReq) {
			return nil
		}
	}
	adsLog.Infof("ADS:CDS: REQ %v version:%s", con.ConID, discReq.VersionInfo)
	err := s.pushCds(con, s.globalPushContext(), versionInfo())
	if err != nil {
		return err
	}
	return nil
}
```



# 生成clusters


BuildClusters



configgen.buildOutboundClusters
 
 
configgen.buildInboundClusters



cb.push.Services 获取所有serviceentry

cb.buildDefaultCluster 构建集群

buildDefaultCluster 构建集群

BuildStatPrefix 构建clustername

setUpstreamProtocol 设置上游集群的协议，当开启sniffing时，可以使用use_downstream_protocol

applyDestinationRule 根据现有cluster结合DestinationRule 设置各个版本


加载authn authz插件p.OnOutboundCluster(inputParams, defaultCluster)

添加blackhole和passthroughcluster

envoyfilter.ApplyClusterPatches 应用 envoyfilter


cdsDiscoveryResponse 响应请求


# 响应请求

pushCds



# pushConnection


initRegistryEventHandlers


AppendServiceHandler


AppendServiceHandler



func (c *Controller) Run(stop <-chan struct{}) {
	if c.networksWatcher != nil {
		c.networksWatcher.AddNetworksHandler(c.reloadNetworkLookup)
		c.reloadNetworkLookup()
	}

	go func() {
		cache.WaitForCacheSync(stop, c.HasSynced)
		c.queue.Run(stop)
	}()

	// To avoid endpoints without labels or ports, wait for sync.
	cache.WaitForCacheSync(stop, c.nodeInformer.HasSynced,
		c.pods.informer.HasSynced,
		c.serviceInformer.HasSynced)

	<-stop
	log.Infof("Controller terminated")
}


queue run


AdsPushAll

startPush 

func (s *DiscoveryServer) Push(req *model.PushRequest)