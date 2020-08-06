# 总览
PipelineRun允许您实例化并执行集群内管道。管道按所需的执行顺序指定一个或多个任务。 PipelineRun按照指定的顺序在管道中执行任务，直到所有任务成功执行或发生故障为止。

注意：PipelineRun自动为管道中的每个任务创建相应的TaskRun。

`Status`字段跟踪PipelineRun的当前状态，并可用于监视进度。 此字段包含每个TaskRun的状态，以及用于实例化此PipelineRun的完整PipelineSpec，以实现全面的可审核性。

# 配置PipelineRun

- 必要字段：
  - apiVersion-指定API版本。例如tekton.dev/v1beta1。
  - kind-指示此资源对象是PipelineRun对象。
  - 元数据-指定唯一标识PipelineRun对象的元数据。例如，一个名字。
  - spec-指定此PipelineRun对象的配置信息。
  - pipelineRef或pipelineSpec-指定目标管道。
- 可选字段：
  - resources-指定要提供的PipelineResources以执行目标Pipeline。
  - params-为管道指定所需的执行参数。
  - serviceAccountName-指定一个ServiceAccount对象，该对象为管道提供特定的执行凭据。
  - serviceAccountNames-将特定的serviceAccountName值映射到管道中的“任务”。这将覆盖为整个管道设置的凭据。
  - taskRunSpec-指定PipelineRunTaskSpec的列表，该列表允许为每个任务设置ServiceAccountName和Pod模板。这将覆盖整个管道的Pod模板集。
  - 超时-指定PipelineRun失败之前的超时。
  - podTemplate-指定Pod模板，用作执行每个任务的Pod的配置基础。

# 指定目标管道

您必须通过引用现有的Pipeline定义或直接将Pipeline定义嵌入PipelineRun中来指定希望PipelineRun执行的目标Pipeline。

要通过引用指定目标管道，请使用pipelineRef字段：

```
spec:
  pipelineRef:
    name: mypipeline
```

要将Pipeline定义嵌入PipelineRun中，请使用pipelineSpec字段：

```
spec:
  pipelineSpec:
    tasks:
    - name: task1
      taskRef:
        name: mytask
```

pipelineSpec示例示例中的Pipeline显示早晨和晚上的问候。 一旦创建并执行它，就可以检查其Pod的日志：

```
kubectl logs $(kubectl get pods -o name | grep pipelinerun-echo-greetings-echo-good-morning)
Good Morning, Bob!

kubectl logs $(kubectl get pods -o name | grep pipelinerun-echo-greetings-echo-good-night)
Good Night, Bob!
```

您还可以在嵌入的管道定义中嵌入task定义

```
spec:
  pipelineSpec:
    tasks:
    - name: task1
      taskSpec:
        steps:
          ...
```


# 指定资源

管道需要使用PipelineResources为构成它的任务提供输入并存储输出。 您必须在PipelineRun定义的spec部分的资源字段中配置这些资源。

管道可能要求您提供许多不同的资源。 例如：

- 当对拉取请求执行管道时，触发系统必须指定git资源的提交。
- 在针对自己的环境手动执行Pipeline时，必须使用git资源设置GitHub分支； 您使用图像资源的图像注册表； 和您的Kubernetes集群使用集群资源。

您可以使用resourceRef字段引用PipelineResources：

```
spec:
  resources:
    - name: source-repo
      resourceRef:
        name: skaffold-git
    - name: web-image
      resourceRef:
        name: skaffold-image-leeroy-web
    - name: app-image
      resourceRef:
        name: skaffold-image-leeroy-app
```

您还可以使用resourceSpec字段将PipelineResource定义嵌入PipelineRun中：

```
spec:
  resources:
    - name: source-repo
      resourceSpec:
        type: git
        params:
          - name: revision
            value: v0.32.0
          - name: url
            value: https://github.com/GoogleContainerTools/skaffold
    - name: web-image
      resourceSpec:
        type: image
        params:
          - name: url
            value: gcr.io/christiewilson-catfactory/leeroy-web
    - name: app-image
      resourceSpec:
        type: image
        params:
          - name: url
            value: gcr.io/christiewilson-catfactory/leeroy-app
```

## 指定Parameters

您可以指定要在执行期间传递给Pipeline的Parameters，包括Pipeline中不同任务的同一参数的不同值。


spec:
  params:
  - name: pl-param-x
    value: "100"
  - name: pl-param-y
    value: "500"
    
如果需要，您可以根据使用情况传入额外的参数。一个示例用例是您的CI系统自动生成PipelineRun，并且它具有要提供给所有PipelineRun的参数。因为您可以传递额外的参数，所以您不必经历检查每个管道并仅提供所需参数的复杂性。

## 指定自定义ServiceAccount凭据

通过在PipelineRun定义的serviceAccountName字段中指定ServiceAccount对象名称，可以使用一组特定的凭据在PipelineRun中执行Pipeline。如果未明确指定，则PipelineRun创建的TaskRun将使用configmap-defaults ConfigMap中指定的凭据执行。如果未指定此默认值，则TaskRun将使用为目标名称空间设置的默认服务帐户执行。


## 将ServiceAccount凭据映射到任务

