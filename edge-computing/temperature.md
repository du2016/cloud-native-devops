# 环境

- 树莓派4b+dt11温度模块(针脚VCC:1 DATA:7对应BCM也就是4 GND:6) ubuntu 19.10
- k8s master 1.17 ubuntu 19.10
- 云边已经配置好，边缘节点ready

> 启动边缘的时间，默认没有启用memorycgroup，会报错进行以下配置
  ```
  vim /boot/firmware/btcmd.txt
  添加以下内容
  cgroup_enable=memory cgroup_memory=1
  reboot
  ```

# 加载device和devicemodel

```
kubectl apply -f https://raw.githubusercontent.com/du2016/kubeedge-examples/master/kubeedge-temperature-demo/crds/device.yaml
kubectl apply -f https://raw.githubusercontent.com/du2016/kubeedge-examples/master/kubeedge-temperature-demo/crds/device.yaml
```

# 加载温度mapper

这里需要和我连接针脚的方式一致，不然需要自行更改代码编译。
```
https://raw.githubusercontent.com/du2016/kubeedge-examples/master/kubeedge-temperature-demo/temperature-mapper/deployment.yaml
```

# 效果

```
kubectl get device temperature -w -o go-template --template='{{ range .status.twins }} {{.reported.value}} {{end}}'
28C
```