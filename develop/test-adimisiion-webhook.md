
# 目标

本篇文章我们将参照官方的测试实例来一步步添加一个Admission Webhook

#配置

## webhook证书生成

由上节的配置我们可以看出，我们的webhook必须使用https,所以我们需要生成一个自签名的https证书，
ca可以使用自定义的也可以共用apiserver的。



脚本可以参照istio的证书生成脚本: `https://raw.githubusercontent.com/istio/istio/release-0.7/install/kubernetes/webhook-create-signed-cert.sh`
```
tmpdir=$(mktemp -d) #生成临时目录
service=sidecar-injector-webhook-svc #指定webhook对应的svc name
namespace=default #指定 运行的命名空间
secret=sidecar-injector-webhook-certs # 指定secret name
csrName=${service}.${namespace}  #指定CertificateSigningRequest 的名称

cat <<EOF >> ${tmpdir}/csr.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF

openssl genrsa -out server-key.pem 2048 #生成私钥
openssl req -new -key ${tmpdir}/server-key.pem -subj "/CN=${service}.${namespace}.svc" -out ${tmpdir}/server.csr -config ${tmpdir}/csr.conf # 生成证书签名请求


通过k8s CSR生成证书
cat <<EOF | kubectl create -f -
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
 name: ${csrName}
spec:
 groups:
 - system:authenticated
 request: $(cat ${tmpdir}/server.csr | base64 | tr -d '\n')
 usages:
 - digital signature
 - key encipherment
 - server auth
EOF

kubectl certificate approve ${csrName} # 通过签名请求
serverCert=$(kubectl get csr ${csrName} -o jsonpath='{.status.certificate}') # 获得证书内容
echo ${serverCert} | openssl base64 -d -A -out ${tmpdir}/server-cert.pem # 写入文件

kubectl create secret generic ${secret} \
       --from-file=key.pem=${tmpdir}/server-key.pem \
       --from-file=cert.pem=${tmpdir}/server-cert.pem \ 生成secret 挂载进webhook的容器内部
```
### 认证apiserver配置



如果我们的webhook需要对请求进行身份认证,那么我们需要对apiserver进行以下配置:
```
--admission-control-config-file=/etc/kubernetes/admission-control.conf
```

/etc/kubernetes/admission-control.conf 文件内容如下:
```
apiVersion: apiserver.k8s.io/v1alpha1
kind: AdmissionConfiguration
plugins:
- name: MutatingAdmissionWebhook
 configuration:
   apiVersion: apiserver.config.k8s.io/v1alpha1
   kind: WebhookAdmission
   kubeConfigFile: /etc/kubeconfig/mutating-admission-webhook.kubeconfig
```
/etc/kubeconfig/mutating-admission-webhook.kubeconfig 配置如下：


```
apiVersion: v1
kind: Config
users:
- name: 'sidecar-injector-webhook-svc.default.svc'
 user:
   token: "testtoken"
```

MutatingWebhookConfiguration 配置
```
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
 name: sidecar-injector-webhook-cfg
 labels:
   app: sidecar-injector
webhooks:
 - name: sidecar-injector.morven.me
   clientConfig:
     service:
       name: sidecar-injector-webhook-svc
       namespace: default
       path: "/mutating-pods"
   rules:
     - operations: [ "CREATE" ]
       apiGroups: [""]
       apiVersions: ["v1"]
       resources: ["pods"]
   namespaceSelector:
     matchLabels:
       sidecar-injector: enabled
```
## 代码实现

