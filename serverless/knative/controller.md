# source

source是遵守源规范的最小资源形态。 这种鸭子类型(动态类型)旨在允许源和进口商的实施者验证自己的资源是否符合期望。 这不是真正的资源。 注意：`Source Specification`正在进行中，可以修改形状和名称，直到被接受为止。

源仓库示例 https://github.com/knative-sandbox/sample-source

Knative Eventing样本源定义了一个简单的源，该源将事件从HTTP服务器转换为CloudEvents，并演示了Knative Eventing编写源的规范样式。


在pkg/apis/samples/v1alpha1/samplesource_types.go中定义资源架构中所需的类型，其中包括资源yaml中将需要的字段，以及将使用源的客户端集和API在控制器中引用的字段


type SampleSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the SampleSource (from the client).
	Spec SampleSourceSpec `json:"spec"`

	// Status communicates the observed state of the SampleSource (from the controller).
	// +optional
	Status SampleSourceStatus `json:"status,omitempty"`
}

// SampleSourceSpec holds the desired state of the SampleSource (from the client).
type SampleSourceSpec struct {
	// ServiceAccountName holds the name of the Kubernetes service account
	// as which the underlying K8s resources should be run. If unspecified
	// this will default to the "default" service account for the namespace
	// in which the SampleSource exists.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Interval is the time interval between events.
	//
	// The string format is a sequence of decimal numbers, each with optional
	// fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time
	// units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Interval string `json:"interval"`

	// Sink is a reference to an object that will resolve to a host
	// name to use as the sink.
	Sink *duckv1.Destination `json:"sink"`
}

// SampleSourceStatus communicates the observed state of the SampleSource (from the controller).
type SampleSourceStatus struct {
	duckv1.Status `json:",inline"`

	// SinkURI is the current active sink URI that has been configured
	// for the SampleSource.
	// +optional
	SinkURI *apis.URL `json:"sinkUri,omitempty"`
}


定义将在status和SinkURI字段中反映的生命周期

const (
	// SampleConditionReady has status True when the SampleSource is ready to send events.
	SampleConditionReady = apis.ConditionReady
    // ...
)

定义将从Reconciler函数调用的函数以设置生命周期条件。这通常是在

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *SampleSourceStatus) InitializeConditions() {
	SampleCondSet.Manage(s).InitializeConditions()
}

...

// MarkSink sets the condition that the source has a sink configured.
func (s *SampleSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		SampleCondSet.Manage(s).MarkTrue(SampleConditionSinkProvided)
	} else {
		SampleCondSet.Manage(s).MarkUnknown(SampleConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *SampleSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	SampleCondSet.Manage(s).MarkFalse(SampleConditionSinkProvided, reason, messageFormat, messageA...)
}

# kafkabinding
publisher->KafkaBinding-> kafkasource -> kafka -> event-display

# kafkachannel

source -> broker -> Trigger  -> filter -> Service -> container
               |
             channel(kafka)
# kafkasource

client -> KafkaTopic  -> KafkaSource -> event-display



source 从指定源中获取数据并发送

controller校验数据并作出相应动作 需要实现Addressable接口，以crd形式存在,使用Deployment和SinkBinding资源来部署并绑定事件源和receive adapter。 同时确保为这些辅助资源正确设置informers

