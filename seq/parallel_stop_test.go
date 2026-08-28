package seq

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// awaitResult 带超时等待执行结果,防止缺陷死锁导致测试长时间挂起
func awaitResult(d time.Duration, f func()) string {
	done := make(chan string, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Sprintf("panic:%v", r)
			}
		}()
		f()
		done <- "ok"
	}()
	select {
	case msg := <-done:
		return msg
	case <-time.After(d):
		return "timeout"
	}
}

// 场景: 有序归并goroutine需捕获消费者提前停止
// Given MapParallel(order=2)后接Take,转换耗时故意不均制造乱序完成
// When 消费者取足元素后发出Stop,停止发生在归并goroutine内
// Then 停止被吸收,不崩溃不死锁,返回前n个元素且保持顺序
func Test_MapParallel_Order2_TakeStop(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		msg := awaitResult(5*time.Second, func() {
			r := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
				MapParallel(func(v int) any {
					time.Sleep(time.Duration(8-v) * time.Millisecond)
					return v * 10
				}, 2, 4).
				Take(3).ToSlice()
			if len(r) != 3 {
				t.Fatalf("Take应返回3个元素,实际%d个:%v", len(r), r)
			}
			for j, e := range r {
				if e.(int) != (j+1)*10 {
					t.Fatalf("元素应按序为%d0,实际:%v", j+1, r)
				}
			}
		})
		if msg != "ok" {
			t.Fatalf("第%d次迭代异常:%s", i, msg)
		}
	}
}

// 场景: First走Take(1)同类路径
// Given MapParallel(order=2)后接First
// When 取第一个元素后发出Stop
// Then 不崩溃不死锁,返回第一个元素
func Test_MapParallel_Order2_FirstStop(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		msg := awaitResult(5*time.Second, func() {
			f := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
				MapParallel(func(v int) any { return v * 10 }, 2, 4).
				First()
			if f == nil || (*f).(int) != 10 {
				t.Fatalf("First应返回10,实际:%v", f)
			}
		})
		if msg != "ok" {
			t.Fatalf("第%d次迭代异常:%s", i, msg)
		}
	}
}

// 场景: BiSeq有序归并goroutine需捕获消费者提前停止
// Given MapVParallel(order=2)后接Take
// When 消费者取足元素后发出Stop
// Then 停止被吸收,不崩溃不死锁
func Test_BiMapVParallel_Order2_TakeStop(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		msg := awaitResult(5*time.Second, func() {
			count := 0
			BiFromSeq(FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}), func(v int) (int, int) {
				return v, v * 10
			}).MapVParallel(func(k int, v int) any {
				time.Sleep(time.Duration(8-k) * time.Millisecond)
				return v
			}, 2, 4).Take(3)(func(k int, v any) {
				count++
			})
			if count != 3 {
				t.Fatalf("Take应消费3个元素,实际%d个", count)
			}
		})
		if msg != "ok" {
			t.Fatalf("第%d次迭代异常:%s", i, msg)
		}
	}
}

// 场景: order=3消费回调持锁期间收到Stop,锁须先归还再recover
// Given MapParallel(order=3)后接Take
// When 消费者取足元素后发出Stop,停止发生在worker持锁调用消费回调时
// Then 不死锁不panic,返回前n个元素且保持顺序
func Test_MapParallel_Order3_TakeStop(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		msg := awaitResult(5*time.Second, func() {
			r := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
				MapParallel(func(v int) any {
					return v * 10
				}, 3, 4).
				Take(3).ToSlice()
			if len(r) != 3 {
				t.Fatalf("Take应返回3个元素,实际%d个:%v", len(r), r)
			}
			for j, e := range r {
				if e.(int) != (j+1)*10 {
					t.Fatalf("元素应按序为%d0,实际:%v", j+1, r)
				}
			}
		})
		if msg != "ok" {
			t.Fatalf("第%d次迭代异常:%s", i, msg)
		}
	}
}

