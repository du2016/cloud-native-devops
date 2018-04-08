

# Container Network Interface

CNI（容器网络接口）是一个云本地计算基金会项目，由用于编写插件以在Linux容器中配置网络接口的规范和库以及许多受支持的插件组成。CNI仅关注容器的网络连接，并在容器被删除时移除分配的资源。由于这个重点，CNI有着广泛的支持，规范很容易实现。

除了规范外，该存储库还包含用于将CNI集成到应用程序中的库的Go源代码以及用于执行CNI插件的示例命令行工具。一个独立的存储设备中包含的参考插件和制造新的插件模板。

模板代码使得它直接为现有的集装箱网络项目创建一个CNI插件。CNI也为从头创建新的集装箱网络项目提供了一个很好的框架。

每个CNI插件必须实现为由容器管理系统（例如rkt或Kubernetes）调用的可执行文件。

CNI插件负责将网络接口插入容器网络命名空间（例如，veth对的一端）
并在主机上进行任何必要的更改（例如将veth的另一端连接到网桥）。
然后它应该将IP分配给接口，并通过调用适当的IPAM插件来设置与“IP地址管理"部分一致的路由。

现有主流CNI插件列表请看[官方列表](https://github.com/containernetworking/cni/blob/master/README.md)

# 为什么要开发CNI？

Linux上的应用程序容器是一个快速发展的领域，在这个领域内，网络并没有得到很好的解决，因为它是高度环境特定的。我们相信许多容器运行时和协调器将试图解决使网络层可插入的相同问题。

为了避免重复，我们认为在网络插件和容器执行之间定义一个通用接口是明智的，因此我们提出了这个规范，以及Go和一组插件库。

# 总则

- 容器运行时必须在调用任何插件之前为容器创建一个新的网络名称空间。
- 然后，运行时必须确定这个容器应属于哪个网络，并为每个网络确定哪些插件必须执行。
- 网络配置采用JSON格式，可以轻松存储在文件中。网络配置包括必填字段，如“名称"和“类型"以及插件（类型）特定的字段。网络配置允许字段在调用之间更改值。为此，可选字段“args"必须包含不同的信息。
- 容器运行时必须通过依次为每个网络执行相应的插件，将容器添加到每个网络。
- 在完成容器生命周期后，运行时必须以相反顺序执行插件（相对于执行它们来添加容器的顺序）以将容器与网络断开连接。
- 容器运行时不能为同一容器调用并行操作，但可以为不同容器调用并行操作。
- 容器运行时必须为容器订购ADD和DEL操作，以便ADD后面总是跟随相应的DEL。DEL可能会跟着额外的DEL，但是，插件应该允许处理多个DEL（即插件DEL应该是幂等的）。
- 容器必须由ContainerID唯一标识。存储状态的插件应该使用主键(network name, container id)。
- 运行时不能相同地调用ADD两次（没有相应的DEL）(network name, container id)。换句话说，给定的容器ID必须只添加到特定网络一次

# 参数

CNI插件必须支持以下操作：
 - 将容器添加到网络
  - 参数
    - 容器ID。运行时分配的容器的唯一明文标识符。一定不能是空的。
    - 网络命名空间路径。这表示要添加的网络名称空间的路径，即/proc/\[pid\]/ns/net或bind-mount/link。
    - 网络配置。这是描述可以连接容器的网络的JSON文档。架构如下所述。
    - 额外的参数。这提供了一种替代机制，允许以每个容器为基础简单配置CNI插件。
    - 容器内接口的名称。这是应该分配给容器（网络命名空间）内创建的接口的名称; 因此它必须符合Linux接口名称上的标准限制。
  - 反馈
    - 接口列表。根据插件的不同，这可以包括沙箱（例如容器或管理程序）接口名称和/或主机接口名称，每个接口的硬件地址以及接口所在的沙箱（如果有）的详细信息。
    - 分配给每个接口的IP配置。分配给沙箱和/或主机接口的IPv4和/或IPv6地址，网关和路由。
    - DNS信息。包含名称服务器，域，搜索域和选项的DNS信息的字典。
 - 从网络删除容器
  - 参数
    - 容器ID，如上所述。
    - 网络命名空间路径，如上所述。
    - 网络配置，如上所述。
    - 额外的参数，如上所述。
    - 如上定义的容器内接口的名称
  - 所有参数应与传递给相应的添加操作的参数相同。
  - 删除操作应释放配置的网络中由所提供的Containerid拥有的所有资源。  
 - 报告版本
  - 参数： 无
  - 返回：有关插件支持的CNI规范版本的信息
    ```
    {
      "cniVersion": "0.3.1", // the version of the CNI spec in use for this output
      "supportedVersions": [ "0.1.0", "0.2.0", "0.3.0", "0.3.1" ] // the list of CNI spec versions that this plugin supports
    }
    ```
运行时必须使用网络类型（请参阅下面的网络配置）作为要调用的可执行文件的名称。
然后运行时应该在预定义的目录列表中查找这个可执行文件（本规范没有规定目录列表）。
一旦找到，它必须使用以下环境变量来调用可执行文件来传递参数：

CNI_COMMAND：表示所需的操作; ADD，DEL或者VERSION。
CNI_CONTAINERID：容器ID
CNI_NETNS：网络名称空间文件的路径
CNI_IFNAME：设置接口名称; 如果插件不能使用这个接口名称，它必须返回一个错误
CNI_ARGS：用户在调用时传入的额外参数。以分号分隔的字母数字键值对; 例如，"FOO = BAR; ABC = 123"
CNI_PATH：搜索CNI插件可执行文件的路径列表。路径由OS特定的列表分隔符分隔; 例如在Linux上'：'在Windows上为';'

JSON格式的网络配置必须通过stdin流式传输到插件。这意味着它没有绑定到磁盘上的特定文件，并且可能包含在调用之间更改的信息。


结果：

请注意，IPAM插件应按IP分配中Result所述返回缩写结构。
插件必须指示成功，返回代码为零，并且在ADD命令的情况下将以下JSON打印到标准输出。
在ips和dns项目应该是相同的输出由IPAM插件（见返回IP分配的详细信息）
除了插件应该在填写interface适当的指标，这是从IPAM插件输出丢失，
因为IPAM插件应该是不知道的接口。

```
{
  "cniVersion": "0.3.1",
  "interfaces": [                                            (this key omitted by IPAM plugins)
      {
          "name": "<name>",
          "mac": "<MAC address>",                            (required if L2 addresses are meaningful)
          "sandbox": "<netns path or hypervisor identifier>" (required for container/hypervisor interfaces, empty/omitted for host interfaces)
      }
  ],
  "ips": [
      {
          "version": "<4-or-6>",
          "address": "<ip-and-prefix-in-CIDR>",
          "gateway": "<ip-address-of-the-gateway>",          (optional)
          "interface": <numeric index into 'interfaces' list>
      },
      ...
  ],
  "routes": [                                                (optional)
      {
          "dst": "<ip-and-prefix-in-cidr>",
          "gw": "<ip-of-next-hop>"                           (optional)
      },
      ...
  ]
  "dns": {
    "nameservers": <list-of-nameservers>                     (optional)
    "domain": <name-of-local-domain>                         (optional)
    "search": <list-of-additional-search-domains>            (optional)
    "options": <list-of-options>                             (optional)
  }
}
```

cniVersion指定插件使用的CNI规范的语义版本2.0。一个插件可以支持多个CNI规范版本（通过VERSION命令报告），这里cniVersion插件返回的结果必须与网络配置中cniVersion指定的一致。如果插件不支持网络配置中的插件，则该插件应返回错误代码1（有关详细信息，请参阅众所周知的错误代码）
interfaces描述插件创建的特定网络接口。如果CNI_IFNAME变量存在，插件必须将该名称用于沙箱/管理程序接口，否则返回错误。
- mac（字符串）：接口的硬件地址。如果L2地址对插件没有意义，那么这个字段是可选的。
- sandbox（字符串）：基于容器/命名空间的环境应该将完整的文件系统路径返​​回到该沙箱的网络名称空间。基于虚拟机管理程序/基于虚拟机的插件应返回创建接口的虚拟化沙箱特有的ID。必须为创建或移入沙盒（如网络名称空间或管理程序/虚拟机）的接口提供此项。

该ips字段是IP配置信息的列表。有关更多信息，请参阅IP知名结构部分。

该dns字段包含由常用DNS信息组成的字典。有关更多信息，请参阅DNS着名的结构部分。

该规范没有声明CNI消费者必须如何处理这些信息。示例包括生成/etc/resolv.conf要注入容器文件系统的文件或在主机上运行DNS转发器。

错误必须用非零返回码表示，并且将以下JSON打印到标准输出：

```
{
  "cniVersion": "0.3.1",
  "code": <numeric-error-code>,
  "msg": <short-error-message>,
  "details": <long-error-message> (optional)
}
```

cniVersion指定插件使用的CNI规范的语义版本2.0。错误代码0-99保留用于众所周知的错误（请参阅众所周知的错误代码部分）。100+的值可以自由用于插件特定的错误。

另外，stderr可以用于非结构化输出，如日志。

# 网络配置
网络配置以JSON格式描述。配置可以存储在磁盘上，也可以通过容器运行时从其他源生成。以下字段是众所周知的，其含义如下：

- cniVersion（字符串）：该配置符合的CNI规范的语义版本2.0。
- name（字符串）：网络名称。这应该在主机（或其他管理域）上的所有容器中都是唯一的。
- type （字符串）：指CNI插件可执行文件的文件名。
- args（字典）：容器运行时提供的可选附加参数。例如，标签字典可以通过将其添加到标签字段下传递给CNI插件args。
- ipMasq（boolean）：可选（如果插件支持）。在此网络的主机上设置IP伪装。如果主机将作为无法路由到分配给容器的IP的子网的网关，则这是必需的。
- ipam：具有IPAM特定值的字典：
  - type （字符串）：指IPAM插件可执行文件的文件名。
- dns：具有DNS特定值的字典：
  - nameservers（字符串列表）：该网络知道的DNS名称服务器的优先级排序列表的列表。列表中的每个条目都是包含IPv4或IPv6地址的字符串。
  - domain （字符串）：用于短主机名查找的本地域。
  - search（字符串列表）：用于短主机名查找的优先级有序搜索域列表。domain大多数解析器都会优先选择。
  - options （字符串列表）：可以传递给解析器的选项列表

插件可能会定义他们接受的其他字段，如果使用未知字段调用，可能会产生错误。这个例外是该args字段可能被用来传递任意数据，如果不理解，应该被插件忽略。

# 示例配置

```
{
   "cniVersion "："0.3.1 "，
   "name "："dbnet "，
   "type "："bridge "，
   //  type  （插件） 特定的
  "bridge "："cni0 "，
   "ipam "：{
     "type "：“主机本地"，
    //  ipam  特定
    “子网“：" 10.1.0.0/16 “，
     "网关“：" 10.1.0.1 “
  }，
  "dns "：{
     "nameservers "：[ "10.1.0.1 " ]
  }
}
```

```
{
   "cniVersion "："0.3.1 "，
   "name "："pci "，
   "type "："ovs "，
   //  type  （插件） 特定的
  "bridge "："ovs0 "，
   "vxlanID "：42，
   "ipam “：{
     " type “："dhcp "，
     “路由"：[{ "DST "："10.3.0.0/16 " }，{ "DST "："10.4.0.0/16 " }]
  }
  //  ARGS  可以 被 忽略 通过 插件
  "ARGS "： {
     “标签"：{
         "appVersion "："1.0 "
    }
  }
}
```

```
{
   "cniVersion "："0.3.1 "，
   "name "："wan "，
   "type "："macvlan "，
   //  ipam  specific 
  "ipam "：{
     "type "："dhcp "，
     "routes "：[ "dst "："10.0.0。0/8 “，" gw“： " 10.0.0.1 "}]
  }，
  "dns "：{
     "nameservers "：[ "10.0.0.1 " ]
  }
}
```

- 网络配置列表

网络配置列表提供了一种机制，可以按照定义的顺序为单个容器运行多个CNI插件，并将每个插件的结果传递给下一个插件。该列表由众所周知的字段和一个或多个标准CNI网络配置列表组成（参见上文）。

该列表以JSON格式描述，可以存储在磁盘上或由容器运行时从其他源生成。以下字段是众所周知的，其含义如下：

- cniVersion（字符串）：此配置列表和所有单独配置符合的CNI规范的语义版本2.0。
name（字符串）：网络名称。这应该在主机（或其他管理域）上的所有容器中都是唯一的。
- plugins （列表）：标准CNI网络配置字典的列表（参见上文）。
当执行插件列表，该运行时必须更换name和cniVersion在每个单独的网络配置字段与列表name和cniVersion字段列表本身的。这确保名称和CNI版本与列表中的所有插件执行相同，从而防止插件之间的版本冲突。运行时也可以将基于能力的密钥作为runtimeConfig插件的顶级密钥JSON中的映射传递，如果插件通告它，则通过capabilities其网络配置的密钥支持特定的功能。传入的密钥runtimeConfig必须与capabilities插件网络配置的密钥中的特定功能的名称匹配。请参阅CONVENTIONS.md了解更多关于功能的信息以及如何通过它们发送给插件runtimeConfig 键。

对于ADD动作，运行时还必须prevResult在第一个插件之后的任何插件的配置JSON中添加一个字段，它必须是以JSON格式（见下文）的上一个插件的结果（如果有的话）。对于ADD动作，插件应该将prevResult字段的内容回显到它们的stdout，以允许后续插件（和运行时）接收结果，除非他们希望修改或抑制先前的结果。允许插件修改或压缩全部或部分prevResult。但是，支持包含该prevResult字段的CNI规范版本的插件必须prevResult通过传递，修改或明确禁止来进行处理。不了解该prevResult领域是违反此规范的。

运行时还必须使用相同的环境执行列表中的每个插件。

对于DEL操作，运行时必须以相反的顺序执行插件。

## 网络配置列表错误处理

当在插件列表上执行动作时发生错误（例如ADD或DEL）时，运行时必须停止执行列表。

如果ADD操作失败，则当运行时决定处理失败时，即使在ADD操作期间没有调用某个插件，它也应该执行DEL操作（与上面指定的ADD相反的顺序）。

即使缺少一些资源，插件通常也应该完成DEL操作而不会出错。例如，即使容器网络名称空间不再存在，IPAM插件通常应释放IP分配并返回成功，除非该网络名称空间对于IPAM管理至关重要。虽然DHCP通常可以在容器网络接口上发送“释放"消息，但由于DHCP租期有一定的生命周期，因此此发布操作不会被视为关键，并且不应返回任何错误。再例如，bridge即使容器网络名称空间和/或容器网络接口不再存在，插件也应将DEL操作委托给IPAM插件并清理自己的资源（如果存在）。


## 示例网络配置列表

```
{
  "cniVersion": "0.3.1",
  "name": "dbnet",
  "plugins": [
    {
      "type": "bridge",
      // type (plugin) specific
      "bridge": "cni0",
      // args may be ignored by plugins
      "args": {
        "labels" : {
            "appVersion" : "1.0"
        }
      },
      "ipam": {
        "type": "host-local",
        // ipam specific
        "subnet": "10.1.0.0/16",
        "gateway": "10.1.0.1"
      },
      "dns": {
        "nameservers": [ "10.1.0.1" ]
      }
    },
    {
      "type": "tuning",
      "sysctl": {
        "net.core.somaxconn": "500"
      }
    }
  ]
}
```

## 网络配置列表运行时示例

鉴于上面显示的网络配置列表JSON ，容器运行时将为ADD操作执行以下步骤。请注意，运行时将配置列表中的字段cniVersion和name字段添加到传递给每个插件的配置JSON，以确保列表中所有插件的版本控制和名称一致。

1 首先bridge用以下JSON 调用插件：

```
{
  "cniVersion": "0.3.1",
  "name": "dbnet",
  "type": "bridge",
  "bridge": "cni0",
  "args": {
    "labels" : {
        "appVersion" : "1.0"
    }
  },
  "ipam": {
    "type": "host-local",
    // ipam specific
    "subnet": "10.1.0.0/16",
    "gateway": "10.1.0.1"
  },
  "dns": {
    "nameservers": [ "10.1.0.1" ]
  }
}
```

2.接下来tuning使用以下JSON 调用插件，包括prevResult包含来自bridge插件的JSON响应的字段：

```
{
  "cniVersion": "0.3.1",
  "name": "dbnet",
  "type": "tuning",
  "sysctl": {
    "net.core.somaxconn": "500"
  },
  "prevResult": {
    "ips": [
        {
          "version": "4",
          "address": "10.0.0.5/32",
          "interface": 0
        }
    ],
    "dns": {
      "nameservers": [ "10.1.0.1" ]
    }
  }
}
```

给定相同的网络配置JSON列表，容器运行时将为DEL操作执行以下步骤。请注意，prevResult由于DEL操作不返回任何结果，因此不需要字段。还要注意，插件是以相反的顺序从ADD操作执行的。

1. 首先tuning用以下JSON 调用插件：

```
{
   “ cniVersion "：“ 0.3.1 "，
   “ name "：“ dbnet "，
   “ type "：“ tuning "，
   “ sysctl "：{
     “ net.core.somaxconn "：“ 500 "
  }
}
```

2. 接下来bridge使用以下JSON 调用插件：

```
{
  "cniVersion": "0.3.1",
  "name": "dbnet",
  "type": "bridge",
  "bridge": "cni0",
  "args": {
    "labels" : {
        "appVersion" : "1.0"
    }
  },
  "ipam": {
    "type": "host-local",
    // ipam specific
    "subnet": "10.1.0.0/16",
    "gateway": "10.1.0.1"
  },
  "dns": {
    "nameservers": [ "10.1.0.1" ]
  }
}
```

# IP分配

作为其操作的一部分，CNI插件需要为接口分配（并维护）一个IP地址，并安装与该接口相关的所有必要路由。这给了CNI插件很大的灵活性，但也给它带来了很大的负担。许多CNI插件需要具有相同的代码来支持用户可能需要的多种IP管理方案（例如dhcp，host-local）。

为了减轻负担并使IP管理策略与CNI插件的类型正交，我们定义了第二种类型的插件--IP地址管理插件（IPAM插件）。然而，CNI插件的责任是在其执行的适当时刻调用IPAM插件。IPAM插件必须确定接口IP /子网，网关和路由，并将此信息返回到“主"插件才能应用。IPAM插件可以通过协议（例如dhcp），存储在本地文件系统中的数据，网络配置文件的“ipam"部分或上述的组合来获得信息。

## IP地址管理（IPAM）接口

像CNI插件一样，IPAM插件通过运行可执行文件来调用。可执行文件在预定义的路径列表中搜索，并通过CNI插件指示CNI_PATH。IPAM插件必须接收所有传入CNI插件的相同环境变量。就像CNI插件一样，IPAM插件通过stdin接收网络配置。

成功必须以零返回码表示，并且将以下JSON打印到标准输出（在ADD命令的情况下）：

```
{
  "cniVersion": "0.3.1",
  "ips": [
      {
          "version": "<4-or-6>",
          "address": "<ip-and-prefix-in-CIDR>",
          "gateway": "<ip-address-of-the-gateway>"  (optional)
      },
      ...
  ],
  "routes": [                                       (optional)
      {
          "dst": "<ip-and-prefix-in-cidr>",
          "gw": "<ip-of-next-hop>"                  (optional)
      },
      ...
  ]
  "dns": {
    "nameservers": <list-of-nameservers>            (optional)
    "domain": <name-of-local-domain>                (optional)
    "search": <list-of-search-domains>              (optional)
    "options": <list-of-options>                    (optional)
  }
}
```

请注意，与常规的CNI插件不同，IPAM插件应返回Result不包含interfaces密钥的缩写结构，因为IPAM插件应该不知道其父插件配置的接口（dhcpIPAM插件专门需要的接口除外）。

cniVersion指定由IPAM插件使用的CNI规范的语义版本2.0。IPAM插件可以支持多种CNI规格版本（如通过VERSION命令报告），这里cniVersionIPAM插件返回的结果必须与网络配置中cniVersion指定的一致。如果IPAM插件不支持网络配置中的插件，则该插件应返回错误代码1（有关详细信息，请参阅众所周知的错误代码）。cniVersion

该ips字段是IP配置信息的列表。有关更多信息，请参阅IP知名结构部分。

该dns字段包含由常用DNS信息组成的字典。有关更多信息，请参阅DNS着名的结构部分。

错误和日志以与CNI插件相同的方式传递。有关详细信息，请参阅CNI插件结果部分。

IPAM插件示例：

- host-local：在指定范围内选择一个未使用（由同一主机上的其他容器）IP。
- dhcp：使用DHCP协议获取和维护租约。DHCP请求将通过创建的容器接口发送; 因此，关联的网络必须支持广播。

## 注意

- 预计路线将添加0度量。
- 默认路由可以通过“0.0.0.0/0"指定。由于另一个网络可能已经配置了默认路由，CNI插件应该准备跳过其默认路由定义。

# 着名的结构
IP地址
```
  "ips": [
      {
          "version": "<4-or-6>",
          "address": "<ip-and-prefix-in-CIDR>",
          "gateway": "<ip-address-of-the-gateway>",      (optional)
          "interface": <numeric index into 'interfaces' list> (not required for IPAM plugins)
      },
      ...
  ]
```
该ips字段是由插件确定的IP配置信息的列表。每个项目都是描述网络接口的IP配置的字典。多个网络接口的IP配置和单个接口的多个IP配置可作为ips列表中的单独项目返回。即使没有严格要求，也应提供插件已知的所有属性。

- version（字符串）：“4"或“6"，对应于条目中地址的IP版本。提供的所有IP地址和网关必须对给定的IP地址和网关有效version。
- address （字符串）：CIDR表示法中的IP地址（例如“192.168.1.3/24"）。
- gateway（字符串）：此子网的默认网关（如果存在）。它不会指示CNI插件使用此网关添加任何路由：要添加的路由通过该routes字段单独指定。CNI bridge插件使用此值的一个示例是将此IP地址添加到Linux网桥以使其成为网关。
- interface（uint）：CNI插件结果interfaces列表中的索引，指示应将此IP配置应用于哪个接口。IPAM插件不应该返回此密钥，因为它们没有关于网络接口的信息。
路线
```
  "routes": [
      {
          "dst": "<ip-and-prefix-in-cidr>",
          "gw": "<ip-of-next-hop>"               (optional)
      },
      ...
  ]
```
- 每个routes条目都是一个包含以下字段的字典。routes条目中的所有IP地址必须是相同的IP版本，即4或6。
  - dst （字符串）：以CIDR表示法指定的目标子网。
  - gw（字符串）：网关的IP。如果省略，则假定默认网关（由CNI插件确定）。

## DNS

```
  "dns": {
    "nameservers": <list-of-nameservers>                 (optional)
    "domain": <name-of-local-domain>                     (optional)
    "search": <list-of-additional-search-domains>        (optional)
    "options": <list-of-options>                         (optional)
  }
```

该dns字段包含由常用DNS信息组成的字典。

- nameservers（字符串列表）：该网络知道的DNS名称服务器的优先级排序列表的列表。列表中的每个条目都是包含IPv4或IPv6地址的字符串。
- domain （字符串）：用于短主机名查找的本地域。
- search（字符串列表）：用于短主机名查找的优先级有序搜索域列表。domain大多数解析器都会优先选择。
- options（字符串列表）：可以传递给解析器的选项列表。有关更多信息，请参阅CNI插件结果部分。

# 众所周知的错误代码

错误代码1-99不得在此处指定以外使用。

1 - 不兼容的CNI版本
2 - 网络配置中不支持的字段。错误消息必须包含不受支持的字段的键和值。