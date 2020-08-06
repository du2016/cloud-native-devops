# 总览

Workspaces允许Tasks声明TaskRuns在运行时需要提供的文件系统部分. TaskRun可以通过多种方式使文件系统的这些部分可用: 使用只读的ConfigMap或Secret，与其他Task共享的现有PersistentVolumeClaim，从提供的VolumeClaimTemplate创建PersistentVolumeClaim，或者仅是在TaskRun完成后被丢弃的emptyDir.

Workspaces 与Volumes类似，不同之处在于 Workspaces 允许Tasks作者在决定使用哪种存储类别时服从用户及其TaskRun.

Workspaces 可以用于以下目的：

- 输入和/或输出的存储
- 在Tasks之间共享数据
- Secrets中保存的凭证的挂载点
- ConfigMap中保存的配置的挂载点
- 组织共享的常用工具的挂载点
- 大量的构建artifacts可以加快工作速度

## Tasks 和TaskRun中的 Workspaces

Tasks指定Workspaces在其Steps在磁盘上的位置. 在运行时，TaskRun提供安装到该Workspaces中的卷的特定详细信息.

关注点的分离提供了很大的灵活性.例如，在隔离环境中，一个TaskRun可能只提供一个emptyDir卷，该卷会快速挂载并在运行结束时消失. 但是，在更复杂的系统中，TaskRun可能会使用PersistentVolumeClaim，该数据已预先填充了要处理的Task数据. 在这两种情况下，Tasks的Workspaces声明均保持不变，并且仅TaskRun中的运行时信息发生更改.

## Pipelines 和 PipelineRuns 的 Workspaces


Pipelines可以使用Workspaces来显示如何通过其Tasks共享存储. 例如，TasksA可能会将源仓库克隆到 Workspaces中，而Task B可能会编译它在该Workspaces中找到的代码. 确保这两个Tasks使用的Workspaces相同是Pipelines的工作，更重要的是，确保它们访问Workspaces的顺序正确.

PipelineRun执行与TaskRun几乎相同的Tasks-它们提供特定的Volume信息以用于每个Pipeline所使用的 Workspaces. PipelineRun具有确保在多个Tasks之间安全正确地共享它们提供的任何卷类型的附加责任.

# 配置Workspaces

本节介绍如何在TaskRun中配置一个或多个Workspaces

## 在Tasks中使用Workspaces


要在Tasks中配置一个或多个Workspaces，请使用以下字段为每个条目添加一个Workspaces列表：

- 名称 - 必需)可用于引用Workspaces的唯一字符串标识符
- description - 信息字符串，描述Workspaces的用途
- readOnly - 一个布尔值，声明Task是否将写入Workspaces.
- mountPath- Workspaces提供给Steps使用时映射到磁盘上某个位置的路径.相对路径将以/workspace开头.如果未提供mountPath，则默认情况下会将Workspaces放置在/workspace/<name>中，其中<name>是Workspaces的唯一名称.

请注意以下几点：

- Task定义可以包含所需的多个Workspaces.建议Task最多使用一个可写Workspaces.
- readOnly Workspaces将其卷挂载为只读.尝试写入readOnly Workspaces将导致错误和失败的TaskRun.
- mountPath可以是绝对的，也可以是相对的.绝对路径以/开头，相对路径以目录名开头.例如，/foobar的mountPath是绝对的，并且在Tasks步骤内的/foobar处公开Workspaces，而foobar的mountPath是相对的，并且在/workspace/foobar处公开Workspaces.

下面是一个示例Task定义，其中包括一个称为message的Workspace，Task向该消息写入一条消息

```
spec:
  steps:
  - name: write-message
    image: ubuntu
    script: |
      #!/usr/bin/env bash
      set -xe
      echo hello! > $(workspaces.messages.path)/message
  workspaces:
  - name: messages
    description: The folder where we write the message to
    mountPath: /custom/path/relative/to/root
```

### 在Tasks中使用Workspaces变量

以下变量可用于Tasks的有关 Workspaces 的信息：

