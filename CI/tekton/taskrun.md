# taskrun

一个TaskRun定义支持以下领域：

- 需要：
  - apiVersion-指定API版本，例如 tekton.dev/v1beta1。
  - kind-将此资源对象标识为TaskRun对象。
  - metadata-指定TaskRun唯一标识的元数据，例如name。
  - spec-指定TaskRun的配置。
  - taskRef或taskSpec -指定TaskRun将执行的tasks。
- 可选的：
  - serviceAccountName-指定一个ServiceAccount对象，该对象提供TaskRun用于执行的自定义凭据。
  - params-指定Task所需的执行参数。
  - resources-指定所需的PipelineResource值。inputs指定输入资源。outputs指定输出资源。
  - timeout-指定TaskRun失败之前的超时时间。
  - podTemplate-指定一个Pod模板作为起点用于配置Task的Pods。
  - workspaces-指定要用于Workspaces声明的物理卷 Task
  
## 指定目标Task

要在Task中指定要执行的代码，使用taskRef字段：

```
spec:
  taskRef:
    name: read-task
```


您还可以使用taskSpec字段直接在TaskRun中嵌入所需的Task定义：

```
spec:
  taskSpec:
    resources:
      inputs:
        - name: workspace
          type: git
    steps:
      - name: build-and-push
        image: gcr.io/kaniko-project/executor:v0.17.1
        # specifying DOCKER_CONFIG is required to allow kaniko to detect docker credential
        env:
          - name: "DOCKER_CONFIG"
            value: "/tekton/home/.docker/"
        command:
          - /kaniko/executor
        args:
          - --destination=gcr.io/my-project/gohelloworld
```


## 指定参数

如果要执行的task包含参数，可以通过parms字段指定，若task中未设置默认值，则必须指定

```
spec:
  params:
    - name: flags
      value: -someflag
```
      
## 指定resource

如果task中设置了resource则必须指定，可以在taskrun中直接指定或者引用已有的 PiplineResource对象

- 引用已有的PiplineResource

```
spec:
  resources:
    inputs:
      - name: workspace
        resourceRef:
          name: java-git-resource
    outputs:
      - name: image
        resourceRef:
          name: my-app-image
```

- 直接指定

```
spec:
  resources:
    inputs:
      - name: workspace
        resourceSpec:
          type: git
          params:
            - name: url
              value: https://github.com/pivotal-nader-ziada/gohelloworld
```

## 指定podtemplate


您可以指定Pod模板配置，该配置将用作Pod的配置起点，您的Task中指定的容器映像将在其中执行。 这使您可以专门为该TaskRun自定义Pod配置。

在下面的例子中，Task指定了一个volumeMount（my-cache）对象，也被TaskRun提供，采用了PersistentVolumeClaim卷。在该SchedulerName字段中还配置了特定的调度程序 。在Pod与常规（非根）用户权限执行。

```
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: mytask
  namespace: default
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
kind: TaskRun
metadata:
  name: mytaskrun
  namespace: default
spec:
  taskRef:
    name: mytask
  podTemplate:
    schedulerName: volcano
    securityContext:
      runAsNonRoot: true
      runAsUser: 1001
    volumes:
    - name: my-cache
      persistentVolumeClaim:
        claimName: my-volume-claim
```
        
## 指定workspaces

如果task指定一个或多个workspaces，则必须将这些workspaces映射到TaskRun定义中的相应物理卷。例如，可以将PersistentVolumeClaim卷映射到workspaces，如下所示：

workspaces:
- name: myworkspace # must match workspace name in the Task
  persistentVolumeClaim:
    claimName: mypvc # this PVC must already exist
  subPath: my-subdir
  

## 指定sidecar

Sidecar是与Steps任务中指定的容器并行运行的容器，以为执行这些tasks提供辅助支持。例如，Sidecar可以运行日志记录守护程序，更新共享卷上文件的服务或网络代理。

Tekton支持将Sidecar注入到属于TaskRun的Pod中，条件是一旦Task中的所有步骤完成执行，则终止在Pod中运行的每个Sidecar。 这可能会导致Pod包含每个受影响的Sidecar（重试计数为1），这与预期的容器映像不同。

