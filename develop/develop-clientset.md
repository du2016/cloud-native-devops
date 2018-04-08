1.使用client-go out-of-cluster

2.如果在集群内部可以使用incluster配置，只需要导入"k8s.io/client-go/1.5/rest" 使用config, err := rest.InClusterConfig()

2.需要将kubeconfig文件放到指定位置

```
package main  
  
import (  
    "flag"  
    "k8s.io/client-go/1.5/kubernetes"  
    "k8s.io/client-go/1.5/pkg/api"  
    "k8s.io/client-go/1.5/pkg/api/unversioned"  
    "k8s.io/client-go/1.5/pkg/api/v1"  
    "k8s.io/client-go/1.5/tools/clientcmd"  
    "log"  
)  
  
var (  
    kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")  
)  
  
func main() {  
    flag.Parse()  
    // uses the current context in kubeconfig  
    config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)  
    if err != nil {  
        panic(err.Error())  
    }  
    // 创建client set  
    clientset, err := kubernetes.NewForConfig(config)  
    if err != nil {  
        panic(err.Error())  
    }  
    // 获取现有的pod数量  
    pods, err := clientset.Core().Pods("").List(api.ListOptions{})  
    check_err(err)  
    log.Printf("there are %d pods in cluster\n", len(pods.Items))  
  
    // 创建pod  
    pod := new(v1.Pod)  
    pod.TypeMeta = unversioned.TypeMeta{Kind: "Pod", APIVersion: "v1"}  
    pod.ObjectMeta = v1.ObjectMeta{Name: "testapi", Namespace: "default", Labels: map[string]string{"name": "testapi"}}  
    pod.Spec = v1.PodSpec{  
        RestartPolicy: v1.RestartPolicyAlways,  
        Containers: []v1.Container{  
            v1.Container{  
                Name:  "testapi",  
                Image: "nginx",  
                Ports: []v1.ContainerPort{  
                    v1.ContainerPort{  
                        ContainerPort: 80,  
                        Protocol:      v1.ProtocolTCP,  
                    },  
                },  
            },  
        },  
    }  
    podname, err := clientset.Core().Pods("default").Create(pod)  
    check_err(err)  
    log.Printf("pod %s have cretae\n", podname.ObjectMeta.Name)  
  
    // 创建namespace  
    ns := new(v1.Namespace)  
    ns.TypeMeta = unversioned.TypeMeta{Kind: "NameSpace", APIVersion: "v1"}  
    ns.ObjectMeta = v1.ObjectMeta{  
        Name: "k8s-test",  
    }  
    ns.Spec = v1.NamespaceSpec{}  
    nsname, err := clientset.Core().Namespaces().Create(ns)  
    check_err(err)  
    log.Printf("namespace %s have cretae\n", nsname.ObjectMeta.Name)  
  
    // 获取现有的pod数量  
    pods, err = clientset.Core().Pods("").List(api.ListOptions{})  
    check_err(err)  
    log.Printf("there are %d pods in cluster\n", len(pods.Items))  
  
    //根据名称获取pod  
    geterpod, err := clientset.Core().Pods("default").Get(podname.ObjectMeta.Name)  
    check_err(err)  
    // 删除pod  
    // 因为关系到时间复杂度 需要加上UID保证唯一性  
    err = clientset.Core().Pods("default").Delete(geterpod.ObjectMeta.Name, &api.DeleteOptions{Preconditions: &api.Preconditions{UID: &geterpod.ObjectMeta.UID}})  
    check_err(err)  
    log.Printf("namespace %s have delete\n", "testapi")  
  
    //根据名称获取namespace  
    geternsname, err := clientset.Core().Namespaces().Get(nsname.ObjectMeta.Name)  
    check_err(err)  
  
    // 删除namespace  
    err = clientset.Core().Namespaces().Delete(geternsname.ObjectMeta.Name, &api.DeleteOptions{Preconditions: &api.Preconditions{UID: &geternsname.ObjectMeta.UID}})  
    check_err(err)  
    log.Printf("namespace %s have delete\n", geternsname.ObjectMeta.Name)  
  
}  
  
func check_err(err error) {  
    if err != nil {  
        log.Fatal("got err from apiserver: %s\n", err)  
    }  
}  
```
