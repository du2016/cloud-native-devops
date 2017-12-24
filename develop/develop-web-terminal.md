

- 首先我们需要获取一个restclient restclient的具体概念请参照我之前的文章
- 通过restclient提交信息指定我们想要执行命令的container，实际上就是一个HTTP request
- 设置request的参数。
- NewExecutor函数将一个普通的request升级为流输出,使用了Google的spdy协议，在返回时使用了chunk头保证长连接
- 发起请求，进行交互
- 在实际使用为了和实际使用界面的大小相匹配，需要监听窗口大小变化，每当变化时由queue的next方法返回。
- 如果为了实现一个简单的命令将输入输出设置为系统的输入输出即可，如果需要实现web terminal，可以使用xterm.js 与ws，将输入

```
req := restclient.Post().  
        Resource("pods").  
        Name(podname).  
        Namespace(namespace).  
        SubResource("exec").  
        Param("container", container).  
        Param("stdin", "true").  
        Param("stdout", "true").  
        Param("stderr", "true").  
        Param("command", "/bin/sh").Param("tty", "true")  
  
    req.VersionedParams(  
        &api.PodExecOptions{  
            Container: container,  
            Command:   []string{"sh"},  
            Stdin:     true,  
            Stdout:    true,  
            Stderr:    true,  
            TTY:       true,  
        },  
        api.ParameterCodec,  
    )  
    executor, err := remotecommand.NewExecutor(  
        config, http.MethodPost, req.URL(),  
    )  
      
    err = executor.Stream(remotecommand.StreamOptions{  
        SupportedProtocols: remotecommandconsts.SupportedStreamingProtocols,  
        Stdin:              r,  
        Stdout:             w,  
        Stderr:             os.Stderr,  
        Tty:                true,  
        TerminalSizeQueue:  t,  
    })  
```