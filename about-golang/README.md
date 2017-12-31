# golang

## goroutine

golang使用go协程实现并发，协程与线程主要区别是它将不再被内核调度，而是交给了程序自己而线程是将自己交给内核调度，
所以需要golang调度器的存在，goroutine以函数作为最小单元，golang语言作者Rob Pike也说，
“Goroutine是一个与其他goroutines 并发运行在同一地址空间的Go函数或方法。
一个运行的程序由一个或更多个goroutine组成。它与线程、协程、进程等不同。它是一个goroutine“.
GOMAXPROCS 限制并发数。

优势：
- 静态编译
- 跨平台
- 语法简单

缺陷：
- GC（标记清除-->三色标记）

sysmon() retake() preemptone()

```
package main

import "fmt"

type Fruit interface {
	Name()
}

type Apple struct{}

func (stu *Apple) Name() {

}

func check() Fruit {
	var apple *Apple
	return apple
}

func main() {
	if check() == (*Apple)(nil) {
		fmt.Println("check() == nil")
	} else {
		fmt.Printf("%#v\n",check())
		fmt.Println("check() != nil")
	}
	fmt.Println()
}
```