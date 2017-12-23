# nfs存储

NFS 存储deploy文件示例
```
cat > deploy.nfs.yaml << EOF
---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: nfs-client-provisioner
spec:
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: nfs-client-provisioner
    spec:
      serviceAccount: nfs-client-provisioner
      containers:
        - name: nfs-client-provisioner
          image: registry.jimubox.com/external_storage/nfs-client-provisioner:v2.0.0
          volumeMounts:
            - name: nfs-client-root
              mountPath: /persistentvolumes
          env:
            - name: PROVISIONER_NAME
              value: fuseim.pri/ifs
            - name: NFS_SERVER
              value: 172.26.6.1
            - name: NFS_PATH
              value: /data/nfs-storage/k8s-storage/ssd
      volumes:
        - name: nfs-client-root
          nfs:
            server: 172.26.6.1
            path: /data/nfs-storage/k8s-storage/ssd
EOF
```