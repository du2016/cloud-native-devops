# pipline总览

PipelineRun允许您实例化并执行集群上的。A 以所需的执行顺序Pipeline指定一个或多个Tasks。一个PipelineRun 执行Tasks中Pipeline，直到所有的顺序，他们被指定Tasks已成功执行或出现故障。

注意： A会PipelineRun自动TaskRuns为中的每个 创建对应Task的内容Pipeline。

该Status字段跟踪的当前状态PipelineRun，并可用于监视进度。此字段包含每个的状态TaskRun，以及PipelineSpec用于实例化此状态的完整状态，以实现PipelineRun完全可审核性。

# 配置 PipelineRun

一个Pipeline定义支持以下领域：

- 需要：
  - apiVersion-指定API版本，例如 tekton.dev/v1beta1。
  - kind-将此资源对象标识为Pipeline对象。
  - metadata-指定唯一标识Pipeline对象的元数据 。例如，一个name。
  - spec-指定此Pipeline对象的配置信息。必须包括：
  - tasks-指定Tasks组成的Pipeline 以及其执行的详细信息。
- 可选的：
  - resources- 仅alpha,指定 PipelineResources由Tasks组成的需要或创建Pipeline。
  - tasks:
    - resources.inputs/resource.outputs
    - from-表示a的数据PipelineResource 来自previous的输出Task。
    - runAfter-表示a Task 应该在其他一个或多个之后执行，Tasks而不进行输出链接。
    - retries-指定Task失败后重试执行的次数。不适用于执行取消。
    - conditions-指定Conditions仅Task 在成功评估后才允许执行。
    - timeout-指定Task失败之前的超时。
  - results-指定Pipeline发出执行结果的位置。
  - description-包含Pipeline对象的详细说明。

# 指定 Resources

pipline需要piplineresources为tasks提供输入存储输出,您可以在Pipeline定义的spec部分的resources字段中声明它们。每个条目都需要唯一的名称和类型。例如：

```
spec:
  resources:
    - name: my-repo
      type: git
    - name: my-image
      type: image
```

## 指定 Workspaces

Workspaces允许指定一个或多个pipline中task运行时所需的volme

```
spec:
  workspaces:
    - name: pipeline-ws1 # The name of the workspace in the Pipeline
  tasks:
    - name: use-ws-from-pipeline
      taskRef:
        name: gen-code # gen-code expects a workspace with name "output"
      workspaces:
        - name: output
          workspace: pipeline-ws1
    - name: use-ws-again
      taskRef:
        name: commit # commit expects a workspace with name "src"
      runAfter:
        - use-ws-from-pipeline # important: use-ws-from-pipeline writes to the workspace first
      workspaces:
        - name: src
          workspace: pipeline-ws1
```


# 指定Parameters

您可以指定要在执行时提供给pipline的全局参数，例如编译flags或artifact名称。参数从其对应的PipelineRun传递到Pipeline，并且可以替换管道中每个Task中指定的模板值。

参数名称：
- 必须仅包含字母数字字符，连字符(-)和下划线(-)
- 必须以字母或下划线(-)开头。

例如，fooIs-Bar_是有效的参数名称，而barIsBa$或0banana不是。

每个声明的参数都有一个类型字段，可以将其设置为数组或字符串。如果在整个执行过程中提供给管道的编译标志的数量不同，则array很有用。如果未指定任何值，则类型字段默认为字符串。提供实际参数值时，将根据类型字段验证其解析的类型。参数的说明和默认字段是可选的。

以下示例说明了Pipeline中Parameters的用法。

以下Pipeline声明了一个名为context的输入参数，并将其值传递给Task以在Task中设置pathToContext参数的值。如果您为默认字段指定一个值，并在PipelineRun中调用此Pipeline而不为上下文指定值，则将使用该值。

```
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: pipeline-with-parameters
spec:
  params:
    - name: context
      type: string
      description: Path to context
      default: /some/where/or/other
  tasks:
    - name: build-skaffold-web
      taskRef:
        name: build-push
      params:
        - name: pathToDockerFile
          value: Dockerfile
        - name: pathToContext
          value: "$(params.context)"
```


以下PipelineRun为上下文提供值

```
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pipelinerun-with-parameters
spec:
  pipelineRef:
    name: pipeline-with-parameters
  params:
    - name: "context"
      value: "/workspace/examples/microservices/leeroy-web"
```


# 为Pipeline添加tasks

您的Pipeline定义必须至少引用一个Task。每个Task内Pipeline必须有一个有效的 name和一个taskRef。例如：

```
tasks:
  - name: build-the-image
    taskRef:
      name: build-push
```

您可以使用PipelineResources作为Tasks中Pipeline的输入和输出。例如：

```
spec:
  tasks:
    - name: build-the-image
      taskRef:
        name: build-push
      resources:
        inputs:
          - name: workspace
            resource: my-repo
        outputs:
          - name: image
            resource: my-image
```

