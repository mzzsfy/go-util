package concurrent

import (
	"errors"
	"fmt"
	"math/rand"
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

func Test_SlidingWindow_BasicFunctionality(t *testing.T) {
	t.Parallel()
	// 测试参数设置
	timeWindow := int64(1000) // 时间窗口1秒
	allowNumber := int32(100) // 允许最大请求数100
	windowNumber := int32(5)  // 窗口数量5

	// 创建滑动窗口实例
	sw, err := NewSlidingWindow(timeWindow, allowNumber, windowNumber)
	if err != nil {
		t.Fatalf("创建滑动窗口失败: %v", err)
	}

	// 测试在时间窗口内能否正确控制请求
	successCount := 0
	for i := 0; i < int(allowNumber)*2; i++ { // 尝试双倍请求次数
		if sw.CanDo() {
			successCount++
		}

		// 如果已经达到了预期数量，就跳出循环
		if int32(successCount) >= allowNumber {
			break
		}
	}

	// 验证成功请求数不超过限制
	if int32(successCount) > allowNumber {
		t.Errorf("请求数量超过限制: 得到 %d, 预期最多 %d", successCount, allowNumber)
	}

	// 此时应该已经到达限制，下一个请求应该失败
	if sw.CanDo() {
		t.Error("达到限制后应该拒绝请求")
	}

	// 使用条件等待窗口滑动，确保 CanDo() 返回 true
	// 每片时长+1s裕量: 并行负载下Windows定时器粒度与调度延迟
	timeout := time.Duration(sw.everyWindowTime+1000) * time.Millisecond
	waitForCondition(t, timeout, func() bool { return sw.CanDo() }, "窗口滑动后应该允许请求")
}

// 添加不同参数的测试用例
func Test_SlidingWindow_DifferentParams(t *testing.T) {
	testCases := []struct {
		name         string
		timeWindow   int64
		allowNumber  int32
		windowNumber int32
	}{
		{"小窗口", 100, 10, 5},
		{"大窗口", 5000, 500, 10},
		{"高频率", 1000, 1000, 5},
		{"低频率", 1000, 10, 10},
		{"默认参数", 1000, 100, 5},
		{"微小窗口", 333, 9, 3},
		{"多窗口", 2220, 110, 10},
		{"快速率", 660, 300, 10},
	}
	t.Parallel()
	for _, tc := range testCases {
		//兼容旧版本循环变量语义,避免并发捕获同一变量
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// 创建滑动窗口实例
			sw, err := NewSlidingWindow(tc.timeWindow, tc.allowNumber, tc.windowNumber)
			if err != nil {
				t.Fatalf("创建滑动窗口失败: %v", err)
			}

			// 测试在时间窗口内能否正确控制请求数量
			successCount := 0
			maxAttempts := int(tc.allowNumber * 2) // 尝试足够多次以测试限制
			for i := 0; i < maxAttempts; i++ {
				if sw.CanDo() {
					successCount++
				}

				// 如果已经达到了预期数量，就跳出循环
				if int32(successCount) >= tc.allowNumber {
					break
				}
			}

			// 验证请求数量是否符合预期
			if int32(successCount) > tc.allowNumber {
				t.Errorf("请求数量超过限制: 得到 %d, 预期最多 %d", successCount, tc.allowNumber)
			}

			// 等待窗口滑动，确保 CanDo() 返回 true
			// 每片时长+1s裕量: 并行负载下Windows定时器粒度与调度延迟
			timeout := time.Duration(sw.everyWindowTime+1000) * time.Millisecond
			waitForCondition(t, timeout, func() bool { return sw.CanDo() }, "窗口滑动后应该允许请求")
		})
	}
}

// 随机参数测试
func TestSlidingWindow_RandomParams(t *testing.T) {
	t.Parallel()
	// 减少随机测试用例数量以提高CI稳定性
	for i := 0; i < 100; i++ {
		// 生成随机参数，使用较小的窗口时间以加快测试
		windowNumber := int32(rand.Intn(10) + 3)                 // 窗口数量 3-12
		allowNumber := int32(rand.Intn(100) + int(windowNumber)) // 允许请求数 >= windowNumber
		// +1 保证严格大于 windowNumber, 满足 NewSlidingWindow 校验
		timeWindow := int64(rand.Intn(1000) + int(windowNumber) + 1) // 时间窗口 > windowNumber，减少范围

		t.Run(fmt.Sprintf("随机测试_%d_%d_%d", windowNumber, allowNumber, timeWindow), func(t *testing.T) {
			t.Parallel()

			// 创建滑动窗口实例
			sw, err := NewSlidingWindow(timeWindow, allowNumber, windowNumber)
			if err != nil {
				t.Fatalf("创建滑动窗口失败: %v", err)
			}

			// 测试在时间窗口内能否正确控制请求数量
			successCount := 0
			maxAttempts := int(allowNumber * 2) // 尝试足够多次以测试限制
			for j := 0; j < maxAttempts; j++ {
				if sw.CanDo() {
					successCount++
				}

				// 如果已经达到了预期数量，就跳出循环
				if int32(successCount) >= allowNumber {
					break
				}
			}

			// 验证请求数量是否符合预期
			if int32(successCount) > allowNumber {
				t.Errorf("请求数量超过限制: 得到 %d, 预期最多 %d", successCount, allowNumber)
			}

			// 等待窗口滑动，确保 CanDo() 返回 true
			// 每片时长+1s裕量: 并行负载下Windows定时器粒度与调度延迟
			timeout := time.Duration(sw.everyWindowTime+1000) * time.Millisecond
			waitForCondition(t, timeout, func() bool { return sw.CanDo() }, "窗口滑动后应该允许请求")
		})
	}
}

