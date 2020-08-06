Deps决定依赖关系

- 主函数Reconcile
- ReconcileKind
- resource注入  pkg/apis/resource/resource.go
- MakePod  pkg/pod/pod.go
- 初始化卷
- 初始化script，生成脚本，用于将script写入到initcontainer,生成tmpfile，执行tmpfile
- 初始化wokingdir
- 解析未指定命令容器的entrypoints
- 将命令与entrypoint传递给orderContainers，

wait_file
wait_file_content
post_file
termination_path

/ko-app/entrypoint
readyAnnotation


- 初始化secret卷
- 将git secret 、docker secret 初始化后作为参数传入init container
- getLimitRangeMinimum 获取limitrange的最小值，然后给container设置request
- 将credentials（/tekton/creds）挂载进每个容器
- 将step container和init container合并
- 运行container