您还可以提供Parameters：

```
spec:
  tasks:
    - name: build-skaffold-web
      taskRef:
        name: build-push
      params:
        - name: pathToDockerFile
          value: Dockerfile
        - name: pathToContext
          value: /workspace/examples/microservices/leeroy-web
```

## 使用from参数

如果管道中的任务需要使用先前任务的输出作为其输入，请使用可选的from参数来指定必须在将其输出作为其输入的任务之前执行的任务列表。 当目标任务执行时，仅使用此列表中最后一个任务生成的所需PipelineResource的版本。 此输出PipelineReource输出的名称必须与在提取它的Task中指定的输入PipelineResource的名称匹配。

在下面的示例中，deploy-app任务提取名为my-image的build-app任务的输出作为其输入。 因此，无论在管道中声明这些任务的顺序如何，build-app Task都将在deploy-app Task之前执行。

```
- name: build-app
  taskRef:
    name: build-push
  resources:
    outputs:
      - name: image
        resource: my-image
- name: deploy-app
  taskRef:
    name: deploy-kubectl
  resources:
    inputs:
      - name: image
        resource: my-image
        from:
          - build-app
```

## 使用runAfter参数

如果您需要任务在管道中按特定顺序执行，但是它们没有需要from参数的资源依赖关系，请使用runAfter参数指示任务必须在其他一个或多个任务之后执行。

在下面的示例中，我们要在构建代码之前对其进行测试。 由于test-app Task没有输出，因此build-app Task使用runAfter来指示test-app必须在其之前运行，而不管它们在管道定义中的引用顺序如何。

```
- name: test-app
  taskRef:
    name: make-test
  resources:
    inputs:
      - name: workspace
        resource: my-repo
- name: build-app
  taskRef:
    name: kaniko-build
  runAfter:
    - test-app
  resources:
    inputs:
      - name: workspace
        resource: my-repo
```


## 使用重试参数

对于管道中的每个任务，您可以指定Tekton失败时应重试其执行的次数。 当任务失败时，相应的TaskRun将其成功条件设置为False。 retries参数指示Tekton在发生这种情况时重试执行任务。

如果您希望Task在执行过程中遇到问题（例如，您知道网络连接性或缺少依赖项会出现问题），请将其retries参数设置为大于0的合适值。如果您未明确指定值 ，Tekton不会尝试再次执行失败的任务。

在下面的示例中，构建映像任务的执行将在失败后重试一次。 如果重试的执行也失败，则任务执行整体会失败。

```
tasks:
  - name: build-the-image
    retries: 1
    taskRef:
      name: build-push
```

## 使用Conditions保护task执行

要仅在满足某些条件时运行任务，可以使用条件字段来保护任务执行。 条件字段使您可以列出对条件资源的一系列引用。 声明的条件在任务运行之前运行。 如果所有条件都成功评估，则运行任务。 如果任何条件失败，则不运行任务，并且TaskRun状态字段ConditionSucceeded设置为False，其原因设置为ConditionCheckFailed。

如下示例，is-master-branch依赖于Conditions资源，deploy task将仅在is-master-branch条件校验成功时执行

```
apiVersion: tekton.dev/v1alpha1
kind: Condition
metadata:
  name: is-master-branch
spec:
  params:
    - name: branch-name
      default: master
  check:
    image: alpine
    git rev-parse --verify $(params.branch)
---

tasks:
  - name: deploy-if-branch-is-master
    conditions:
      - conditionRef: is-master-branch
        params:
          - name: branch-name
            value: master
    taskRef:
      name: deploy
```

与常规任务失败不同，条件失败不会自动使整个PipelineRun失败
- 仍然运行不依赖于Task（通过from或runAfter）的其他任务。

## 配置故障超时

您可以使用管道内“Task”规范中的“Timeout”字段来设置执行该管道的“TaskRun”中执行该任务的TaskRun的超时。 超时值是符合Go的ParseDuration格式的持续时间。 例如，有效值为1h30m，1h，1m和60s。

```
spec:
  tasks:
    - name: build-the-image
      taskRef:
        name: build-push
      Timeout: "0h1m30s"
```
      
## 在Task级别配置执行结果

任务在执行时可以发出结果.您可以通过变量替换将这些结果值用作管道中后续任务中的参数值。 Tekton推断任务顺序，以便发出引用结果结果的任务在消耗它们的任务之前执行。

在下面的示例中，先前任务名称Task的结果声明为bar-result：

```
params:
  - name: foo
    value: "$(tasks.previous-task-name.results.bar-result)"
```
    
# 在Pipeline级别配置执行结果

您可以将管道配置为在执行期间发出结果，以引用其中每个任务发出的结果。

在下面的示例中，管道使用名称sum指定一个结果条目，该名称引用第二个添加Task发出的结果

```
results:
    - name: sum
      description: the sum of all three operands
      value: $(tasks.second-add.results.sum)
```


# 配置task执行顺序

