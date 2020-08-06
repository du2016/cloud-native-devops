# ConvertService 

将Service转换为serviceentry,
根据annotation.NetworkingExportTo.Name进行判断 

defaultServiceExportTo 定义了默认的export范围

serviceEntry,vs,dr 使用exportTo 定义了暴露的范围

使用sidecar控制export范围

# 暴露范围


k8s service对应 cluster,cluster同时会根据vs,dr生成
k8s的ep对应ep
Destination Rule -> envoyroute


# CDS

```
func (ps *PushContext) Services(proxy *Proxy) []*Service {
	// If proxy has a sidecar scope that is user supplied, then get the services from the sidecar scope
	// sidecarScope.config is nil if there is no sidecar scope for the namespace
	if proxy != nil && proxy.SidecarScope != nil && proxy.Type == SidecarProxy {
		return proxy.SidecarScope.Services()
	}

	out := make([]*Service, 0)

	// First add private services and explicitly exportedTo services
	if proxy == nil {
		for _, privateServices := range ps.privateServicesByNamespace {
			out = append(out, privateServices...)
		}
	} else {
		out = append(out, ps.privateServicesByNamespace[proxy.ConfigNamespace]...)
		out = append(out, ps.servicesExportedToNamespace[proxy.ConfigNamespace]...)
	}

	// Second add public services
	out = append(out, ps.publicServices...)

	return out
}
```