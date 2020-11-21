# cert-manager

cert-manager是本地Kubernetes证书管理控制器。它可以帮助从各种来源颁发证书，例如Let's Encrypt,HashiCorp Vault,Venafi,简单的签名密钥对或自签名。

它将确保证书有效并且是最新的，并在到期前尝试在配置的时间续订证书。

它大致基于kube-lego的工作， 并从kube-cert-manager等其他类似项目中借鉴了一些智慧 。

![](https://cert-manager.io/images/high-level-overview.svg)

# 安装

kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.0.3/cert-manager-legacy.yaml
kubectl get pods --namespace cert-manager

# 安装nginx

helm install quickstart stable/nginx-ingress

# 部署服务

https://cert-manager.io/docs/tutorials/acme/ingress/