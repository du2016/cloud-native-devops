# 介绍

在envoy中有两种方式可以根据地域进行流量转发

- 区域感知路由
- 局部加权负载均衡

两种方式为互斥关系,区域感知路由根据地域进行流量转发,而局部加权负载均衡根据不同地域的权重及ep优先级进行流量转发,下面我们将针对两种方式进行深入讲解.

# 区域感知路由

## 术语

始发/上游群集：Envoy将请求从始发群集路由到上游群集.
本地区域：同一区域,包含原始群集和上游群集中的主机子集.
区域感知路由：将请求的尽力而为路由到本地区域中的上游群集主机.

## 条件

在原始群集和上游群集中的主机属于不同区域的部署中,Envoy执行区域感知路由.在执行区域感知路由之前,有几个先决条件：

1. 原始群集和上游群集都不处于紧急模式.
2. 启用区域感知路由.
3. 上游群集具有足够的主机.

```
cluster.CommonLbConfig = &v2.Cluster_CommonLbConfig{
		LocalityConfigSpecifier: &v2.Cluster_CommonLbConfig_ZoneAwareLbConfig_{
			ZoneAwareLbConfig: &v2.Cluster_CommonLbConfig_ZoneAwareLbConfig{ // 开启区域路由感知,对应要求2
				RoutingEnabled: &_type.Percent{
					Value: 100, // 设置区域亲和的流量百分比
				},
				MinClusterSize: &wrappers.UInt64Value{Value: 1}, // 对应要求3,最小集群数量1
			},
		},
		HealthyPanicThreshold: &_type.Percent{Value: 0}, // 对应要求1,关闭紧急模式
	}
```
- 原始群集具有与上游群集相同的区域数.

该要求意味着你在每一个可用区内都必须要部署至少一个envoy,对应可用区的envoy承载流向对应可用区的流量,在真实环境中最好可以实现自注册,在k8s上部署很容易实现,但是在我们内部实现网关的过程中,在容器内部部署性能太差,我们直接在虚拟机上进行部署,这就要求必须对接自己的服务发现,我们在实现过程中通过将虚机节点同步到k8s ep实现；如果可用区比较少的话也可以通过静态配置进行配置


## 流量百分比决定条件

区域感知路由的目的是向上游群集中的本地区域发送尽可能多的流量,同时在所有上游主机之间大致每秒保持相同数量的请求(取决于负载平衡策略).
Envoy尝试将尽可能多的流量推送到本地上游区域,只要上游群集中每个主机的请求数量保持大致相同即可.Envoy是路由到本地区域还是执行跨区域路由,取决于本地群集中原始群集和上游群集中正常主机的百分比.关于原始集群和上游集群之间的局部区域中的百分比关系,有两种情况：

- 原始群集本地区域百分比大于上游群集中的百分比.在这种情况下,我们无法将所有请求从始发集群的本地区域路由到上游集群的本地区域,因为这将导致所有上游主机之间的请求不平衡.相反,Envoy计算可以直接路由到上游群集本地区域的请求的百分比.其余请求被路由到跨区域.根据区域的剩余容量选择特定区域(该区域将获得一些本地区域流量,并且可能具有Envoy可以用于跨区域流量的其他容量).

- 起始群集本地区域百分比小于上游群集中的百分比.在这种情况下,上游群集的本地区域可以从原始群集的本地区域获取所有请求,并且还具有一定的空间以允许来自原始群集中其他区域的流量(如果需要).

## 如何配置

- envoy必须配置zone信息
- 必须将local_cluster_name设置为源集群.
- 源群集和目标群集的定义都必须具有EDS类型.

```
node:
  id: 1
  cluster: envoy
  locality:
    zone: "us-central1-b"
    region: "us-central1"
cluster_manager:
  local_cluster_name: local_cluster
static_resources:
  clusters:
  - name: local_cluster
    connect_timeout: 0.25s
    type: STATIC
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: local_cluster
      endpoints:
      - locality:
          region: "us-central1"
          zone: 'us-central1-b'
        lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8000
      - locality:
          region: "us-central1"
          zone: 'us-central1-a'
        lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8003
```


对于cds的配置代码如下,策略主要取决于流量百分比和最小集群数
```
&v2.Cluster_CommonLbConfig{
    LocalityConfigSpecifier: &v2.Cluster_CommonLbConfig_ZoneAwareLbConfig_{
        ZoneAwareLbConfig: &v2.Cluster_CommonLbConfig_ZoneAwareLbConfig{
            RoutingEnabled: &_type.Percent{
                Value: 100,
            },
            MinClusterSize: &wrappers.UInt64Value{Value: 1},
        },
    },
}
```

对于eds的配置如下,我们可以从k8s的节点或者高本版的endpointslice中获取位置信息,来进行字段填充

```
&envoy_api_v2_endpoint.LocalityLbEndpoints{
    LbEndpoints: ep,
    Locality: &envoy_api_v2_core.Locality{
        Region: zonetoRegin[zone],
        Zone:   zone,
    },
}
```

# 局部加权负载均衡

确定如何对不同区域和地理位置上的流量分配进行加权的一种方法是使用在LocalityLbEndpoints消息中通过EDS提供显式加权 .这种方法是和区域感知路由是互斥的,因为对于本地化的LB,我们依靠在管理服务器上提供本地权重,而不是在区域感知路由中使用的Envoy端启发式路由.

当所有端点均可用时,使用加权循环调度来选择位置,其中将位置权重用于加权.当某个地点的某些端点不可用时,我们将调整地点权重以反映这一点.与优先级级别一样,我们假设有一个 预留空间因子(默认值为1.4),这意味着当本地中只有少量端点不可用时,我们不执行任何权重调整.

## 配置

在cds中使用Cluster_CommonLbConfig_LocalityWeightedLbConfig,启用局部加权负载均衡

```
cluster.CommonLbConfig = &v2.Cluster_CommonLbConfig{
    LocalityConfigSpecifier: &v2.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{
        LocalityWeightedLbConfig: &v2.Cluster_CommonLbConfig_LocalityWeightedLbConfig{
        },
    },
    HealthyPanicThreshold: &_type.Percent{Value: 0},
}
```

eds配置 相对于区域感知路由,这里多了权重和优先级

```
&envoy_api_v2_endpoint.LocalityLbEndpoints{
    LbEndpoints: ep,
    Locality: &envoy_api_v2_core.Locality{
        Region: zonetoRegin[zone],
        Zone:   zone,
    },
    LoadBalancingWeight: &wrappers.UInt32Value{Value: 80},
    Priority: 0,
}
```

eds配置中还需要设置该eds的空间因子,不设置为140

```
&v2.ClusterLoadAssignment_Policy{
    OverprovisioningFactor: &wrappers.UInt32Value{Value: 140},
}
```

## 流量百分比决定条件

遵循以下伪算法

```
availability(L_X) = 140 * available_X_upstreams / total_X_upstreams
effective_weight(L_X) = locality_weight_X * min(100, availability(L_X))
load to L_X = effective_weight(L_X) / Σ_c(effective_weight(L_c))
```

在选择优先级之后,将进行局部加权选择.负载均衡器遵循以下步骤：

- 选择优先级.
- 在(1)的优先级内选择位置.
- 在(2)中,使用群集中指定的负载均衡器选择端点.

# 总结

使用区域感知路由或者局部加权负载均衡对于使用云的全球化业务非常有用,在保证可用性的基础上,尽量减少跨区域流量,从而节约流量成本,istio中也实现了这两个功能.
