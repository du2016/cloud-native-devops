restclient 是dynamic client和clientset的基础，支持json与protobuf，可以访问所有资源，实现对自定义thirdpartresource资源的获取

示例代码：




[plain] view plain copy print?
package main  
  
import (  
    "flag"  
    "k8s.io/client-go/pkg/api"  
    "k8s.io/client-go/pkg/api/v1"  
    "k8s.io/client-go/pkg/runtime"  
    "k8s.io/client-go/pkg/runtime/schema"  
    "k8s.io/client-go/pkg/runtime/serializer"  
    "k8s.io/client-go/rest"  
    "k8s.io/client-go/tools/clientcmd"  
    "log"  
)  
  
func main() {  
    log.SetFlags(log.Llongfile)  
    kubeconfig := flag.String("kubeconfig", "./config", "Path to a kube config. Only required if out-of-cluster.")  
    flag.Parse()  
    config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)  
    if err != nil {  
        log.Fatalln(err)  
    }  
    groupversion := schema.GroupVersion{  
        Group:   "k8s.io",  
        Version: "v1",  
    }  
    config.GroupVersion = &groupversion  
    config.APIPath = "/apis"  
    config.ContentType = runtime.ContentTypeJSON  
    config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}  
    restclient, err := rest.RESTClientFor(config)  
    if err != nil {  
        log.Fatalln(err)  
    }  
    e := examples{}  
    err = restclient.Get().  
        Resource("examples").  
        Namespace("default").  
        Name("example1").  
        Do().Into(&e)  
    log.Println(e)  
}  
