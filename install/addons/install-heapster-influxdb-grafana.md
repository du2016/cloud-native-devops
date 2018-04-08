
# 使用主仓库

```
cd /opt/kubernetes-repo/cluster/addons/cluster-monitoring/influxdb

cat <<EOF | kubectl create -f -
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: heapster
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:heapster
subjects:
- kind: ServiceAccount
  name: heapster
  namespace: kube-system
EOF

kubectl create -f influxdb-grafana-controller.yaml
kubectl create -f grafana-service.yaml
kubectl create -f influxdb-service.yaml
kubectl create -f heapster-service.yaml
pip install jinja2
alias render_template='python -c "from jinja2 import Template; import sys; print(Template(sys.stdin.read()).render());"'
NODECOUNT=`kubectl get nodes | grep -v 'NAME'| wc -l`
```

### heaster 资源基于动态配置

```
单位 Mi
metrics_memory 200 + num_nodes * 4
eventer_memory 200 + num_nodes * 0.5
sed "s/pillar.get('num_nodes', -1)/$NODECOUNT/g" heapster-controller.yaml | render_template | kubectl create -f -
```


# 使用heapster仓库

```
git clone https://github.com/kubernetes/heapster/ heapster-repo
cd kubernetes-repo/cluster/addons/cluster-monitoring/influxdb
```

####  创建rbac

```
 kubectl create -f rbac/heapster-rbac.yaml
 kubectl create -f grafana.yaml
 kubectl create -f heapster.yaml
 kubectl create -f influxdb.yaml
 
 
alias render_template='python -c "from jinja2 import Template; import sys; print(Template(sys.stdin.read()).render());"'
```
