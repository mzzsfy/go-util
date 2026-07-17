package concurrent

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueueSizeIntegrity(t *testing.T) {
	const rounds = 100
	const producers = 4
	const n = 500
	for round := 0; round < rounds; round++ {
		q := newSegQueue[int]()
		var produced, consumed int64
		var pwg sync.WaitGroup
		pwg.Add(producers)
		for i := 0; i < producers; i++ {
			go func() {
				defer pwg.Done()
				for j := 0; j < n; j++ {
					q.Enqueue(j)
					atomic.AddInt64(&produced, 1)
				}
			}()
		}
		pwg.Wait()
		for {
			_, ok := q.Dequeue()
			if !ok {
				break
			}
			consumed++
		}
		if produced != consumed {
			t.Fatalf("round %d: produced=%d consumed=%d size=%d", round, produced, consumed, q.Size())
		}
		if q.Size() != 0 {
			t.Fatalf("round %d: size=%d after drain", round, q.Size())
		}
	}
}

func TestQueueConcurrentProdCons(t *testing.T) {
	const producers = 4
	const consumers = 4
	const n = 2000
	for _, opt := range []struct {
		name string
		o    Opt[int]
	}{
		{"chan", WithTypeChan[int](128)},
		{"ring", WithTypeRing[int](1024)},
		{"seg", WithTypeSegment[int]()},
	} {
		t.Run(opt.name, func(t *testing.T) {
			q := NewQueue(opt.o)
			var consumed int64
			var doneFlag int32
			target := int64(producers * n)
			var pwg, cwg sync.WaitGroup
			pwg.Add(producers)
			cwg.Add(consumers)
			for i := 0; i < producers; i++ {
				go func() {
					defer pwg.Done()
					for j := 0; j < n; j++ {
						q.Enqueue(j)
					}
				}()
			}
			for i := 0; i < consumers; i++ {
				go func() {
					defer cwg.Done()
					for atomic.LoadInt32(&doneFlag) == 0 {
						_, ok := q.Dequeue()
						if ok {
							if atomic.AddInt64(&consumed, 1) == target {
								atomic.StoreInt32(&doneFlag, 1)
								return
							}
							continue
						}
						runtime.Gosched()
					}
				}()
			}
			done := make(chan struct{})
			go func() {
				cwg.Wait()
				close(done)
			}()
			select {
			case <-time.After(10 * time.Second):
				c := atomic.LoadInt64(&consumed)
				t.Fatalf("%s deadlock: target=%d consumed=%d size=%d", opt.name, target, c, q.Size())
			case <-done:
				c := atomic.LoadInt64(&consumed)
				if c != target {
					t.Fatalf("%s mismatch: target=%d consumed=%d size=%d", opt.name, target, c, q.Size())
				}
			}
		})
	}
}
