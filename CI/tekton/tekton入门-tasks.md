# 介绍

task是steps的集合，可以在持续集成流程中按照特定的顺序执行，task在k8s集群中以pod的方式运行，task可以在其明明空间中可用，clustertask可以在集群范围内使用


# Task配置

task配置字段如下

- 必选：
  - apiVersion-指定API版本。例如，tekton.dev/v1beta1。
  - kind-将此资源对象标识为Task对象。
  - metadata-指定资源对象元数据的唯一标识。例如name。
  - spec-指定此Task资源对象的配置信息。
  - steps-指定要在Task中运行的一个或多个容器镜像。
- 可选的：
  - description-Task的信息描述。
  - params-指定Task的执行参数。
  - resources-仅用于alpha,指定您的任务需要或创建的PipelineResources。
  - inputs-指定Task中提取的资源。
  - outputs-指定Task产生的资源。
  - workspaces-指定Task所需卷的路径。
  - results-指定Tasks将其执行结果写入的文件。
  - volumes-指定一个或多个卷，被task中的step访问。
  - stepTemplate-制定所有Task中step 容器所需的基础选项。
  - sidecars-制定Task中与step容器一起运行的Sidecar，。

## ClusterTask

ClusterTask作用域为集群，Task作用域为命名空间。若要在pipline中使用ClusterTask，需要制指定其类型为 `kind: ClusterTask`

## Steps定义

Steps是对容器镜像的引用，该容器镜像通过input产生特定output，要将Steps添加到Task你需要定义 一个steps字段包含一系列step, step根据其排列顺序决定执行顺序。

对于steps中的容器需要满足以下条件：

- 容器镜像必须满足容器镜像合约
- 每个容器都将运行到第一次运行出现故障为止
- 如果容器镜像在任务中的所有容器镜像中没有最大的资源请求，则CPU、内存和临时存储资源请求将设置为零，或者，如果指定，则设置为通过该命名空间中的LimitRanges设置的最小值。这可以确保执行任务的Pod只请求足够的资源来运行任务中的单个容器镜像，而不是一次为任务中的所有容器镜像累计资源

### 保留目录

Tekton运行的所有任务都有几个目录将被视为特殊目录

- /workspace-此目录是资源和工作空间的安装目录。通过变量替换，任务作者可以使用这些路径
- /tekton-此目录用于Tekton特定功能：
  /tekton/results是写入结果的位置，任务作者可以通过$(results.name.path)使用该路径还有其他子文件夹是Tekton的实现细节，用户不应依赖其特定行为，因为将来可能会更改

### 在Step中运行脚本

step可以指定script字段，其包含了一个脚本的主体，使用该脚本就像该脚本存储在容器中一样，所有的参数都将传递给该脚本。该参数与command字段互斥。

如果脚本中不指定 shebang,则默认指定为：

```
#!/bin/sh
set -xe
```

您可以通过在前面指定指定的解析器的shebang来覆盖此默认前导。该解析器必须存在于该步骤的容器镜像中。

以下实例是一个 bash 脚本：

```
steps:
- image: ubuntu  # contains bash
  script: |
    #!/usr/bin/env bash
    echo "Hello from Bash!"
```

## 指定 Parameters

您可以指定要在执行时提供给Task的参数，例如编译标志或工件名称。参数从其对应的TaskRun传递给Task。

参数名称需要满足以下条件：

- 必须仅包含字母数字字符，连字符(-)和下划线(_)
- 必须以字母或下划线(_)开头。


每个生命都有一个type字段，可以设置为array或者string，默认为string

以下示例展示了通过定义param传入container args：

```
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: task-with-parameters
spec:
  params:
    - name: flags
      type: array
    - name: someURL
      type: string
  steps:
    - name: build
      image: my-builder
      args: ["build", "$(params.flags[*])", "url=$(params.someURL)"]
```

在taskrun中传入task parm

```
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: run-with-parameters
spec:
  taskRef:
    name: task-with-parameters
  params:
    - name: flags
      value:
        - "--set"
        - "arg1=foo"
        - "--randomflag"
        - "--someotherflag"
    - name: someURL
      value: "http://google.com"
```

## 指定Resource

通过中可以指定 PipelineResources 实体用于定义输入和输出资源

使用input字段为任务提供所需要执行的上下文或数据，如果任务的输出是下一个任务的输入，则必须在 `/workspace/output/resource_name/`处使用该数据，例如：



> 注意： 如果task依赖于输出资源，则 task step字段中的容器无法在路径/workspace/output上挂载任何内容
>
```

apiVersion: tekton.dev/v1alpha1
kind: PipelineResource
metadata:
  name: test
  namespace: default
spec:
  params:
  - name: url
    value: https://github.com/du2016/jaeger-doc-zh
  type: git

---

apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: test-res
spec:
  resources:
    inputs:
    - name: tar-artifact  # 默认的容器内挂载路径
      targetPath: customworkspace #覆盖默认
      type: git
    outputs:
    - name: tar-artifact
      type: git
  steps:
   - name: untar
     image: ubuntu
     command: ["/bin/bash"]
     args: ['-c', 'mkdir -p /workspace/tar-scratch-space/ && tar -xvf /workspace/customworkspace/rules_docker-master.tar -C /workspace/tar-scratch-space/']
   - name: edit-tar
     image: ubuntu
     command: ["/bin/bash"]
     args: ['-c', 'echo crazy > /workspace/tar-scratch-space/rules_docker-master/crazy.txt']
   - name: tar-it-up
     image: ubuntu
     command: ["/bin/bash"]
     args: ['-c', 'cd /workspace/tar-scratch-space/ && tar -cvf /workspace/customworkspace/rules_docker-master.tar rules_docker-master']
```


