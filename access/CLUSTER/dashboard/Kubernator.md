kubectl create ns kubernator
kubectl -n kubernator run --image=smpio/kubernator --port=80 kubernator
kubectl -n kubernator expose deploy kubernator
kubectl proxy


http://localhost:8001/api/v1/namespaces/kubernator/services/kubernator/proxy/