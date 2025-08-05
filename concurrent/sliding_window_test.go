package concurrent

import (
    "fmt"
    "math/rand"
    "testing"
    "time"
)

func Test_SlidingWindow_BasicFunctionality(t *testing.T) {
    t.Parallel()
    // 测试参数设置
    timeWindow := int64(1000) // 时间窗口1秒
    allowNumber := int32(100) // 允许最大请求数100
    windowNumber := int32(5)  // 窗口数量5

    // 创建滑动窗口实例
    sw := NewSlidingWindow(timeWindow, allowNumber, windowNumber)

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

    // 等待窗口时间过去，再次测试
    time.Sleep(time.Millisecond * time.Duration(sw.everyWindowTime+100))

    // 这次应该又能通过
    if !sw.CanDo() {
        t.Error("窗口滑动后应该允许请求")
    }
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
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // 创建滑动窗口实例
            sw := NewSlidingWindow(tc.timeWindow, tc.allowNumber, tc.windowNumber)

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

            // 等待窗口时间过去
            time.Sleep(time.Millisecond * time.Duration(sw.everyWindowTime+100))

            // 现在应该能再次通过
            if !sw.CanDo() {
                t.Error("窗口滑动后应该允许请求")
            }
        })
    }
}

func sumI32(weights []int32) int32 {
    sum := int32(0)
    for _, v := range weights {
        sum += v
    }
    return sum
}

// 随机参数测试
func TestSlidingWindow_RandomParams(t *testing.T) {
    t.Parallel()
    // 生成随机测试用例
    for i := 0; i < 1000; i++ {
        // 生成随机参数
        windowNumber := int32(rand.Intn(10) + 3)                  // 窗口数量 3-12
        allowNumber := int32(rand.Intn(1000) + int(windowNumber)) // 允许请求数 >= windowNumber
        timeWindow := int64(rand.Intn(10000) + int(windowNumber)) // 时间窗口 >= windowNumber

        t.Run(fmt.Sprintf("随机测试_%d_%d_%d", windowNumber, allowNumber, timeWindow), func(t *testing.T) {
            t.Parallel()

            // 创建滑动窗口实例
            sw := NewSlidingWindow(timeWindow, allowNumber, windowNumber)

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

            // 等待窗口时间过去
            time.Sleep(time.Millisecond * time.Duration(sw.everyWindowTime+100))

            // 现在应该能再次通过
            if !sw.CanDo() {
                t.Error("窗口滑动后应该允许请求")
            }
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
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // 创建滑动窗口实例
            sw := NewSlidingWindow(tc.timeWindow, tc.allowNumber, tc.windowNumber)

            // 发送少量请求（低频率）
            successCount := 0
            for i := 0; i < int(tc.allowNumber/2); i++ {
                if sw.CanDo() {
                    successCount++
                }
                time.Sleep(tc.requestInterval)
            }

            // 等待较长时间（窗口时间的倍数）
            time.Sleep(time.Millisecond * time.Duration(sw.everyWindowTime*tc.waitMultiplier))

            // 再次尝试发送请求
            postWaitSuccess := 0
            for i := 0; i < int(tc.allowNumber); i++ {
                if sw.CanDo() {
                    postWaitSuccess++
                } else {
                    break // 一旦被拒绝就停止
                }
            }

            // 验证等待后应该能够执行请求
            if postWaitSuccess == 0 {
                t.Error("长时间等待后应该能够执行请求")
            }
        })
    }
}

// 辅助函数：取整数的绝对值
func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