- $(workspaces.<name>.path)-指定Workspaces的路径，其中<name>是Workspaces的名称.
- $(workspaces.<name>.claim)-指定用作Workspaces的卷源的PersistentVolumeClaim的名称，其中<name>是Workspaces的名称.如果使用了PersistentVolumeClaim以外的其他卷源，则返回一个空字符串.
- $(workspaces.<名称>.volume)-指定为Workspaces提供的卷的名称，其中<name>是Workspaces的名称.
将Tasks中的Workspaces映射到TaskRun
执行包含Workspaces列表的Tasks的TaskRun必须将这些Workspaces绑定到实际的物理卷.为此，TaskRun包含其自己的Workspaces列表.列表中的每个条目都包含以下字段：

name-(必需)为其提供卷的Task中的Workspaces的名称
subPath-卷上的一个可选子目录，用于存储该Workspaces的数据
该条目还必须包含一个VolumeSource.有关更多信息，请参见在Workspaces中使用VolumeSources.

注意：-在执行TaskRun之前，卷上必须存在subPath，否则执行将失败. -执行关联的TaskRun时，在Task中声明的Workspaces必须可用.否则，TaskRun将失败.

### 使用Workspaces的TaskRun定义的示例

以下示例说明了如何在TaskRun定义中指定Workspaces，为Task的Workspacesmyworkspace提供了emptyDir：

```
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  generateName: example-taskrun-
spec:
  taskRef:
    name: example-task
  workspaces:
    - name: myworkspace # this workspace name must be declared in the Task
      emptyDir: {}      # emptyDir volumes can be used for TaskRuns, but consider using a PersistentVolumeClaim for PipelineRuns
```

### 在Pipelines中使用Workspaces

当各个Tasks声明需要运行的Workspaces时，Pipelines将决定在其Tasks之间共享哪些Workspaces.要在Pipelines中声明共享Workspaces，必须在Pipelines定义中添加以下信息：

- 您的PipelineRun提供的Workspaces列表.使用Workspaces字段在Pipelines定义中指定目标Workspaces，如下所示.列表中的每个条目都必须具有唯一的名称。

- Pipelines和Tasks定义之间的Workspaces名称映射。

下面的示例定义了一个Pipelines，其中包含一个名为pipeline-ws1的Workspaces.此Workspaces绑定在两个Tasks中-首先是由gen-code Task声明的输出Workspaces，然后是作为commit Task声明的src Workspaces.如果PipelineRun提供的Workspaces是PersistentVolumeClaim，则这两个Tasks可以在该Workspaces中共享数据。


spec:
  workspaces:
    - name: pipeline-ws1 # Name of the workspace in the Pipeline
  tasks:
    - name: use-ws-from-pipeline
      taskRef:
        name: gen-code # gen-code expects a workspace named "output"
      workspaces:
        - name: output
          workspace: pipeline-ws1
    - name: use-ws-again
      taskRef:
        name: commit # commit expects a workspace named "src"
      workspaces:
        - name: src
          workspace: pipeline-ws1
      runAfter:
        - use-ws-from-pipeline # important: use-ws-from-pipeline writes to the workspace first
        
在Workspaces绑定中包含一个 `subPath`，以针对不同的Tasks挂载同一卷的不同部分。请参阅此类Pipelines的完整示例，该示例将数据写入同一卷上的两个相邻目录。

在Pipeline中指定的subPath将被追加到在PipelineRunWorkspaces声明中指定的任何subPath。因此，如果PipelineRun声明了将Pipelines绑定到带有/bar subPath的Task的Pipelines，则声明带有/foo subPath的Workspaces将最终挂载该卷的/foo/bar目录。

# Affinity Assistant 和 在Pipelines中指定Workspaces顺序

在Tasks之间共享Workspaces要求您定义这些Tasks写入或读取该Workspaces的顺序。使用Pipelines定义中的runAfter字段来定义何时应执行Tasks。有关更多信息，请参见runAfter文档。

当PersistentVolumeClaim用作PipelineRun中Workspaces的卷源时，将创建一个Affinity Assistant.Affinity Assistant充当共享同一Workspaces的TaskRun pod的占位符。共享Workspace的PipelineRun中所有TaskRun pod都将与Affinity Assistant pod调度到同一节点。这意味着Affinity Assistant与例如为TaskRun pod配置的其他相似性规则。如果PipelineRun具有配置的自定义PodTemplate，则还将在Affinity Assistant pod上设置NodeSelector和Tolerations字段.PipelineRun完成后，将删除Affinity Assistant。可以通过设置禁用Affinity Assistant功能门来禁用Affinity Assistant。

