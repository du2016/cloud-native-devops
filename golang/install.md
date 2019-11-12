# 下载安装包

https://golang.google.cn/dl/

# 下载

```
wget https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz
tar xf go1.13.4.linux-amd64.tar.gz
mv go /usr/local/
```

# 配置环境

设置gopath 为/go
gobin则为/go/bin

将以下内容添加到/etc/profile
```
export GOPATH=/go
export PATH=/usr/local/go/bin:$GOPATH/bin
```

执行
```
source /etc/profile
```