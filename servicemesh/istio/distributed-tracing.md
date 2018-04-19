# 分布式链路追踪

## zipkin


kubectl apply -f install/kubernetes/addons/zipkin.yaml

## Jaeger

kubectl apply -n istio-system -f https://raw.githubusercontent.com/jaegertracing/jaeger-kubernetes/master/all-in-one/jaeger-all-in-one-template.yml


## 要求

传入请求收集下列header传入后出：
- x-request-id
- x-b3-traceid
- x-b3-spanid
- x-b3-parentspanid
- x-b3-sampled
- x-b3-flags
- x-ot-span-context