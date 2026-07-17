package concurrent

import (
	"testing"
)

func Test_SegBasic(t *testing.T) {
	q := newSegQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)
	if q.Size() != 3 {
		t.Fatalf("size=%d want=3", q.Size())
	}
	v, ok := q.Dequeue()
	if !ok || v != 1 {
		t.Fatalf("v=%d ok=%v", v, ok)
	}
	v, ok = q.Dequeue()
	if !ok || v != 2 {
		t.Fatalf("v=%d ok=%v", v, ok)
	}
	v, ok = q.Dequeue()
	if !ok || v != 3 {
		t.Fatalf("v=%d ok=%v", v, ok)
	}
	_, ok = q.Dequeue()
	if ok {
		t.Fatal("should be empty")
	}
}

func Test_SegMultiSegment(t *testing.T) {
	q := newSegQueue[int]()
	n := segSize * 3
	for i := 0; i < n; i++ {
		q.Enqueue(i)
	}
	if q.Size() != n {
		t.Fatalf("size=%d want=%d", q.Size(), n)
	}
	for i := 0; i < n; i++ {
		v, ok := q.Dequeue()
		if !ok || v != i {
			t.Fatalf("i=%d v=%d ok=%v", i, v, ok)
		}
	}
	_, ok := q.Dequeue()
	if ok {
		t.Fatal("should be empty")
	}
}

func TestSegPingPong(t *testing.T) {
	q1 := newSegQueue[int]()
	q2 := newSegQueue[int]()
	done := make(chan struct{})
	n := 1000
	go func() {
		defer close(done)
		for i := 0; i < n; i++ {
			for {
				_, ok := q1.Dequeue()
				if ok {
					break
				}
			}
			q2.Enqueue(1)
		}
	}()
	for i := 0; i < n; i++ {
		q1.Enqueue(1)
		for {
			_, ok := q2.Dequeue()
			if ok {
				break
			}
		}
	}
	<-done
}