> 注意：Affinity Assistant使用的Pod间亲和力和反亲和力需要大量处理，这可能会大大减慢大型集群中的调度。我们不建议在超过数百个节点的集群中使用它们

> 注意：Pod反亲和性要求对节点进行一致的标记，换句话说，集群中的每个节点都必须具有与拓扑关键字匹配的适当标签。如果某些或所有节点缺少指定的topologyKey标签，则可能导致意外行为。

### 在PipelineRun中指定Workspaces
为了使PipelineRun执行包含一个或多个Workspaces的Pipelines，需要绑定该Workspaces名称到卷使用其自己的workspaces字段。该列表中的每个条目必须对应于Pipelines中的Workspaces声明。Workspaces列表中的每个条目都必须指定以下内容：

- 名称-（必需）在为其提供卷的Pipelines定义中指定的Workspaces的名称。
- subPath-（可选）卷上将存储该Workspace数据的目录。执行TaskRun时此目录必须存在，否则执行将失败。

该条目还必须包含一个VolumeSource。有关更多信息，请参见在Workspaces中使用VolumeSources。

> 注意：如果PipelineRun在运行时未提供Pipelines指定的Workspaces，则该PipelineRun将失败。

### 使用Workspaces的示例PipelineRun定义

在下面的示例中，提供了volumeClaimTemplate来说明如何为在Pipelines中声明的名为myworkspace的Workspaces创建PersistentVolumeClaim。使用volumeClaimTemplate时，将为每个PipelineRun创建一个新的PersistentVolumeClaim，并允许用户指定例如卷的大小和StorageClass。

```
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: example-pipelinerun-
spec:
  pipelineRef:
    name: example-pipeline
  workspaces:
    - name: myworkspace # this workspace name must be declared in the Pipeline
      volumeClaimTemplate:
        spec:
          accessModes:
            - ReadWriteOnce # access mode may affect how you can use this volume in parallel tasks
          resources:
            requests:
              storage: 1Gi
```

# 在Workspaces中指定VolumeSources

每个Workspaces条目只能使用一种类型的VolumeSource。每种类型的配置选项都不同。Workspaces支持以下字段：


## 使用PersistentVolumeClaims作为VolumeSource

PersistentVolumeClaim卷是在Pipelines内的Tasks之间共享数据的理想选择。请注意，为PersistentVolumeClaim配置的访问模式会影响如何将卷用于Pipelines中的并行Tasks。有关更多信息，请参见在Pipelines中指定Workspaces顺序。有两种方法可以将PersistentVolumeClaims用作VolumeSource。

### volumeClaimTemplate

volumeClaimTemplate是为每个PipelineRun或TaskRun创建的PersistentVolumeClaim卷的模板。当从PipelineRun或TaskRun中的模板创建卷时，将在删除PipelineRun或TaskRun时将其删除。

```
workspaces:
- name: myworkspace
  volumeClaimTemplate:
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
```

### persistentVolumeClaim

persistentVolumeClaim字段引用现有的persistentVolumeClaim卷.该示例仅公开该PersistentVolumeClaim的子目录my-subdir

```
workspaces:
- name: myworkspace
  persistentVolumeClaim:
    claimName: mypvc
  subPath: my-subdir
```
  
## 使用其他类型的VolumeSource

### emptyDir

emptyDir字段引用一个emptyDir卷，该卷包含一个临时目录，该目录仅在调用它的TaskRun中存在.emptyDir卷不适合在Pipelines内的Tasks之间共享数据.但是，它们对于单个TaskRun效果很好，其中需要在Task的各个步骤之间共享emptyDir中存储的数据，并在执行后将其丢弃。

```
workspaces:
- name: myworkspace
  emptyDir: {}
```
  
### configMap

configMap字段引用configMap卷.将configMap用作Workspaces具有以下限制：

configMap卷源始终安装为只读.步骤无法向其写入内容，如果尝试则将出错。
您要用作Workspaces的configMap必须在提交TaskRun之前存在。
configMap的大小限制为1MB。

```
workspaces:
- name: myworkspace
  configmap:
    name: my-configmap
```
    
### secret
secret 字段引用secret卷.使用secret卷具有以下限制：

secret卷源始终安装为只读.步骤无法向其写入内容，如果尝试则将出错。
您要用作Workspaces的secret必须在提交TaskRun之前存在。
secret的大小限制为1MB。
    
```
workspaces:
- name: myworkspace
  secret:
    secretName: my-secret
```