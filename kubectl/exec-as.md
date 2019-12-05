# exec-as
默认由于cri没有实现user接口kubectl只能以默认用于进入容器，exec-as通过插件实现了此功能

# 安装插件
kubectl krew install exec-as

# 使用

kubectl exec-as  -u root rabbitmq-0 -- pwd