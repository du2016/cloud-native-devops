# 介绍

k8s中资源限制分别为requests(需求)和limits(限制)

### 分类
request: 保留的资源

limit: 最大占用的资源

查看pod内容器的资源限制

```bash
kubectl get pods --namespace=test test -o template --template='{{range .spec.containers}}{{.resources}}{{end}}'
```

### requests与limit值范围如下：

0 <= requests <= NodeAllocatable

requests <= limits <= Infinity

### 根据request与limit的比较大小分为以下三种qos

- Guaranteed
  - pod中所有容器都必须统一设置limits，并且设置参数都一致，如果有一个容器要设置requests，那么所有容器都要设置，并设置参数同limits一致，那么这个pod的QoS就是Guaranteed级别。
- Burstable
  - pod中只要有一个容器的requests和limits的设置不相同，该pod的QoS即为Burstable。
- Best-Effort
  - 如果对于全部的resources来说requests与limits均未设置，该pod的QoS即为Best-Effort。

查看qosClass：
```bash
kubectl get pods --namespace=test test -o template --template='{{.status.qosClass}}'
```

判断代码如下：
```
if len(requests) == 0 && len(limits) == 0 {
		return v1.PodQOSBestEffort
	}
	// Check is requests match limits for all resources.
	if isGuaranteed {
		for name, req := range requests {
			if lim, exists := limits[name]; !exists || lim.Cmp(req) != 0 {
				isGuaranteed = false
				break
			}
		}
	}
	if isGuaranteed &&
		len(requests) == len(limits) {
		return v1.PodQOSGuaranteed
	}
```

### oom killer kill顺序
Best-Effort -> Burstable -> Guaranteed

### 算法

分数越低越不容器被杀掉

- Guaranteed：直接返回 -998
- Burstable: oomScoreAdjust := 1000 - (1000*memoryRequest)/memoryCapacity(占用资源越多越不容易被杀掉)
- Best-Effort: 直接返回1000

判断代码如下：

```
func GetContainerOOMScoreAdjust(pod *v1.Pod, container *v1.Container, memoryCapacity int64) int {
	switch v1qos.GetPodQOS(pod) {
	case v1.PodQOSGuaranteed:
		// Guaranteed containers should be the last to get killed.
		return guaranteedOOMScoreAdj
	case v1.PodQOSBestEffort:
		return besteffortOOMScoreAdj
	}

	// Burstable containers are a middle tier, between Guaranteed and Best-Effort. Ideally,
	// we want to protect Burstable containers that consume less memory than requested.
	// The formula below is a heuristic. A container requesting for 10% of a system's
	// memory will have an OOM score adjust of 900. If a process in container Y
	// uses over 10% of memory, its OOM score will be 1000. The idea is that containers
	// which use more than their request will have an OOM score of 1000 and will be prime
	// targets for OOM kills.
	// Note that this is a heuristic, it won't work if a container has many small processes.
	memoryRequest := container.Resources.Requests.Memory().Value()
	oomScoreAdjust := 1000 - (1000*memoryRequest)/memoryCapacity
	// A guaranteed pod using 100% of memory can have an OOM score of 10. Ensure
	// that burstable pods have a higher OOM score adjustment.
	if int(oomScoreAdjust) < (1000 + guaranteedOOMScoreAdj) {
		return (1000 + guaranteedOOMScoreAdj)
	}
	// Give burstable pods a higher chance of survival over besteffort pods.
	if int(oomScoreAdjust) == besteffortOOMScoreAdj {
		return int(oomScoreAdjust - 1)
	}
	return int(oomScoreAdjust)
}
```