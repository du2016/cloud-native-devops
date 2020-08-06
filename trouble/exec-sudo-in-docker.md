# 背景

使用的是公有云，最近要对k8s版本进行升级，在升级之后发发现从我们的web terminal 进入到容器，拥有sudo权限的用户无法进行sudo命令,即使使用root通过docker exec 进入到容器，依旧无法sudo

```
sudo: pam_open_session: Permission denied
sudo: policy plugin failed session initialization
```

# 定位

进入到容器中我们查看ulimit -a 如下

```
core file size          (blocks, -c) 5242880
data seg size           (kbytes, -d) unlimited
scheduling priority             (-e) 0
file size               (blocks, -f) unlimited
pending signals                 (-i) 3806
max locked memory       (kbytes, -l) 82000
max memory size         (kbytes, -m) unlimited
open files                      (-n) 1048576
pipe size            (512 bytes, -p) 8
POSIX message queues     (bytes, -q) 819200
real-time priority              (-r) 0
stack size              (kbytes, -s) 8192
cpu time               (seconds, -t) unlimited
max user processes              (-u) unlimited
virtual memory          (kbytes, -v) unlimited
file locks                      (-x) unlimited
```

我们基础镜像里面的/etc/security/limits.conf配置如下

```
* soft core unlimited
* hard core unlimited
* soft nofile 1048576
* hard nofile 1048576
root soft nofile 1048576
root hard nofile 1048576
* soft nproc 102400
* hard nproc 102400
```

可见我们在/etc/security/limits.conf配置文件中的配置并未生效，
查看psp，公有云也未做psp相关的初始配置，通过docker inspect查看，并没有相关ulimit设置，
最终查看systemd发现docker.service配置多了一行

```
LimitCORE=5368709120
```

这里的值是单位是字节 和 内部ulimit看到的有所差异，ulimit看到的是block数

systemd 中有关limit的配置对照表如下

指令 | 等价的ulimit 命令 | 单位
--- | --------------- | ---
LimitCPU | ulimit -t | 秒
LimitFSIZE= | ulimit -f | 字节
LimitDATA=  | ulimit -d | 字节
LimitSTACK=	 | ulimit -s | 字节
LimitCORE=	 | ulimit -c | 字节
LimitRSS=	 | ulimit -m | 字节
LimitNOFILE=  | ulimit -n | 文件描述符的数量
LimitAS= | ulimit -v | 字节
LimitNPROC= | ulimit -u | 进程的数量
LimitMEMLOCK= | ulimit -l | 字节
LimitLOCKS= | ulimit -x | 锁的数量
LimitSIGPENDING= | ulimit -i | 信号队列的长度(排队的信号数量)
LimitMSGQUEUE= | ulimit -q | 字节
LimitNICE= | ulimit -e | 谦让度
LimitRTPRIO= | ulimit -r | 实时优先级
LimitRTTIME= | 不存在 | 微秒

由此可见最终生效的是systemd下core file size 的值，那为什么会出现上面的报错呢？


# pam_limits.sod

查看linux sudo pam 配置如下

```
#%PAM-1.0
auth       include      system-auth
account    include      system-auth
password   include      system-auth
session    optional     pam_keyinit.so revoke
session    required     pam_limits.so
```

可以看到sudo加载了pam_limits.so模块，而limits.conf 文件实际是 Linux PAM（插入式认证模块，Pluggable Authentication Modules）中 pam_limits.so 的配置文件

有关pam类型如下

![](http://img.rocdu.top/20200528/pam-type.png)

由此可知当我们执行sudo时触发了pam_limits.so模块的某些限制，导致执行失败,
实际上pam_limits.so的实现主要包括以下步骤：

- 解析配置文件 /etc/security/limits.conf及/etc/security/limits.d下的*.conf文件
- setup_limits调用setrlimits生效配置


parm_limits 的[说明文档](http://www.linux-pam.org/Linux-PAM-html/sag-pam_limits.html)

# setrlimit和getrlimit系统调用

pam_limits.so进行了setrlimit和getrlimit系统调用,setrlimit和getrlimit的定义如下

```
int getrlimit(int resource, struct rlimit *rlim);
int setrlimit(int resource, const struct rlimit *rlim);
```

在linux系统中，Resouce limit指在一个进程的执行过程中，它所能得到的资源的限制，比如进程的core file的最大值，虚拟内存的最大值等。
Resouce limit的大小可以直接影响进程的执行状况。其有两个最重要的概念：soft limit 和 hard limit。

soft limit和hard limit概念如下：

- soft limit是指内核所能支持的资源上限。对于RLIMIT_COREsoft limit最大能是unlimited。

- hard limit在资源中只是作为soft limit的上限，当你设置hard limit后，你以后设置的soft limit只能小于hard limit。

```
struct rlimit {
　　rlim_t rlim_cur;　　//soft limit
　　rlim_t rlim_max;　　//hard limit
};
```

根据man文档在进行setrlimit系统调用时操作系统会检查新的值是否超过当前hard limit,对于root没有这种限制

返回错误码如下
EFAULT：rlim指针指向的空间不可访问
EINVAL：参数无效
EPERM：增加资源限制值时，权能不允许

EPERM对应的返回为：Operation not permitted 
这和我们手动执行ulimit的返回一致

setrlimit [man文档](https://linux.die.net/man/2/setrlimit)

# 容器内的root如何突破限制

在docker没有添加 CAP_SYS_RESOURCE 时，才可以突破内核上限，所以docker内部的root并不是真正的root

```
docker run --cap-add CAP_SYS_RESOURCE 
# 或者
docker run --privileged
```

这样容器内部的root用户就可以突破该ulimit限制


# 容器ulimit设置原则

- dockerd和容器都设置为unlimited
- 容器设置的limit比宿主机小
- 为容器添加CAP_SYS_RESOURCE capability
- 容器开启privileged(不推荐)
- 不设置，以dockerd为准


扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)