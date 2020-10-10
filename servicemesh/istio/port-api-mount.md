# 挂载目录

- istio(cm istio) -> /etc/istio/config    istio的配置文件
- cacerts(secret  cacerts) -> /etc/cacerts  认证的证书
- inject(cm istio-sidecar-injector) -> /var/lib/istio/inject  注入规则  INJECTION_WEBHOOK_CONFIG_NAME
- istiod-service-account  具有istiod-istio-system clusterrole 权限

# api

/inject 用于mutatingwebhookconfigurations注入   initSidecarInjector

/validate  用于validatingwebhookconfigurations   initConfigValidation

/httpsReady 健康检查  initSecureWebhookServer

/ready 健康检查 initIstiodAdminServer 

# port
15053 dns的端口  agent dns port
15013 IstioAgentDNSListenerPort  envoy dns port