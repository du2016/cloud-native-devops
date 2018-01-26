# prometheus
![aa](./architecture.svg)


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
