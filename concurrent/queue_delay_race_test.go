package concurrent

import (
	"sync/atomic"
	"testing"
	"time"
)

// 消费者退出与生产者start的CAS竞态: 退出窗口期入队的元素不得永久滞留
// 概率性复现窗口(消费者退出临界区数条指令), 循环压测放大
// delay取5ms: 避免1us级sleep触发Windows定时器病态(见报告), 不影响竞态覆盖
// 轮数受每轮delay下限约束: 600轮*5ms=3s, 需低于CI包级timeout
func Test_DelayQueueRestartRace(t *testing.T) {
	t.Parallel()
	q := newDelayQueue[int](5 * time.Millisecond)
	var got int64
	deadline := time.Now().Add(5 * time.Second)
	rounds := 0
	for rounds < 600 {
		rounds++
		q.Enqueue(rounds)
		// 等待元素到达out; 滞留即失败
		for {
			if _, ok := q.Dequeue(); ok {
				atomic.AddInt64(&got, 1)
				break
			}
			if time.Now().After(deadline) {
				// 元素可能在途中(消费者已取走正sleep等待到期), 宽限确认是否真滞留
				// 真滞留时宽限期内依然取不到, 不会掩盖bug
				graceDeadline := time.Now().Add(200 * time.Millisecond)
				recovered := false
				for time.Now().Before(graceDeadline) {
					if _, ok := q.Dequeue(); ok {
						atomic.AddInt64(&got, 1)
						recovered = true
						break
					}
					time.Sleep(2 * time.Millisecond)
				}
				if recovered {
					break
				}
				dq := q.(*delayQueue[int])
				t.Fatalf("元素滞留: round=%d got=%d in=%d out=%d running=%d",
					rounds, atomic.LoadInt64(&got), dq.in.Size(), dq.out.Size(),
					atomic.LoadInt32(&dq.running))
			}
			time.Sleep(2 * time.Millisecond)
		}
	}
}
