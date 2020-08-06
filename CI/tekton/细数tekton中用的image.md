tekton中以pod为Task的运行单元，而Task中的step实际就是一个个pod,其中用到了许多容器用于进行初始化动作，本文将分析各个容器在tekton task运行时起到的作用


# entrypoint-image

包含entrypoint 可执行文件的image，默认值"override-with-entrypoint:latest"，在task pod启动时，会将/ko-app/entrypoint拷贝到具体step的/tekton/tools/entrypoint目录，作为首先调用的命令，将使用该命令调用真正的命令

entrypoint镜像主要有以下六个参数

- entrypoint  真正要运行的entrypoint
- wait_file   要等待的文件
- wait_file_content 等待的文件需要有具体内容
- post_file  执行完成之后写入的文件
- termination_path 终止时写入的文件
- results  包含task results的文件列表

# nop-image

用于停止sidecar，"tianon/true",没有任何逻辑，直接替换sidecar 容器完成更新

# affinity-assistant-image

Affinity Assistant(亲和助理)，用于在使用动态PV作为workspaces时保证tasks调度到同一个节点

默认是nginx,没有任何逻辑

# gitImage

包含git命令的image，"override-with-git:latest"

包含以下参数
- url  git 远程仓库地址
- revision  版本
- refspec  revision从哪个refspec
- path 本地存储代码的路径
- sslVerify 是否开启ssl检查
- submodules 初始化并获取的submodules
- depth 执行shallow clone的深度
- terminationMessagePath  终止信息写入的文件

# credsImage

用于生产成credentials的image，"override-with-creds:latest",

包含两个部分：

1.basicDockerBuilder

包含以下三个参数：
  - basic-docker secret和路径的列表
  - docker-config  从docker config.json获取配置
  - docker-cfg 从 .dockercfg获取配置

从而生成docker /tekton/creds/.docker/config.json

2.gitConfigBuilder

包含以下两个参数
- basic-git
- ssh-git 

根据名称将sshConfig写入到 /tekton/creds/.ssh/下，同时添加到/tekton/creds/.ssh/config和known_hosts

根据名称写入到.gitconfig,.git-credentials

# kubeconfigWriterImage

包含kubeconfig writer二进制文件的容器映像，"override-with-kubeconfig-writer:latest"

两个参数：

- clusterConfig 当在集群外时需要提供的json格式的cluster配置
- destinationDir kubeconfig写入的目标目录

# shellImage

包含shell的二进制镜像，默认"busybox"，主要用于运行初始化脚本，
例如task中支持的script功能，就是通过运行busybox将script写入文件，达到运行的目的

# gsutilImage

包含gsutil的镜像，默认"google/cloud-sdk"

用于创建gcs类型的storage作为piplineresource

# buildGCSFetcherImage

包含GCS fetcher 二进制文件的镜像，默认"gcr.io/cloud-builders/gcs-fetcher:latest"

上面的基本一样，是gcs的子类型，它类似于GCSResource，但添加了其他功能从而与本地构建兼容。

# prImage

包含PR二进制文件的容器镜像，"override-with-pr:latest"

参数： 
- url  pull request的url
- path pull request的目录
- mode  默认download,pull request的模式
- provider  要使用的SCM provider
- insecure-skip-tls-verify 是否跳过ssl校验



# imageDigestExporterImage

包含image digest导出器二进制文件的容器映像，"override-with-imagedigest-exporter-image:latest"，用于到处镜像的digest

参数：
- images 镜像列表
- terminationMessagePath 默认值/tekton/termination


扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
