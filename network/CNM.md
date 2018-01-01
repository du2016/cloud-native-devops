通过kubelet的--network-plugin=cni命令行选择项来选择CNI插件。
kubelet从--cni-conf-dir（默认为/etc/cni/net.d）中读取文件，
并使用该文件中的CNI配置去设置每个pod网络。CNI配置文件必须与CNI相匹配，
并且任何所需的CNI插件的配置必须引用目前的--cni-bin-dir(默认为/opt/cni/bin)