// 场景: BiSeq order=3消费回调持锁期间收到Stop
// Given MapVParallel(order=3)后接Take
// When 消费者取足元素后发出Stop
// Then 不死锁不panic,返回前n个元素且保持顺序
func Test_BiMapVParallel_Order3_TakeStop(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		msg := awaitResult(5*time.Second, func() {
			count := 0
			BiFromSeq(FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}), func(v int) (int, int) {
				return v, v * 10
			}).MapVParallel(func(k int, v int) any {
				return v
			}, 3, 4).Take(3)(func(k int, v any) {
				count++
			})
			if count != 3 {
				t.Fatalf("Take应消费3个元素,实际%d个", count)
			}
		})
		if msg != "ok" {
			t.Fatalf("第%d次迭代异常:%s", i, msg)
		}
	}
}

// 场景: order=1路径Stop哨兵不得作为任务错误存储,否则Wait在源头stopRecover退栈后重抛裸指针
// Given MapParallel(order=1)经Sync保证消费不竞争后接Take
// When 消费者取足元素后发出Stop
// Then 正常返回已消费部分,无panic
func Test_MapParallel_Order1_TakeStop(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		msg := awaitResult(5*time.Second, func() {
			r := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
				MapParallel(func(v int) any {
					return v * 10
				}, 1, 4).Sync().
				Take(3).ToSlice()
			if len(r) != 3 {
				t.Fatalf("Take应返回3个元素,实际%d个:%v", len(r), r)
			}
			//order=1为完成序消费,不保证源顺序,只校验元素取自源
			for _, e := range r {
				v := e.(int)
				if v < 10 || v > 80 || v%10 != 0 {
					t.Fatalf("元素应为源元素转换值,实际:%v", r)
				}
			}
		})
		if msg != "ok" {
			t.Fatalf("第%d次迭代异常:%s", i, msg)
		}
	}
}

// 场景: BiSeq order=1路径Stop哨兵不得作为任务错误存储
// Given MapVParallel(order=1)经Sync保证消费不竞争后接Take
// When 消费者取足元素后发出Stop
// Then 正常返回已消费部分,无panic
func Test_BiMapVParallel_Order1_TakeStop(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		msg := awaitResult(5*time.Second, func() {
			count := 0
			BiFromSeq(FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}), func(v int) (int, int) {
				return v, v * 10
			}).MapVParallel(func(k int, v int) any {
				return v
			}, 1, 4).Sync().Take(3)(func(k int, v any) {
				count++
			})
			if count != 3 {
				t.Fatalf("Take应消费3个元素,实际%d个", count)
			}
		})
		if msg != "ok" {
			t.Fatalf("第%d次迭代异常:%s", i, msg)
		}
	}
}

// 场景: Parallel路径Stop哨兵不得作为任务错误存储
// Given Parallel后经Sync保证消费不竞争再接Take
// When 消费者取足元素后发出Stop
// Then 正常返回已消费部分,无panic
func Test_Parallel_TakeStop(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		msg := awaitResult(5*time.Second, func() {
			r := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
				Parallel(4).Sync().
				Take(3).ToSlice()
			if len(r) != 3 {
				t.Fatalf("Take应返回3个元素,实际%d个:%v", len(r), r)
			}
			//Parallel为完成序消费,不保证源顺序,只校验元素取自源
			for _, e := range r {
				if e < 1 || e > 8 {
					t.Fatalf("元素应为源元素,实际:%v", r)
				}
			}
		})
		if msg != "ok" {
			t.Fatalf("第%d次迭代异常:%s", i, msg)
		}
	}
}

