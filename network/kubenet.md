kubenet是一个在Linux上非常基本，简单的网络插件。它本身不会实现更高级的功能，如跨node网络或网络策略。
它通常与云提供商一起使用，为跨nodes或者单node环境通信设置路由规则。


--network-plugin=cni规定了cni网络插件的使用，
CNI插件二进制文件位于--cni-bin-dir（默认是/opt/cni/bin），
其配置位于--cni-conf-dir（默认是/etc/cni/net.d）。

--network-plugin=kubenet规定了kubenet网络插件的使用，
CNI的bridge和host-local插件，位于/opt/cni/bin或network-plugin-dir

--network-plugin-mtu=9001规定MTU的使用，目前只能在kubenet网络插件中使用。