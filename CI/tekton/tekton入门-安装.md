# 介绍

Tekton是Kubernetes原生的持续集成和交付CI/CD解决方案。它允许开发人员跨云提供商和本地系统构建、测试和部署

包含以下四个组件

- Pipelines
- triggers
- cli
- dashboard 

## 概念模型

### steps tasks piplines

step是CI/CD工作流中的具体操作
task是step的集合
pipline是tasks的集合

![](http://img.rocdu.top/20200703/concept-tasks-pipelines.png)

### 输入输出

task和pipline可能都有自己的输入输出，在tekton成为输入输出资源

Tekton支持许多不同类型的资源，包括：

- git：一个git仓库
- 提取请求：git存储库中的特定提取请求
- 镜像：容器镜像
- 集群：Kubernetes集群
- 存储：Blob存储中的对象或目录，例如Google Cloud Storage
- CloudEvent：A CloudEvent

![](http://img.rocdu.top/20200703/concept-resources.png)

### TaskRuns和PipelineRuns

pipelineRun，顾名思义，是一个具体的执行流水线，taskRun是一个具体的执行任务

TaskRuns 和 pipelineRuns将资源与task或Pipline连接

可以手动创建taskRuns或pipelineRuns，这会触发Tekton立即运行任务或管道。
或者，可以要求Tekton组件（例如Tekton Triggers）根据需要自动创建运行。例如，您可能希望在每次将新的拉取请求checked 到git仓库

![](http://img.rocdu.top/20200703/concept-runs.png)


## 工作原理

Tekton Pipelines的核心是包装每个task,更具体地说，Tekton Pipelines将entrypoint 二进制文件注入到步骤容器中，该容器将在系统准备就绪时执行您指定的命令。

Tekton Pipelines使用Kubernetes注释跟踪管道的状态,这些注释以Kubernetes Downward API的文件形式映射在每个步骤容器中 。该entrypoint二进制密切关注映射文件，如果一个特定的注释显示为文件才会开始提供的命令。例如，当您要求Tekton在一个任务中连续运行两个步骤时，entrypoint注入第二步容器的二进制文件将闲置等待，直到注释报告第一步容器已成功完成。

此外，Tekton Pipelines调度一些容器在您的task容器之前和之后自动运行，以支持特定的内置功能，例如检索输入资源以及将输出上传到Blob存储解决方案。您还可以通过taskRuns和pipelineRuns跟踪它们的运行状态。在运行task容器之前，系统还执行许多其他操作来设置环境。

# 安装

```
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
```

如果要修改默认的 storageclasses以及pv大小

```
kubectl create configmap config-artifact-pvc \
                         --from-literal=size=10Gi \
                         --from-literal=storageClassName=manual \
                         -o yaml -n tekton-pipelines | kubectl replace -f -
```


如果要修改默认的serviceaccount 

```
kubectl create configmap config-defaults \
                         --from-literal=default-service-account=YOUR-SERVICE-ACCOUNT \
                         -o yaml -n tekton-pipelines | kubectl replace -f -
```


# 安装CLI

```
brew tap tektoncd/tools
brew install tektoncd/tools/tektoncd-cli
```

# 创建一个Task

```
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: hello
spec:
  steps:
    - name: hello
      image: ubuntu
      command:
        - echo
      args:
        - "Hello World!"
```

将上面的YAML写入名为的文件task-hello.yaml，并将其应用于您的Kubernetes集群：

```
kubectl apply -f task-hello.yaml
```

要使用Tekton运行此任务，您需要创建一个TaskRun，这是另一个Kubernetes对象，用于指定的运行时信息Task。

要查看该TaskRun对象，您可以运行以下Tekton CLI（tkn）命令：

```
tkn task start hello --dry-run
```

运行上面的命令后，TaskRun应显示以下定义：

```
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  generateName: hello-run-
spec:
  taskRef:
    name: hello
```
    

运行hello task

```
tkn task start hello
```

使用kubectl 运行task

```
# use tkn's --dry-run option to save the TaskRun to a file
tkn task start hello --dry-run > taskRun-hello.yaml
# create the TaskRun
kubectl create -f taskRun-hello.yaml
```

扫描关注我:

![微信](http://img.rocdu.top/20200703/qrcode_for_gh_7457c3b1bfab_258.jpg)

