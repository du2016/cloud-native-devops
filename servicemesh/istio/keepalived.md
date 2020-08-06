
downstream 配置

[文档](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/core/socket_option.proto#core-socketoption)

level:

```
#define IPPROTO_IP 0 /* dummy for IP */
#define IPPROTO_ICMP 1 /* control message protocol */
#define IPPROTO_IGMP 2 /* internet group management protocol */
#define IPPROTO_GGP 3 /* gateway^2 (deprecated) */
#define IPPROTO_TCP 6 /* tcp */
#define IPPROTO_PUP 12 /* pup */
#define IPPROTO_UDP 17 /* user datagram protocol */
#define IPPROTO_IDP 22 /* xns idp */
#define IPPROTO_ND 77 /* UNOFFICIAL net disk proto */
#define IPPROTO_RAW 255 /* raw IP packet */
#define IPPROTO_MAX 256
```

listener下的配置
```
socket_options:
    - level: 6 #TCP
      name: 18
      int_value: 30000  #30秒
      state: STATE_LISTENING # STATE_PREBIND/STATE_BOUND/STATE_LISTENING
```


upstreanm 配置

[文档](https://www.envoyproxy.io/docs/envoy/v1.14.3/api-v2/api/v2/core/address.proto#envoy-api-msg-core-tcpkeepalive)

cluster下UpstreamConnectionOptions:
```
"tcp_keepalive": {
  "keepalive_probes": "{...}",
  "keepalive_time": "{...}",
  "keepalive_interval": "{...}"
}