# 介绍

xds-relay是面向xDS兼容客户端和服务器的轻量级缓存，聚合和低延迟分发层。在生产环境中下发envoy策略需要传输大量数据，xds-relay通过在一个开放源代码的地方实施本文中概述的所有分布式系统最佳实践，可以帮助大型Envoy部署实现高可用性。

在xds-relay主要实现以下功能：

- 从当前状态到增量的转化，减少下发频率
- 缓存上游更新，下发策略给envoy
- 优雅切换原有xds server

# 使用xds-relay实现xds策略下发

## 先决条件

- envoy 可以使用getenvoy进行安装
- jq
- curl
- 下载xds-relay代码

```
git clone https://github.com/envoyproxy/xds-relay
```

## 运行xds server

```
make build-example-management-server
```

这将在bin目录下生成一个名为example-management-server的二进制文件。该二进制文件运行基于go-control-plane SnapshotCache的简单管理服务器。它每10秒产生一批带有随机版本的xDS数据。

运行xdsserver

```
./bin/example-management-server
```

## 运行xds-relay实例

下一步是配置xds-relay服务器。为此，我们需要提供2个文件：

- 聚合规则文件
- 引导文件

您可以在example/config-files目录中找到每个文件的一个示例，分别为aggregation-rules.yaml和xds-relay-bootstrap.yaml。

运行
```
./bin/xds-relay -a example/config-files/aggregation-rules.yaml -c example/config-files/xds-relay-bootstrap.yaml -m serve
```

## 运行envoy实例

我们将使用它们将envoy实例连接到xds-relay。打开2个终端窗口并运行：

```
envoy -c example/config-files/envoy-bootstrap-1.yaml # on the first window
envoy -c example/config-files/envoy-bootstrap-2.yaml # on the second window
```

## 验证

访问本地的6070端口查看cds信息

```
curl -s 0:6070/cache/staging_cds | jq '(.Cache[0].Resp.Resources.Clusters | map({"name": .name})) as $resp_clusters | (.Cache[0].Requests | map({"version_info": .version_info, "node.id": .node.id, "node.cluster": .node.cluster})) as $reqs | {"response": {"version": .Cache[0].Resp.VersionInfo, "clusters": $resp_clusters}, "requests": $reqs}'
```

Envoy还公开了一个端点，使我们可以研究配置数据的当前状态。如果我们仅关注xds-relay中继的动态集群信息，则可以使用curl通过运行以下命令来检查envoy的cluster信息：

```
curl -s 0:19000/config_dump | jq '.configs | (.[1].dynamic_active_clusters | map({"version": .version_info, "cluster": .cluster.name})) as $clusters | {"clusters": $clusters}'
```

可以看到两边版本是对应的

# 总结

现有控制面，面临着需要下发大量数据，全量更新等问题，xds-relay在现有go-control-plane基础上，增加缓存策略，减少对现有控制面的压力，虽然istio实现了export-to的功能，能够减少下发的策略数量，但是配置比较复杂，很容易出现问题，xds-relay的诞生从另一个角度使大规模策略下发成为可能，将进一步促进servicemesh的落地。但是现在xds-relay还处于开发阶段,还有很大的局限性例如不支持ads，希望xds-relay早日达到生产可用。