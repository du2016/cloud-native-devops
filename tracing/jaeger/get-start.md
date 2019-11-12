本文将讲解jaeger基本概念，基于golang的代码实现以及注入原理

# jaeger 概述

组件概念：

- jaeger-client
- jaeger-agent 将client发送的span发送到collector
- jaeger-collector 收集数据并存储或发送到队列
- jaeger ingester 读取kafka队列写入存储
- jaeger-query 查询数据展示tracer

逻辑概念：

- span 具体的某个操作，包含以下属性

  - 操作名称
  - 开始时间
  - 执行时长
  - logs # 捕获指定时间的消息，或调试输出
  - tags # 不被继承,查询过滤理解追踪数据
  
- Trace 是一个完整的执行过程，是span的有向无环图
- SpanContext # 传递给下级span的信息trace_id，span_id，parentId等
- Baggage 存储在SpanContext的键值集合，在一个链路上全局传输


# 应用代码实现

## 单span代码
```
    // 从环境变量获取配置参数
	cfg,err:=jaegercfg.FromEnv()
	if err!=nil {
		log.Println(err)
	}
	cfg.Sampler=&jaegercfg.SamplerConfig{
		Type:  "const",// 使用const采样器
		Param: 1, // 采样所有追踪
	}
    // 设置服务名
	cfg.ServiceName = "jaeger tracer demo"
    // 根据创建Tracer
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Println(err)
	}
	defer closer.Close()
    //设置全局tracer
	opentracing.SetGlobalTracer(tracer)
    //创建一个span
	parentSpan:=tracer.StartSpan("root")
	defer parentSpan.Finish()
	parentSpan.LogFields(
		tracelog.String("hello","world"),
	)
	parentSpan.LogKV("foo","bar")
    // 创建一个childspan
	childspan:=tracer.StartSpan("child span",opentracing.ChildOf(parentSpan.Context()))
	defer childspan.Finish()
```


## 夸进程传播

使用上下文传递
```
ctx := opentracing.ContextWithSpan(context.Background(), span)
span, _ := opentracing.StartSpanFromContext(ctx, "req svc2")
```

## 跨请求传播

将span注入header

```
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
```

从header获取span上下文，并根据上下文创建新span

```
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("get haha", ext.RPCServerOption(spanCtx))
```

## 跨应用多span示例代码

有两个服务组成，svc1,svc2
有三个span组成，svc1 sayhello --> svc1 req svc2 --> svc2 get bagger

svc1代码

```
package main

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go/config"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Println(err)
	}
	cfg.ServiceName = "svc1"
	cfg.Sampler = &config.SamplerConfig{
		Type:  "const",
		Param: 1,
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Println(err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	span := tracer.StartSpan("say hello")
	span.SetTag("role", "root")
	span.LogKV("hello", "world")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)
	testchildspan(ctx)
}
func testchildspan(ctx context.Context) {
	url := "http://localhost:8000"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	span, _ := opentracing.StartSpanFromContext(ctx, "req svc2")
	defer span.Finish()
	span.SetTag("role", "childspan")
	span.SetBaggageItem("haha", "heihei")

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	log.Println(resp.Status)
}
```

svc2代码

```
package main

import (
	"net/http"
	"log"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go/config"
)

func main(){
	http.HandleFunc("/",test)
	http.ListenAndServe(":8000",nil)
}

func test(w http.ResponseWriter,r *http.Request){
	log.Println(r.Header,r.URL)

	cfg,err:=config.FromEnv()
	if err!=nil {
		log.Println(err)
	}
	cfg.ServiceName="svc2"
	cfg.Sampler=&config.SamplerConfig{
		Type:  "const",
		Param: 1,
	}
	tracer,closer,err:=cfg.NewTracer()
	if err!=nil {
		log.Println(err)
	}
	defer closer.Close()

	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("get haha", ext.RPCServerOption(spanCtx))
	defer span.Finish()
	log.Println(span.BaggageItem("haha"))

	span.LogFields(
		otlog.String("event", "string-format"),
		otlog.String("value", "hello wrold"),
	)
	w.Write([]byte("hello wrold"))
}
```


# header数据解析

对于上述代码我们发现svc1发送给svc2有如下特殊header
Uber-Trace-Id:[518ee099f68f3974:17531754249a513a:518ee099f68f3974:1] 
Uberctx-Haha:[heihei]
那这些header是如何根据tracer信息添加呢

## 根据传入的format获取对应的injector

Inject代码实现
```
func (t *Tracer) Inject(ctx opentracing.SpanContext, format interface{}, carrier interface{}) error {
	c, ok := ctx.(SpanContext)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}
	if injector, ok := t.injectors[format]; ok {
		return injector.Inject(c, carrier)
	}
	return opentracing.ErrUnsupportedFormat
}
```

我们看到是根据format拿到Injector,默认支持三种类型

- binary
- TextMap
- HTTPHeaders

其中HTTPHeaders基于TextMap传播

## Propagator

两种Propagator实现了Injector接口

- BinaryPropagator
- TextMapPropagator
BinaryPropagator的不再讲述

### TextMapPropagator 

```
func (p *TextMapPropagator) Inject(
	sc SpanContext,
	abstractCarrier interface{},
) error {
	textMapWriter, ok := abstractCarrier.(opentracing.TextMapWriter)
	if !ok {
		return opentracing.ErrInvalidCarrier
	}

	// 不要使用trace context对字符串进行编码
    // 设置trace 上下文header
    // 默认值TraceContextHeaderName = "uber-trace-id"
	textMapWriter.Set(p.headerKeys.TraceContextHeaderName, sc.String())
    // 设置baggage
	for k, v := range sc.baggage {
		safeKey := p.addBaggageKeyPrefix(k)
		safeVal := p.encodeValue(v)
		textMapWriter.Set(safeKey, safeVal)
	}
	return nil
}
```

## carrier

Propagator通过carrier注入提取数据
binary直接传入实现io.Writer接口对象即可
TextMapPropagator有两种carrier

- HTTPHeadersCarrier
- TextMapCarrier

### HTTPHeadersCarrier

```
type HTTPHeadersCarrier http.Header

// 符合TextMapWriter接口
func (c HTTPHeadersCarrier) Set(key, val string) {
	h := http.Header(c)
	h.Set(key, val)
}

// 符合TextMapReader接口
func (c HTTPHeadersCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range c {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
```

### TextMapCarrier

```
type TextMapCarrier map[string]string

// ForeachKey conforms to the TextMapReader interface.
func (c TextMapCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, v := range c {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Set implements Set() of opentracing.TextMapWriter
func (c TextMapCarrier) Set(key, val string) {
	c[key] = val
}
```

## 结果

TraceContextHeaderName 默认值为 "uber-trace-id"
设置该值代码为：

```
func (c SpanContext) String() string {
	if c.traceID.High == 0 {
		return fmt.Sprintf("%x:%x:%x:%x", c.traceID.Low, uint64(c.spanID), uint64(c.parentID), c.samplingState.stateFlags.Load())
	}
	return fmt.Sprintf("%x%016x:%x:%x:%x", c.traceID.High, c.traceID.Low, uint64(c.spanID), uint64(c.parentID), c.samplingState.stateFlags.Load())
}
```
由此看出该header包含了traceID,spanID,parentID

扫描关注我的公众号:

![微信](http://q08i5y6c2.bkt.clouddn.com/qrcode_for_gh_7457c3b1bfab_258.jpg)