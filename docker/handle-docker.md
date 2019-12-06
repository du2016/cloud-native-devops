# 手动挂载镜像


- 导出镜像为tar包

```
docker save -o busybox.tar busybox
```

- 解压镜像

```
tar xf busybox.tar
```

- 查看文件

```
ls 
020584afccce44678ec82676db80f68d50ea5c766b6e9d9601f7b5fc86dfb96d.json  busybox.tar
2ec9d8fd000cf6929df9aa7a58b6d872f5588e097d18d91cadc2a08dd563d590       manifest.json
79cf3ae80ddd8a3f5e5f09849fa9b8b35ade13b5c9571b534d3651c6d50668e2       repositories
b534869c81f05ce6fbbdd3a3293e64fd032e059ab4b28a0e0d5b485cf904be4b.json
```

- 解压layer

```
mkdir layer1 layer2
tar xf 2ec9d8fd000cf6929df9aa7a58b6d872f5588e097d18d91cadc2a08dd563d590/layer.tar -C layer1
tar xf 79cf3ae80ddd8a3f5e5f09849fa9b8b35ade13b5c9571b534d3651c6d50668e2/layer.tar -C layer2
```

- 创建文件系统

选用overlay fs 

```
mkdir ./rootfs/{merged,diff,work} -p

mount -t overlay overlay -o lowerdir=./layer1:./layer2,upperdir=./rootfs/diff,workdir=./rootfs/work ./rootfs/merged

overlayfs有以下概念：
merged: 挂载点
diff 是upper
work 是work
```

# 基于golang创建命名空间挂载rootfs

```
package main

import (
    "flag"
    "fmt"
    "os"
    "os/exec"
    "syscall"
    "path/filepath"

    "github.com/docker/docker/pkg/reexec"
)

func init() {
    // 使用reexec注册函数，reexec提供了执行自身的方法，从而对默认的初始命名空间信息进行修改
    reexec.Register("docker-shim", nsInitialisation)
    if reexec.Init() {
        os.Exit(0)
    }
}

func nsInitialisation() {
    newrootPath := os.Args[1]

    // 挂载proc
    if err := mountProc(newrootPath); err != nil {
        fmt.Printf("Error mounting /proc - %s\n", err)
        os.Exit(1)
    }
    
    // 执行pivot_root
    if err := pivotRoot(newrootPath); err != nil {
        fmt.Printf("Error running pivot_root - %s\n", err)
        os.Exit(1)
    }
    
    // 配置hostname
    if err := syscall.Sethostname([]byte("my-docker")); err != nil {
        fmt.Printf("Error setting hostname - %s\n", err)
        os.Exit(1)
    }
    nsRun()
}

func nsRun() {
    cmd := exec.Command("/bin/sh")

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        fmt.Printf("Error running the /bin/sh command - %s\n", err)
        os.Exit(1)
    }
}

func main() {
    var rootfsPath string
    flag.StringVar(&rootfsPath, "rootfs", "./rootfs/merged", "Path to the root filesystem to use")
    flag.Parse()

    cmd := reexec.Command("docker-shim", rootfsPath)

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 创建命名空间
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWNS |
            syscall.CLONE_NEWUTS |
            syscall.CLONE_NEWIPC |
            syscall.CLONE_NEWPID |
            syscall.CLONE_NEWNET |
            syscall.CLONE_NEWUSER,
        // UID映射
        UidMappings: []syscall.SysProcIDMap{
            {
                ContainerID: s.Getuid(),0,
                HostID:      o
                Size:        1,
            },
        },
        // GID映射
        GidMappings: []syscall.SysProcIDMap{
            {
                ContainerID: 0,
                HostID:      os.Getgid(),
                Size:        1,
            },
        },
    }

    if err := cmd.Run(); err != nil {
        fmt.Printf("Error running the reexec.Command - %s\n", err)
        os.Exit(1)
    }
}

// 相当于 pivot_root . .pivot_root
func pivotRoot(newroot string) error {
    putold := filepath.Join(newroot, "/.pivot_root")

    // 挂载newroot到自身，目的是为了确保newroot和putold处于不同目录，
    // MS_BIND：执行bind挂载，使文件或者子目录树在文件系统内的另一个点上可视。
    // MS_REC： 创建递归绑定挂载，递归更改传播类型
    if err := syscall.Mount(newroot, newroot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
        return err
    }

    // 创建put_old目录
    if err := os.MkdirAll(putold, 0700); err != nil {
        return err
    }

    // 调用 pivot_root
    if err := syscall.PivotRoot(newroot, putold); err != nil {
        return err
    }

    // 切换到根目录，不然的话为pivot_root后的当前目录
    if err := os.Chdir("/"); err != nil {
        return err
    }

    // 卸载pivot_root
    putold = "/.pivot_root"
    if err := syscall.Unmount(putold, syscall.MNT_DETACH); err != nil {
        return err
    }

    // 删除put_old目录
    if err := os.RemoveAll(putold); err != nil {
        return err
    }

    return nil
}

// 挂载
func mountProc(newroot string) error {
    source := "proc"
    target := filepath.Join(newroot, "/proc")
    fstype := "proc"
    flags := 0
    data := ""
 
    // 创建Proc目录
    os.MkdirAll(target, 0755)
    // 挂载proc
    // mount -t proc proc ./proc
    if err := syscall.Mount(source, target, fstype, uintptr(flags), data); err != nil {
        return err
    }

    return nil
}
```

```
go build main.go -o docker
./docker
```

# 创建网络


为进程命名空间添加虚拟设备

```
# 查找ID
ps -ef | grep docker-shim
root     32133 32128  0 13:07 pts/1    00:00:00 docker-shim ./rootfs/merged
root     32149 14085  0 13:08 pts/0    00:00:00 grep --color=auto docker-shim

# 查看PID命名空间
nsenter -t 32133 -n ip link show
ip link add veth0 type veth peer name veth1
ip link set veth0 netns 32133

查看命名空间内网卡信息
nsenter -t 32133 -n ip link show
```


在宿主机上执行为veth1配置IP
```
ip addr add 100.100.1.1/24 dev veth1
ip link set veth1 up
```

在docker内执行为容器配置IP，测试网络


```
ip link set veth0 name  eth0
ip addr add 100.100.1.2/24 dev eth0
ip link set eth0 up
ip link set lo up
ip route add default via 100.100.1.1
ping 100.100.1.2
```

# 为容器配置NAT规则

```
iptables -t nat -A POSTROUTING -s 100.100.1.0/255.255.255.0 -o eth0 -j MASQUERADE
iptables -A FORWARD -i eth0 -o veth1 -j ACCEPT
iptables -A FORWARD -o eth0 -i veth1 -j ACCEPT
```


# 验证

```
echo "nameserver 8.8.8.8" > /etc/resolv.conf
ping baidu.com
```
此时容器已经可以正常上网了

# 总结

本文通过golang创建命名空间，挂载文件系统，通过pivot_root切换文件系统root，通过linux命令创建虚拟设备对并配置网络，
通过iptables实现网络的NAT，从而快速实现一个简单的容器。