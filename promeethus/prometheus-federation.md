# federation

通过federate接口来收集数据，该接口

## 分层联邦

使prometheus具有横向扩展能力，形成树形架构，上级从下级收集汇总的时间序列


##跨服务联邦

从另外的server上面获取指定的数据，从而经过计算形成报警查询标准


## 配置示例


```
 job_name: 'federate'
  scrape_interval: 15s

  honor_labels: true
  metrics_path: '/federate'

  params:
    'match[]':
      - '{job="prometheus"}'
      - '{__name__=~"job:.*"}'

  static_configs:
    - targets:
      - 'source-prometheus-1:9090'
      - 'source-prometheus-2:9090'
      - 'source-prometheus-3:9090'
```