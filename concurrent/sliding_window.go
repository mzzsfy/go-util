package concurrent

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrWindowNumber 窗口数量非法
	ErrWindowNumber = errors.New("windowNumber must > 2")
	// ErrWindowTime 时间窗口长度非法
	ErrWindowTime = errors.New("time must > windowNumber")
	// ErrAllowNumber 允许请求数非法
	ErrAllowNumber = errors.New("allowNumber must >= windowNumber")
)

// SlidingWindow 简单的时间滑动窗口实现
type SlidingWindow struct {
	// 所有窗口，记录当前计数
	counters []int32
	// 每个窗口的限制
	limits []int32
	// 时间窗口，毫秒
	time int64
	// 每秒允许尝试数量
	allowNumber int32
	// 窗口数量
	windowNumber int32
	// 每个窗口负责的时间
	everyWindowTime int64
	// 分片满的时间
	sliceFullTime []int64
	// 当前指针
	sliceIndex int32

	// 保护 sliceIndex 和 sliceFullTime 的修改
	mutex sync.Mutex

	Now func() int64
}

// NewSlidingWindow 创建滑动窗口
// windowTime: 时间窗口长度(毫秒)
// allowNumber: 允许的最大请求数
// windowNumber: 窗口数量
// 参数非法时返回 ErrWindowNumber/ErrWindowTime/ErrAllowNumber
// Now 注入假时钟时, 时钟值须不早于构造时刻, 否则窗口永不滑动导致首片填满后永久拒绝
func NewSlidingWindow(windowTime int64, allowNumber, windowNumber int32) (*SlidingWindow, error) {
	if windowNumber <= 2 {
		return nil, ErrWindowNumber
	}

	if windowTime <= int64(windowNumber) {
		return nil, ErrWindowTime
	}

	if allowNumber < windowNumber {
		return nil, ErrAllowNumber
	}

	// 分片时间戳初始化为当前时间, 零值会与真实UnixMilli差值过大, 启动期触发立即滑动导致超发
	nowMilli := time.Now().UnixMilli()
	fullTime := make([]int64, windowNumber)
	for i := range fullTime {
		fullTime[i] = nowMilli
	}

	sw := &SlidingWindow{
		time:            windowTime,
		allowNumber:     allowNumber,
		windowNumber:    windowNumber,
		everyWindowTime: windowTime / int64(windowNumber),
		counters:        make([]int32, windowNumber),
		limits:          make([]int32, windowNumber),
		sliceFullTime:   fullTime,
	}

	weights := make([]int32, windowNumber)
	for i := range weights {
		weights[i] = 1
	}

	values := avgByWeight(allowNumber, weights...)

	for i := int32(0); i < windowNumber; i++ {
		sw.limits[i] = values[i]
	}

	return sw, nil
}

// CanDo 获取能否执行
func (sw *SlidingWindow) CanDo() bool {
	// 获取当前时间
	var now int64
	if sw.Now == nil {
		now = time.Now().UnixMilli()
	} else {
		now = sw.Now()
	}

	for {
		// 获取当前分片
		index := atomic.LoadInt32(&sw.sliceIndex)

		// 预检查计数是否可能超过限制，避免不必要的原子操作
		currentCount := atomic.LoadInt32(&sw.counters[index])
		limit := atomic.LoadInt32(&sw.limits[index])

		// 如果当前计数已经等于或超过限制，则直接处理滑动逻辑
		if currentCount >= limit {
			// 处理窗口滑动逻辑
			if sw.handleWindowSliding(index, now) {
				continue // 窗口已滑动，重试
			} else {
				return false // 无法滑动窗口，拒绝请求
			}
		}

		// 原子增加计数器
		count := atomic.AddInt32(&sw.counters[index], 1)

		if count <= limit {
			// 有空间，返回true
			return true
		} else {
			// 没空间，先回退计数器
			atomic.AddInt32(&sw.counters[index], -1)

			// 处理窗口滑动逻辑
			if sw.handleWindowSliding(index, now) {
				continue // 窗口已滑动，重试
			} else {
				return false // 无法滑动窗口，拒绝请求
			}
		}
	}
}

// handleWindowSliding 处理窗口滑动逻辑
func (sw *SlidingWindow) handleWindowSliding(currentIndex int32, now int64) bool {
	prevIndex := (currentIndex + sw.windowNumber - 1) % sw.windowNumber

	// 获取上个分片满的时间
	last := atomic.LoadInt64(&sw.sliceFullTime[prevIndex])
	difference := now - last

	// 与当前时间对比，过了一个窗口时间，清空已经过期的分片，指针指向下一个
	if difference >= sw.everyWindowTime {
		sw.mutex.Lock()
		// 简单做个防修改
		if last == atomic.LoadInt64(&sw.sliceFullTime[prevIndex]) {
			// 记录当前分片已满
			atomic.StoreInt64(&sw.sliceFullTime[prevIndex], now)
			num := difference / sw.everyWindowTime
			if num > int64(sw.windowNumber) {
				num = int64(sw.windowNumber)
			}

			// 清空已经过期的分片
			for l := int64(1); l <= num; l++ {
				idx := (prevIndex + int32(l)) % sw.windowNumber
				atomic.StoreInt32(&sw.counters[idx], 0)
			}

			idx := atomic.LoadInt32(&sw.sliceIndex)
			nextIndex := (idx + int32(num)) % sw.windowNumber
			atomic.StoreInt32(&sw.sliceIndex, nextIndex)
		}
		sw.mutex.Unlock()
		return true // 窗口已滑动
	}

	return false // 窗口未滑动
}

// avgByWeight 按权重分配总量
func avgByWeight(total int32, weights ...int32) []int32 {
	if len(weights) == 0 {
		panic("拆分个数不能为0")
	}

	totalWeights := int32(0)
	for _, weight := range weights {
		if weight < 0 {
			panic("权重只能为正数")
		}
		totalWeights += weight
	}

	//总权重为0语义: 无法按权重分配, 返回全0
	if totalWeights <= 0 {
		return make([]int32, len(weights))
	}

	usedValue := int32(0)
	res := make([]int32, len(weights))

	for i := 0; i < len(res); i++ {
		allocated := total * weights[i] / totalWeights
		res[i] = allocated
		usedValue += allocated
	}

	// 有余数
	if total != usedValue {
		direction := int32(-1)
		singleValue := int32(1)
		if total < 0 {
			direction = 1
			singleValue = -1
		}

		remainder := total - usedValue
		for i := 0; remainder != 0; i++ {
			// 权重为0则不分值
			if weights[i] > 0 {
				remainder += direction
				res[i] += singleValue
			}
		}
	}

	return res
}
