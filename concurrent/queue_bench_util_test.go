package concurrent

import (
	"runtime"
)

// benchConcurrency 自适应并发数: 取CPU核数,最少1,最多16
func benchConcurrency() int {
	n := runtime.NumCPU()
	if n < 1 {
		n = 1
	}
	if n > 16 {
		n = 16
	}
	return n
}

// qBuilder 队列基准测试构造器
type qBuilder struct {
	name string
	newQ func() Queue[int]
}

var allQ = []qBuilder{
	{"seg", func() Queue[int] { return newSegQueue[int]() }},
	{"ring", func() Queue[int] { return newRingQueue[int](1024) }},
	{"chan", func() Queue[int] { return newChanQueue[int](128) }},
}

func drainQ(q Queue[int]) {
	for {
		_, ok := q.Dequeue()
		if !ok {
			return
		}
	}
}
