# kubectl-plugins

kubectl-plugins 是在 v1.8.0 发行版中作为 alpha 功能正式引入的。 
因此，尽管插件功能的某些部分已经在以前的版本中可用，建议使用 1.8.0 或更高版本的 kubectl 版本.

# 安装 kubectl 插件

一个插件只不过是一组文件：至少一个 plugin.yaml 描述符，以及可能有一个或多个二进制文件、脚本或资产文件。 
要安装一个插件，将这些文件复制到 kubectl 搜索插件的文件系统中的某个位置.

请注意，Kubernetes 不提供包管理器或类似的东西来安装或更新插件，
因此您有责任将插件文件放在正确的位置。我们建议每个插件都位于自己的目录下，
因此安装一个以压缩文件形式发布的插件就像将其解压到 插件加载器 部分指定的某个位置一样简单。

# 插件加载器

插件加载器负责在下面指定的文件系统位置搜索插件文件，并检查插件是否提供运行所需的最小信息量。放在正确位置但未提供最少信息的文件将被忽略，例如没有不完整的 plugin.yaml 描述符。

## 插件搜索顺序

插件加载器使用以下搜索顺序：

如果指定了 ${KUBECTL_PLUGINS_PATH} ，搜索在这里停止。
${XDG_DATA_DIRS}/kubectl/plugins
~/.kube/plugins
如果存在 KUBECTL_PLUGINS_PATH 环境变量，则加载器将其用作查找插件的唯一位置。 KUBECTL_PLUGINS_PATH 环境变量是一个目录列表。在 Linux 和 Mac 中，列表是冒号分隔的。在 Windows 中，列表是以分号分隔的。

如果 KUBECTL_PLUGINS_PATH 不存在，加载器将搜索这些额外的位置：

首先，根据指定的一个或多个目录 [XDG系统目录结构]（https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html） 规范。具体来说，加载器定位由 XDG_DATA_DIRS 环境变量指定的目录， 然后在里面搜索 kubectl/plugins 目录。 如果未指定 XDG_DATA_DIRS ，则默认为 /usr/local/share:/usr/share 。

其次，用户的 kubeconfig 目录下的 plugins 目录。在大多数情况下，就是 ~/.kube/plugins 。

```
# Loads plugins from both /path/to/dir1 and /path/to/dir2
KUBECTL_PLUGINS_PATH=/path/to/dir1:/path/to/dir2 kubectl plugin -h
```


## 编写 kubectl 插件

您可以使用任何允许编写命令行命令的编程语言或脚本编写插件。 一个插件不一定需要有一个二进制组件。 它完全可以依靠操作系统实用程序 像 echo、sed 或 grep 。或者可以依靠 kubectl 二进制文件。

kubectl 插件的唯一的强需求是 plugin.yaml 描述符文件。该文件负责声明注册插件所需的最小属性，并且必须位于 插件搜索顺序 部分指定的其中一个位置下。

plugin.yaml 描述符
描述符文件支持以下属性：
```
name: "targaryen"                 # 必须项：插件命令名称，在 'kubectl' 下调用
shortDesc: "Dragonized plugin"    # 必须项: 该命令的简短描述，以获得帮助
longDesc: ""                      # the command long description, for help
example: ""                       # 该命令的长描述，寻求帮助
command: "./dracarys"             # 必须项：运行插件时要调用的命令、二进制文件或脚本
flags:                            # 插件支持的参数
  - name: "heat"                  # 每个参数的必须项：参数名称
    shorthand: "h"                # 参数名称的简短版本
    desc: "Fire heat"             # 每个参数的必须项：参数描述
    defValue: "extreme"           # 参数的默认值
tree:                             # 允许子命令的声明
  - ...                           # 子命令支持相同的一组属性
```
上面的描述符声明了 kubectl plugin targaryen 插件，它有一个名为 -h | --heat 的参数。 当插件被调用时，它会调用与描述符文件位于同一目录中的 dracarys 二进制文件或脚本。 访问运行时属性 部分描述了 dracarys 命令如何访问参数值和其他运行时上下文。


## 推荐的目录结构
建议每个插件在文件系统中都有自己的子目录，最好使用与插件命令相同的名称。该目录必须包含 plugin.yaml 描述符以及它可能需要的任何二进制文件、脚本、资产或其他依赖项。

例如，targaryen 插件的目录结构可能如下所示：

```
~/.kube/plugins/
└── targaryen
    ├── plugin.yaml
    └── dracarys
```

# 访问运行时属性

在大多数使用情况下，您为编写插件而编写的二进制文件或脚本文件必须能够访问由插件框架提供的一些上下文信息。例如，如果您在描述符文件中声明了参数，则您的插件必须能够在运行时访问用户提供的参数值。全局标志也是如此。插件框架负责做这件事，所以插件编写者不需要担心解析参数。这也确保了插件和常规 kubectl 命令之间的最佳一致性。

插件可以通过环境变量访问运行时上下文属性。因此，要访问通过参数提供的值，只需使用适当的函数调用二进制文件或脚本查找适当环境变量的值即可。

支持的环境变量是：

- KUBECTL_PLUGINS_CALLER: 在当前命令调用中使用的 kubectl 二进制文件的完整路径。作为一个插件编写者，您不必实现逻辑来认证和访问 Kubernetes API。相反，您可以通过像 kubectl get --raw=/apis 这样的命令来调用 kubectl 来获得您需要的信息。

- KUBECTL_PLUGINS_CURRENT_NAMESPACE: 当前名称空间是此调用的上下文。这是要使用的实际名称空间，这意味着它已经通过kubeconfig、--namespace 全局参数、环境变量等提供的优先级处理。

- KUBECTL_PLUGINS_DESCRIPTOR_*: 在 plugin.yaml 描述符中声明的每个属性对应的一个环境变量。 例如，KUBECTL_PLUGINS_DESCRIPTOR_NAME ， KUBECTL_PLUGINS_DESCRIPTOR_COMMAND。

- KUBECTL_PLUGINS_GLOBAL_FLAG_*: kubectl 支持的每个全局参数对应的一个环境变量。 例如，KUBECTL_PLUGINS_GLOBAL_FLAG_NAMESPACE ， KUBECTL_PLUGINS_GLOBAL_FLAG_V。

- KUBECTL_PLUGINS_LOCAL_FLAG_*: plugin.yaml 描述符中声明的每个本地参数对应的一个环境变量。例如前面的 targaryen 示例中的 KUBECTL_PLUGINS_LOCAL_FLAG_HEAT。



欢迎加入QQ群：k8s开发与实践（482956822）一起交流k8s技术