您可以在管道中连接任务，以便它们在有向非循环图（DAG）中执行。 流水线中的每个任务都成为图上的一个节点，可以与一条边相连，这样一个任务将先于另一个边运行，并且流水线的执行进度到完成而不会陷入无限循环。

使用以下命令完成此操作：

- 每个任务使用的PipelineResources上的from子句
- 相应任务上的runAfter子句
- 通过将一个任务的结果链接到另一个任务的参数

例如，管道定义如下

```
- name: lint-repo
  taskRef:
    name: pylint
  resources:
    inputs:
      - name: workspace
        resource: my-repo
- name: test-app
  taskRef:
    name: make-test
  resources:
    inputs:
      - name: workspace
        resource: my-repo
- name: build-app
  taskRef:
    name: kaniko-build-app
  runAfter:
    - test-app
  resources:
    inputs:
      - name: workspace
        resource: my-repo
    outputs:
      - name: image
        resource: my-app-image
- name: build-frontend
  taskRef:
    name: kaniko-build-frontend
  runAfter:
    - test-app
  resources:
    inputs:
      - name: workspace
        resource: my-repo
    outputs:
      - name: image
        resource: my-frontend-image
- name: deploy-all
  taskRef:
    name: deploy-kubectl
  resources:
    inputs:
      - name: my-app-image
        resource: my-app-image
        from:
          - build-app
      - name: my-frontend-image
        resource: my-frontend-image
        from:
          - build-frontend
```

根据下图执行：
```

        |            |
        v            v
     test-app    lint-repo
    /        \
   v          v
build-app  build-frontend
   \          /
    v        v
    deploy-all
```

- lint-repo和test-app Task没有from和runafter子句并开始同时执行。
- 一旦测试应用程序完成，由于build-app和build-frontend都运行在test-app Task之后，因此它们同时开始执行。
- 全部部署任务在build-app和build-frontend都完成后执行，因为它会从这两者中提取PipelineResources。
- 一旦lint-repo和deploy-all全部完成执行，整个管道就完成了执行。

# 添加Finally

您可以在finally部分下指定一个或多个最终任务的列表。 无论任务是成功还是错误，都保证在任务下的所有PipelineTasks完成之后并行执行最终任务。 最终任务与tasks部分下的PipelineTasks非常相似，并且遵循相同的语法。 每个最终任务必须具有有效的名称和taskRef或taskSpec。 例如：

```
spec:
  tasks:
    - name: tests
      taskRef:
        Name: integration-test
  finally:
    - name: cleanup-test
      taskRef:
        Name: cleanup
```

## 在最终任务中指定工作区

finally任务可以指定PipelineTasks可能已经利用的工作空间，例如Secrets中保存的凭据的挂载点。为了支持该要求，您可以在Workspaces字段中为与任务相似的Final tasks指定一个或多个workspaces。

```
spec:
  resources:
    - name: app-git
      type: git
  workspaces:
    - name: shared-workspace
  tasks:
    - name: clone-app-source
      taskRef:
        name: clone-app-repo-to-workspace
      workspaces:
        - name: shared-workspace
          workspace: shared-workspace
      resources:
        inputs:
          - name: app-git
            resource: app-git
  finally:
    - name: cleanup-workspace
      taskRef:
        name: cleanup-workspace
      workspaces:
        - name: shared-workspace
          workspace: shared-workspace
```

## 为final tasks指定参数

```
spec:
  tasks:
    - name: tests
      taskRef:
        Name: integration-test
  finally:
    - name: report-results
      taskRef:
        Name: report-results
      params:
        - name: url
          value: "someURL"
```

## 使用finally的PipelineRun状态

不设置 `finally`:

| `PipelineTasks` under `tasks` | `PipelineRun` status | Reason |
| ----------------------------- | -------------------- | ------ |
| all `PipelineTasks` successful | `true` | `Succeeded` |
| one or more `PipelineTasks` skipped and rest successful | `true` | `Completed` |
| single failure of `PipelineTask` | `false` | `failed` |

设置 `finally`:

| `PipelineTasks` under `tasks` | Final Tasks | `PipelineRun` status | Reason |
| ----------------------------- | ----------- | -------------------- | ------ |
| all `PipelineTask` successful | all final tasks successful | `true` | `Succeeded` |
| all `PipelineTask` successful | one or more failure of final tasks | `false` | `Failed` |
| one or more `PipelineTask` [skipped](conditions.md) and rest successful | all final tasks successful | `true` | `Completed` |
| one or more `PipelineTask` [skipped](conditions.md) and rest successful | one or more failure of final tasks | `false` | `Failed` |
| single failure of `PipelineTask` | all final tasks successful | `false` | `failed` |
| single failure of `PipelineTask` | one or more failure of final tasks | `false` | `failed` |


扫描关注我:

![微信](http://img.rocdu.top/20200703/qrcode_for_gh_7457c3b1bfab_258.jpg)
