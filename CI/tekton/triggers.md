# trigger

trigger使用户能够将事件有效负载中的字段映射到资源模板中。 换句话说，这允许事件既可以建模也可以将实例化为Kubernetes资源。 对于tektoncd/pipeline，这使得将配置封装到PipelineRuns和PipelineResources中变得容易。

![](http://img.rocdu.top/20200707/TriggerFlow.png)

# 安装

kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml

# TriggerTemplates

TriggerTemplate是可以模板化资源的资源。 TriggerTemplate具有可在资源模板中任何位置替换的参数。

```
apiVersion: triggers.tekton.dev/v1alpha1
kind: TriggerTemplate
metadata:
  name: pipeline-template
spec:
  params:
  - name: gitrevision
    description: The git revision
    default: master
  - name: gitrepositoryurl
    description: The git repository url
  - name: message
    description: The message to print
    default: This is the default message
  - name: contenttype
    description: The Content-Type of the event
  resourcetemplates:
  - apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      generateName: simple-pipeline-run-
    spec:
      pipelineRef:
        name: simple-pipeline
      params:
      - name: message
        value: $(params.message)
      - name: contenttype
        value: $(params.contenttype)
      resources:
      - name: git-source
        resourceSpec:
          type: git
          params:
          - name: revision
            value: $(tt.params.gitrevision)
          - name: url
            value: $(tt.params.gitrepositoryurl)
```

# TriggerBindings

按照名称，TriggerBindings绑定事件/触发器。 使用TriggerBindings可以捕获事件中的字段并将其存储为参数。 故意将TriggerBindings与TriggerTemplates分开以鼓励它们之间的重用。


apiVersion: triggers.tekton.dev/v1alpha1
kind: TriggerBinding
metadata:
  name: pipeline-binding
spec:
  params:
  - name: gitrevision
    value: $(body.head_commit.id)
  - name: gitrepositoryurl
    value: $(body.repository.url)
  - name: contenttype
    value: $(header.Content-Type)
    
TriggerBindings连接到EventListener内的TriggerTemplates，在该模板上实际实例化Pod，以“侦听”各个事件

## 多个绑定

在EventListener中，您可以将多个绑定指定为触发器的一部分。 这使您可以创建可重用的绑定，这些绑定可以与各种触发器混合并匹配。 例如，触发器具有一个绑定，该绑定提取事件信息，而另一个绑定提供部署环境信息：

```
apiVersion: triggers.tekton.dev/v1alpha1
kind: TriggerBinding
metadata:
  name: event-binding
spec:
  params:
    - name: gitrevision
      value: $(body.head_commit.id)
    - name: gitrepositoryurl
      value: $(body.repository.url)
---
apiVersion: triggers.tekton.dev/v1alpha1
kind: TriggerBinding
metadata:
  name: prod-env
spec:
  params:
    - name: environment
      value: prod
---
apiVersion: triggers.tekton.dev/v1alpha1
kind: TriggerBinding
metadata:
  name: staging-env
spec:
  params:
    - name: environment
      value: staging
---
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: listener
spec:
  triggers:
    - name: prod-trigger
      bindings:
        - name: event-binding
        - name: prod-env
      template:
        name: pipeline-template
    - name: staging-trigger
      bindings:
        - name: event-binding
        - name: staging-env
      template:
        name: pipeline-template
```

# EventListener

EventListener是Kubernetes的自定义资源，它允许用户以声明的方式处理带有JSON负载的基于HTTP的传入事件。 EventListeners公开了传入事件指向的可寻址“接收器”。 用户可以声明TriggerBindings来从事件中提取字段，并将其应用于TriggerTemplates以创建Tekton资源。 此外，EventListeners允许使用事件拦截器进行轻量级的事件处理。

- webhook
- GitHub
- GitLab
- Bitbucket
- CEL

以gitlab为例如下，绑定trigger template和TriggerBindings及eventlisteners关联起来

```
  podTemplate: {}
  serviceAccountName: tekton-triggers-gitlab-sa
  triggers:
  - bindings:
    - kind: TriggerBinding
      ref: gitlab-push-binding
    interceptors:
    - gitlab:
        eventTypes:
        - Push Hook
        secretRef:
          secretKey: secretToken
          secretName: gitlab-secret
    name: gitlab-push-events-trigger
    template:
      name: gitlab-echo-template
```


# 使用gitlab拦截器执行任务

```
git clone https://github.com/tektoncd/triggers

kubectl apply -f examples/gitlab/

kubectl port-forward \
 "$(kubectl get pod --selector=eventlistener=gitlab-listener -oname)" \
  8080
  
  
curl -v \
-H 'X-GitLab-Token: 1234567' \
-H 'X-Gitlab-Event: Push Hook' \
-H 'Content-Type: application/json' \
--data-binary "@examples/gitlab/gitlab-push-event.json" \
http://localhost:8080

kubectl get taskruns | grep gitlab-run-
```

扫描关注我:

![微信](http://img.rocdu.top/20200707/qrcode_for_gh_7457c3b1bfab_258.jpg)



扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
