# kubelet bootstrap kubeconfig
```
export KUBE_APISERVER="https://172.26.6.131:6443"

kubectl config set-cluster kubernetes \
  --certificate-authority=/etc/kubernetes/ssl/ca.pem \
  --embed-certs=true \
  --server=${KUBE_APISERVER} \
  --kubeconfig=/etc/kubernetes/bootstrap.kubeconfig

kubectl config set-credentials kubelet-bootstrap \
  --token=${BOOTSTRAP_TOKEN} \
  --kubeconfig=/etc/kubernetes/bootstrap.kubeconfig

kubectl config set-context default \
  --cluster=kubernetes \
  --user=kubelet-bootstrap \
  --kubeconfig=/etc/kubernetes/bootstrap.kubeconfig

kubectl config use-context default --kubeconfig=/etc/kubernetes/bootstrap.kubeconfig
```

# kube-proxy kubeconfig
```
export KUBE_APISERVER="https://172.26.6.131:6443"
kubectl config set-cluster kubernetes \
  --certificate-authority=/etc/kubernetes/ssl/ca.pem \
  --embed-certs=true \
  --server=${KUBE_APISERVER} \
  --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig
kubectl config set-credentials kube-proxy \
  --client-certificate=/etc/kubernetes/ssl/kube-proxy.pem \
  --client-key=/etc/kubernetes/ssl/kube-proxy-key.pem \
  --embed-certs=true \
  --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig
kubectl config set-context default \
  --cluster=kubernetes \
  --user=kube-proxy \
  --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig
kubectl config use-context default --kubeconfig=/etc/kubernetes/kube-proxy.kubeconfig
```

# 生成kubectl kubeconfig
```
export KUBE_APISERVER="https://172.26.6.131:6443"
kubectl config set-cluster kubernetes \
  --certificate-authority=ca.pem \
  --embed-certs=true \
  --server=${KUBE_APISERVER} \
  --kubeconfig=/root/.kube/config
kubectl config set-credentials admin \
  --client-certificate=admin.pem \
  --embed-certs=true \
  --client-key=admin-key.pem \
    --kubeconfig=/root/.kube/config
kubectl config set-context kubernetes \
  --cluster=kubernetes \
  --user=admin \
    --kubeconfig=/root/.kube/config
kubectl config use-context kubernetes   --kubeconfig=/root/.kube/config
```
