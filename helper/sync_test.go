package helper

import (
	"sync"
	"testing"
	"time"
)

// Test_NewWaitGroup 初始计数器生效, 计数归零后 Wait 返回
func Test_NewWaitGroup(t *testing.T) {
	t.Parallel()
	wg := NewWaitGroup(2)
	if wg == nil {
		t.Fatal("NewWaitGroup 不应返回 nil")
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	// 计数未归零时 Wait 必然阻塞, 不依赖真实时间推进
	select {
	case <-done:
		t.Fatal("计数未归零时 Wait 不应返回")
	case <-time.After(100 * time.Millisecond):
	}
	wg.Done()
	wg.Done()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("计数归零后 Wait 未返回")
	}
}

// Test_NewWaitGroup_ZeroInit 初始计数为零时 Wait 立即返回
func Test_NewWaitGroup_ZeroInit(t *testing.T) {
	t.Parallel()
	wg := NewWaitGroup(0)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("初始计数为零时 Wait 应立即返回")
	}
}

// Test_NewWaitGroup_ConcurrentDone 并发 Done 全部完成后 Wait 返回
func Test_NewWaitGroup_ConcurrentDone(t *testing.T) {
	t.Parallel()
	const goroutines = 8
	wg := NewWaitGroup(goroutines)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	var fire sync.WaitGroup
	fire.Add(1)
	for i := 0; i < goroutines; i++ {
		go func() {
			fire.Wait()
			wg.Done()
		}()
	}
	// 同时放行全部 Done, 保证并发释放
	fire.Done()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("并发 Done 后 Wait 未返回")
	}
}
