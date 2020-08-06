# tekton controllers概述

tekton中主要实现了以下两个controller，用于接收处理具体的任务运行事件
- taskrun
- pipelinerun

由于pipelinerun最终将转化为taskrun，将首先对pipelinerun的处理逻辑进行说明。

# MainWithConfig

MainWithConfig是程序的入口，负责整体的初始化工作，
注册了clients,informers,InformerFactories,controllers

- CheckK8sClientMinimumVersionOrDie 对k8s版本进行检查
- SetupConfigMapWatchOrDie 检测所需的config map,此处根据SYSTEM_NAMESPACE进行判断，同时也可以指定ResourceLabelEnvKey，指定label selector，用于watch 配置，初始化controllers

```
      - name: SYSTEM_NAMESPACE
        valueFrom:
          fieldRef:
            apiVersion: v1
            fieldPath: metadata.namespace
```

- 通过配置 config-leader-election 决定选举相关参数，默认不开启选举
- 如果开启选举调用 RunLeaderElected开启选举
- 接下来启动所有的cache informers用于缓存对象变更事件，启动对应的controller


# pipelinerun controller

- pipelinerun.NewController(*namespace, images) 通过namespace,images两个参数返回一个func(context.Context, configmap.Watcher) *controller.Impl 函数，该函数返回结果实现了knative的controller.Impl,实际上就是一个k8s controller，最终被上文的MainWithConfig调用执行
- 熟悉kubebuilder的都之道一个controller具体逻辑都在Reconciler（协调器）内

# Reconcile

- Reconcile 接收 ctx和key作为参数，其中key为 namespace/name
- 通过namespace,name获取具体的piplinerun
- 若没有被删除则执行ReconcileKind方法

# ReconcileKind

- GetCondition获取
- 若piplinerun还未执行，则初始化piplinerun的condition
- 若pipline已经停止则执行相关清理操作
- 若piplinerun被终止，则调用cancelPipelineRun终止相关的taskrun
- 根据updatePipelineRunStatusFromInformer设置Piplinerun状态

# reconcile

- GetPipelineData获取pipelineMeta, pipelineSpec  
- dag.Build 构建pipeline task的 DAG,构建Finally的DAG
- GetResourcesFromBindings 获取providedResources
- 解析tasks返回 ResolvedPipelineRunTask
- 检查pvc,Affinity Assistant，InitializeArtifactStorage初始化
- runNextSchedulableTask 调用下一个Task
- GetNextTasks  获取下一个Tasks
- createTaskRun 创建taskrun
- 便利nextRprts 创建taskrun




