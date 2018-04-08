---
title: "如何实现自己的k8s调度器"
date: 2018-03-15T11:18:29+08:00
draft: false
---

# 调度器介绍

scheduler 是k8s master的一部分

# 自定义调度器方式

- 添加功能重新编译
- 实现自己的调度器（multi-scheduler）
- scheduler调用扩展程序实现最终调度（Kubernetes scheduler extender）

# 添加调度功能

k8s中的[调度算法介绍](https://github.com/kubernetes/community/blob/master/contributors/devel/scheduler_algorithm.md)

[预选](https://github.com/kubernetes/kubernetes/blob/5d6722259204d5677d5af2f38ac3cf5640c2bb2d/pkg/scheduler/algorithm/predicates/predicates.go)
[优选](http://releases.k8s.io/HEAD/pkg/scheduler/algorithm/priorities/)



# 实现自己的调度器(配置多个scheduler)

scheduler以插件形式存在，集群中可以存在多个scheduler，可以显式指定scheduler

## 配置pod使用自己的调度器

下面pod显式指定使用my-scheduler调度器

```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  schedulerName: my-scheduler
  containers:
  - name: nginx
    image: nginx:1.10
```

## 官方给出的shell版本scheduler示例

```
#!/bin/bash
SERVER='localhost:8001'
while true;
do
    for PODNAME in $(kubectl --server $SERVER get pods -o json | jq '.items[] | select(.spec.schedulerName == "my-scheduler") | select(.spec.nodeName == null) | .metadata.name' | tr -d '"')
;
    do
        NODES=($(kubectl --server $SERVER get nodes -o json | jq '.items[].metadata.name' | tr -d '"'))
        NUMNODES=${#NODES[@]}
        CHOSEN=${NODES[$[ $RANDOM % $NUMNODES ]]}
        curl --header "Content-Type:application/json" --request POST --data '{"apiVersion":"v1", "kind": "Binding", "metadata": {"name": "'$PODNAME'"}, "target": {"apiVersion": "v1", "kind"
: "Node", "name": "'$CHOSEN'"}}' http://$SERVER/api/v1/namespaces/default/pods/$PODNAME/binding/
        echo "Assigned $PODNAME to $CHOSEN"
    done
    sleep 1
done
```

# 影响pod调度的因素
https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/

## 预选

过滤不符合运行条件的node

## 优选

对node进行打分

## 抢占
Kubernetes 1.8 及其以后的版本中可以指定 Pod 的优先级。优先级表明了一个 Pod 相对于其它 Pod 的重要性。
当 Pod 无法被调度时，scheduler 会尝试抢占（驱逐）低优先级的 Pod，使得这些挂起的 pod 可以被调度。
在 Kubernetes 未来的发布版本中，优先级也会影响节点上资源回收的排序。

1.9+支持pdb，优先支持PDB策略，但在无法抢占其他pod的情况下，配置pdb策略的pod依旧会被抢占



# Kubernetes scheduler extender

## scheduler策略配置

```
  {
  "kind" : "Policy",
  "apiVersion" : "v1",
  "predicates" : [
  	{"name" : "PodFitsHostPorts"},
  	{"name" : "PodFitsResources"},
  	{"name" : "NoDiskConflict"},
  	{"name" : "MatchNodeSelector"},
  	{"name" : "HostName"}
  	],
  "priorities" : [
  	{"name" : "LeastRequestedPriority", "weight" : 1},
  	{"name" : "BalancedResourceAllocation", "weight" : 1},
  	{"name" : "ServiceSpreadingPriority", "weight" : 1},
  	{"name" : "EqualPriority", "weight" : 1}
  	],
  "extenders" : [
  	{
          "urlPrefix": "http://localhost/scheduler",
          "apiVersion": "v1beta1",
          "filterVerb": "predicates/always_true",
          "bindVerb": "",
          "prioritizeVerb": "priorities/zero_score",
          "weight": 1,
          "enableHttps": false,
          "nodeCacheCapable": false
          "httpTimeout": 10000000000
  	}
      ],
  "hardPodAffinitySymmetricWeight" : 10
  }
```

## 包含extender的配置

```
// ExtenderConfig保存用于与扩展器通信的参数。如果动词是未指定/空的即认为该扩展器选择不提供该扩展。
type ExtenderConfig struct {
	// 访问该extender的url前缀
	URLPrefix string `json:"urlPrefix"`
	//过滤器调用的动词，如果不支持则为空。当向扩展程序发出过滤器调用时，此谓词将附加到URLPrefix
	FilterVerb string `json:"filterVerb,omitempty"`
	//prioritize调用的动词，如果不支持则为空。当向扩展程序发出优先级调用时，此谓词被附加到URLPrefix。
	PrioritizeVerb string `json:"prioritizeVerb,omitempty"`
	//优先级调用生成的节点分数的数字乘数，权重应该是一个正整数
	Weight int `json:"weight,omitempty"`
	//绑定调用的动词，如果不支持则为空。在向扩展器发出绑定调用时，此谓词会附加到URLPrefix。
	//如果此方法由扩展器实现，则将pod绑定动作将由扩展器返回给apiserver。只有一个扩展可以实现这个功能
	BindVerb string
	// EnableHTTPS指定是否应使用https与扩展器进行通信
	EnableHTTPS bool `json:"enableHttps,omitempty"`
	// TLSConfig指定传输层安全配置
	TLSConfig *restclient.TLSClientConfig `json:"tlsConfig,omitempty"`
	// HTTPTimeout指定对扩展器的调用的超时持续时间，过滤器超时无法调度pod。Prioritize超时被忽略
	//k8s或其他扩展器优先级被用来选择节点
	HTTPTimeout time.Duration `json:"httpTimeout,omitempty"`
	//NodeCacheCapable指定扩展器能够缓存节点信息
	//所以调度器应该只发送关于合格节点的最少信息
	//假定扩展器已经缓存了群集中所有节点的完整详细信息
	NodeCacheCapable bool `json:"nodeCacheCapable,omitempty"`
	// ManagedResources是由扩展器管理的扩展资源列表.
	// - 如果pod请求此列表中的至少一个扩展资源，则将在Filter，Prioritize和Bind（如果扩展程序是活页夹）
	//阶段将一个窗格发送到扩展程序。如果空或未指定，所有pod将被发送到这个扩展器。
	// 如果pod请求此列表中的至少一个扩展资源，则将在Filter，Prioritize和Bind（如果扩展程序是活页夹）阶段将一个pod发送到扩展程序。如果空或未指定，所有pod将被发送到这个扩展器。
	ManagedResources []ExtenderManagedResource `json:"managedResources,omitempty"`
}
```

通过k8s predicates和pod过滤的节点集传递给扩展器上的FilterVerb端点的参数。
通过k8s predicates和扩展predicates以及pod过滤的节点集传递给扩展器上的PrioritizeVerb端点的参数。

```
// ExtenderArgs代表被扩展器用于为pod filter/prioritize node所需要的参数
type ExtenderArgs struct {
	// 被调度的pod
	Pod   api.Pod      `json:"pod"`
	// 可被调度的候选列表
	Nodes api.NodeList `json:"nodes"`
}
```

"filter"被调用时返回节点列表(schedulerapi.ExtenderFilterResult)，
"prioritize"返回节点的优先级(schedulerapi.HostPriorityList).
 
"filter"可以根据对应动作对节点列表进行剪裁，"prioritize"返回的分数将添加到k8s最终分数（通过其优先函数进行计算），用于最终宿主选择。
 
“bind”调用用于将pod绑定到节点的代理绑定到扩展器。它可以选择由扩展器实现。当它被实现时，
它是向apiserver发出绑定调用的扩展器的响应。 Pod名称，名称空间，UID和节点名称被传递给扩展器

ExtenderBindingArgs表示将pod绑定到节点的扩展器的参数

```
type ExtenderBindingArgs struct {
	// 将被绑定的pod
	PodName string
	// 将被绑定的namespace
	PodNamespace string
	// poduid
	PodUID types.UID
	// 最终调度到的pod
	Node string
}
```


# 实现

```
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api/v1"
	"log"
	"net/http"
)

var (
	kubeconfig string = "xxx"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("hellowrold"))
	})
	http.HandleFunc("/predicates/test", testPredicateHandler)
	http.HandleFunc("/prioritize/test", testPrioritizeHandler)
	http.HandleFunc("/bind/test", BindHandler)
	http.ListenAndServe(":8880", nil)
}

func testPredicateHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	body := io.TeeReader(r.Body, &buf)
	log.Println(buf.String())

	var extenderArgs schedulerapi.ExtenderArgs
	var extenderFilterResult *schedulerapi.ExtenderFilterResult

	if err := json.NewDecoder(body).Decode(&extenderArgs); err != nil {
		extenderFilterResult = &schedulerapi.ExtenderFilterResult{
			Nodes:       nil,
			FailedNodes: nil,
			Error:       err.Error(),
		}
	} else {
		extenderFilterResult = predicateFunc(extenderArgs)
	}
	if resultBody, err := json.Marshal(extenderFilterResult); err != nil {
		panic(err)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resultBody)
	}

}

func testPrioritizeHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	body := io.TeeReader(r.Body, &buf)
	var extenderArgs schedulerapi.ExtenderArgs
	var hostPriorityList *schedulerapi.HostPriorityList
	if err := json.NewDecoder(body).Decode(&extenderArgs); err != nil {
		panic(err)
	}
	if list, err := prioritizeFunc(extenderArgs); err != nil {
		panic(err)
	} else {
		hostPriorityList = list
	}
	if resultBody, err := json.Marshal(hostPriorityList); err != nil {
		panic(err)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resultBody)
	}
}

func predicateFunc(args schedulerapi.ExtenderArgs) *schedulerapi.ExtenderFilterResult {
	pod := args.Pod
	canSchedule := make([]v1.Node, 0, len(args.Nodes.Items))
	canNotSchedule := make(map[string]string)
	for _, node := range args.Nodes.Items {
		result, err := func(pod v1.Pod, node v1.Node) (bool, error) {
			return true, nil
		}(pod, node)
		if err != nil {
			canNotSchedule[node.Name] = err.Error()
		} else {
			if result {
				canSchedule = append(canSchedule, node)
			}
		}
	}
	result := schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: canSchedule,
		},
		FailedNodes: canNotSchedule,
		Error:       "",
	}
	return &result
}

func prioritizeFunc(args schedulerapi.ExtenderArgs) (*schedulerapi.HostPriorityList, error) {
	nodes := args.Nodes.Items
	var priorityList schedulerapi.HostPriorityList
	priorityList = make([]schedulerapi.HostPriority, len(nodes))
	for i, node := range nodes {
		priorityList[i] = schedulerapi.HostPriority{
			Host:  node.Name,
			Score: 0,
		}
	}
	return &priorityList, nil
}

func BindHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	body := io.TeeReader(r.Body, &buf)
	var extenderBindingArgs schedulerapi.ExtenderBindingArgs
	if err := json.NewDecoder(body).Decode(&extenderBindingArgs); err != nil {
		panic(err)
	}
	b := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{Namespace: extenderBindingArgs.PodNamespace, Name: extenderBindingArgs.PodName, UID: extenderBindingArgs.PodUID},
		Target: v1.ObjectReference{
			Kind: "Node",
			Name: extenderBindingArgs.Node,
		},
	}
	bind(b)

}

func bind(b *v1.Binding) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset.CoreV1().Pods(b.Namespace).Bind(b)
}

```


参考：
https://github.com/kubernetes/community/blob/master/contributors/devel/scheduler.md
https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scheduling/scheduler_extender.md
https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/
https://github.com/kubernetes/kubernetes-docs-cn/blob/master/docs/concepts/overview/extending.md
