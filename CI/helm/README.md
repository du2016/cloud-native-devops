# HELM

## 安装

helm init

## 替换serviceaccount

kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'

## 安装mysql服务

helm install stable/mysql --name=mysql

查看状态

helm status mysql

## 安装前定义chart

helm inspect values stable/mariadb

## 指定参数

$ echo '{mariadbUser: user0, mariadbDatabase: user0db}' > config.yaml
$ helm install -f config.yaml stable/mariadb


## 升级

helm upgrade  -f config.yaml mysql stable/mariadb


## 下载到本地

helm fetch stable/mysql

## 创建chart

helm create test

## 打包

helm package test

## 安装

helm install ./test.tar

