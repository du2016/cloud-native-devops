# 介绍
Docker 是一个开源的应用容器引擎，让开发者可以打包他们的应用以及依赖包到一个可移植的镜像中，
然后发布到任何流行的 Linux或Windows 机器上，也可以实现虚拟化。容器是完全使用沙箱机制，相互之间不会有任何接口

# 特点

- 灵活：即使最复杂的应用程序也可以容器化。
- 轻量级：容器利用并共享主机内核，在系统资源方面比虚拟机更有效。
- 可移植：您可以在本地构建，部署到云并在任何地方运行。
- 松散耦合：容器是高度自给自足并封装的容器，使您可以在不破坏其他容器的情况下更换或升级它们。
- 可扩展：您可以在数据中心内增加并自动分发容器副本。
- 安全：容器将积极的约束和隔离应用于流程，而无需用户方面的任何配置。

# 版本

docker分为社区版本 docker-ce和商业版本docker-ee

# 概念

- container 

容器是图像的可运行实例。您可以使用Docker API或CLI创建，启动，停止，移动或删除容器。
您可以将容器连接到一个或多个网络，将存储连接到它，甚至根据其当前状态创建新映像。

默认情况下，容器与其他容器及其主机之间的隔离程度相对较高。您可以控制容器的网络，
存储或其他基础子系统与其他容器或与主机的隔离程度。

容器由其映像以及在创建或启动时为其提供的任何配置选项定义。
删除容器后，未存储在持久性存储中的状态更改将消失。


- image

镜像是用于创建docker容器指令的只读模板。
通常，一个image基于另一个image，并进行一些其他自定义。例如，您可以基于该ubuntu image构建image，

- docker registry
Docker registry存储Docker映像。Docker Hub是任何人都可以使用的公共registry，
并且Docker配置为默认在Docker Hub上查找映像。您甚至可以运行自己的私人registry。
如果使用Docker数据中心（DDC），则其中包括Docker可信注册表（DTR）。

使用docker pull或docker run命令时，所需的镜像将从配置的registry中提取。使用该docker push命令时，
会将映像推送到配置的registry。

- network 网络
- volume 数据卷
- docker cli 

Docker客户端（docker）是许多Docker用户与Docker交互的主要方式。
当您使用诸如之类的命令时docker run，客户端会将这些命令发送到dockerd，以执行这些命令。
该docker命令使用Docker API。Docker客户端可以与多个守护程序通信。

- api

通过api实现daemon操作的rest接口

- docker daemon
Docker守护程序（dockerd）侦听Docker API请求并管理Docker对象，例如镜像，容器，网络和卷。
守护程序还可以与其他守护程序通信以管理Docker服务


https://docs.docker.com/engine/images/engine-components-flow.png

- service

服务实现了跨宿主机实现服务扩容的能力，需要和多个worker的swarm集群一起工作，

# 容器与镜像的关系

从根本上讲，一个容器不过是一个正在运行的进程，并对其应用了一些附加的封装功能，
以使其与主机和其他容器隔离。容器隔离的最重要方面之一是每个容器都与自己的私有文件系统进行交互。
该文件系统由Docker 镜像提供。
镜像包括运行应用程序所需的所有内容-代码或二进制文件，运行时，依赖关系以及所需的任何其他文件系统对象。

# docker用到的底层原理

- namespace
- cgroup
- union fs

# docker和虚拟机的简单对比

容器在Linux上本地运行，并与其他容器共享主机的内核。它运行一个离散进程，
不占用任何其他可执行文件更多的内存，从而使其轻巧。

相比之下，虚拟机（VM）运行具有虚拟机管理程序对主机资源的虚拟访问权的成熟`guest`操作系统。通常，
VM会产生大量开销，超出了应用程序逻辑所消耗的开销。


https://docs.docker.com/images/Container%402x.png
https://docs.docker.com/images/VM%402x.png

