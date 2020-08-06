# 依赖

- lua 5.1版本
- envoy基于exported symbols编译

# 支持的功能

> 随着生产中使用过滤器，预计该列表会随着时间的推移而扩展。API的表面故意保持较小。目的是使脚本极其简单和安全地编写。假定非常复杂或高性能的用例使用本机C ++筛选器API。

- 在请求流和/或响应流中流式传输时检查标头，正文和尾部。

- 标题和尾部的修改。

- 阻塞并缓冲整个请求/响应正文以进行检查。

- 对上游主机执行出站异步HTTP调用。可以在缓冲正文数据的同时执行此类调用，以便在调用完成时可以修改上游头。

- 执行直接响应并跳过进一步的过滤器迭代。例如，脚本可以进行上游HTTP身份验证调用，然后直接以403响应代码进行响应。


# 示例

```
function envoy_on_request(request_handle)
  -- Wait for the entire request body and add a request header with the body size.
  request_handle:headers():add("request_body_size", request_handle:body():length())
end

-- Called on the response path.
function envoy_on_response(response_handle)
  -- Wait for the entire response body and a response header with the body size.
  response_handle:headers():add("response_body_size", response_handle:body():length())
  -- Remove a response header named 'foo'
  response_handle:headers():remove("foo")
end
```

# 流句柄API
当Envoy在配置中加载脚本时，它将查找该脚本定义的两个全局函数：

```
function envoy_on_request(request_handle)
end

function envoy_on_response(response_handle)
end
```

脚本可以定义这两个功能之一或全部。在请求路径中，Envoy将作为协程运行envoy_on_request，并将句柄传递给请求API。
在响应路径中，Envoy将作为协程运行envoy_on_response，将句柄传递给响应API。