## 指定LimitRange

为了仅消耗从被调用的任务一次执行一个步骤所需的最少资源，Tekton仅从每个步骤中请求CPU，内存和临时存储的最大值。 这足够了，因为步骤只能在Pod中一次执行一个。 最大值以外的请求都设置为零。

当在其中执行TaskRun的名称空间中存在LimitRange参数并且为容器资源请求指定了最小值时，Tekton将搜索名称空间中存在的所有LimitRange值，并使用最小值而不是0。

## 指定失败超时时间


您可以使用`timeout`字段,以分钟为单位设置TaskRun的所需超时值。 如果未在TaskRun中指定此值，则将应用全局默认超时值。 如果将超时设置为0，则TaskRun在遇到错误时立即失败。

首次安装Tekton时，全局默认超时设置为60分钟。 您可以使用config/config-defaults.yaml中的default-timeout-minutes字段设置其他全局默认超时值。

超时值是符合Go的ParseDuration格式的持续时间.例如，有效值为1h30m，1h，1m和60s。 如果将全局超时设置为0，则所有没有单独设置超时的TaskRun都会在遇到错误时立即失败。

### 指定`ServiceAccount`凭据

通过在TaskRun定义的serviceAccountName字段中指定ServiceAccount对象名称，可以使用一组特定的凭据在TaskRun中执行Task。 如果未明确指定，则TaskRun将使用configmap-defaults ConfigMap中指定的凭据执行。如果未指定此默认值，则TaskRuns将使用为目标名称空间设置的默认服务帐户执行。

## 监控执行状态

执行TaskRun时，其状态字段会累积有关每个步骤以及TaskRun整体执行情况的信息。该信息包括开始和停止时间，退出代码，容器映像的全限定名称以及相应的摘要。

> Note! 如果Kubernetes已经OOM杀死了任何Pod，即使TaskRun的退出代码为0也被标记为失败。

下表显示了如何读取TaskRun的总体状态：

`status`|`reason`|`completionTime` is set|Description
:-------|:-------|:---------------------:|:--------------
Unknown|Started|No|TaskRun刚刚被控制器拉取
Unknown|Pending|No|TaskRun正在等待状态为Pod的Pod
Unknown|Running|No|TaskRun已通过验证并开始执行其工作
Unknown|TaskRunCancelled|No|用户请求取消TaskRun。取消尚未完成。
True|Succeeded|Yes|TaskRun成功完成
False|Failed|Yes|TaskRun失败，因为步骤之一失败
False|\[Error message\]|No|TaskRun遇到非永久错误，并且仍在运行。它可能最终会成功.
False|\[Error message\]|Yes|TaskRun因永久错误而失败（通常是验证）
False|TaskRunCancelled|Yes|TaskRun已成功取消.
False|TaskRunTimeout|Yes|TaskRun超时.

### 监控steps

如果在TaskRun调用的任务中定义了多个步骤，则可以使用以下命令在steps.results字段中监视其执行状态，其中<name>是目标TaskRun的名称：

状态中还包含用于实例化TaskRun的确切任务规范，以实现全面的可审核性。

### steps
相应的状态以在任务定义中指定步骤的顺序显示在status.steps列表中。

### 监测结果

如果在调用的任务中指定了一个或多个结果字段，则TaskRun的执行状态将包括`Task Results`部分，其中`Results`逐字显示，包括原始行返回和空格。 例如：

```
Status:
  # […]
  Steps:
  # […]
  Task Results:
    Name:   current-date-human-readable
    Value:  Thu Jan 23 16:29:06 UTC 2020

    Name:   current-date-unix-timestamp
    Value:  1579796946
```

### 取消TaskRun

要取消当前正在执行的TaskRun，请更新其定义以将其标记为已取消。这样做时，与该TaskRun相关联的所有正在运行的Pod都将被删除。例如：

```
apiVersion: tekton.dev/v1alpha1
kind: TaskRun
metadata:
  name: go-example-git
spec:
  # […]
  status: "TaskRunCancelled"
```

扫描关注我:

![微信](http://img.rocdu.top/20200703/qrcode_for_gh_7457c3b1bfab_258.jpg)

