# 自定义资源

本页阐释了自定义资源的概念，它是对Kubernetes API的扩展。

# 介绍

一种资源就是Kubernetes API中的一个端点，它存储着某种API 对象的集合。 
例如，内建的pods资源包含Pod对象的集合。

自定义资源是对Kubernetes API的一种扩展，它对于每一个Kubernetes集群不一定可用。
换句话说，它代表一个特定Kubernetes的定制化安装。

在一个运行中的集群内，自定义资源可以通过动态注册出现和消失，集群管理员可以独立于集群本身更新自定义资源。
一旦安装了自定义资源，用户就可以通过kubectl创建和访问他的对象，就像操作内建资源pods那样。

# 自定义控制器

自定义资源本身让你简单地存储和索取结构化数据。
只有当和控制器结合后，他们才成为一种真正的declarative API。 控制器将结构化数据解释为用户所期望状态的记录，并且不断地采取行动来实现和维持该状态。

定制化控制器是用户可以在运行中的集群内部署和更新的一个控制器，它独立于集群本身的生命周期。 定制化控制器可以和任何一种资源一起工作，当和定制化资源结合使用时尤其有效。

# 创建CRD


```
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: crontabs.stable.example.com
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: stable.example.com
  # list of versions supported by this CustomResourceDefinition
  versions:
    - name: v1
      # Each version can be enabled/disabled by Served flag.
      served: true
      # One and only one version must be marked as the storage version.
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            json:
              x-kubernetes-preserve-unknown-fields: true
              type: object
              properties:
                spec:
                  type: object
                  properties:
                    cronSpec:
                      type: string
                    image:
                      type: string
                    replicas:
                      type: integer
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: crontabs
    # singular name to be used as an alias on the CLI and for display
    singular: crontab
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CronTab
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - ct 
```