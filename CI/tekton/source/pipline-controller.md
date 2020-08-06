
## step
### injection clients

- "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
- "github.com/tektoncd/pipeline/pkg/client/resource/clientset/versioned"
- cloudevents.Client
- "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
- "k8s.io/client-go/kubernetes"


### informer factories

- "github.com/tektoncd/pipeline/pkg/client/informers/externalversions"
- "github.com/tektoncd/pipeline/pkg/client/resource/informers/externalversions"
- "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
- "k8s.io/client-go/informers"
- "k8s.io/client-go/informers"

### informers

- clustertask
- contadion
- pipline
- piplinerun
- task
- taskrun
- pod
- secret
- configmap