# cloudhub

cloud用于从云端向边缘发布消息

initHubConfig 初始化edgehub的配置

# DispatchMessage

从beehive接收消息，发送到指定的channel

# StartCloudHub

根据配置判断启用了哪个protocol，然后启动

对于websocket和quic做了封装
https://github.com/kubeedge/viaduct

# startuds

启动Unix socket用于和csiduiver通信



