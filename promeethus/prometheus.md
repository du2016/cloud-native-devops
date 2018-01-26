# prometheus基于yaml格式配置

```
global:
  scrape_interval:     15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - 172.26.6.1:9093
rule_files:
  - "first_rules.yml"
  - "second_rules.yml"
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
```

## global全局配置
scrape_interval 数据收集频率默认一分钟
scrape_timeout 收集超市时间 默认10秒
evaluation_interval 执行规则的频率（报警，数据的预先计算）
external_labels 所有序列的通知标签

## rule_files配置规则和报警策略

### 语法检查

```
go get github.com/prometheus/prometheus/cmd/promtool
promtool check rules /path/to/example.rules.yml
```

### 报警条件
```
groups:
- name: test-alert
  rules:
  - alert: testing
    expr: up{instance="localhost:9090",job="prometheus"} > 0
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: High request latency
```
#### 使用模板报警

模板报警介绍 https://prometheus.io/docs/prometheus/latest/configuration/template_examples/
```
groups:
- name: example
  rules:

  # Alert for any instance that is unreachable for >5 minutes.
  - alert: InstanceDown
    expr: up == 0
    for: 5m
    labels:
      severity: page
    annotations:
      summary: "Instance {{ $labels.instance }} down"
      description: "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 5 minutes."

  # Alert for any instance that has a median request latency >1s.
  - alert: APIHighRequestLatency
    expr: api_http_request_latencies_second{quantile="0.5"} > 1
    for: 10m
    annotations:
      summary: "High request latency on {{ $labels.instance }}"
      description: "{{ $labels.instance }} has a median request latency above 1s (current value: {{ $value }}s)"
```

### 数据计算规则
```
groups:
- name: test
  rules:
  - record: test:test
    expr: up{instance="localhost:9090",job="prometheus"}
```

## scrape_configs获取数据相关配置

分为静态配置和服务发现

```
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
```
scrape_interval 覆盖全局配置
scrape_timeout 覆盖全局配置
metrics_path 获取对应api接口 默认是/metrics
honor_labels 用来解决当标签冲突时的情况，布尔类型 true则保留原始值，忽略服务端设置的标记，false则原始标签前添加exported_前缀，external_labels不收影响
scheme 请求方法
params 添加参数
basic_auth 认证
bearer_token 请求秘钥
bearer_token_file 秘钥文件
tls_config 请求证书
proxy_url 代理url

[服务发现](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#<scrape_config>)
azure_sd_configs 
consul_sd_configs
dns_sd_configs
ec2_sd_configs
openstack_sd_configs
file_sd_configs
gce_sd_configs
kubernetes_sd_configs
marathon_sd_configs
nerve_sd_configs
serverset_sd_configs
triton_sd_configs
static_configs

relabel_configs 根据服务发现元数据进行label https://prometheus.io/docs/prometheus/latest/configuration/configuration/#<relabel_config>
metric_relabel_configs
alert_relabel_configs

## alertmanager_config
同scrape_configs

## remote_read&write

url
write_relabel_configs
basic_auth
bearer_token
bearer_token_file
tls_config
proxy_url


[后端存储支持列表](https://prometheus.io/docs/operating/integrations/#remote-endpoints-and-storage)
[示例](https://github.com/prometheus/prometheus/tree/master/documentation/examples/remote_storage/remote_storage_adapter)


## alertmanager webhook
[支持列表](https://prometheus.io/docs/operating/integrations/#alertmanager-webhook-receiver)
指定webhook alertmgr发送告警到指定接口

## promgen
方便的创建报警规则和配置文件
https://github.com/line/promgen

## Prometheus operator
https://github.com/coreos/prometheus-operator