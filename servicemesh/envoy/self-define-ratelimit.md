# envoy ratelimit

envoy 可以继承一个全局grpc ratelimit 服务，称之为为`rate limit service`，

go-control-plane 是一个官方实现的golang 库`github.com/envoyproxy/go-control-plane`

go-control-plane中关于rls的pb文件为`envoy/service/ratelimit/v2/rls.pb.go`

其包含了一个RegisterRateLimitServiceServer方法，将一个限流器实现注册到grpcserver

```
func RegisterRateLimitServiceServer(s *grpc.Server, srv RateLimitServiceServer) {
	s.RegisterService(&_RateLimitService_serviceDesc, srv)
}
```

而RateLimitServiceServer是一个接口

```
type RateLimitServiceServer interface {
	ShouldRateLimit(context.Context, *RateLimitRequest) (*RateLimitResponse, error)
}
```

由此看出我们重点需要实现一个ShouldRateLimit方法

对于ShouldRateLimit，接收RateLimitRequest，返回RateLimitResponse


对于RateLimitRequest结构体如下，

```
type RateLimitRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Domain      string                           `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	Descriptors []*ratelimit.RateLimitDescriptor `protobuf:"bytes,2,rep,name=descriptors,proto3" json:"descriptors,omitempty"`
	HitsAddend  uint32                           `protobuf:"varint,3,opt,name=hits_addend,json=hitsAddend,proto3" json:"hits_addend,omitempty"`
}
```

其包含了Descriptors，也就是限流信息描述，可以包含多个Descriptor，HitsAddend就是命中累加次数

```
type RateLimitDescriptor struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entries []*RateLimitDescriptor_Entry `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
}
```
每个Descriptor 可以包含多个Entry，Descriptor是限流的最小单元，对于Descriptor下所有的Entry，无论任何一个达到阈值，都应触发限流

```
type RateLimitDescriptor_Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}
```

Entry包含具体的key value

对于key，envoy包含五种类型：

- source_cluster（根据source_cluster限流）
- destination_cluster （根据destination_cluster限流）
- request_headers （根据request_headers限流）
- remote_address （根据remote_address限流）
- generic_key （根据generic_key限流）
- header_value_match （根据header 正则匹配进行限流）


# 实现限流器

我们将通过redis实现一个基于固定窗口的限流实现

这里我们实现了一个不限流的ShouldRateLimit方法实现。

## 定义限流结构、方法

```
type ratelimitService struct{}
func (r ratelimitService) ShouldRateLimit(ctx context.Context, request *pb.RateLimitRequest) (*pb.RateLimitResponse, error) {
    return  &pb.RateLimitResponse{
        OverallCode: pb.RateLimitResponse_OK,
    }, nil
}
```

## 注册限流实现

在main函数中，将我们的限流器注册到grpcserver，调用`reflection.Register(s)`方便我们使用grpcurl进行调试。

```
func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println(err)
	}
	s := grpc.NewServer()
	pb.RegisterRateLimitServiceServer(s, &ratelimitService{})
	reflection.Register(s)
	s.Serve(listener)
}
```


## 添加限流逻辑


这里将通过redis固定窗口实现限流器，限制每分钟不能超过2个请求，超过则处罚限流

```
func (r ratelimitService) ShouldRateLimit(ctx context.Context, request *pb.RateLimitRequest) (*pb.RateLimitResponse, error) {
	now := (time.Now().Unix()/60)*60
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil {
		log.Println(err)
		return nil, err
	}
	defer conn.Close()
	var uq string
	if request.Descriptors[0].Entries[0].Value!=""{
		uq = request.Domain+"_"+request.Descriptors[0].Entries[0].Key +"_" +request.Descriptors[0].Entries[0].Value
	}else {
		uq = request.Domain+"_"+request.Descriptors[0].Entries[0].Key
	}
	uq+=fmt.Sprint(now)
	reply,err:=redis.String(conn.Do("GET", uq))
	if  err!= nil&& reply!="" {
		return &pb.RateLimitResponse{OverallCode: pb.RateLimitResponse_UNKNOWN,}, err
	}
	if count,_:=strconv.Atoi(fmt.Sprint(reply));count>2 {
		return  &pb.RateLimitResponse{OverallCode: pb.RateLimitResponse_OVER_LIMIT}, nil
	}
	if _, err := conn.Do("INCR", uq); err != nil {
		return &pb.RateLimitResponse{OverallCode: pb.RateLimitResponse_UNKNOWN}, err
	}
	return  &pb.RateLimitResponse{OverallCode: pb.RateLimitResponse_OK,}, nil
}
```

我们通过 timestamp除去60获取一个时间窗口，在时间窗口内将访问次数进行累加，当达到阈值返回overlimit,这里并没有进行ttl设置，生产级别实现需要对rediskey 设置ttl，自动删除过期的key,这个使用OverallCode进行统一返回，实际上我们针对每个Descriptor可以进行单独设置，并且可以设置limit_remaining，让客户端可以获取当前剩余的可访问次数


## grpccurl请求

``````
grpcurl -plaintext -d '{"domain":"contour","descriptors":[{"entries":[{"key":"generic_key","value":"apis"}]}]}'  127.0.0.1:8080 envoy.service.ratelimit.v2.RateLimitService/ShouldRateLimit
{
  "overallCode": "OVER_LIMIT",
  "statuses": [
    {
      "code": "OVER_LIMIT",
      "currentLimit": {
        "requestsPerUnit": 2,
        "unit": "MINUTE"
      }
    }
  ]
}
```

扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)
