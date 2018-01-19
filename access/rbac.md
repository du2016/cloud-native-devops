/etc/kubernetes/token.csv
ddsadsasad,test,10002,"test"

kubectl --namespace=kube-system get role secret-reader -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: 2018-01-19T07:58:45Z
  name: secret-reader
  namespace: kube-system
  resourceVersion: "769941"
  selfLink: /apis/rbac.authorization.k8s.io/v1/namespaces/kube-system/roles/secret-reader
  uid: 9329c605-fcee-11e7-afbc-fa0ef1535b00
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - watch
  - list
  

kubectl --namespace=kube-system get rolebinding read-secrets -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: 2018-01-19T07:55:58Z
  name: read-secrets
  namespace: kube-system
  resourceVersion: "769905"
  selfLink: /apis/rbac.authorization.k8s.io/v1/namespaces/kube-system/rolebindings/read-secrets
  uid: 2f68f5a5-fcee-11e7-afbc-fa0ef1535b00
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: secret-reader
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: test
  
  
curl --cacert /etc/kubernetes/ssl/ca.pem -X GET https://172.26.6.1:6443/api/v1/namespaces/kube-system/secrets  -H 'Authorization: Bearer ddsadsasad'





