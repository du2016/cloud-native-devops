设备白名单控制器

## 说明

实施一个cgroup来跟踪和执行开放和mknod限制
在设备文件上。设备cgroup关联设备访问
每个cgroup列入白名单。白名单条目具有4个字段。
“类型”是a（全部），c（字符）或b（块）。“全部”表示适用
适用于所有类型以及所有主要和次要数字。主要和次要是
整数或*。访问是由r组成的
（读取），w（写入）和m（mknod）。

根设备cgroup以rwm开头为“ all”。子设备
cgroup获取父级的副本。管理员然后可以删除
白名单中的设备或添加新条目。儿童cgroup可以
永远不会收到其父级拒绝的设备访问权限。

2.用户界面

使用devices.allow添加条目，并使用删除条目
devices.deny。例如

	echo 'c 1:3 mr' > /sys/fs/cgroup/1/devices.allow

允许cgroup 1读取并mknod该设备通常称为
/ dev / null。在做

	echo a > /sys/fs/cgroup/1/devices.deny

将删除默认的'a *：* rwm'条目。在做

	echo a > /sys/fs/cgroup/1/devices.allow

会将'a *:* rwm'条目添加到白名单。

3.安全性

任何任务都可以在cgroup之间移动。这显然不会
足够，但我们可以决定适当限制的最佳方法
人们对此有所了解。我们可能只想
需要CAP_SYS_ADMIN，至少与
CAP_MKNOD。我们可能只想拒绝迁移到
不是当前版本的后代。或者我们可能要使用
CAP_MAC_ADMIN，因为我们确实试图锁定root用户。

需要CAP_SYS_ADMIN来修改白名单或移动另一个
新cgroup的任务。（再次，我们可能要更改它）。

不能向cgroup授予比cgroup更大的权限
父母有。

4.层次结构

设备cgroup通过确保cgroup从不拥有更多内容来维护层次结构
访问权限比其父级高。每次将条目写入
cgroup的devices.deny文件，其所有子项都将删除该条目
从他们的白名单中，所有本地设置的白名单条目都将是
重新评估。如果本地设置的白名单条目之一将提供
比cgroup的父级拥有更多访问权限，它将被从白名单中删除。

例：

      A
     / \
        乙

    group        behavior	exceptions
    A            allow		"b 8:* rwm", "c 116:1 rw"
    B            deny		"c 1:3 rwm", "c 116:2 rwm", "b 3:* rwm"

如果设备在组A中被拒绝：
	＃echo“ c 116：* r”> A / devices.deny
它会向下传播，并在重新验证B的条目后，将白名单条目
“ c 116：2 rwm”将被删除：

    group        whitelist entries                        denied devices
    A            all                                      "b 8:* rwm", "c 116:* rw"
    B            "c 1:3 rwm", "b 3:* rwm"                 all the rest

如果父母的例外发生变化并且不允许本地例外
将会被删除。

请注意，不会传播新的白名单条目：
      一种
     / \
        乙

    group        whitelist entries                        denied devices
    A            "c 1:3 rwm", "c 1:5 r"                   all the rest
    B            "c 1:3 rwm", "c 1:5 r"                   all the rest

添加“ c *：3 rwm”时：
	# echo "c *:3 rwm" >A/devices.allow

结果：

    group        whitelist entries                        denied devices
    A            "c *:3 rwm", "c 1:5 r"                   all the rest
    B            "c 1:3 rwm", "c 1:5 r"                   all the rest

但现在可以将新条目添加到B：

	# echo "c 2:3 rwm" >B/devices.allow
	# echo "c 50:3 r" >B/devices.allow
甚至
	# echo "c *:3 rwm" >B/devices.allow

通过在device.allow或devices.deny中写入“ a”来允许或拒绝所有操作
一旦设备cgroups有子代，则不可能。

4.1层次结构（内部实现）

设备cgroups在内部使用行为（ALLOW，DENY）和
例外清单。内部状态由同一用户控制
接口，以保持与以前的仅白名单的兼容性
实施。删除或添加将减少访问权限的异常
设备将向下传播到层次结构。
对于每个传播的异常，将根据以下规则重新评估有效规则
根据当前父母的访问规则。