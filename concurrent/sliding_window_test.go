package concurrent

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// waitForCondition 等待条件满足，超时则 fatal
func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal(msg)
}

// fillAndCheckLimit 连续请求, 放行数不得超过 allowNumber, 且达到上限后应拒绝
func fillAndCheckLimit(t *testing.T, sw *SlidingWindow) int {
	t.Helper()
	allowed := 0
	for i := 0; i < int(sw.allowNumber)*2; i++ {
		if sw.CanDo() {
			allowed++
		}
	}
	if int32(allowed) != sw.allowNumber {
		t.Fatalf("窗口内放行数=%d, 预期 %d", allowed, sw.allowNumber)
	}
	if sw.CanDo() {
		t.Fatal("达到上限后应拒绝请求")
	}
	return allowed
}

func Test_SlidingWindow_BasicFunctionality(t *testing.T) {
	t.Parallel()
	sw, err := NewSlidingWindow(1000, 100)
	if err != nil {
		t.Fatalf("创建滑动窗口失败: %v", err)
	}

	fillAndCheckLimit(t, sw)

	// 等待整个窗口过期后恢复放行
	waitForCondition(t, time.Duration(sw.windowTime+1000)*time.Millisecond, func() bool { return sw.CanDo() }, "窗口过期后应该允许请求")
}

func Test_SlidingWindow_DifferentParams(t *testing.T) {
	testCases := []struct {
		name        string
		timeWindow  int64
		allowNumber int32
	}{
		{"小窗口", 100, 10},
		{"大窗口", 5000, 500},
		{"高频率", 1000, 1000},
		{"低频率", 1000, 10},
		{"默认参数", 1000, 100},
		{"微小窗口", 333, 9},
		{"多窗口", 2220, 110},
		{"快速率", 660, 300},
	}
	t.Parallel()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sw, err := NewSlidingWindow(tc.timeWindow, tc.allowNumber)
			if err != nil {
				t.Fatalf("创建滑动窗口失败: %v", err)
			}

			fillAndCheckLimit(t, sw)

			waitForCondition(t, time.Duration(sw.windowTime+1000)*time.Millisecond, func() bool { return sw.CanDo() }, "窗口过期后应该允许请求")
		})
	}
}

// 随机参数测试
func TestSlidingWindow_RandomParams(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		allowNumber := int32(rand.Intn(100) + 1)
		timeWindow := int64(rand.Intn(1000) + 1)

		t.Run(fmt.Sprintf("随机测试_%d_%d", allowNumber, timeWindow), func(t *testing.T) {
			t.Parallel()
			sw, err := NewSlidingWindow(timeWindow, allowNumber)
			if err != nil {
				t.Fatalf("创建滑动窗口失败: %v", err)
			}

			fillAndCheckLimit(t, sw)

			waitForCondition(t, time.Duration(sw.windowTime+1000)*time.Millisecond, func() bool { return sw.CanDo() }, "窗口过期后应该允许请求")
		})
	}
}

// 低频率请求后长时间等待的测试
func Test_SlidingWindow_LowFrequencyLongWait(t *testing.T) {
	testCases := []struct {
		name            string
		timeWindow      int64
		allowNumber     int32
		waitMultiplier  int64
		requestInterval time.Duration
	}{
		{"低频短窗口", 500, 3, 2, time.Millisecond * 100},
		{"低频中等窗口", 1000, 20, 3, time.Millisecond * 200},
		{"极低频短窗口", 500, 5, 5, time.Second * 1},
	}
	t.Parallel()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sw, err := NewSlidingWindow(tc.timeWindow, tc.allowNumber)
			if err != nil {
				t.Fatalf("创建滑动窗口失败: %v", err)
			}

			for i := 0; i < int(tc.allowNumber/2); i++ {
				sw.CanDo()
				time.Sleep(tc.requestInterval)
			}

			timeout := time.Duration(sw.windowTime*tc.waitMultiplier+200) * time.Millisecond
			waitForCondition(t, timeout, func() bool { return sw.CanDo() }, "长时间等待后应该能够执行请求")
		})
	}
}

// 启动期不得超发: 首个窗口最多放行 allowNumber 个
func Test_SlidingWindow_StartupNoOvershoot(t *testing.T) {
	t.Parallel()
	sw, err := NewSlidingWindow(60*1000, 100)
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	allowed := 0
	for i := 0; i < 100*2; i++ {
		if sw.CanDo() {
			allowed++
		}
	}
	if int32(allowed) != sw.allowNumber {
		t.Fatalf("启动期应仅放行窗口配额 %d, 实际 %d", sw.allowNumber, allowed)
	}
}

// 固定假时钟下任意滑动窗口内的放行数不得超过 allowNumber
func Test_SlidingWindow_FakeClockStrictWindow(t *testing.T) {
	t.Parallel()
	const (
		windowTime  = int64(1000)
		allowNumber = int32(100)
	)
	sw, err := NewSlidingWindow(windowTime, allowNumber)
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	now := int64(1_700_000_000_000)
	sw.Now = func() int64 { return now }

	// 窗口内放行数不得超过配额
	burst := 0
	for i := 0; i < 1000; i++ {
		if sw.CanDo() {
			burst++
		}
	}
	if int32(burst) != allowNumber {
		t.Fatalf("固定时钟下放行数=%d, 预期 %d", burst, allowNumber)
	}

	// 窗口未过期时不得恢复
	now += windowTime - 1
	if sw.CanDo() {
		t.Fatal("窗口未过期不应恢复放行")
	}

	// 窗口过期后可再次放行
	now++
	for i := 0; i < 1000; i++ {
		if sw.CanDo() {
			burst++
		}
	}
	if int32(burst) != allowNumber*2 {
		t.Fatalf("窗口过期后放行总数=%d, 预期 %d", burst, allowNumber*2)
	}
}

// 并发请求下不得超发
func Test_SlidingWindow_ConcurrentNoOvershoot(t *testing.T) {
	t.Parallel()
	sw, err := NewSlidingWindow(60*1000, 100)
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	sw.Now = func() int64 { return 1_700_000_000_000 }

	var allowed int32
	var wg sync.WaitGroup
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if sw.CanDo() {
					atomic.AddInt32(&allowed, 1)
				}
			}
		}()
	}
	wg.Wait()
	if allowed != sw.allowNumber {
		t.Fatalf("并发放行数=%d, 预期 %d", allowed, sw.allowNumber)
	}
}

// Test_SlidingWindow_InvalidWindowTime 测试窗口长度为非正数时返回ErrWindowTime
func Test_SlidingWindow_InvalidWindowTime(t *testing.T) {
	t.Parallel()
	for _, windowTime := range []int64{0, -1} {
		if _, err := NewSlidingWindow(windowTime, 10); !errors.Is(err, ErrWindowTime) {
			t.Errorf("窗口长度 %d 应返回 ErrWindowTime, 得到 %v", windowTime, err)
		}
	}
}

// Test_SlidingWindow_InvalidAllowNumber 测试允许数为非正数时返回ErrAllowNumber
func Test_SlidingWindow_InvalidAllowNumber(t *testing.T) {
	t.Parallel()
	for _, allowNumber := range []int32{0, -1} {
		if _, err := NewSlidingWindow(1000, allowNumber); !errors.Is(err, ErrAllowNumber) {
			t.Errorf("允许数 %d 应返回 ErrAllowNumber, 得到 %v", allowNumber, err)
		}
	}
}
