# envoy dns filter

DNS filter使Envoy可以将转发的DNS查询解析为任何已配置域的权威服务器.filter的配置指定Envoy会回答的名称和地址，以及从外部向未知域发送查询所需的配置.filter支持本地和外部DNS解析。如果名称查找与静态配置的域或配置的群集名称不匹配，Envoy可以将查询引至外部解析器以寻求answer。用户可以选择指定Envoy将用于外部解析的DNS服务器。用户可以通过省略客户端配置对象来禁用外部DNS解析.过滤器支持每个过滤器配置.

[envoy dns filter官方文档](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/udp_filters/dns_filter.html)

# dns拦截

根据配置可以将dns请求拦截到以下端口:

- envoy 端口 15013
- pilot-agent拦截端口  15053

通过以下环境变量决定使用哪种方式：

```
ISTIO_META_DNS_CAPTURE 环境变量代表通过envoy代理dns请求
DNS_AGENT 代表通过pilot-agent代理dns请求
```

dns 证书通过istio chiron 模块调用k8s csr接口进行签发。

# envoy dns listener

由envoy处理dns解析

## buildSidecarListeners

通过buildSidecarListeners，pilot-discovery下发listener配置

```
func (configgen *ConfigGeneratorImpl) buildSidecarListeners(push *model.PushContext, builder *ListenerBuilder) *ListenerBuilder {
	if push.Mesh.ProxyListenPort > 0 {
		// Any build order change need a careful code review
		builder.buildSidecarInboundListeners(configgen).
			buildSidecarOutboundListeners(configgen).
			buildHTTPProxyListener(configgen).
			buildVirtualOutboundListener(configgen).
			buildVirtualInboundListener(configgen).
			buildSidecarDNSListener(configgen)
	}

	return builder
}
```


## 构建dnsfilter

构建envoy dns filter

```
	inlineDNSTable := configgen.buildInlineDNSTable(node, push)

	dnsFilterConfig := &dnsfilter.DnsFilterConfig{
		StatPrefix: "dns",
		ServerConfig: &dnsfilter.DnsFilterConfig_ServerContextConfig{
			ConfigSource: &dnsfilter.DnsFilterConfig_ServerContextConfig_InlineDnsTable{InlineDnsTable: inlineDNSTable},
		},
		ClientConfig: &dnsfilter.DnsFilterConfig_ClientContextConfig{
			ResolverTimeout: ptypes.DurationProto(resolverTimeout),
			// no upstream resolves. Envoy will use the ambient ones
			MaxPendingLookups: 256, // arbitrary
		},
	}
```

## configgen.buildInlineDNSTable

将serviceentry对应的域名写入virtualDomains，

```
apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  name: details-svc
spec:
  hosts:
  - details.bookinfo.com
  location: MESH_INTERNAL
  ports:
  - number: 80
    name: http
    protocol: HTTP
  resolution: DNS
  workloadSelector:
    labels:
      app: details-legacy

---

apiVersion: networking.istio.io/v1beta1
kind: WorkloadEntry
metadata:
  name: details-svc
spec:
  serviceAccount: details-legacy
  address: 1.2.3.4
  labels:
    app: details-legacy
    instance-id: vm1
```

将转化为以下配置

```
listener_filters:
  name: envoy.filters.udp.dns_filter
  typed_config:
    "@type": "type.googleapis.com/envoy.extensions.filters.udp.dns_filter.v3alpha.DnsFilterConfig"
    stat_prefix: "dns_filter_prefix"
    client_config:
      resolution_timeout: 10s
      max_pending_lookups: 256
    server_config:
      inline_dns_table:
        virtual_domains
          - name: details.bookinfo.com
            endpoint:
              address_list:
                address:
                  - 1.2.3.4
```

如上配置未指定upstream_resolvers，envoy将默认使用系统的resolvers

# pilot-agent dns劫持

由pilot-agent执行解析操作

## InitDNSAgent

通过k8s ca 创建证书，连接到istiod 的dns port，通过miekg/dns实现dnsserver，

IstioDNS实现了dns.Handler interface，

```
type Handler interface {
	ServeDNS(w ResponseWriter, r *Msg)
}
```

监听dnsport,转发到istiod,来获取解析记录

```
dnsSrv := dns.InitDNSAgent(proxyConfig.DiscoveryAddress,
    role.DNSDomain, sa.RootCert,
    []string{".global."})
dnsSrv.StartDNS(dns.DNSAgentAddr, nil)
```


# istiod dns server

初始化，注册启动函数

```
func (s *Server) initDNSServer(args *PilotArgs) {
	if dns.DNSAddr.Get() != "" {
		log.Info("initializing DNS server")
		if err := s.initDNSTLSListener(dns.DNSAddr.Get(), args.ServerOptions.TLSOptions); err != nil {
			log.Warna("error initializing DNS-over-TLS listener ", err)
		}

		// Respond to CoreDNS gRPC queries.
		s.addStartFunc(func(stop <-chan struct{}) error {
			if s.DNSListener != nil {
				dnsSvc := dns.InitDNS()
				dnsSvc.StartDNS(dns.DNSAddr.Get(), s.DNSListener)
			}
			return nil
		})
	}
}
```

该server 逻辑和pilot-agent dns逻辑相同，实际上也是一个转发作用，具体处理的dns是coredns


# 总结

通过DNS拦截，istio可以自由下发ServiceEntry的服务名称，从而不依赖于外部解析，实现服务的istio接入。


