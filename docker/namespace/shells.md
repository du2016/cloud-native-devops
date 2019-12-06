# nsenter

nsenter 用于在指定进程的命名空间内执行命令

```
 -t, --target <pid>     要获取名字空间的目标进程
 -m, --mount[=<file>]   enter mount namespace
 -u, --uts[=<file>]     enter UTS namespace (hostname etc)
 -i, --ipc[=<file>]     enter System V IPC namespace
 -n, --net[=<file>]     enter network namespace
 -p, --pid[=<file>]     enter pid namespace
 -U, --user[=<file>]    enter user namespace
 -S, --setuid <uid>     set uid in entered namespace
 -G, --setgid <gid>     set gid in entered namespace
     --preserve-credentials do not touch uids or gids
 -r, --root[=<dir>]     set the root directory
 -w, --wd[=<dir>]       set the working directory
 -F, --no-fork          执行 <程序> 前不 fork
 -Z, --follow-context   set SELinux context according to --target PID
```

# ip netns exec

通过指定网络命名空间名称执行命令