接下来我们参照官方的测试实例实现一个 sidecar inject MutatingAdmissionWebhook。
官方的测试实例地址：
[https://github.com/kubernetes/kubernetes/tree/v1.15.0/test/images/webhook](https://github.com/kubernetes/kubernetes/tree/v1.15.0/test/images/webhook)
```
package main

import (
  "crypto/tls"
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "net/http"
  "k8s.io/api/admission/v1beta1"
  admissionv1beta1 "k8s.io/api/admission/v1beta1"
  admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/runtime"
  "k8s.io/apimachinery/pkg/runtime/serializer"
  utilruntime "k8s.io/apimachinery/pkg/util/runtime"
  "k8s.io/klog"
  // TODO: try this library to see if it generates correct json patch
)

var scheme = runtime.NewScheme() //用于序列化
var codecs = serializer.NewCodecFactory(scheme) //提供检索序列化方式的函数

const (
  //返回给apiserver的patcher
  podsInjectContainerPatch string = `{"op":"add","path":"/spec/containers/-","value":[{"image":"envoy","name":"envoy","resources":{}}]}`
  addInjectAnnotationPatch string = `{"op":"add","path":"/metadata/annotations","value": {"sidecar-injector-webhook/status": "true"}}`
  updateInjectAnnotationPatch string = `{"op":"replace","path":"/metadata/annotations/sidecar-injector-webhook/status","value":"true"}`
  injectAnnotation string = "sidecar-injector-webhook/status"
)

func init() {
  // 将数据类型注册到scheme
  utilruntime.Must(corev1.AddToScheme(scheme))
  utilruntime.Must(admissionv1beta1.AddToScheme(scheme))
  utilruntime.Must(admissionregistrationv1beta1.AddToScheme(scheme))
}

// 创建AdmissionResponse
func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
  return &v1beta1.AdmissionResponse{
     Result: &metav1.Status{
        Message: err.Error(),
     },
  }
}


type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

// 使用对应http.handlefunc 处理请求
func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
  var body []byte
  if r.Body != nil {
     if data, err := ioutil.ReadAll(r.Body); err == nil {
        body = data
     }
  }
 
   // 校验token
   token:=r.Header.Get("Authorization")
   if token!=fmt.Sprintf("Bearer %s",bearToken) {
       klog.Errorf("unexpect bear token: %s", token)
       return
   }

  // 校验content type 只接收json
  contentType := r.Header.Get("Content-Type")
  if contentType != "application/json" {
     klog.Errorf("contentType=%s, expect application/json", contentType)
     return
  }

  klog.V(2).Info(fmt.Sprintf("handling request: %s", body))

  // 接收到的AdmissionReview
  requestedAdmissionReview := v1beta1.AdmissionReview{}

  // 用于返回的AdmissionReview
  responseAdmissionReview := v1beta1.AdmissionReview{}

  deserializer := codecs.UniversalDeserializer() //返回runtime.Decoder 用于转换成为k8s的runtime.Object
  if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil { //将body转换为AdmissionReview
     klog.Error(err)
     responseAdmissionReview.Response = toAdmissionResponse(err) //返回包含错误的AdmissionResponse
  } else {
     klog.Info("s1")
     responseAdmissionReview.Response = admit(requestedAdmissionReview) //通过对应handlefunc处理
  }

  responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID //响应 UUID需要同请求UUID相同

  klog.V(2).Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

  respBytes, err := json.Marshal(responseAdmissionReview) //序列化为JSON
  if err != nil {
     klog.Error(err)
  }
  if _, err := w.Write(respBytes); err != nil {
     klog.Error(err)
  }
}

func serveMutatePods(w http.ResponseWriter, r *http.Request) {
  serve(w, r, mutatePods)
}

func main() {
  klog.InitFlags(nil)
  flag.Parse()
  http.HandleFunc("/mutating-pods", serveMutatePods)
  sCert, err := tls.LoadX509KeyPair("/etc/certs/cert.pem", "/etc/certs/key.pem")
  if err != nil {
     klog.Fatal(err)
  }
  server := &http.Server{
     Addr: ":443",
     TLSConfig: &tls.Config{
        Certificates: []tls.Certificate{sCert},
     },
  }
  server.ListenAndServeTLS("", "")
}

func mutatePods(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
  klog.V(2).Info("mutating pods")
  podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
  if ar.Request.Resource != podResource { //判断类型是否相同
     klog.Errorf("expect resource to be %s %s", podResource,ar.Request.Resource)
     return nil
  }

  raw := ar.Request.Object.Raw
  pod := corev1.Pod{}
  deserializer := codecs.UniversalDeserializer()
  if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil { //反序列化为pod
     klog.Error(err)
     return toAdmissionResponse(err)
  }
  reviewResponse := v1beta1.AdmissionResponse{}
  reviewResponse.Allowed = true
  klog.Info(reviewResponse)
  annotations:=pod.Annotations

  if annotations==nil {
     annotations = map[string]string{}
  }
  if v,ok:=annotations[injectAnnotation];ok { // 没有或不是true 进行注入
     klog.Info(v)
     if v!="true" {
        reviewResponse.Patch = []byte(fmt.Sprintf("{%s,%s}",podsInjectContainerPatch,updateInjectAnnotationPatch)) //设置patcher
        pt := v1beta1.PatchTypeJSONPatch
        reviewResponse.PatchType = &pt
     }
  }else {
     reviewResponse.Patch = []byte(fmt.Sprintf("[%s,%s]",podsInjectContainerPatch,addInjectAnnotationPatch)) //设置patcher
     pt := v1beta1.PatchTypeJSONPatch
     reviewResponse.PatchType = &pt
  }
  klog.Info(string(reviewResponse.Patch))
  return &reviewResponse
}
```