# Beehive

在看kubeedge的源码过程中发现对beehive的理解不够深刻，所以又回来看了一下beehive的官方文档

# 概述

Beehive是基于go-channel的消息传递框架，用于KubeEdge模块之间的通信。
如果已注册其他beehive模块的名称或该模块组的名称已知，则在蜂箱中注册的模块可以与其他蜂箱模块进行通信。Beehive支持以下模块操作：

- 添加模块
- 将模块添加到组
- 清理（​​从蜂巢核心和所有组中删除模块）

Beehive支持以下消息操作：

- 发送到模块/组
- 通过模块接收
- 发送同步到模块/组
- 发送对同步消息的响应

# 消息格式

消息分为三部分

1. header：

- ID：消息ID（字符串）
- ParentID：如果是对同步消息的响应，则说明parentID存在（字符串）
- TimeStamp：生成消息的时间（整数）
- sync：标志，指示消息是否为同步类型（布尔型）

2. Route：

- Source：消息的来源（字符串）
- Group：必须将消息广播到的组（字符串）
- Operation：对资源的操作（字符串）
- Resource：要操作的资源（字符串）

3. content：消息的内容（interface{}）

# 注册模块

1. 在启动edgecore时，每个模块都会尝试将其自身注册到beehive内核。

2. Beehive核心维护一个名为modules的映射，该映射以模块名称为键，模块接口的实现为值。

3. 当模块尝试向蜂巢核心注册自己时，beehive 内核会从已加载的modules.yaml配置文件中进行检查，
以检查该模块是否已启用。如果启用，则将其添加到模块映射中，否则将其添加到禁用的模块映射中。

# channel上下文结构字段

*  channels - channels是字符串（键）的映射，它是模块的名称和消息的通道（值），用于将消息发送到相应的模块。
*  chsLock - channels map的锁
*  typeChannels  - typeChannels是一个字符串（key）的映射，它是组(将字符串(key)映射到message的chan(value)，
是该组中每个模块的名称到对应通道的映射。
*  typeChsLock - typeChannels map的锁 
*  anonChannels - anonChannels是消息的字符串（父id）到chan（值）的映射，将用于发送同步消息的响应。 
*  anonChsLock - anonChannels map的锁

# 模块操作

## 添加模块
   
*  添加模块操作首先创建一个消息类型的新通道。
*  然后，将模块名称（键）及其通道（值）添加到通道上下文结构的通道映射中。
*  例如：添加边缘模块

```
coreContext.Addmodule(“edged”)
```

## 将模块添加到组中

*  首先，addModuleGroup从通道映射中获取模块的通道。
*  然后，将模块及其通道添加到typeChannels映射中，其中key是组，值是map中的映射（key是模块名称，value是通道）。
*  例如：在边缘组中添加边缘。这里的第一个边缘是模块名称，第二个边缘是组名称。

```
coreContext.AddModuleGroup(“edged”,”edged”)
```

## CleanUp

* CleanUp从通道映射中删除该模块，并从所有组（typeChannels映射）中删除该模块。
* 然后，关闭与模块关联的通道。
* 例如：清理边缘模块

```
coreContext.CleanUp(“edged”)
```

# 消息操作

##发送给模块

* 发送从通道映射中获取模块的通道。
* 然后，将消息放入通道。
* 例如：发送消息到edged。

```
coreContext.Send(“edged”,message) 
```

## 发送给一组

* SendToGroup从typeChannels映射获取所有模块（映射）。
* 然后，在地图上进行迭代，并在地图中所有模块的通道上发送消息。
* 例如：要发送到边缘组中所有模块的消息。

```
coreContext.SendToGroup(“edged”,message) message will be sent to all modules in edged group.
```

## 通过模块接收

* 接收从通道图获取模块的通道。

* 然后，它等待消息到达该通道并返回消息。如果有错误，则返回错误。
* 例如：接收边缘模块的消息

```
msg, err := coreContext.Receive("edged")
```

## SendSync到模块

* SendSync具有3个参数（模块，消息和超时持续时间）
* SendSync首先从channels map中获取模块的channel。
* 然后，将消息放入channel。
* 然后创建一个新的消息channel，并将其添加到anonChannels映射中，其中键是messageID。
* 然后，它等待在它创建的anonChannel上接收到消息（响应），直到超时。
* 如果在超时之前收到消息，则返回错误为nil的消息，否则返回超时错误。
* 例如：以60秒的超时时间将同步发送到边缘

```
response, err := coreContext.SendSync("edged",message,60*time.Second)
```

## SendSync到组

* 从typeChannels映射中获取组的模块列表。
* 创建一个消息channel，其大小等于该组中的模块数，然后将anonChannels映射作为值放入，键为messageID。
* 在所有模块的channel上发送消息。
* 等到超时。如果anonChannel的长度=该组中的模块数，请检查通道中的所有消息是否具有parentID = messageID。如果没有返回错误，则返回nil错误。
* 如果达到超时，则返回超时错误。
* 例如：以60秒的超时时间向edged发送同步消息

```
err := coreContext.SendToGroupSync("edged",message,60*time.Second)
```

##SendResp到同步消息

* SendResp用于发送同步消息的响应。
* 发送响应的messageID必须在响应消息的parentID中。
* 调用SendResp时，它将检查响应消息的parentID是否存在anonChannels。
* 如果channel存在，则在该channel上发送消息（response）。
* 否则将记录错误。

```
coreContext.SendResp(respMessage)
```