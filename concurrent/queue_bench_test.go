package concurrent

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkQueueEnqueueScalability 多协程入队吞吐量,消费者数=CPU核数
func BenchmarkQueueEnqueueScalability(b *testing.B) {
	n := benchConcurrency()
	for _, tc := range allQ {
		b.Run(tc.name, func(b *testing.B) {
			q := tc.newQ()
			bq := BlockQueueWrapper(q)
			var over int32
			for i := 0; i < n; i++ {
				go func() {
					for atomic.LoadInt32(&over) == 0 {
						bq.DequeueBlock(time.Second)
					}
				}()
			}
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					bq.Enqueue(1)
				}
			})
			atomic.StoreInt32(&over, 1)
		})
	}
	b.Run("chan", func(b *testing.B) {
		ch := make(chan int, 128)
		var over int32
		for i := 0; i < n; i++ {
			go func() {
				for atomic.LoadInt32(&over) == 0 {
					<-ch
				}
			}()
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ch <- 1
			}
		})
		atomic.StoreInt32(&over, 1)
	})
}

// BenchmarkQueuePingPong 1生产1消费阻塞往返延迟
func BenchmarkQueuePingPong(b *testing.B) {
	for _, tc := range allQ {
		b.Run(tc.name, func(b *testing.B) {
			q1 := tc.newQ()
			bq1 := BlockQueueWrapper(q1)
			q2 := tc.newQ()
			bq2 := BlockQueueWrapper(q2)
			done := make(chan struct{})
			go func() {
				defer close(done)
				for i := 0; i < b.N; i++ {
					bq1.DequeueBlock()
					bq2.Enqueue(1)
				}
			}()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bq1.Enqueue(1)
				bq2.DequeueBlock()
			}
			b.StopTimer()
			<-done
		})
	}
	b.Run("chan", func(b *testing.B) {
		ch1 := make(chan int, 1)
		ch2 := make(chan int, 1)
		done := make(chan struct{})
		go func() {
			defer close(done)
			for i := 0; i < b.N; i++ {
				<-ch1
				ch2 <- 1
			}
		}()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch1 <- 1
			<-ch2
		}
		b.StopTimer()
		<-done
	})
}

// BenchmarkQueueBlockNotify 阻塞队列通知机制性能
func BenchmarkQueueBlockNotify(b *testing.B) {
	q := BlockQueueWrapper[int](newSegQueue[int]())
	var wg sync.WaitGroup
	// 消费者数=min(CPU核数, 10)
	n := benchConcurrency()
	if n > 10 {
		n = 10
	}
	var received int64
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				_, ok := q.DequeueBlock(time.Millisecond * 10)
				if !ok {
					return
				}
				if atomic.AddInt64(&received, 1) >= int64(b.N) {
					return
				}
			}
		}()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
	}
	deadline := time.After(5 * time.Second)
	for atomic.LoadInt64(&received) < int64(b.N) {
		select {
		case <-deadline:
			b.Fatalf("timeout: received=%d want=%d", received, b.N)
		default:
			runtime.Gosched()
		}
	}
	b.StopTimer()
	wg.Wait()
}
