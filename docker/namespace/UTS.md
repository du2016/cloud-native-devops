
使用unshare创建新的UTS并运行

```
hostname
unshare -u /bin/sh
hostname test
hostname
exit
hostname
```

查看当前进程所属的UTS号
readlink /proc/self/ns/uts 


centos设置user namespace数量
echo 12345 > /proc/sys/user/max_user_namespaces
