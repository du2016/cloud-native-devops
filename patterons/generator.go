// 生成器
package main

import "log"

func Count(start, end int) <-chan int {
	ch := make(chan int)

	go func(ch chan int) {
		for i := start; i <= end; i ++ {
			ch <- i
		}
		close(ch)
	}(ch)

	return ch
}

func main()  {
	aa:=Count(1,5)
	for i :=range aa {
		log.Println(i)
	}
}