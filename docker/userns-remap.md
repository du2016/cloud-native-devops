默认docker没有为容器启用user namespace,docker中的root就是系统的root


只不过通过PID和NS做了隔离，想要改变映射需要以下配置



设置运行用户和ID范围
echo "dockeruser:165536:65536" > /etc/subuid
echo "dockeruser:165536:65536" > /etc/subgid

/etc/docker/daemon.json
{
  "userns-remap": "dockeruser"
}


默认情况下

readlink /proc/$$/ns/user会发现在docker中以root和宿主机root效果一样