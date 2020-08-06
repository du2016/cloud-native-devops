constructProxyConfig 构造proxy config,

getDNSDomain 根据registry生成域名

NewServer

initWorkloadSdsService

initWorkloadSdsService

startXDS 创建xdsclient和xdsserver
  
  
type SecretDiscoveryServiceServer interface {
    DeltaSecrets(SecretDiscoveryService_DeltaSecretsServer) error
    // SDS API
    StreamSecrets(SecretDiscoveryService_StreamSecretsServer) error
    // 获取secret
    FetchSecrets(context.Context, *envoy_api_v2.DiscoveryRequest) (*envoy_api_v2.DiscoveryResponse, error)
}

StreamSecrets



SecretManager

type SecretManager interface {
	// GenerateSecret generates new secret and cache the secret.
	// Current implementation constructs the SAN based on the token's 'sub'
	// claim, expected to be in the K8S format. No other JWTs are currently supported
	// due to client logic. If JWT is missing/invalid, the resourceName is used.
	GenerateSecret(ctx context.Context, connectionID, resourceName, token string) (*SecretItem, error)

	// ShouldWaitForIngressGatewaySecret indicates whether a valid ingress gateway secret is expected.
	ShouldWaitForGatewaySecret(connectionID, resourceName, token string, fileMountedCertsOnly bool) bool

	// SecretExist checks if secret already existed.
	// This API is used for sds server to check if coming request is ack request.
	SecretExist(connectionID, resourceName, token, version string) bool

	// DeleteSecret deletes a secret by its key from cache.
	DeleteSecret(connectionID, resourceName string)
}

# 证书生成流程

根据参数生成csr
pkiutil.GenCSR(options) 

# sendRetriableRequest

- CSRSign

```
发送请求到认证中心签发证书
func (c *citadelClient) CSRSign


func (c *istioCertificateServiceClient) CreateCertificate(ctx context.Context, in *IstioCertificateRequest, opts ...grpc.CallOption) 
请求下面的接口进行证书签发
/istio.v1.auth.IstioCertificateService/CreateCertificate"
```

- ExchangeToken


generateSecret 生成secret 


sendRetriableRequest




https://github.com/envoyproxy/data-plane-api/blob/master/xds_protocol.rst#id35