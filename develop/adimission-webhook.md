# 简介

Admission webhooks 是接收准入请求http回调并且进行处理，分为两种类型:

- validating admission Webhook

- mutating admission webhook

mutating admission webhook 先于validating admission Webhook被调用，可以由mutating admission webhook先对 对象进行修改设置默认值，然后validating admission Webhook可以拒绝请求以执行自定义的 admission 策略

# admission webhook controller 处理流程

接受请求–>解析成为AdmissionReview–>解析AdmissionRequest请求资源–> 解析成期望的资源对象–>根据资源对象的现有数据生成pacher—>判断请求AdmissionReview是否包含UUID—返回AdmissionReview

# 配置文件

## 配置文件字段介绍

```
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: <name of this configuration object>
webhooks:
- name: <webhook name, e.g., pod-policy.example.io>
  failurePolicy: Fail #Ignore或Fail
  rules:
  - apiGroups:# 列出一个活多个api group,空代表core *代表所有
    - ""
    apiVersions:# api版本
    - v1
    operations:# 列出需要匹配的动作
    - CREATE
    resources:# 列出一个或多个资源
    - pods
    scope: "Namespaced"# 指定匹配范围有效值为Cluster、Namespaced、*
  clientConfig:
    url: #webhook地址 必须为https
    service:
      namespace: #命名空间
      name: #服务名称 端口必须为443
    caBundle: #pem编码的ca证书，用于签署webhook使用的服务器证书，默认apiserver的系统根证书
  admissionReviewVersions:
  - v1beta1 #版本 默认v1beta1
  timeoutSeconds: 1#请求超时时间，默认30，1-30秒
  namespaceSelector:# 选择过滤哪些ns下的对象
    matchExpressions:
      key: environment
      operator: In
      values:
      - prod
      - staging
  sideEffects: # 明这个webhook是否有副作用，有效值 Unknown, None, Some, NoneOnDryRun，如果具有dryrun属性，切sideEffects为unknown或some，将自动拒绝执行。
```

## 验证apiserver


认证类型：基本身份验证，不记名令牌、证书

- 启动apiserver时，通过 –admission-control-config-file 参数指定许可控制配置文件的位置。

- 在准入控制配置文件中，指定 MutatingAdmissionWebhook 控制器和 ValidatingAdmissionWebhook 控制器应该读取凭据的位置。

- 在kubeconfig中指定凭据

准入配置文件：

```
apiVersion: apiserver.k8s.io/v1alpha1
kind: AdmissionConfiguration
plugins:
- name: ValidatingAdmissionWebhook
  configuration:
    apiVersion: apiserver.config.k8s.io/v1alpha1
    kind: WebhookAdmission
    kubeConfigFile: <path-to-kubeconfig-file>
- name: MutatingAdmissionWebhook
  configuration:
    apiVersion: apiserver.config.k8s.io/v1alpha1
    kind: WebhookAdmission
    kubeConfigFile: <path-to-kubeconfig-file>
```

认证kubeconfig配置：

```
apiVersion: v1
kind: Config
users:
# webhook 的dns名称格式为<service name>.<namespace>.svc,或URL
- name: 'webhook1.ns1.svc'
  user: # 证书认证
    client-certificate-data: <pem encoded certificate>
    client-key-data: <pem encoded key>
# The `name` supports using * to wildmatch prefixing segments.
- name: '*.webhook-company.org'
  user: #用户密码认证
    password: <password>
    username: <name>
# '*'匹配所有.
- name: '*'
  user:
    token: <token> # 令牌认证
```