// 低频率请求后长时间等待的测试
func Test_SlidingWindow_LowFrequencyLongWait(t *testing.T) {
	testCases := []struct {
		name            string
		timeWindow      int64
		allowNumber     int32
		windowNumber    int32
		waitMultiplier  int64         // 等待时间是窗口时间的倍数
		requestInterval time.Duration // 请求间隔
	}{
		{"低频短窗口", 500, 3, 3, 2, time.Millisecond * 100},
		{"低频中等窗口", 1000, 20, 5, 3, time.Millisecond * 200},
		{"极低频短窗口", 500, 5, 5, 5, time.Second * 1},
	}
	t.Parallel()
	for _, tc := range testCases {
		//兼容旧版本循环变量语义,避免并发捕获同一变量
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// 创建滑动窗口实例
			sw, err := NewSlidingWindow(tc.timeWindow, tc.allowNumber, tc.windowNumber)
			if err != nil {
				t.Fatalf("创建滑动窗口失败: %v", err)
			}

			// 发送少量请求（低频率）
			successCount := 0
			for i := 0; i < int(tc.allowNumber/2); i++ {
				if sw.CanDo() {
					successCount++
				}
				time.Sleep(tc.requestInterval)
			}

			// 等待窗口滑动，确保至少有一个请求可以成功
			timeout := time.Duration(sw.everyWindowTime*tc.waitMultiplier+200) * time.Millisecond
			var postWaitSuccess int32
			waitForCondition(t, timeout, func() bool {
				if sw.CanDo() {
					postWaitSuccess++
					return true
				}
				return false
			}, "长时间等待后应该能够执行请求")

			// 验证等待后能够执行请求
			if postWaitSuccess == 0 {
				t.Error("长时间等待后应该能够执行请求")
			}
		})
	}
}

// Test_SlidingWindow_StartupNoOvershoot 启动期不得因零值分片时间戳触发立即滑动导致超发
func Test_SlidingWindow_StartupNoOvershoot(t *testing.T) {
	t.Parallel()
	// 窗口时间远大于测试耗时, 排除真实时间流逝导致的正常滑动
	sw, err := NewSlidingWindow(60*1000, 100, 5)
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	allowed := 0
	for i := 0; i < 100*2; i++ {
		if sw.CanDo() {
			allowed++
		}
	}
	// 零值分片时间戳会触发一次立即滑动, 放行数达到首窗口限额的两倍
	if int32(allowed) != sw.limits[0] {
		t.Fatalf("启动期应仅放行首窗口限额 %d, 实际 %d", sw.limits[0], allowed)
	}
}

// Test_SlidingWindow_InvalidWindowNumber 测试窗口数量<=2时返回ErrWindowNumber
func Test_SlidingWindow_InvalidWindowNumber(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		timeWindow   int64
		allowNumber  int32
		windowNumber int32
	}{
		{"窗口数为1", 1000, 100, 1},
		{"窗口数为2", 1000, 100, 2},
		{"窗口数为0", 1000, 100, 0},
		{"窗口数为-1", 1000, 100, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSlidingWindow(tc.timeWindow, tc.allowNumber, tc.windowNumber)
			if !errors.Is(err, ErrWindowNumber) {
				t.Errorf("窗口数 %d 应返回 ErrWindowNumber, 得到 %v", tc.windowNumber, err)
			}
		})
	}
}

// Test_SlidingWindow_InvalidWindowTime 测试时间小于等于窗口数量时返回ErrWindowTime
func Test_SlidingWindow_InvalidWindowTime(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		timeWindow   int64
		allowNumber  int32
		windowNumber int32
	}{
		{"时间窗口为0", 0, 10, 5},
		{"时间窗口小于窗口数", 3, 10, 5},
		{"时间窗口等于窗口数", 5, 10, 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSlidingWindow(tc.timeWindow, tc.allowNumber, tc.windowNumber)
			if !errors.Is(err, ErrWindowTime) {
				t.Errorf("时间窗口 %d 小于或等于窗口数 %d 应返回 ErrWindowTime, 得到 %v", tc.timeWindow, tc.windowNumber, err)
			}
		})
	}
}

// Test_SlidingWindow_InvalidAllowNumber 测试允许数小于窗口数时返回ErrAllowNumber
func Test_SlidingWindow_InvalidAllowNumber(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		timeWindow   int64
		allowNumber  int32
		windowNumber int32
	}{
		{"允许数小于窗口数", 1000, 2, 5},
		{"允许数为0", 1000, 0, 5},
		{"允许数为负数", 1000, -1, 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSlidingWindow(tc.timeWindow, tc.allowNumber, tc.windowNumber)
			if !errors.Is(err, ErrAllowNumber) {
				t.Errorf("允许数 %d 小于窗口数 %d 应返回 ErrAllowNumber, 得到 %v", tc.allowNumber, tc.windowNumber, err)
			}
		})
	}
}

// 全零权重: 总权重为0时应返回全0分配, 而非panic
func Test_AvgByWeight_AllZeroWeights(t *testing.T) {
	t.Parallel()
	res := avgByWeight(100, 0, 0, 0)
	if len(res) != 3 {
		t.Fatalf("len=%d want=3", len(res))
	}
	for i, v := range res {
		if v != 0 {
			t.Fatalf("res[%d]=%d want=0", i, v)
		}
	}
}

// 混合零权重: 零权重不分值, 非零权重按比例分配
func Test_AvgByWeight_MixedZeroWeights(t *testing.T) {
	t.Parallel()
	res := avgByWeight(10, 0, 1, 1)
	if res[0] != 0 || res[1]+res[2] != 10 {
		t.Fatalf("res=%v want [0 5 5]", res)
	}
	if res[1] != 5 || res[2] != 5 {
		t.Fatalf("res=%v want [0 5 5]", res)
	}
}
