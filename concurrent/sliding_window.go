package concurrent

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrWindowTime 时间窗口长度非法
	ErrWindowTime = errors.New("windowTime must > 0")
	// ErrAllowNumber 允许请求数非法
	ErrAllowNumber = errors.New("allowNumber must > 0")
)

// SlidingWindow 时间滑动窗口限流器
// 记录窗口内已放行请求的时间戳, 窗口外的记录在判定时剔除, 队列满则拒绝
type SlidingWindow struct {
	// 时间窗口长度, 毫秒
	windowTime int64
	// 窗口内允许的最大请求数
	allowNumber int32
	// 已放行请求的时间戳环形队列
	times []int64
	// 队首下标
	head int
	// 队列长度
	size int

	// 保护 times/head/size
	mutex sync.Mutex

	// 时钟函数, 返回毫秒时间戳; nil 时使用 time.Now().UnixMilli()
	// 假时钟须单调递增, 时钟回拨期间不剔除记录、不恢复配额
	Now func() int64
}

// NewSlidingWindow 创建滑动窗口
// windowTime: 时间窗口长度(毫秒), 须大于0, 否则返回 ErrWindowTime
// allowNumber: 窗口内允许的最大请求数, 须大于0, 否则返回 ErrAllowNumber
func NewSlidingWindow(windowTime int64, allowNumber int32) (*SlidingWindow, error) {
	if windowTime <= 0 {
		return nil, ErrWindowTime
	}
	if allowNumber <= 0 {
		return nil, ErrAllowNumber
	}
	return &SlidingWindow{
		windowTime:  windowTime,
		allowNumber: allowNumber,
		times:       make([]int64, allowNumber),
	}, nil
}

// CanDo 获取能否执行
// 允许则记录当前时间并返回true, 窗口内已达上限则返回false
// 并发安全, 可在多goroutine同时调用
func (sw *SlidingWindow) CanDo() bool {
	var now int64
	if sw.Now == nil {
		now = time.Now().UnixMilli()
	} else {
		now = sw.Now()
	}

	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	// 剔除窗口外的记录
	for sw.size > 0 {
		oldest := sw.times[sw.head]
		if now-oldest < sw.windowTime {
			break
		}
		sw.head = (sw.head + 1) % len(sw.times)
		sw.size--
	}

	if sw.size >= len(sw.times) {
		return false
	}

	sw.times[(sw.head+sw.size)%len(sw.times)] = now
	sw.size++
	return true
}
