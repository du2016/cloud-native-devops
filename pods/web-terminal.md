基于gin+sockjs实现k8s pod web terminal


# 路由定义

由于sockjs会动态生成路由参数用来记录回话id,所以这里需要使用参数路由，beego的也是一样，很早之前写过一个beego的，在github上可以找到

```
r.Any("/pod/exec/*path", cluster.ContainerTerminal)
```

# 定义一个sockjsterminal对象

```

# 这里写的比较粗糙，实际上可以定义通信消息对象，规范化字段
func (self TerminalSockjs) Read(p []byte) (int, error) {
	var reply string
	var msg map[string]uint16
	reply, err := self.Conn.Recv()
	if err != nil {
		return 0, err
	}
	if err := json.Unmarshal([]byte(reply), &msg); err != nil {
		return copy(p, reply), nil
	} else {
		self.SizeChan <- &remotecommand.TerminalSize{
			Width:  msg["cols"],
			Height: msg["rows"],
		}
		return 0, nil
	}
}

func (self TerminalSockjs) Write(p []byte) (int, error) {
	err := self.Conn.Send(string(p))
	return len(p), err
}

# resize
func (self *TerminalSockjs) Next() *remotecommand.TerminalSize {
	size := <-self.SizeChan
	log.Printf("terminal size to width: %d height: %d", size.Width, size.Height)
	return size
}

type TerminalSockjs struct {
	Conn      sockjs.Session
	SizeChan  chan *remotecommand.TerminalSize
	Cluster   uint
	Namespace string
	Pod       string
	Container string
}
```

实现了读写方法，用来处理k8s exec接口的读写

# sockjs handler定义

```
func Handler(t *TerminalSockjs, cmd []string) error {
	client, err := kubeconn.GetClientset(t.Cluster)
	if err != nil {
		log.Println(err)
		return err
	}
	clientConfig, err := kubeconn.GetClientconfig(t.Cluster)
	if err != nil {
		log.Println(err)
		return err
	}
	restclient := client.CoreV1().RESTClient()
	fn := func() error {
		req := restclient.Post().
			Resource("pods").
			Name(t.Pod).
			Namespace(t.Namespace).
			SubResource("exec")
		req.VersionedParams(
			&v1.PodExecOptions{
				Container: t.Container,
				Command:   cmd,
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
			},
			scheme.ParameterCodec,
		)
		executor, err := remotecommand.NewSPDYExecutor(
			clientConfig, http.MethodPost, req.URL(),
		)
		if err != nil {
			return err
		}
		return executor.Stream(remotecommand.StreamOptions{
			Stdin:             t,
			Stdout:            t,
			Stderr:            t,
			Tty:               true,
			TerminalSizeQueue: t,
		})
	}
	return fn()
}
```

具体的handler
```
	Sockjshandler := func(session sockjs.Session) {
		log.Println("ContainerTerminal2")
		t := &term.TerminalSockjs{
			Conn:      session,
			SizeChan:  make(chan *remotecommand.TerminalSize),
			Cluster:   uint(self.MustGet("cid").(int64)),
			Namespace: namespace,
			Pod:       pod,
			Container: container,
		}
		if casbin.Enforcer.HasRoleForUser(self.MustGet("user").(db.User).Email, "admin") == true {
			if err := term.Handler(t, []string{"/bin/bash"}); err != nil {
				err := term.Handler(t, []string{"/bin/sh"})
				log.Println(t.Conn.Close(200, "client close"), err)
			}
		} else {
			if err := term.Handler(t, []string{"/bin/bash", "-c", "echo 'dev ALL=(root) NOPASSWD:/usr/local/bin/jstack,/usr/local/bin/jmap,/usr/local/bin/jstat'> /etc/sudoers;useradd dev;su dev"}); err != nil {
				log.Println(err)
				err := term.Handler(t, []string{"/bin/sh", "-c", "echo 'dev ALL=(root) NOPASSWD:/usr/local/bin/jstack,/usr/local/bin/jmap,/usr/local/bin/jstat'> /etc/sudoers;useradd dev;su dev"})
				log.Println(t.Conn.Close(200, "client close"), err)
			}
		}
	}
```

这里为了探测使用的shell所以定义了一个重试,另外结合平台，可以实现不同用户进入容器的用户，以及可以执行sudo的命令，示例为dev用户可执行sudo命令的命令列表，实际使用可以存入数据库动态生成sudo文件，实现细粒度的控制。

# sockjs handler

```
sockjs.NewHandler("/api/cluster/pod/exec", sockjs.Options{
		Websocket:       true,
		JSessionID:      nil,
		SockJSURL:       "https://cdn.bootcss.com/sockjs-client/1.3.0/sockjs.min.js",
		HeartbeatDelay:  25 * time.Second,
		DisconnectDelay: 5 * time.Second,
		ResponseLimit:   128 * 1024,
	}, Sockjshandler).ServeHTTP(self.Writer, self.Request)
```


前端结合xterm.js就可以实现webterminal，具体前端比较简单，可以看xterm.js的官方文档
