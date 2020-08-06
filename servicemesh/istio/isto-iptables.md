# 环境

这里我们是在本机使用kind，安装



# 查看容器对应的虚拟设备对

```
for container in `crictl ps -q`
do
iflink=`crictl exec $container cat /sys/class/net/eth0/iflink`
iflink=`echo $iflink|tr -d '\r'`
veth=`grep -l $iflink /sys/class/net/*/ifindex`
veth=`echo $veth|sed -e 's;^.*net/\(.*\)/ifindex$;\1;'`
echo $container:$veth
done
```

# 查看iptables规则

```
ip netns exec cni-bf783dac-fe05-cb35-4d5a-848449119b19 iptables -L -t nat

-A PREROUTING -p tcp -j ISTIO_INBOUND                          # PREROUTING全部转发到INBOUND,PREROUTING发生在流入的数据包进入路由表之前
-A OUTPUT -p tcp -j ISTIO_OUTPUT                               # 由本机产生的数据向外转发的
-A ISTIO_INBOUND -p tcp -m tcp --dport 22 -j RETURN            # 22 15090  15021 15020的不转发到ISTIO_REDIRECT 
-A ISTIO_INBOUND -p tcp -m tcp --dport 15090 -j RETURN         
-A ISTIO_INBOUND -p tcp -m tcp --dport 15021 -j RETURN
-A ISTIO_INBOUND -p tcp -m tcp --dport 15020 -j RETURN
-A ISTIO_INBOUND -p tcp -j ISTIO_IN_REDIRECT                   # 剩余的流量都转发到ISTIO_REDIRECT
-A ISTIO_IN_REDIRECT -p tcp -j REDIRECT --to-ports 15006       # 转发到15006
-A ISTIO_OUTPUT -s 127.0.0.6/32 -o lo -j RETURN                # 127.0.0.6是InboundPassthroughBindIpv4，代表原地址是passthrough的流量都直接跳过,不劫持
-A ISTIO_OUTPUT ! -d 127.0.0.1/32 -o lo -m owner --uid-owner 1337 -j ISTIO_IN_REDIRECT  #lo网卡出流量，目标地址不是localhost的，且为同用户的流量进入ISTIO_IN_REDIRECT
-A ISTIO_OUTPUT -o lo -m owner ! --uid-owner 1337 -j RETURN    # lo网卡出流量 非同用户的不劫持
-A ISTIO_OUTPUT -m owner --uid-owner 1337 -j RETURN            # 剩下的同用户的都跳过
-A ISTIO_OUTPUT ! -d 127.0.0.1/32 -o lo -m owner --gid-owner 1337 -j ISTIO_IN_REDIRECT  # lo网卡出流量，目标地址非本地，同用户组的流量进入ISTIO_IN_REDIRECT
-A ISTIO_OUTPUT -o lo -m owner ! --gid-owner 1337 -j RETURN    # lo网卡出流量非同组的不劫持
-A ISTIO_OUTPUT -m owner --gid-owner 1337 -j RETURN            # 剩余的同用户的不劫持
-A ISTIO_OUTPUT -d 127.0.0.1/32 -j RETURN                      # 剩余的目标地址为127的不劫持
-A ISTIO_OUTPUT -j ISTIO_REDIRECT                              # 剩下的都进入 ISTIO_REDIRECT
-A ISTIO_REDIRECT -p tcp -j REDIRECT --to-ports 15001          # 转达到15001 outbond
COMMIT
```

# 请求流程分析

现在有httpbin和sleep两个服务，如果httpbin要访问sleep

- httpbin访问sleep:80端口
- iptables拦截转发到15001 的15001端口

## virtualOutbound Listener

```
{
  "@type": "type.googleapis.com/envoy.api.v2.Listener",
  "name": "virtualOutbound",
  "address": {
    "socket_address": {
      "address": "0.0.0.0",
      "port_value": 15001
    }
  },
  "filter_chains": [
    {
      "filters": [
        {
          "name": "istio.stats", # 为指标添加istio_前缀
          "typed_config": {
            "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
            "type_url": "type.googleapis.com/envoy.extensions.filters.network.wasm.v3.Wasm",
            "value": {
              "config": {
                "root_id": "stats_outbound",
                "vm_config": {
                  "vm_id": "tcp_stats_outbound",
                  "runtime": "envoy.wasm.runtime.null",
                  "code": {
                    "local": {
                      "inline_string": "envoy.wasm.stats"
                    }
                  }
                },
                "configuration": "{\n \"debug\": \"false\",\n \"stat_prefix\": \"istio\"\n}\n"
              }
            }
          }
        },
        {
          "name": "envoy.tcp_proxy",
          "typed_config": {
            "@type": "type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy",
            "stat_prefix": "PassthroughCluster",
            "cluster": "PassthroughCluster",
            "access_log": [
              {
                "name": "envoy.file_access_log",
                "typed_config": {
                  "@type": "type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog",
                  "path": "/dev/stdout",
                  "format": "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \"%DYNAMIC_METADATA(istio.mixer:status)%\" \"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n"
                }
              }
            ]
          }
        }
      ],
      "name": "virtualOutbound-catchall-tcp"
    }
  ],
  "use_original_dst": true, 
  "traffic_direction": "OUTBOUND"
}
```

use_original_dst： 如果使用iptables重定向连接，则代理在其上接收连接的端口可能与原始目标地址不同。 当此标志设置为true时，侦听器将重定向到与原始目标地址关联的侦听器的重定向连接。 如果没有与原始目标地址关联的侦听器，则连接由接收该侦听器的侦听器处理。 默认为false。

我们原本请求的是 sleep:80,则在PassthroughCluster之后重新匹配符合sleep:80的规则

## PassthroughCluster

```
{
  "@type": "type.googleapis.com/envoy.api.v2.Cluster",
  "name": "PassthroughCluster",
  "type": "ORIGINAL_DST",
  "connect_timeout": "10s",
  "lb_policy": "CLUSTER_PROVIDED",
  "circuit_breakers": {
    "thresholds": [
      {
        "max_connections": 4294967295,
        "max_pending_requests": 4294967295,
        "max_requests": 4294967295,
        "max_retries": 4294967295
      }
    ]
  },
  "filters": [
    {
      "name": "istio.metadata_exchange",
      "typed_config": {
        "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
        "type_url": "type.googleapis.com/envoy.tcp.metadataexchange.config.MetadataExchange",
        "value": {
          "protocol": "istio-peer-exchange"
        }
      }
    }
  ]
}
```

## 匹配sleep:80的listener

```
{
  "@type": "type.googleapis.com/envoy.api.v2.Listener",
  "name": "0.0.0.0_80",
  "address": {
    "socket_address": {
      "address": "0.0.0.0",
      "port_value": 80
    }
  },
  "filter_chains": [
    {
      "filter_chain_match": {
        "application_protocols": [
          "http/1.0",
          "http/1.1",
          "h2c"
        ]
      },
      "filters": [
        {
          "name": "envoy.http_connection_manager",
          "typed_config": {
            "@type": "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
            "stat_prefix": "outbound_0.0.0.0_80",
            "rds": {
              "config_source": {
                "ads": {
                }
              },
              "route_config_name": "80"
            },
            "http_filters": [
              {
                "name": "istio.metadata_exchange",
                "typed_config": {
                  "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
                  "type_url": "type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm",
                  "value": {
                    "config": {
                      "vm_config": {
                        "runtime": "envoy.wasm.runtime.null",
                        "code": {
                          "local": {
                            "inline_string": "envoy.wasm.metadata_exchange"
                          }
                        }
                      },
                      "configuration": "{}\n"
                    }
                  }
                }
              },
              {
                "name": "istio.alpn",
                "typed_config": {
                  "@type": "type.googleapis.com/istio.envoy.config.filter.http.alpn.v2alpha1.FilterConfig",
                  "alpn_override": [
                    {
                      "alpn_override": [
                        "istio-http/1.0",
                        "istio"
                      ]
                    },
                    {
                      "upstream_protocol": "HTTP11",
                      "alpn_override": [
                        "istio-http/1.1",
                        "istio"
                      ]
                    },
                    {
                      "upstream_protocol": "HTTP2",
                      "alpn_override": [
                        "istio-h2",
                        "istio"
                      ]
                    }
                  ]
                }
              },
              {
                "name": "envoy.cors",
                "typed_config": {
                  "@type": "type.googleapis.com/envoy.config.filter.http.cors.v2.Cors"
                }
              },
              {
                "name": "envoy.fault",
                "typed_config": {
                  "@type": "type.googleapis.com/envoy.config.filter.http.fault.v2.HTTPFault"
                }
              },
              {
                "name": "istio.stats",
                "typed_config": {
                  "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
                  "type_url": "type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm",
                  "value": {
                    "config": {
                      "root_id": "stats_outbound",
                      "vm_config": {
                        "vm_id": "stats_outbound",
                        "runtime": "envoy.wasm.runtime.null",
                        "code": {
                          "local": {
                            "inline_string": "envoy.wasm.stats"
                          }
                        }
                      },
                      "configuration": "{\n \"debug\": \"false\",\n \"stat_prefix\": \"istio\"\n}\n"
                    }
                  }
                }
              },
              {
                "name": "envoy.router",
                "typed_config": {
                  "@type": "type.googleapis.com/envoy.config.filter.http.router.v2.Router"
                }
              }
            ],
            "tracing": {
              "client_sampling": {
                "value": 100
              },
              "random_sampling": {
                "value": 100
              },
              "overall_sampling": {
                "value": 100
              }
            },
            "access_log": [
              {
                "name": "envoy.file_access_log",
                "typed_config": {
                  "@type": "type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog",
                  "path": "/dev/stdout",
                  "format": "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \"%DYNAMIC_METADATA(istio.mixer:status)%\" \"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n"
                }
              }
            ],
            "use_remote_address": false,
            "generate_request_id": true,
            "upgrade_configs": [
              {
                "upgrade_type": "websocket"
              }
            ],
            "stream_idle_timeout": "0s",
            "normalize_path": true
          }
        }
      ]
    },
    {
      "filter_chain_match": {
      },
      "filters": [
        {
          "name": "istio.stats",
          "typed_config": {
            "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
            "type_url": "type.googleapis.com/envoy.extensions.filters.network.wasm.v3.Wasm",
            "value": {
              "config": {
                "root_id": "stats_outbound",
                "vm_config": {
                  "vm_id": "tcp_stats_outbound",
                  "runtime": "envoy.wasm.runtime.null",
                  "code": {
                    "local": {
                      "inline_string": "envoy.wasm.stats"
                    }
                  }
                },
                "configuration": "{\n \"debug\": \"false\",\n \"stat_prefix\": \"istio\"\n}\n"
              }
            }
          }
        },
        {
          "name": "envoy.tcp_proxy",
          "typed_config": {
            "@type": "type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy",
            "stat_prefix": "PassthroughCluster",
            "cluster": "PassthroughCluster",
            "access_log": [
              {
                "name": "envoy.file_access_log",
                "typed_config": {
                  "@type": "type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog",
                  "path": "/dev/stdout",
                  "format": "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \"%DYNAMIC_METADATA(istio.mixer:status)%\" \"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n"
                }
              }
            ]
          }
        }
      ],
      "metadata": {
        "filter_metadata": {
          "pilot_meta": {
            "fallthrough": true
          }
        }
      },
      "name": "PassthroughFilterChain"
    }
  ],
  "deprecated_v1": {
    "bind_to_port": false
  },
  "listener_filters": [
    {
      "name": "envoy.listener.tls_inspector",
      "typed_config": {
        "@type": "type.googleapis.com/envoy.config.filter.listener.tls_inspector.v2.TlsInspector"
      }
    },
    {
      "name": "envoy.listener.http_inspector",
      "typed_config": {
        "@type": "type.googleapis.com/envoy.config.filter.listener.http_inspector.v2.HttpInspector"
      }
    }
  ],
  "listener_filters_timeout": "0.100s",
  "traffic_direction": "OUTBOUND",
  "continue_on_listener_filters_timeout": true
}
```

## 匹配 route 80

因为配置较多 我们值展示对应sleep的route config

```
{
  "name": "sleep.foo.svc.cluster.local:80",
  "domains": [
    "sleep.foo.svc.cluster.local",
    "sleep.foo.svc.cluster.local:80",
    "sleep",
    "sleep:80",
    "sleep.foo.svc.cluster",
    "sleep.foo.svc.cluster:80",
    "sleep.foo.svc",
    "sleep.foo.svc:80",
    "sleep.foo",
    "sleep.foo:80",
    "10.97.250.188",
    "10.97.250.188:80"
  ],
  "routes": [
    {
      "match": {
        "prefix": "/"
      },
      "route": {
        "cluster": "outbound|80||sleep.foo.svc.cluster.local",
        "timeout": "0s",
        "retry_policy": {
          "retry_on": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
          "num_retries": 2,
          "retry_host_predicate": [
            {
              "name": "envoy.retry_host_predicates.previous_hosts"
            }
          ],
          "host_selection_retry_max_attempts": "5",
          "retriable_status_codes": [
            503
          ]
        },
        "max_grpc_timeout": "0s"
      },
      "decorator": {
        "operation": "sleep.foo.svc.cluster.local:80/*"
      },
      "name": "default"
    }
  ],
  "include_request_attempt_count": true
}
```

这里我们可以看到最终请求到了outbound|80||sleep.foo.svc.cluster.local 这个cluster,只有一个ep 10.244.1.12:80

```
outbound|80||sleep.foo.svc.cluster.local::default_priority::max_connections::4294967295
outbound|80||sleep.foo.svc.cluster.local::default_priority::max_pending_requests::4294967295
outbound|80||sleep.foo.svc.cluster.local::default_priority::max_requests::4294967295
outbound|80||sleep.foo.svc.cluster.local::default_priority::max_retries::4294967295
outbound|80||sleep.foo.svc.cluster.local::high_priority::max_connections::1024
outbound|80||sleep.foo.svc.cluster.local::high_priority::max_pending_requests::1024
outbound|80||sleep.foo.svc.cluster.local::high_priority::max_requests::1024
outbound|80||sleep.foo.svc.cluster.local::high_priority::max_retries::3
outbound|80||sleep.foo.svc.cluster.local::added_via_api::true
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::cx_active::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::cx_connect_fail::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::cx_total::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::rq_active::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::rq_error::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::rq_success::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::rq_timeout::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::rq_total::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::hostname::
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::health_flags::healthy
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::weight::1
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::region::
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::zone::
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::sub_zone::
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::canary::false
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::priority::0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::success_rate::-1.0
outbound|80||sleep.foo.svc.cluster.local::10.244.1.12:80::local_origin_success_rate::-1.0
```

## sleep接收请求

sleep接收到请求将被iptables重定向到inboud port 15006


为了选择过滤器链，传入连接必须满足其所有条件，连接的属性由网络堆栈和/或侦听器过滤器设置。

以下顺序适用：

- 目的端口。
- 目的IP地址。
- 服务器名称（例如TLS协议的SNI），
- 传输协议。
- 应用协议（例如用于TLS协议的ALPN）。


prefix_ranges如果为非空，则在侦听器绑定到0.0.0.0/::或指定use_original_dst时，指定IP地址和前缀长度以匹配地址。

## 流量到达sleep进行匹配

```
{
  "filter_chain_match": {
    "prefix_ranges": [
      {
        "address_prefix": "10.244.1.12",
        "prefix_len": 32
      }
    ],
    "destination_port": 80
  },
  "filters": [
    {
      "name": "istio.metadata_exchange",
      "typed_config": {
        "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
        "type_url": "type.googleapis.com/envoy.tcp.metadataexchange.config.MetadataExchange",
        "value": {
          "protocol": "istio-peer-exchange"
        }
      }
    },
    {
      "name": "envoy.http_connection_manager",
      "typed_config": {
        "@type": "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
        "stat_prefix": "inbound_10.244.1.12_80",
        "route_config": {
          "name": "inbound|80|http|sleep.foo.svc.cluster.local",
          "virtual_hosts": [
            {
              "name": "inbound|http|80",
              "domains": [
                "*"
              ],
              "routes": [
                {
                  "match": {
                    "prefix": "/"
                  },
                  "route": {
                    "cluster": "inbound|80|http|sleep.foo.svc.cluster.local",
                    "timeout": "0s",
                    "max_grpc_timeout": "0s"
                  },
                  "decorator": {
                    "operation": "sleep.foo.svc.cluster.local:80/*"
                  },
                  "name": "default"
                }
              ]
            }
          ],
          "validate_clusters": false
        },
        "http_filters": [
          {
            "name": "istio.metadata_exchange",
            "typed_config": {
              "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
              "type_url": "type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm",
              "value": {
                "config": {
                  "vm_config": {
                    "runtime": "envoy.wasm.runtime.null",
                    "code": {
                      "local": {
                        "inline_string": "envoy.wasm.metadata_exchange"
                      }
                    }
                  },
                  "configuration": "{}\n"
                }
              }
            }
          },
          {
            "name": "istio_authn",
            "typed_config": {
              "@type": "type.googleapis.com/istio.envoy.config.filter.http.authn.v2alpha1.FilterConfig",
              "policy": {
                "peers": [
                  {
                    "mtls": {
                    }
                  }
                ]
              }
            }
          },
          {
            "name": "envoy.cors",
            "typed_config": {
              "@type": "type.googleapis.com/envoy.config.filter.http.cors.v2.Cors"
            }
          },
          {
            "name": "envoy.fault",
            "typed_config": {
              "@type": "type.googleapis.com/envoy.config.filter.http.fault.v2.HTTPFault"
            }
          },
          {
            "name": "istio.stats",
            "typed_config": {
              "@type": "type.googleapis.com/udpa.type.v1.TypedStruct",
              "type_url": "type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm",
              "value": {
                "config": {
                  "root_id": "stats_inbound",
                  "vm_config": {
                    "vm_id": "stats_inbound",
                    "runtime": "envoy.wasm.runtime.null",
                    "code": {
                      "local": {
                        "inline_string": "envoy.wasm.stats"
                      }
                    }
                  },
                  "configuration": "{\n \"debug\": \"false\",\n \"stat_prefix\": \"istio\"\n}\n"
                }
              }
            }
          },
          {
            "name": "envoy.router",
            "typed_config": {
              "@type": "type.googleapis.com/envoy.config.filter.http.router.v2.Router"
            }
          }
        ],
        "tracing": {
          "client_sampling": {
            "value": 100
          },
          "random_sampling": {
            "value": 100
          },
          "overall_sampling": {
            "value": 100
          }
        },
        "server_name": "istio-envoy",
        "access_log": [
          {
            "name": "envoy.file_access_log",
            "typed_config": {
              "@type": "type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog",
              "path": "/dev/stdout",
              "format": "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% \"%DYNAMIC_METADATA(istio.mixer:status)%\" \"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n"
            }
          }
        ],
        "use_remote_address": false,
        "generate_request_id": true,
        "forward_client_cert_details": "APPEND_FORWARD",
        "set_current_client_cert_details": {
          "subject": true,
          "dns": true,
          "uri": true
        },
        "upgrade_configs": [
          {
            "upgrade_type": "websocket"
          }
        ],
        "stream_idle_timeout": "0s",
        "normalize_path": true
      }
    }
  ],
  "transport_socket": {
    "name": "envoy.transport_sockets.tls",
    "typed_config": {
      "@type": "type.googleapis.com/envoy.api.v2.auth.DownstreamTlsContext",
      "common_tls_context": {
        "alpn_protocols": [
          "h2",
          "http/1.1"
        ],
        "tls_certificate_sds_secret_configs": [
          {
            "name": "default",
            "sds_config": {
              "api_config_source": {
                "api_type": "GRPC",
                "grpc_services": [
                  {
                    "envoy_grpc": {
                      "cluster_name": "sds-grpc"
                    }
                  }
                ]
              }
            }
          }
        ],
        "combined_validation_context": {
          "default_validation_context": {
          },
          "validation_context_sds_secret_config": {
            "name": "ROOTCA",
            "sds_config": {
              "api_config_source": {
                "api_type": "GRPC",
                "grpc_services": [
                  {
                    "envoy_grpc": {
                      "cluster_name": "sds-grpc"
                    }
                  }
                ]
              }
            }
          }
        }
      },
      "require_client_certificate": true
    }
  },
  "name": "10.244.1.12_80"
}
```


## 流量到达sleep的 inbound|80|http|sleep.foo.svc.cluster.local cluster

```
{
  "version_info": "2020-07-17T08:45:26Z/18",
  "cluster": {
    "@type": "type.googleapis.com/envoy.api.v2.Cluster",
    "name": "inbound|80|http|sleep.foo.svc.cluster.local",
    "type": "STATIC",
    "connect_timeout": "10s",
    "circuit_breakers": {
      "thresholds": [
        {
          "max_connections": 4294967295,
          "max_pending_requests": 4294967295,
          "max_requests": 4294967295,
          "max_retries": 4294967295
        }
      ]
    },
    "load_assignment": {
      "cluster_name": "inbound|80|http|sleep.foo.svc.cluster.local",
      "endpoints": [
        {
          "lb_endpoints": [
            {
              "endpoint": {
                "address": {
                  "socket_address": {
                    "address": "127.0.0.1",
                    "port_value": 80
                  }
                }
              }
            }
          ]
        }
      ]
    }
  },
  "last_updated": "2020-07-17T08:45:29.404Z"
}
```

## InboundPassthroughClusterIpv4

```
这是一个最后生效的cluster，如果入流量，没有匹配到规则，即我们访问了一个没有暴露到svc,的端口，则透传到服务容器
```

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
