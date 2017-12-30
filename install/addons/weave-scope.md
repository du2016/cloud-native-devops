Weave Scope 是一个图形化工具，用于查看你的 containers、 pods、services等

[scope on k8s doc](https://www.weave.works/docs/scope/latest/installing/\#k8s)

```bash
# with weave cloud
kubectl apply --namespace kube-system -f "https://cloud.weave.works/k8s/scope.yaml?service-token=<token>&k8s-version=$(kubectl version | base64 | tr -d '\n')")
# without weave cloud
kubectl apply --namespace kube-system -f "https://cloud.weave.works/k8s/scope.yaml?k8s-version=$(kubectl version | base64 | tr -d '\n')"
```