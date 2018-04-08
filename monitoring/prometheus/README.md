# prometheus
![aa](architecture.svg)


## 特点
prometheus是一个开源系统监控报警系统

- 尺寸数据
- 强大的查询
- 可视化
- 高效存储
- 操作简单
- 精确报警
- 客户端库多语言支持
- 集成其他服务

## 组件

- prometheus server 用于收集存储时需数据
- 应用程序的客户端库
- push gateway 用于瞬时任务
- 各种exporter
- altermanager
- 其他工具


## 基于pull方式

对于需要push的时候可以使用push gateway


## 四种数据类型

- Counter

```
Counter用于累计值，例如记录请求次数、任务完成数、错误发生次数。一直增加，不会减少。重启进程后，会被重置。
例如：http_response_total{method=”GET”,endpoint=”/api/tracks”} 100，10秒后抓取http_response_total{method=”GET”,endpoint=”/api/tracks”} 100。
```

- Gauge

```
Gauge常规数值，例如 温度变化、内存使用变化。可变大，可变小。重启进程后，会被重置。
例如： memory_usage_bytes{host=”master-01″} 100 < 抓取值、memory_usage_bytes{host=”master-01″} 30、memory_usage_bytes{host=”master-01″} 50、memory_usage_bytes{host=”master-01″} 80 < 抓取值。
```

- Histogram

```
Histogram（直方图）可以理解为柱状图的意思，常用于跟踪事件发生的规模，例如：请求耗时、响应大小。它特别之处是可以对记录的内容进行分组，提供count和sum全部值的功能。
例如：{小于10=5次，小于20=1次，小于30=2次}，count=7次，sum=7次的求和值。
```

- Summary

```
Summary和Histogram十分相似，常用于跟踪事件发生的规模，例如：请求耗时、响应大小。同样提供 count 和 sum 全部值的功能。
例如：count=7次，sum=7次的值求值。

它提供一个quantiles的功能，可以按%比划分跟踪的结果。例如：quantile取值0.95，表示取采样值里面的95%数据。
```


```
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/scrape: true
  labels:
    run: nginx
  name: nginx-aaa
  namespace: default
spec:
  containers:
  - image: nginx
    imagePullPolicy: Always
    name: nginx
```