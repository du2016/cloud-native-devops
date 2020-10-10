# 基本概念

配置：
- mesh 作用于 mesh侧也就是envoy的配置
- pilot 配置 通过参数定义
- 其他 通过env设置

初始化方法:

# bootstrap.NewServer

## 初始化Environment

- aggregate.NewController() 

创建了一个Aggregate controller 用于存储registry,并根据变化更新，主要管理以下registry:

```
- consul
- k8s
- serviceentry
- simple（用于测试）
```

- NewPushContext

创建PushContext用于追踪push状态,PushContext还记录了服务的暴露范围网格的配置等等，以便于在下发配置时进行过滤，在下发完成后会记录下发的状态，每次下发完成都会修改下发状态，PushContext数据可以通过debug接口进行查看


## Server的初始化

- getClusterID(args)

根据参数获取ID，若无且registry包含k8s则设置为kubernetes

- NewDiscoveryServer

根据plugins，Environment创建DiscoveryServer，pugins是网络插件列表，支持authn,authz,mixer

- filewatcher.NewWatcher

初始化一个filewatcher,后续用于监听配置文件变化


- initMeshConfiguration

新建filewatcher,根据`/etc/istio/config/mesh`生成MeshConfiguration，当文件变化时，重新读取文件内容


- initKubeClient

根据meshconfig(ConfigSources)或command的registries参数进行判断判断是否包含k8s，获取resetconfig,


使用resetconfig初始化一系列informer

- restClient
- restconfig
- clientset
- shareinformer
- metadatainformer
- dynamicinformer
- istioclient(clientset操作istio config/rbac/networking/security group下的资源)
- istioInformer
- serviceapisinformer
- ext clientset（用于操作crd）


- initMeshNetworks

根据/etc/istio/config/meshNetworks进行 初始化,提供有关网格内的网络集以及如何路由到每个网络中的端点的信息。

- initMeshHandlers

当mesh或者network配置发生变化时进行全量更新

```
func (s *Server) initMeshHandlers() {
	log.Info("initializing mesh handlers")
	// When the mesh config or networks change, do a full push.
	s.environment.AddMeshHandler(func() {
		// Inform ConfigGenerator about the mesh config change so that it can rebuild any cached config, before triggering full push.
		s.XDSServer.ConfigGenerator.MeshConfigChanged(s.environment.Mesh())
		s.XDSServer.ConfigUpdate(&model.PushRequest{
			Full:   true,
			Reason: []model.TriggerReason{model.GlobalUpdate},
		})
	})
	s.environment.AddNetworksHandler(func() {
		s.XDSServer.ConfigUpdate(&model.PushRequest{
			Full:   true,
			Reason: []model.TriggerReason{model.GlobalUpdate},
		})
	})
}
```
- GetDiscoveryAddress

获取istiod的主机名和端口

- initControllers


通过以下环境变量获取当前命名空间

```
POD_NAMESPACE=istio-system
```


初始化 initCertController 

```
certs(secret),根据env,初始化secret controller，watch secret的变化，当被删除的时间进行重建，更新时对证书进行验证，过期/与CA不一致则进行 refresh 操作
```

初始化 initConfigController 
```
初始化config store
1. MCP
2. configsource
3. k8s config store

```

初始化initServiceControllers

从k8s中获取 ingress、endponit(也可能是consul或其他registry)，

然后通过NewServiceDiscovery进行转换，这里有两个判断,通过环境变量设置

PILOT_ENABLE_SERVICEENTRY_SELECT_PODS     serviceentry能够筛选到pod
PILOT_ENABLE_K8S_SELECT_WORKLOAD_ENTRIES  service selector包含workload entries  

- initJwtPolicy

初始化jwt policy

  - third-party-jwt
        使用istio的token /var/run/secrets/tokens/istio-token 作为jwttoken 对agent 进行验证
  - first-party-jwt
  
    使用k8s的token /var/run/secrets/kubernetes.io/serviceaccount/token 作为jwttoken 对agent 进行验证


- maybeCreateCA

