# 挂载目录

- istio(cm istio) -> /etc/istio/config    istio的配置文件
- cacerts(secret  cacerts) -> /etc/cacerts  认证的证书
- inject(cm istio-sidecar-injector) -> /var/lib/istio/inject  注入规则  INJECTION_WEBHOOK_CONFIG_NAME
- istiod-service-account  具有istiod-istio-system clusterrole 权限

# 暴露范围

k8s service对应 cluster,cluster同时会根据vs,dr生成
k8s的ep对应ep
Destination Rule -> envoyroute

# 设置

只允许访问注册的outbound流量
istioctl install --set profile=demo --set 'meshConfig.outboundTrafficPolicy.mode=REGISTRY_ONLY'

设置日志输出
istioctl install --set profile=demo --set meshConfig.accessLogFile="/dev/stdout" 


args代表了所有的命令行参数

# api

/inject 用于mutatingwebhookconfigurations注入   initSidecarInjector

/validate  用于validatingwebhookconfigurations   initConfigValidation

/httpsReady 健康检查  initSecureWebhookServer

/ready 健康检查 initIstiodAdminServer 


15053 dns的端口  agent dns port
15013 IstioAgentDNSListenerPort  envoy dns port