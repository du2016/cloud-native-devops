Docker通过从一个Dockerfile文本文件中读取指令来自动构建映像，
该文本文件按顺序包含构建给定映像所需的所有命令。
Dockerfile遵循特定的格式和指令集，您可以在Dockerfile参考中找到。

# 简介

Docker可以通过阅读Docker中的指令来自动构建映像 Dockerfile。A Dockerfile是一个文本文档，
其中包含用户可以在命令行上调用以组装图像的所有命令。
使用docker build 用户可以创建自动构建，该构建连续执行多个命令行指令。


# escape

转义docker内的字符默认是反斜杠
通常不需要设置，当编译win系统

```
FROM microsoft/nanoserver
COPY testfile.txt c:\\
RUN dir c:\
```

这里会将\视为转义字符所以将找不到c:\路径，这个时间需要使用其他字符为转义字符

需要用以下设置：

```
# escape=`
FROM microsoft/nanoserver
COPY testfile.txt c:\\
RUN dir c:\
```

# .dockerignore

类似于gitignore,忽略文件不进行操作

```
*.md
README-secret.md
!README*.md
```

# 参数

## From

FROM指令为后续指令设置Base Image

```
From ubuntu
```

## MAINTAINER

设置生成的images的作者字段


```
MAINTAINER tianpeng.du
From ubuntu
```

## RUN 

在当前image之上的新层中执行任何命令，并提交结果，生成的已提交image将用于Dockerfile中的下一步

- command 形式  RUN <command>
- exec形式  ['execable','parm','parm']

## CMD

CMD的主要目的是为执行容器提供默认值，
配置方式同RUN

- 有ENTRYPOINT 作为ENTRYPOINT参数应以JSON数组格式指定
- 无ENTRYPOINT 执行写入的配置

在Dockerfile中只能有一个CMD指令。如果您列出多个CMD，则只有最后一个CMD将生效。

```
MAINTAINER tianpeng.du
From ubuntu
CMD sleep 1000
```

## LABEL

向image添加元数据

```
MAINTAINER tianpeng.du
From ubuntu
LABEL version="1.0"
CMD sleep 1000
```

## EXPOSE

EXPOSE指令通知Docker容器在运行时侦听指定的网络端口。

```
MAINTAINER tianpeng.du
From ubuntu
EXPOSE 8080
LABEL version="1.0"
CMD sleep 1000
```

## ENV

设置环境变量

```
MAINTAINER tianpeng.du
From ubuntu
EXPOSE 8080
ENV test 123
LABEL version="1.0"
CMD sleep 1000
```

## ADD

ADD指令从<src>复制新文件，目录或远程文件URL，并将它们添加到容器的文件系统，路径<dest>。

```
MAINTAINER tianpeng.du
From ubuntu
EXPOSE 8080
ENV test 123
ADD test.txt /media/
ADD abc.com/test.txt /media/
LABEL version="1.0"
CMD sleep 1000
```

- <src>路径必须在构建的上下文中
- 如果<src>是URL并且<dest>以尾部斜杠结尾，则从URL中推断文件名，并将文件下载到<dest>/<filename>
- 如果<src>是识别的压缩格式（identity，gzip，bzip2或xz）的本地tar存档，则将其解包为目录
- 如果<src>是任何其他类型的文件，它会与其元数据一起单独复制。在这种情况下，如果<dest>以尾部斜杠/结尾，它将被认为是一个目录，并且<src>的内容将被写在<dest>/base(<src>)。
- 如果直接或由于使用通配符指定了多个<src>资源，则<dest>必须是目录，并且必须以斜杠/结尾。
- 如果<dest>不以尾部斜杠结尾，它将被视为常规文件，<src>的内容将写在<dest>。
- 如果<dest>不存在，则会与其路径中的所有缺少的目录一起创建。

## COPY

两种形式： 

- COPY <src>... <dest> 
- COPY ["<src>",... "<dest>"]

基本和ADD类似，不过COPY的<src>不能为URL。

```
MAINTAINER tianpeng.du
From ubuntu
EXPOSE 8080
ENV test 123
COPY test.txt /media/
ADD abc.com/test.txt /media/
LABEL version="1.0"
CMD sleep 1000
```

## ENTRYPOINT
ENTRYPOINT允许您配置容器，运行执行的可执行文件。

两种形式： 
- ENTRYPOINT “executable”, “param1”, “param2” 
- ENTRYPOINT command param1 param2 (shell 形式)

MAINTAINER tianpeng.du
From ubuntu
EXPOSE 8080
ENV test 123
COPY test.txt /media/
ADD abc.com/test.txt /media/
LABEL version="1.0"
ENTRYPOINT sleep
CMD ["1000"]

> 覆盖entrypoint docker run --entrypoint=/bin/sh xxx 

## VOLUME

VOLUME指令创建具有指定名称的挂载点，并将其标记为从本机主机或其他容器保留外部挂载的卷。

## USER

USER指令设置运行image时使用的用户名或UID

## WORKDIR

WORKDIR指令为Dockerfile中的任何RUN，CMD，ENTRYPOINT，COPY和ADD指令设置工作目录。

## ARG

ARG指令定义一个变量，用户可以使用docker build命令使用--build-arg <varname> = <value>标志，
在构建时将其传递给构建器。如果用户指定了一个未在Dockerfile中定义的构建参数，构建将输出错误。

## ONBUILD

ONBUILD指令在image被用作另一个构建的基础时，向image添加要在以后执行的*trigger*指令

## STOPSIGNAL

STOPSIGNAL指令设置将发送到容器以退出的系统调用信号

## HEALTHCHECK

- HEALTHCHECK [OPTIONS] CMD command (通过在容器中运行命令来检查容器运行状况) 
- HEALTHCHECK NONE (禁用从基本映像继承的任何运行状况检查)

## SHELL

SHELL指令允许用于命令的shell形式的默认shell被覆盖