根据参数确定命名空间和TrustDomain(默认为cluster.local)
是否需要创建CA，该ca实际就是名为cacerts的secret中,如果没有将根据上述TrustDomain和k8s svc domain自行创建CA,
CA证书会根据证书类型以及CheckInterval判断是否需要定期轮换


- initIstiodCerts

/var/run/secrets/istio-dns/ 没有则创建用于istio dns的证书，流程和上面CA类似

- setPeerCertVerifier

初始化server.peerCertVerifier，根据domain和ca进行验证


- initSecureDiscoveryService 

初始化SecureDiscoveryService，注册poliot grpc接口，监听15012端口

主要实现AggregatedDiscoveryServiceServer接口


// AggregatedDiscoveryServiceServer is the server API for AggregatedDiscoveryService service.
type AggregatedDiscoveryServiceServer interface {
	StreamAggregatedResources(AggregatedDiscoveryService_StreamAggregatedResourcesServer) error
	DeltaAggregatedResources(AggregatedDiscoveryService_DeltaAggregatedResourcesServer) error
}

```
- StreamAggregatedResources  // https://github.com/istio/istio/blob/master/pilot/pkg/xds/ads.go#L223
- DeltaAggregatedResources   // 未实现
```

  - StreamAggregatedResources 
  认证agent信息，关于认证方式可以查看 [](证书验证原理.md)

主要有两个channel

```
- reqChannel

envoy连接上主动请求,返回ads信息

- pushChannel

配置变化poliot主动推送，全量推送xds
```


- initSecureWebhookServer

初始化webhookserver,"/httpsReady"接口，生成httpsReadyClient httpclient,获取、ready的状态用于健康检查

- initSidecarInjector

创建mutatingwebhook用户实现自动注入，/var/lib/istio/inject/config go tmpl,/var/lib/istio/inject/config用于渲染的值，

inject.NewWebhook(parameters)

监听路径为/inject的请求

- initConfigValidation

创建validatingwebhook 用于校验cr的规范

server.New(params) 监听
/validate
/admitpilot
/admitmixer


- initIstiodAdminServer

监听8080处理/ready

initMonitor

监听15014处理/metrics和"/version"请求


- initRegistryEventHandlers

初始化用于监听config和service变化时的event handler

当service配置变化时全量更新

```
serviceHandler := func(svc *model.Service, _ model.Event) {
    pushReq := &model.PushRequest{
        Full: true,
        ConfigsUpdated: map[model.ConfigKey]struct{}{{
            Kind:      gvk.ServiceEntry,
            Name:      string(svc.Hostname),
            Namespace: svc.Attributes.Namespace,
        }: {}},
        Reason: []model.TriggerReason{model.ServiceUpdate},
    }
    s.EnvoyXdsServer.ConfigUpdate(pushReq)
}
if err := s.ServiceController().AppendServiceHandler(serviceHandler); err != nil {
    return fmt.Errorf("append service handler failed: %v", err)
}
```


当除k8s外的registry变化时，全量更新

```
instanceHandler := func(si *model.ServiceInstance, _ model.Event) {
    // TODO: This is an incomplete code. This code path is called for consul, etc.
    // In all cases, this is simply an instance update and not a config update. So, we need to update
    // EDS in all proxies, and do a full config push for the instance that just changed (add/update only).
    s.EnvoyXdsServer.ConfigUpdate(&model.PushRequest{
        Full: true,
        ConfigsUpdated: map[model.ConfigKey]struct{}{{
            Kind:      gvk.ServiceEntry,
            Name:      string(si.Service.Hostname),
            Namespace: si.Service.Attributes.Namespace,
        }: {}},
        Reason: []model.TriggerReason{model.ServiceUpdate},
    })
}
```

如果设置IngressControllerMode，且ingress变化，则进行全量更新

- initDiscoveryService

初始化15010端口监听 grpc plain text请求


- initDNSServer

15053监听dns请求，基于miekg/dns实现,貌似coredns也是用的这个

- initNamespaceController

初始化namespace controller，查看istio-ca-root-cert cm对变更进行merge,删除则创建，命名空间状态变化写入cm

- RunAndWait

运行各种informer