
https://k8smeetup.github.io/docs/tasks/tools/install-kubectl/

```
# 指定tty大小
kubectl exec -it loan-admin1-kkjnx --namespace=qa-18 env COLUMNS=`tput cols` LINES=`tput lines` TERM=xterm /bin/bash

# 格式化字段
kubectl get pods loan2-g2xhz --namespace=qa-18 -o jsonpath='{.status.hostIP}'
kubectl get pods loan2-g2xhz --namespace=qa-18 -o template --template='{{.status.hostIP}}'
kubectl get pods loan2-g2xhz --namespace=qa-18 -o custom-columns=NAME:.metadata.name,hostip:.status.hostIP
```

https://k8smeetup.github.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
https://k8smeetup.github.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/
https://k8smeetup.github.io/docs/user-guide/kubectl/v1.8/


```
cat > test.tmpl << EOF
NAME                    HOSTIP
metadata.name           status.hostIP
EOF
kubectl get pods loan2-g2xhz --namespace=qa-18 -o custom-columns-file=test
```

 kubectl run -i --tty busybox --image=busybox --restart=Never -- sh
 
 kubectl config view -o jsonpath='{.contexts[*].name}'