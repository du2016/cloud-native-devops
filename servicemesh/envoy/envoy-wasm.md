# 使用wasm扩展envoy

# envoy wasm 介绍
WebAssembly是一种沙盒技术，可用于扩展Istio代理（Envoy）。Proxy-Wasm沙箱API取代了Mixer作为Istio中的主要扩展机制。

WebAssembly沙箱目标：

- 效率 -扩展增加了低延迟，CPU和内存开销。
- 功能 -扩展可以执行策略，收集遥测和执行有效载荷突变。
- 隔离 -一个插件中的编程错误或崩溃确实会影响其他插件。
- 配置 -使用与其他Istio API一致的API配置插件。扩展名可以动态配置。
- Operator -可以扩展扩展并将其部署为仅日志，失败打开或失败关闭。
- 扩展开发人员 -该插件可以用几种编程语言编写。

istio社区基于官方envoy的基础上fork 了`https://github.com/istio/envoy`,在wasm分支以实现istio wasm支持，当前官方envoy暂未支持wasm

# 架构

- 筛选器服务提供程序接口（SPI），用于为筛选器构建Proxy-Wasm插件。
- Envoy中嵌入了Sandbox V8 Wasm Runtime。
- headers, trailers 和 metadata的host api。
- 调出用于gRPC和HTTP调用的API。
- Stats和Logging API，用于度量和监视

![扩展envoy](http://img.rocdu.top/20200615/extending.png)

# 通过js生成wasm实现envoy header的修改

## 代码实现
使用 solo.io提供的[proxy-runtime](https://github.com/solo-io/proxy-runtime)
通过js来实现wasm逻辑

```
git clone https://github.com/solo-io/proxy-runtime
mkdir wasm-addheader
npm install --save-dev assemblyscript
npx asinit .
```

修改package.json

```
"asbuild:untouched": "asc assembly/index.ts -b build/untouched.wasm -t build/untouched.wat --validate --use abort=abort_proc_exit --sourceMap --debug",
"asbuild:optimized": "asc assembly/index.ts -b build/optimized.wasm -t build/optimized.wat --validate --use abort=abort_proc_exit --sourceMap --optimize",
```

修改依赖为 proxy-runtime的本地路径

```
"@solo-io/proxy-runtime": "file:/root/proxy-runtime"
```

该示例判断配置的header value是否存在，不存在设置则添加 `hello: world`header,
设置则为  `hello: configvalue` 
通过逻辑实现：

```
export * from "@solo-io/proxy-runtime/proxy"; // this exports the required functions for the proxy to interact with us.
import { RootContext, Context, RootContextHelper, ContextHelper, registerRootContext, FilterHeadersStatusValues, stream_context } from "@solo-io/proxy-runtime";

class AddHeaderRoot extends RootContext {
  configuration : string;

  createContext(context_id: u32): Context {
    return ContextHelper.wrap(new AddHeader(context_id, this));
  }
}

class AddHeader extends Context {
  root_context : AddHeaderRoot;
  constructor(context_id: u32, root_context:AddHeaderRoot){
    super(context_id, root_context);
    this.root_context = root_context;
  }
  onResponseHeaders(a: u32): FilterHeadersStatusValues {
    const root_context = this.root_context;
    if (root_context.getConfiguration() == "") {
      stream_context.headers.response.add("hello", "world!");
    } else {
      stream_context.headers.response.add("hello", root_context.getConfiguration());
    }
    return FilterHeadersStatusValues.Continue;
  }
}

registerRootContext((context_id: u32) => { return RootContextHelper.wrap(new AddHeaderRoot(context_id)); }, "add_header");
```

## 编译

```
npm run asbuild
```

# 配置envoy

我们配置了一个返回admin接口的listener，在filter中添加了我们的 add_header wasm plugin
```
admin:
  access_log_path: /tmp/admin_access.log
  address:
    socket_address: { address: 127.0.0.1, port_value: 9901 }

static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address: { address: 0.0.0.0, port_value: 10000 }
    filter_chains:
    - filters:
      - name: envoy.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
          stat_prefix: ingress_http
          use_remote_address: true
          codec_type: AUTO
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match: { prefix: "/" }
                route: { cluster: some_service }
          http_filters:
          - name: envoy.filters.http.wasm
            config:
              config:
                name: "add_header"
                root_id: "add_header"
                configuration: "what ever you want"
                vm_config:
                  vm_id: "my_vm_id"
                  runtime: "envoy.wasm.runtime.v8"
                  code:
                    local:
                      filename: /media/optimized.wasm
                  allow_precompiled: false
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
                port_value: 9901
```


# 测试访问

我们将编译好的文件通过volume挂在进容器

```
docker run -it --net=host --entrypoint=bash -v /media:/media istio/proxyv2:latest
envoy -c /media/t.yaml
```

通过curl测试访问：

```
curl http://172.16.233.129:10000/ -I
HTTP/1.1 200 OK
content-type: text/html; charset=UTF-8
cache-control: no-cache, max-age=0
x-content-type-options: nosniff
date: Fri, 05 Jun 2020 16:00:57 GMT
server: envoy
x-envoy-upstream-service-time: 0
hello: what ever you want
transfer-encoding: chunked
```

可以看到返回的header已经包含 `hello: what ever you want`

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)