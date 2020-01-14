# envoy rate limit介绍

envoy中有以下限速方式：

- 全局限速

Envoy的全局请求限速服务器，检查是否接受。
全局意味着所有代理都将使用一个计数器作为评估请求的基础。
每个代理都请求一个上游速率限制服务（在此示例中为Lyfts），该服务将在envoy外部运行以决定请求。

- 本地限速

本地速率限制计数器在处理请求的单个envoy代理的上下文中运行。这意味着每个代理都跟踪其管理的连接并应用限速策略（即[熔断](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/circuit_breaking)）。
最新的版本添加了一个使用自身令牌桶进行[本地限速功能](https://github.com/envoyproxy/envoy/pull/9354)

# 环境准备

## 安装envoy

```
brew tap tetratelabs/getenvoy
brew install getenvoy
```

## 启动redis

```
docker run -p 6379:6379 redis
```

## 启动上游服务

```
python -m SimpleHTTPServer 1234
```

# 使用lyft/ratelimit进行限速

## 启动ratelimit

```
export USE_STATSD=false 
export LOG_LEVEL=debug 
export REDIS_SOCKET_TYPE=tcp 
export REDIS_URL=localhost:6379 
export RUNTIME_ROOT="./" 
export RUNTIME_SUBDIRECTORY=ratelimit
git clone https://github.com/lyft/ratelimit.git

cat >> config.yaml < EOF
domain: ratelimiter
descriptors:
- key: header_match
  value: lyft-rate-limit
  rate_limit:
    unit: minute
    requests_per_unit: 2
EOF

cd ratelimit
go get -v github.com/Masterminds/glide
glide install
go run src/service_cmd/main.go
```

## 启动envoy

```
cat >> config.yaml < EOF
admin:
  access_log_path: /tmp/admin_access.log
  address:
    socket_address: { address: 127.0.0.1, port_value: 9901 }

static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address: { address: 127.0.0.1, port_value: 10000 }
    filter_chains:
    - filters:
      - name: envoy.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
          stat_prefix: ingress_http
          codec_type: AUTO
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["ratelimiter"]
              routes:
              - match: { prefix: "/" }
                route: { cluster: some_service }
              rate_limits:
              - actions:
                - header_value_match:
                    descriptor_value: lyft-rate-limit
                    expect_match: false
                    headers:
                    - name: ":path"
                      exact_match: "/"
                stage: 0
          http_filters:
          - name: envoy.rate_limit
            config:
              stage: 0
              domain: "ratelimiter"
              request_type: external
              failure_mode_deny: true
              rate_limit_service:
                grpc_service:
                  envoy_grpc:
                    cluster_name: rate_limit_service
          - name: envoy.local_rate_limit
            config:
              token_bucket:
                max_tokens: 10
                fill_interval: 1s
          - name: envoy.router
  clusters:
  - name: some_service
    connect_timeout: 0.25s
    type: STATIC
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: some_service
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 1234
  - name: rate_limit_service
    connect_timeout: 0.25s
    type: static
    lb_policy: round_robin
    http2_protocol_options: {}
    hosts:
    - socket_address:
        address:  127.0.0.1
        port_value: 8081
EOF

envoy -c config.yaml
```

# 验证

前两次正常，第三次发现返回429，限速正常
```
$ curl -I -H 'HOST: ratelimiter' 127.0.0.1:10000
HTTP/1.1 429 Too Many Requests
x-envoy-ratelimited: true
date: Tue, 14 Jan 2020 07:14:35 GMT
server: envoy
transfer-encoding: chunked
```

扫描关注我:

![微信](http://img.rocdu.top/qrcode_for_gh_7457c3b1bfab_258.jpg)