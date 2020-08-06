```
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-cm
data:
  envoy-yml: |-
    admin:
      access_log_path: /tmp/admin_access.log
      address:
        socket_address:
          protocol: TCP
          address: 127.0.0.1
          port_value: 9901
    static_resources:
      listeners:
      - name: listener_0
        address:
          socket_address:
            protocol: TCP
            address: 0.0.0.0
            port_value: 10000
        filter_chains:
        - filters:
          - name: envoy.http_connection_manager
            typed_config:
              "@type": type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
              stat_prefix: ingress_http
              route_config:
                name: local_route
                virtual_hosts:
                - name: local_service
                  domains: ["*"]
                  routes:
                  - match:
                      prefix: "/"
                    route:
                      host_rewrite: httpbin.org
                      cluster: service_httpbin
                  rate_limits:
                    - stage: 0
                      actions:
                        - {destination_cluster: {}}
              http_filters:
                - name: envoy.rate_limit
                  config:
                    domain: foo
                    stage: 0
                    rate_limit_service:
                      grpc_service:
                        envoy_grpc:
                          cluster_name: rate_limit_cluster
                        timeout: 0.25s
                - name: envoy.router
      clusters:
      - name: service_httpbin
        connect_timeout: 0.5s
        type: LOGICAL_DNS
        # Comment out the following line to test on v6 networks
        dns_lookup_family: V4_ONLY
        lb_policy: ROUND_ROBIN
        load_assignment:
          cluster_name: service_httpbin
          endpoints:
          - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: httpbin.org
                    port_value: 80
      - name: rate_limit_cluster
        type: LOGICAL_DNS
        connect_timeout: 0.25s
        lb_policy: ROUND_ROBIN
        http2_protocol_options: {}
        load_assignment:
          cluster_name: rate_limit_cluster
          endpoints:
          - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: sentinel-rls-service
                    port_value: 10245
---
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: envoy-deployment-basic
  labels:
    app: envoy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: envoy
  template:
    metadata:
      labels:
        app: envoy
    spec:
      containers:
        - name: envoy
          image: envoyproxy/envoy:v1.12.0
          ports:
            - containerPort: 10000
          command: ["envoy"]
          args: ["-c", "/tmp/envoy/envoy.yaml"]
          volumeMounts:
            - name: envoy-config
              mountPath: /tmp/envoy
      volumes:
        - name: envoy-config
          configMap:
            name: envoy-cm
            items:
              - key: envoy-yml
                path: envoy.yaml
---
apiVersion: v1
kind: Service
metadata:
  name: envoy-service
  labels:
    name: envoy-service
spec:
  type: ClusterIP
  ports:
    - port: 10000
      targetPort: 10000
      protocol: TCP
  selector:
    app: envoy
```


```
apiVersion: v1
kind: ConfigMap
metadata:
  name: sentinel-rule-cm
data:
  rule-yaml: |-
    domain: foo
    descriptors:
      - resources:
        - key: "destination_cluster"
          value: "service_httpbin"
        count: 1
---
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: sentinel-rls-server
  labels:
    app: sentinel
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sentinel
  template:
    metadata:
      labels:
        app: sentinel
    spec:
      containers:
        - name: sentinelserver
          # You could replace the image with your own image here
          image: "registry.cn-hangzhou.aliyuncs.com/sentinel-docker-repo/sentinel-envoy-rls-server:latest"
          imagePullPolicy: Always
          ports:
            - containerPort: 10245
            - containerPort: 8719
          volumeMounts:
            - name: sentinel-rule-config
              mountPath: /tmp/sentinel
          env:
            - name: SENTINEL_RLS_RULE_FILE_PATH
              value: "/tmp/sentinel/rule.yaml"
      volumes:
        - name: sentinel-rule-config
          configMap:
            name: sentinel-rule-cm
            items:
              - key: rule-yaml
                path: rule.yaml
---
apiVersion: v1
kind: Service
metadata:
  name: sentinel-rls-service
  labels:
    name: sentinel-rls-service
spec:
  type: ClusterIP
  ports:
    - port: 8719
      targetPort: 8719
      name: sentinel-command
    - port: 10245
      targetPort: 10245
      name: sentinel-grpc
  selector:
    app: sentinel
```