// 场景: 任务panic经生产中止路径退出时,常驻worker与归并goroutine须全部回收
// Given 四条并行路径(order=1/2/3与Parallel)各自任务panic
// When 连续多次触发生产中止
// Then 每次panic值重抛,goroutine数量回落至基线
func Test_ParallelCore_TaskPanic_NoGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		for _, order := range []int{1, 2, 3} {
			msg := awaitResult(5*time.Second, func() {
				//order=1消费竞争,按语义以Sync保证消费不竞争
				FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).MapParallel(func(v int) any {
					if v == 2 {
						panic("boom")
					}
					return v
				}, order, 4).Sync().ToSlice()
			})
			if msg != "panic:boom" {
				t.Fatalf("第%d次迭代order=%d,期望panic:boom,实际:%s", i, order, msg)
			}
		}
		msg := awaitResult(5*time.Second, func() {
			FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).Parallel(4)(func(v int) {
				if v == 2 {
					panic("boom")
				}
			})
		})
		if msg != "panic:boom" {
			t.Fatalf("第%d次迭代Parallel,期望panic:boom,实际:%s", i, msg)
		}
	}
	//worker经wg.Wait确认退出,归并goroutine退出有微小滞后,给一次调度余量
	for k := 0; k < 100; k++ {
		runtime.Gosched()
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("goroutine泄漏,基线%d,当前%d", before, runtime.NumGoroutine())
}

// 场景: order=2生产中止路径中归并goroutine须可退出
// Given MapParallel(order=2)任务panic,生产者经Add快速失败中止
// When 中止时results通道未关闭
// Then 归并goroutine经兜底关闭退出,goroutine数量回落至基线
func Test_MapParallel_Order2_Panic_NoGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		msg := awaitResult(5*time.Second, func() {
			FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
				MapParallel(func(v int) any {
					time.Sleep(time.Duration(8-v) * time.Millisecond)
					if v == 2 {
						panic("boom")
					}
					return v
				}, 2, 4).ToSlice()
		})
		if msg != "panic:boom" {
			t.Fatalf("第%d次迭代,期望panic:boom,实际:%s", i, msg)
		}
	}
	for k := 0; k < 100; k++ {
		runtime.Gosched()
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("goroutine泄漏,基线%d,当前%d", before, runtime.NumGoroutine())
}

// 场景: order=3任务panic需推进序号并重抛
// Given MapParallel(order=3)且元素2的转换panic
// When 消费完整序列
// Then 不发生worker永久等待,panic值经Wait重抛
func Test_MapParallel_Order3_PanicRethrow(t *testing.T) {
	t.Parallel()
	for i := 0; i < 200; i++ {
		msg := awaitResult(3*time.Second, func() {
			FromSlice([]int{1, 2, 3, 4}).MapParallel(func(v int) any {
				if v == 2 {
					panic("boom")
				}
				return v
			}, 3, 4).ToSlice()
		})
		if msg != "panic:boom" {
			t.Fatalf("第%d次迭代,期望panic重抛boom,实际:%s", i, msg)
		}
	}
}

// 场景: order=3后半段任务panic也需推进序号
// Given MapParallel(order=3)且最后一个元素转换panic
// When 消费完整序列
// Then 不死锁,panic值重抛
func Test_MapParallel_Order3_LastPanicRethrow(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		msg := awaitResult(3*time.Second, func() {
			FromSlice([]int{1, 2, 3, 4}).MapParallel(func(v int) any {
				if v == 4 {
					panic("tail-boom")
				}
				return v
			}, 3, 4).ToSlice()
		})
		if msg != "panic:tail-boom" {
			t.Fatalf("第%d次迭代,期望panic重抛tail-boom,实际:%s", i, msg)
		}
	}
}

// 场景: BiSeq order=3任务panic需推进序号并重抛
// Given MapVParallel(order=3)且元素panic
// When 消费完整序列
// Then 不死锁,panic值经Wait重抛
func Test_BiMapVParallel_Order3_PanicRethrow(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		msg := awaitResult(3*time.Second, func() {
			count := 0
			BiFromSeq(FromSlice([]int{1, 2, 3, 4}), func(v int) (int, int) {
				return v, v
			}).MapVParallel(func(k int, v int) any {
				if k == 2 {
					panic("bi-boom")
				}
				return v
			}, 3, 4)(func(k int, v any) {
				count++
			})
		})
		if msg != "panic:bi-boom" {
			t.Fatalf("第%d次迭代,期望panic重抛bi-boom,实际:%s", i, msg)
		}
	}
}
