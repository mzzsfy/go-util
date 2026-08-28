package concurrent

import (
	"testing"
	"time"
)

// BenchmarkQueueDequeueTimeoutExpire 空队列超时出队, 每次必然走 timer 创建路径
// 衡量超时出队的 timer 分配成本(allocs/op)
func BenchmarkQueueDequeueTimeoutExpire(b *testing.B) {
	for _, tc := range allQ {
		b.Run(tc.name, func(b *testing.B) {
			bq := BlockQueueWrapper(tc.newQ())
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bq.DequeueBlock(time.Microsecond * 50)
			}
		})
	}
	b.Run("chanraw", func(b *testing.B) {
		q := BlockQueueWrapper(newChanQueue[int](1))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			q.DequeueBlock(time.Microsecond * 50)
		}
	})
}

// BenchmarkQueueDequeueTimeoutHit 消费者阻塞等待期间生产者入队, 命中路径
func BenchmarkQueueDequeueTimeoutHit(b *testing.B) {
	for _, tc := range allQ {
		b.Run(tc.name, func(b *testing.B) {
			bq := BlockQueueWrapper(tc.newQ())
			go func() {
				for i := 0; i < b.N; i++ {
					bq.Enqueue(1)
					time.Sleep(time.Microsecond * 100)
				}
			}()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bq.DequeueBlock(time.Second)
			}
			b.StopTimer()
		})
	}
}
