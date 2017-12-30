# Kubespray 安装

## 介绍
Kubespray 由一系列的 Ansible playbook、生成 inventory 的命令行工具以及生成 OS/Kubernetes 集群配置管理任务的专业知识构成。Kubespray 提供：

- 一个高可用集群
- 可组合的属性
- 支持主流的 Linux 发行版
- 持续集成测试

[github](https://github.com/kubernetes-incubator/kubespray)

## 依赖
- Ansible >= v2.3
- Jinja >= 2.9
- python-netaddr

[aws](https://github.com/kubernetes-incubator/kubespray/tree/master/contrib/terraform/aws)
[openstack](https://github.com/kubernetes-incubator/kubespray/tree/master/contrib/terraform/openstack)

## 安装

### 定义inventory

```bash
# 定义集群节点IP
declare -a IPS=(10.10.1.3 10.10.1.4 10.10.1.5)
# 生成inventory到inventory/inventory.cfg
CONFIG_FILE=inventory/inventory.cfg python3 contrib/inventory_builder/inventory.py ${IPS[@]}
```
inventory/group_var 包含了一系列配置参数

inventory/cluster.yml 为具体的安装playbook

inventory/scale.yml 添加节点的playbook

### 部署

```bash
ansible-playbook -i inventory/inventory.cfg cluster.yml -b -v --private-key=~/.ssh/private_key
```

### 添加节点

```bash
ansible-playbook -i my_inventory/inventory.cfg scale.yml -b -v --private-key=~/.ssh/private_key
```

### 网络检查应用

> deploy_netchecker 设置为 true

```bash
# 访问nodeport
curl http://localhost:31081/api/v1/connectivity_check
```

### 升级集群

```bash
VERSION=v1.9.0
ansible-playbook cluster.yml -i inventory/inventory.cfg -e kube_version=${VERSION}
```

### 重置节点

```bash
ansible-playbook -i my_inventory/inventory.cfg reset.yml -b -v --private-key=~/.ssh/private_key
```