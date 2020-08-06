# 基本概念

配置：
- mesh 作用于 mesh侧也就是envoy的配置
- pilot 配置 通过参数定义
- 其他 通过env设置

初始化方法:

bootstrap.NewServer 

# Environment

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


# Server的初始化

- getClusterID(args)

根据参数获取ID，若无且registry包含k8s则设置为kubernetes

- NewDiscoveryServer

根据plugins，Environment创建DiscoveryServer，pugins是网络插件列表，支持authn,authz,mixer

- filewatcher.NewWatcher

初始化一个filewatcher,后续用于监听配置文件变化


# initMeshConfiguration

新建filewatcher,根据/etc/istio/config/mesh生成MeshConfiguration，当文件变化时，重新读取文件内容


# initKubeClient

根据meshconfig判断是否需要初始化kubelet,判断mesh config ConfigSources和Registry是否包含k8s，来对kubecient进行初始化rest client

# initMeshNetworks

根据/etc/istio/config/meshNetworks进行 初始化,提供有关网格内的网络集以及如何路由到每个网络中的端点的信息。

# initMeshHandlers

当mesh或者network配置发生变化时进行全量更新

# initControllers 

初始化 certs、ingress、endponit(也可能是consul或其他registry)，
然后通过NewServiceDiscovery进行转换，这里有两个判断,通过环境变量设置

PILOT_ENABLE_SERVICEENTRY_SELECT_PODS     serviceentry能够筛选到pod
PILOT_ENABLE_K8S_SELECT_WORKLOAD_ENTRIES  service selector包含workload entries  


# initGenerators 

初始化xds需要用到的生成器


# initJwtPolicy

初始化jwt policy

- third-party-jwt
使用istio的token /var/run/secrets/tokens/istio-token
- first-party-jwt
使用k8s的token /var/run/secrets/kubernetes.io/serviceaccount/token


# maybeCreateCA

是否需要创建CA，该ca实际就是secret中的cacerts,如果没有将自行创建


# initIstiodCerts
/var/run/secrets/istio-dns/ 没有则创建用于istio dns的证书

# setPeerCertVerifier

初始化spiffe Verifier

# initSecureDiscoveryService 

初始化SecureDiscoveryService，注册poliot grpc接口，监听15012端口

主要实现下面两个方法

```
- StreamAggregatedResources
- DeltaAggregatedResources
```

StreamAggregatedResources 主要有两个channel

```
- reqChannel
envoy连接上主动请求

- pushChannel

配置变化poliot主动推送
```


# initSecureWebhookServer

初始化webhookserver,"/httpsReady"接口用于健康检查

# initSidecarInjector

创建mutatingwebhook用户实现自动注入，/var/lib/istio/inject/config go tmpl,/var/lib/istio/inject/config用于渲染的值，通过NewWebhook，

inject.NewWebhook(parameters)

监听路径为/inject的请求

# initConfigValidation

创建validatingwebhook 用于校验cr的规范

server.New(params) 监听
/validate
/admitpilot
/admitmixer


# initIstiodAdminServer

监听8080处理/ready

initMonitor

监听15014处理/metrics和"/version"请求


# initRegistryEventHandlers

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

# initDiscoveryService

初始化15010端口监听 grpc plain text请求


# initDNSServer

15053监听dns请求 

# initNamespaceController

初始化namespace controller

