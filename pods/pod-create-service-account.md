# 创建serviceaccount


kubectl create serviceaccount test

自动会为serviceaccount创建token
kubectl get secret我们可以看到对应的token

# 手动为serviceaccount创建secret

```
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  annotations:
    kubernetes.io/service-account.name: test
type: kubernetes.io/service-account-token
EOF
```

# 取消自动挂载token

## 取消某个serviceaccount的自动挂载


```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-robot
automountServiceAccountToken: false
```

## 取消某个pod的自动挂载

```
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  serviceAccountName: build-robot
  automountServiceAccountToken: false
  ...
```

# 将imagepullsecret添加到serviceaccount

## 创建image pull secret

kubectl create secret docker-registry <name> --docker-server=DOCKER_REGISTRY_SERVER --docker-username=DOCKER_USER --docker-password=DOCKER_PASSWORD --docker-email=DOCKER_EMAIL

## 添加imagePullSecrets字段

```
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: 2015-08-07T22:02:39Z
  name: default
  namespace: default
  uid: 052fb0f4-3d50-11e5-b066-42010af0d7b6
secrets:
- name: default-token-uudge
imagePullSecrets:
- name: myregistrykey
```

# serviceaccount卷投影

```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    volumeMounts:
    - mountPath: /var/run/secrets/tokens
      name: vault-token
  serviceAccountName: build-robot
  volumes:
  - name: vault-token
    projected:
      sources:
      - serviceAccountToken:
          path: vault-token
          expirationSeconds: 7200
          audience: vault
```
serviceaccount卷投影可以设置手中和有效期