## 指定workspace

workspace 允许指定task运行期间需要的一个或者多个卷,建议最多使用一个可写卷

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

## 发出结果

使用results字段可以指定一个或多个文件来存储其执行结果，这些文件存储在/tekton/results中，如果results中指定了文件，则该目录自动创建，要指定文件，需要指定name和
description字段


以下实例指定了两个result文件：

```
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: print-date
  annotations:
    description: |
      A simple task that prints the date
spec:
  results:
    - name: current-date-unix-timestamp
      description: The current date in unix timestamp format
    - name: current-date-human-readable
      description: The current date in human readable format
  steps:
    - name: print-date-unix-timestamp
      image: bash:latest
      script: |
        #!/usr/bin/env bash
        date +%s | tee /tekton/results/current-date-unix-timestamp
    - name: print-date-human-readable
      image: bash:latest
      script: |
        #!/usr/bin/env bash
        date | tee /tekton/results/current-date-human-readable
```

结果可以在task或pipline中使用

> task result的最大值受k8s container termination message 功能限制，目前限制为4096字节，结果被写入编码为json的终止消息中。可以通过kubectl describe 看到


## 指定volumes

除了指定输入和输出外，还可以为task中的step指定一个或多个volume

可以通过volume执行以下操作：

- 挂载k8s secret
- 创建一个emptydir从而为多个steps共享临时数据
- 将configmap作为挂载源
- 将宿主机的dockersocket挂载进容器从而将Dockerfile构建为镜像。不过建议使用 Google出的kaniko，脱离dockerd构建镜像

## 指定一个step template

stepTemplate 字段指定 容器配置作为所有steps的起点，当发生冲突时，template中的配置将被step中的配置覆盖

```
stepTemplate:
  env:
    - name: "FOO"
      value: "bar"
steps:
  - image: ubuntu
    command: [echo]
    args: ["FOO is ${FOO}"]
  - image: ubuntu
    command: [echo]
    args: ["FOO is ${FOO}"]
    env:
      - name: "FOO"
        value: "baz"
```


## 指定 sidecar

sidecar字段指定一系列与task中step一起运行的容器，可以通过sidecar实现很多功能例如 docker in docker或者在测试时运行一个mock apiserver，sidecar容器早于 task执行，并在task执行完成后删除，sidecar也指定script字段运行脚本

如下，通过sidecar字段实现了docker-in-docker 功能
```
steps:
  - image: docker
    name: client
    script: |
        #!/usr/bin/env bash
        cat > Dockerfile << EOF
        FROM ubuntu
        RUN apt-get update
        ENTRYPOINT ["echo", "hello"]
        EOF
        docker build -t hello . && docker run hello
        docker images
    volumeMounts:
      - mountPath: /var/run/
        name: dind-socket
sidecars:
  - image: docker:18.05-dind
    name: server
    securityContext:
      privileged: true
    volumeMounts:
      - mountPath: /var/lib/docker
        name: dind-storage
      - mountPath: /var/run/
        name: dind-socket
volumes:
  - name: dind-storage
    emptyDir: {}
  - name: dind-socket
    emptyDir: {}
```

> 如果sidecar在接受停止信号时正在执行命令，sidecar会继续运行从而导致task执行失败

## 变量替换

### params和resources可以通过变量替换

- shell $(params.<name>)  获取param
- $(outputs.resources.<name>.<key>)  在Task中获取resource
- $(resources.<name>.<key>)  在 Condition中获取resource
- $(resources.inputs.<name>.path) 获取本地资源路径

### 替换数组参数

可以使用*运算符扩展array参数，为此，请将[*]添加到参数，以将该数组插入到引用的位置

例如`["first", "$(params.array-param[*])", "last"]` 可以转化为 `["first", "some", "array", "elements", "last"]`


必须在完全孤立的字段才可以引用array参数。array以任何其他方式引用参数将导致错误。例如，如果build-args是array类型的参数，则以下示例是无效的，因为该字符串在step中未隔离：

```
 - name: build-step
      image: gcr.io/cloud-builders/some-image
      args: ["build", "additionalArg $(params.build-args[*])"]
```


在非array中引用build-args也是不允许的

```
 - name: build-step
      image: "$(params.build-args[*])"
      args: ["build", "args"]
```

有效的引用build-args如下：

```
 - name: build-step
      image: gcr.io/cloud-builders/some-image
      args: ["build", "$(params.build-args[*])", "additionalArg"]
```


### 替换workspace路径

您可以按以下方式替换任务中指定的工作区的路径：

```
$(workspaces.myworkspace.path)
```


由于卷名是随机的，并且仅在执行任务时设置，因此您也可以按以下方式替换卷名：

```
$(workspaces.myworkspace.volume)
```


您可以通过参数化卷名称和类型来代替它们。 Tekton支持流行的卷类型，例如ConfigMap，Secret和PersistentVolumeClaim。 

扫描关注我:

![微信](http://img.rocdu.top/20200703/qrcode_for_gh_7457c3b1bfab_258.jpg)

