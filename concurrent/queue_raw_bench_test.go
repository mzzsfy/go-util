package concurrent

import (
	"sync/atomic"
	"testing"
)

// BenchmarkRaw_1P1C 单生产者单消费者吞吐量
func BenchmarkRaw_1P1C(b *testing.B) {
	for _, tc := range allQ {
		b.Run(tc.name, func(b *testing.B) {
			q := tc.newQ()
			var done int32
			go func() {
				for atomic.LoadInt32(&done) == 0 {
					q.Dequeue()
				}
				drainQ(q)
			}()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q.Enqueue(i)
			}
			b.StopTimer()
			atomic.StoreInt32(&done, 1)
			drainQ(q)
		})
	}
	b.Run("chan", func(b *testing.B) {
		ch := make(chan int, 128)
		var done int32
		go func() {
			for atomic.LoadInt32(&done) == 0 {
				<-ch
			}
		}()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch <- i
		}
		b.StopTimer()
		atomic.StoreInt32(&done, 1)
	})
}

// BenchmarkRaw_PingPong 1生产1消费往返延迟
func BenchmarkRaw_PingPong(b *testing.B) {
	for _, tc := range allQ {
		b.Run(tc.name, func(b *testing.B) {
			q1 := tc.newQ()
			q2 := tc.newQ()
			done := make(chan struct{})
			go func() {
				defer close(done)
				for i := 0; i < b.N; i++ {
					for {
						_, ok := q1.Dequeue()
						if ok {
							break
						}
					}
					q2.Enqueue(1)
				}
			}()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q1.Enqueue(1)
				for {
					_, ok := q2.Dequeue()
					if ok {
						break
					}
				}
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

// BenchmarkRaw_MPMC 多生产多消费吞吐量
// 每个goroutine入队后尝试出队一次,测量并发下的综合吞吐
func BenchmarkRaw_MPMC(b *testing.B) {
	for _, tc := range allQ {
		b.Run(tc.name, func(b *testing.B) {
			q := tc.newQ()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					q.Enqueue(1)
					q.Dequeue()
				}
			})
		})
	}
	b.Run("chan", func(b *testing.B) {
		ch := make(chan int, 128)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ch <- 1
				<-ch
			}
		})
	})
}

// BenchmarkRaw_Scale 不同并发度下的MPMC吞吐量
func BenchmarkRaw_Scale(b *testing.B) {
	cpuCount := benchConcurrency()
	levels := []int{1}
	for l := 2; l <= cpuCount*2; l *= 2 {
		levels = append(levels, l)
	}
	for _, n := range levels {
		for _, tc := range allQ {
			b.Run(tc.name+"/"+itoa(n)+"G", func(b *testing.B) {
				q := tc.newQ()
				b.ResetTimer()
				b.SetParallelism(n)
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						q.Enqueue(1)
						q.Dequeue()
					}
				})
			})
		}
		b.Run("chan/"+itoa(n)+"G", func(b *testing.B) {
			ch := make(chan int, 128)
			b.ResetTimer()
			b.SetParallelism(n)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					ch <- 1
					<-ch
				}
			})
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 4)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
