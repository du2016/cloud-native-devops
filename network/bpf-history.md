# 名词解释

## BPF

全称是 Berkeley Packet，用于过滤(filter)网络报文(packet)的架构，主要有两种方式

- 过滤(Filter): 根据外界输入的规则过滤报文；
- 复制(Copy)：将符合条件的报文由内核空间复制到用户空间；

场景
- 性能调优
- 内核监控
-流量控制

## eBPF

起源于bpf,称之为cbpf,提供了内核的数据包过滤机制，3.15内核开始支持，3.17进行规范化，放置在内核kernel/bpf下
代码长度上限 4kb

## Bcc

ebpf的编译工具集合,前端提供python/lua调用，本身通过c语言实现，集成llvm/clang，将ebpf代码注入，提供一些更人性化的函数给用户使用，比如函数的注入等，里面提供了很多xdp的例子

## XDP

意思是eXpress Data Path，它能够在网络包进入用户态直接对网络包进行过滤或者处理。
XDP依赖eBPF技术。

## LLVM

是构架编译器(compiler)的框架系统，以C++编写而成，用于优化以任意程序语言编写的程序的编译时间(compile-time)、链接时间(link-time)、运行时间(run-time)以及空闲时间(idle-time)，对开发者保持开放，并兼容已有脚本。


## CFG(Computation Flow Graph)

将过滤器构筑于一套基于 if-else的控制流(flow graph)之上,BPF 采用该报文过滤设计

## BPF JIT

3.0内核对BPF进行优化提速,打开方式
echo 1 > /proc/sys/net/core/bpf_jit_enable

## ebpf map

在用户控件建立map用于和bpf程序交互，内核控件亦可以访问，
一次通信4字节


# bcc tools

![bcc tools](https://github.com/iovisor/bcc/raw/master/images/bcc_tracing_tools_2019.png)

# 查看已加载BPF项目