如果在指定执行凭据时需要更多粒度，请使用serviceAccountNames字段将特定serviceAccountName值映射到管道中的特定Task。这将覆盖您在上一节中为管道设置的全局serviceAccountName。

例如，如果您指定以下映射：

```
spec:
  serviceAccountName: sa-1
  serviceAccountNames:
    - taskName: build-task
      serviceAccountName: sa-for-build
```

## 指定Pod模板

您可以指定Pod模板配置，该配置将用作Pod的配置起点，您的Tasks中指定的容器映像将在其中执行。 这使您可以专门为每个TaskRun自定义Pod配置。

在以下示例中，任务定义了一个名为my-cache的volumeMount对象。 PipelineRun使用persistentVolumeClaim为任务配置此对象，并以用户1001的身份执行该对象。

```
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: mytask
spec:
  steps:
    - name: writesomething
      image: ubuntu
      command: ["bash", "-c"]
      args: ["echo 'foo' > /my-cache/bar"]
      volumeMounts:
        - name: my-cache
          mountPath: /my-cache
---
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: mypipeline
spec:
  tasks:
    - name: task1
      taskRef:
        name: mytask
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: mypipelinerun
spec:
  pipelineRef:
    name: mypipeline
  podTemplate:
    securityContext:
      runAsNonRoot: true
      runAsUser: 1001
    volumes:
    - name: my-cache
      persistentVolumeClaim:
        claimName: my-volume-claim
```

## 指定taskRunSpecs

指定PipelineTaskRunSpec的列表，其中包含TaskServiceAccountName，TaskPodTemplate和PipelineTaskName。 根据TaskName将规范映射到相应的Task，PipelineTask将与配置的TaskServiceAccountName和TaskPodTemplate一起运行，覆盖管道范围内的ServiceAccountName和podTemplate配置，例如：

```
spec:
   podTemplate:
    securityContext:
      runAsUser: 1000
      runAsGroup: 2000
      fsGroup: 3000
  taskRunSpecs:
    - pipelineTaskName: build-task
      taskServiceAccountName: sa-for-build
      taskPodTemplate:
        nodeSelector:
          disktype: ssd
```

如果与此管道一起使用，则构建任务将使用特定于任务的PodTemplate（其中nodeSelector的磁盘类型等于ssd）

## 指定 Workspaces

如果管道指定一个或多个Workspaces，则必须将这些Workspaces映射到PipelineRun定义中的相应物理卷。例如，可以将PersistentVolumeClaim卷映射到Workspaces，如下所示：

```
workspaces:
- name: myworkspace # must match workspace name in Task
  persistentVolumeClaim:
    claimName: mypvc # this PVC must already exist
  subPath: my-subdir
```

## 指定LimitRange

为了仅消耗从被调用的任务一次执行一个步骤所需的最少资源，Tekton仅从每个步骤中请求CPU，内存和临时存储的最大值。 这足够了，因为步骤只能在Pod中一次执行一个。 最大值以外的请求都设置为零。

当在其中执行PipelineRun的名称空间中存在LimitRange参数并且为容器资源请求指定了最小值时，Tekton将搜索名称空间中存在的所有LimitRange值，并使用最小值而不是0。

## 配置故障超时

您可以使用timeout字段以分钟为单位设置PipelineRun的所需超时值。如果未在PipelineRun中指定此值，则将应用全局默认timeout值。如果将timeout设置为0，则遇到错误时PipelineRun将立即失败。

首次安装Tekton时，全局默认超时设置为60分钟。您可以使用config/config-defaults.yaml中的`default-timeout-minutes`字段设置其他全局默认timeout值。

timeout是符合Go的ParseDuration格式的持续时间。例如，有效值为1h30m，1h，1m和60s。如果将全局超时设置为0，则所有没有单独设置超时的PipelineRun都会在遇到错误时立即失败。

## 监视执行状态
当您执行PipelineRun时，其状态字段会累积有关每个TaskRun以及整个PipelineRun的执行信息。此信息包括与TaskRun关联的管道任务的名称，TaskRun的完整状态以及有关可能与TaskRun关联的条件的详细信息。

`status`|`reason`|`completionTime` is set|Description
:-------|:-------|:---------------------:|--------------:
Unknown|Started|No|The `PipelineRun` has just been picked up by the controller.
Unknown|Running|No|The `PipelineRun` has been validate and started to perform its work.
Unknown|PipelineRunCancelled|No|The user requested the PipelineRun to be cancelled. Cancellation has not be done yet.
True|Succeeded|Yes|The `PipelineRun` completed successfully.
True|Completed|Yes|The `PipelineRun` completed successfully, one or more Tasks were skipped.
False|Failed|Yes|The `PipelineRun` failed because one of the `TaskRuns` failed.
False|\[Error message\]|No|The `PipelineRun` encountered an non-permanent error, but it's still running and it may ultimately succeed.
False|\[Error message\]|Yes|The `PipelineRun` failed with a permanent error (usually validation).
False|PipelineRunCancelled|Yes|The `PipelineRun` was cancelled successfully.
False|PipelineRunTimeout|Yes|The `PipelineRun` timed out.

扫描关注我:

![微信](http://img.rocdu.top/20200703/qrcode_for_gh_7457c3b1bfab_258.jpg)

