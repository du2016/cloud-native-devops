# coredns

[coredns](https://github.com/coredns/deployment/tree/master/kubernetes)

```
后面可以有三个参数 svc-cidr pod-cidr domain
$ ./deploy.sh 10.3.0.0/12 172.17.0.0/16 | kubectl apply -f -
$ kubectl delete --namespace=kube-system deployment kube-dns
```