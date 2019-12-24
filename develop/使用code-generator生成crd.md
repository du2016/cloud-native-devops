# code-generator

用于生成k8s风格的api代码

# 生成器

- client-gen
- conversion-gen
- deepcopy-gen
- defaulter-gen
- go-to-protobuf
- import-boss
- informer-gen
- lister-gen
- openapi-gen
- register-gen
- set-gen

## client-gen

在`pkg/apis/${GROUP}/${VERSION}/types.go`中使用，使用`// +genclient`标记对应类型生成的客户端，
如果与该类型相关联的资源不是命名空间范围的(例如PersistentVolume),
则还需要附加`// + genclient：nonNamespaced`标记，

- `// +genclient` - 生成默认的客户端动作函数（create, update, delete, get, list, update, patch, watch以及
是否生成updateStatus取决于.Status字段是否存在）。
- `// +genclient:nonNamespaced` - 所有动作函数都是在没有名称空间的情况下生成
- `// +genclient:onlyVerbs=create,get` - 指定的动作函数被生成.
- `// +genclient:skipVerbs=watch` - 生成watch以外所有的动作函数.
- `// +genclient:noStatus` - 即使`.Status`字段存在也不生成updateStatus动作函数

[官方文档](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/generating-clientset.md)

## conversion-gen

conversion-gen是用于自动生成在内部和外部类型之间转换的函数的工具。一般的转换代码生成任务涉及三套程序包：

- 一套包含内部类型的程序包，
- 一套包含外部类型的程序包，
- 单个目标程序包（即，生成的转换函数所在的位置，以及开发人员授权的转换功能所在的位置）。
包含内部类型的包在Kubernetes的常规代码生成框架中扮演着称为`peer package`的角色。

使用方法

- 标记转换内部软件包 `// +k8s:conversion-gen=<import-path-of-internal-package>`
- 标记转换外部软件包`// +k8s:conversion-gen-external-types=<import-path-of-external-package>`
- 标记不转换对应注释或结构 `// +k8s:conversion-gen=false`

[官方文档](https://github.com/kubernetes/code-generator/blob/master/cmd/conversion-gen/main.go)

## deepcopy-gen

deepcopy-gen是用于自动生成DeepCopy函数的工具，使用方法：

- 在文件中添加注释`// +k8s:deepcopy-gen=package`
- 为单个类型添加自动生成`// +k8s:deepcopy-gen=true`
- 为单个类型关闭自动生成`// +k8s:deepcopy-gen=false`

## defaulter-gen

用于生成Defaulter函数

- 为包含字段的所有类型创建defaulters，`// +k8s:defaulter-gen=<field-name-to-flag>`
- 所有都生成`// +k8s:defaulter-gen=true|false`

## go-to-protobuf

通过go struct生成pb idl

## import-boss

在给定存储库中强制执行导入限制

## informer-gen

生成informer

## lister-gen

生成对应的lister方法

## openapi-gen
生成openAPI定义

使用方法：

- `+k8s:openapi-gen=true` 为指定包或方法开启
- `+k8s:openapi-gen=false` 指定包关闭
## register-gen
生成register


# 手动添加基础代码

ip是我们定义的新的资源

- pkg/apis/ip/register.go内容如下

```
package ip

// GroupName is the group name used in this package
const (
	GroupName = "rocdu.top"
)
```

- pkg/apis/ip/v1/doc.go内容如下

```
// +k8s:deepcopy-gen=package

// Package v1 is the v1 version of the API.
// +groupName=rocdu.top
package v1
```

- pkg/apis/ip/v1/types.go内容如下,该文件包含了资源的数据结构，对应yaml
```
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ip is a res to get  node/pod from ip
type Ip struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IpSpec `json:"spec"`
}

type IpSpec struct {
	Pod  string `json:"pod,omitempty"`
	Node string `json:"node,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IpCrdList is a list of ip
type IpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Ip `json:"items"`
}
```

- boilerplate.go.txt该文件是文件开头统一的注释

```
boilerplate.go.txt

/*
/*
@Time : 2019/12/23 3:08 下午
@Author : tianpeng.du
@File : types
@Software: GoLand
*/
```

- pkg/apis/ip/v1/register.go 

```
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/du2016/code-generator/pkg/apis/ip"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: ip.GroupName, Version: "v1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder initializes a scheme builder
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is a global function that registers this API group & version to a scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Ip{},
		&IpList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
```

# 生成代码

```
./vendor/k8s.io/code-generator/generate-groups.sh all github.com/du2016/code-generator/pkg/client github.com/du2016/code-generator/pkg/apis ip:v1
```

# 使用crd informer

```
informer := externalversions.NewSharedInformerFactoryWithOptions(clientset, 10*time.Second, externalversions.WithNamespace("default"))
go informer.Start(nil)

IpCrdInformer:=informer.Rocdu().V1().Ips()
cache.WaitForCacheSync(nil,IpCrdInformer.Informer().HasSynced)
```


https://github.com/kubernetes/code-generator
https://github.com/kubernetes/sample-controller
