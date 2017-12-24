client-go除了提供clientset的连接方式，还提供了dynamic client 和restful api的连接方式与apiserver交互

通过dynamic client可以访问所有资源（包括thirdpartresource所能提供的资源）


```
package main  
  
import (  
    "encoding/json"  
    "flag"  
    "k8s.io/client-go/1.5/dynamic"  
    "k8s.io/client-go/1.5/pkg/api/unversioned"  
    "k8s.io/client-go/1.5/pkg/api/v1"  
    "k8s.io/client-go/1.5/pkg/runtime"  
    "k8s.io/client-go/1.5/pkg/watch"  
    "k8s.io/client-go/1.5/rest"  
    "k8s.io/client-go/1.5/tools/clientcmd"  
    "log"  
    "reflect"  
)  
  
var (  
    kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")  
)  
  
func main() {  
    log.SetFlags(log.Llongfile)  
    flag.Parse()  
    //获取Config  
    config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)  
    if err != nil {  
        log.Println(err)  
    }  
    //指定gv  
    gv := &unversioned.GroupVersion{"", "v1"}  
    //指定resource  
    resource := &unversioned.APIResource{Name: "pods", Namespaced: true}  
  
    //指定GroupVersion  
    config.ContentConfig = rest.ContentConfig{GroupVersion: gv}  
    //默认的是/api 需要手动指定  
    config.APIPath = "/api"  
    //创建新的dynamic client  
    cl, err := dynamic.NewClient(config)  
    if err != nil {  
        log.Println(err)  
    }  
  
    //根据APIResource获取  
    obj, err := cl.Resource(resource, "default").Get("golang")  
    if err != nil {  
        log.Println(err)  
    }  
    pod := v1.Pod{}  
    b, err := json.Marshal(obj.Object)  
    if err != nil {  
        log.Println(err)  
    }  
    json.Unmarshal(b, &pod)  
    log.Println(pod.Name)  
  
    //创建pod  
    conf := make(map[string]interface{})  
    conf = map[string]interface{}{  
        "apiVersion": "v1",  
        "kind":       "Pod",  
        "metadata": map[string]interface{}{  
            "name": "golang1",  
        },  
        "spec": map[string]interface{}{  
            "containers": []map[string]interface{}{  
                map[string]interface{}{  
                    "image": "golang",  
                    "command": []string{  
                        "sleep",  
                        "3600",  
                    },  
                    "name": "golang1",  
                },  
            },  
        },  
    }  
    podobj := runtime.Unstructured{Object: conf}  
    _, err = cl.Resource(resource, "default").Create(&podobj)  
    if err != nil {  
        log.Println(err)  
    }  
    // 删除一个pod,删除资源前最好获取UUID  
    cl.Resource(resource, "default").Delete("golang1", &v1.DeleteOptions{})  
  
    // 获取列表  
    got, err := cl.Resource(resource, "default").List(&v1.ListOptions{})  
    if err != nil {  
        log.Println(err)  
    }  
    js, err := json.Marshal(reflect.ValueOf(got).Elem().Interface())  
    if err != nil {  
        log.Println(err)  
    }  
    podlist := v1.PodList{}  
    err = json.Unmarshal(js, &podlist)  
    if err != nil {  
        log.Println(err)  
    }  
    log.Println(podlist.Items[0].Name)  
  
    // 获取thirdpart resource  
    gvthird := &unversioned.GroupVersion{"test.io", "v1"}  
    thirdpartresource := &unversioned.APIResource{Name: "podtoservices", Namespaced: true}  
    config.ContentConfig = rest.ContentConfig{GroupVersion: gvthird}  
    config.APIPath = "/apis"  
    clthird, err := dynamic.NewClient(config)  
    if err != nil {  
        log.Println(err)  
    }  
    objthird, err := clthird.Resource(thirdpartresource, "default").Get("redis-slave-360xf")  
    if err != nil {  
        log.Println(err)  
    }  
    log.Println(objthird)  
  
    //watch一个resource  
    watcher, err := clthird.Resource(thirdpartresource, "").Watch(&unversioned.TypeMeta{})  
    if err != nil {  
        log.Println(err)  
    }  
  
    c := watcher.ResultChan()  
    for {  
        select {  
        case e := <-c:  
            getptrstring(e)  
        }  
    }  
}  
  
func getptrstring(e watch.Event) {  
    v := reflect.ValueOf(e.Object)  
    log.Printf("Type: %s --- Obj: %s", e.Type, v.Elem().Interface())  
}  

```