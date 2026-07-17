package pool

import (
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func Test_bufferPool(t *testing.T) {
    bp := NewBufferPool()
    bp.SetMaxCap(1024)

    b := bp.Get()
    b.WriteString("Hello, World!")
    bp.Put(b)

    b2 := bp.Get()
    if b2.String() != "" {
        t.Errorf("Expected empty buffer, got %s", b2.String())
    }
}

func Test_bytePool(t *testing.T) {
    bp := NewSimpleBytesPool()
    bp.SetMaxCap(1024)
    bp.SetInitCap(512)

    b := bp.Get()
    b.Write([]byte("Hello, World!"))
    bp.Put(b)

    b2 := bp.Get()
    if len(b2.buf) != 0 {
        t.Errorf("Expected empty buffer, got %s", string(b2.buf))
    }
}

var (
    shortStr []string
    midStr   []string
    longStr  []string
)

func init() {
    rand.Seed(time.Now().UnixNano())
    for i := 0; i < 1000; i++ {
        shortStr = append(shortStr, strings.Repeat(strconv.Itoa(rand.Int()), 1))
        midStr = append(midStr, strings.Repeat(strconv.Itoa(rand.Int()), 10))
        longStr = append(longStr, strings.Repeat(strconv.Itoa(rand.Int()), 100))
    }
}

func BenchmarkBufferPool(b *testing.B) {
    bp := NewBufferPool()
    bp.SetMaxCap(1024)
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            for z := range shortStr {
                buf := bp.Get()
                buf.WriteString(shortStr[z])
                bp.Put(buf)
                buf = bp.Get()
                buf.WriteString(shortStr[z])
                bp.Put(buf)
                buf = bp.Get()
                buf.WriteString(shortStr[z])
                bp.Put(buf)
            }
        }
    })
}

func BenchmarkBytePool(b *testing.B) {
    bp := NewSimpleBytesPool()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            for z := range shortStr {
                buf := bp.Get()
                buf.WriteString(shortStr[z])
                bp.Put(buf)
                buf = bp.Get()
                buf.WriteString(shortStr[z])
                bp.Put(buf)
                buf = bp.Get()
                buf.WriteString(shortStr[z])
                bp.Put(buf)
            }
        }
    })
}

// TestBytePoolConcurrentConfig 并发读写 BytePool 配置字段, 验证无数据竞争
func TestBytePoolConcurrentConfig(t *testing.T) {
	t.Parallel()
	bp := NewSimpleBytesPool()
	const workers = 8

	// 并发修改配置
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				bp.SetMaxCap(16 + i*10)
				bp.SetInitCap(16 + i)
			}
		}(i)
	}
	// 并发使用池
	var wg2 sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			for j := 0; j < 100; j++ {
				b := bp.Get()
				b.WriteString("test")
				bp.Put(b)
			}
		}()
	}
	wg.Wait()
	wg2.Wait()
}

// TestBufferPoolConcurrentConfig 并发读写 BufferPool 配置字段, 验证无数据竞争
func TestBufferPoolConcurrentConfig(t *testing.T) {
	t.Parallel()
	bp := NewBufferPool()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				bp.SetMaxCap(16 + j)
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				b := bp.Get()
				b.WriteString("test")
				bp.Put(b)
			}
		}()
	}
	wg.Wait()
}

// TestBytes_Write 验证 Write 方法正确追加字节数据
func Test_Bytes_Write(t *testing.T) {
	t.Parallel()
	var b Bytes

	n, err := b.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write 不应返回错误: %v", err)
	}
	if n != 5 {
		t.Fatalf("Write 应返回写入字节数 5, got %d", n)
	}
	if b.String() != "hello" {
		t.Fatalf("写入后内容应为 'hello', got %q", b.String())
	}

	// 追加写入
	n, err = b.Write([]byte(" world"))
	if err != nil {
		t.Fatalf("Write 不应返回错误: %v", err)
	}
	if n != 6 {
		t.Fatalf("Write 应返回写入字节数 6, got %d", n)
	}
	if b.String() != "hello world" {
		t.Fatalf("追加后内容应为 'hello world', got %q", b.String())
	}
}

// TestBytes_WriteString 验证 WriteString 方法正确追加字符串
func Test_Bytes_WriteString(t *testing.T) {
	t.Parallel()
	var b Bytes

	n, err := b.WriteString("abc")
	if err != nil {
		t.Fatalf("WriteString 不应返回错误: %v", err)
	}
	if n != 3 {
		t.Fatalf("WriteString 应返回写入字节数 3, got %d", n)
	}
	if b.String() != "abc" {
		t.Fatalf("写入后内容应为 'abc', got %q", b.String())
	}

	// 追加
	b.WriteString("def")
	if b.String() != "abcdef" {
		t.Fatalf("追加后内容应为 'abcdef', got %q", b.String())
	}
}

// TestBytes_WriteByte 验证 WriteByte 方法逐字节写入
func Test_Bytes_WriteByte(t *testing.T) {
	t.Parallel()
	var b Bytes

	for _, c := range []byte{'X', 'Y', 'Z'} {
		if err := b.WriteByte(c); err != nil {
			t.Fatalf("WriteByte(%c) 不应返回错误: %v", c, err)
		}
	}
	if b.String() != "XYZ" {
		t.Fatalf("写入后内容应为 'XYZ', got %q", b.String())
	}
}

// TestBytes_LenCap 验证 Len 和 Cap 方法返回正确值
func Test_Bytes_LenCap(t *testing.T) {
	t.Parallel()
	var b Bytes

	// 空 buffer
	if b.Len() != 0 {
		t.Fatalf("空 buffer Len 应为 0, got %d", b.Len())
	}
	if b.Cap() != 0 {
		t.Fatalf("空 buffer Cap 应为 0, got %d", b.Cap())
	}

	// 写入后检查
	b.Write([]byte("12345"))
	if b.Len() != 5 {
		t.Fatalf("写入 5 字节后 Len 应为 5, got %d", b.Len())
	}
	if b.Cap() < 5 {
		t.Fatalf("Cap 应 >= 5, got %d", b.Cap())
	}
}

// TestBytes_Reset 验证 Reset 清空内容但保留底层数组容量
func Test_Bytes_Reset(t *testing.T) {
	t.Parallel()
	var b Bytes

	b.WriteString("some data")
	prevCap := b.Cap()

	b.Reset()
	if b.Len() != 0 {
		t.Fatalf("Reset 后 Len 应为 0, got %d", b.Len())
	}
	if b.String() != "" {
		t.Fatalf("Reset 后 String 应为空, got %q", b.String())
	}
	// 底层数组应被复用, 容量不变
	if b.Cap() != prevCap {
		t.Fatalf("Reset 后 Cap 应保持 %d 不变, got %d", prevCap, b.Cap())
	}

	// Reset 后仍可正常写入
	b.WriteString("new")
	if b.String() != "new" {
		t.Fatalf("Reset 后再写入应为 'new', got %q", b.String())
	}
}

// TestBytes_Bytes 验证 Bytes 方法返回原始切片引用
func Test_Bytes_Bytes(t *testing.T) {
	t.Parallel()
	var b Bytes

	b.Write([]byte{1, 2, 3})
	raw := b.Bytes()
	if len(raw) != 3 {
		t.Fatalf("Bytes 应返回长度 3, got %d", len(raw))
	}
	// 验证返回的是底层数组的引用
	raw[0] = 99
	if b.Bytes()[0] != 99 {
		t.Fatal("Bytes 应返回底层数组的引用, 修改应反映到 Bytes 对象")
	}
}

// TestBytes_String 验证 String 方法返回内容的字符串副本
func Test_Bytes_String(t *testing.T) {
	t.Parallel()
	var b Bytes

	b.WriteString("test")
	s := b.String()
	if s != "test" {
		t.Fatalf("String 应返回 'test', got %q", s)
	}
	// 修改返回值不影响原 buffer
	s = "modified"
	if b.String() != "test" {
		t.Fatalf("修改 String 返回值不应影响原 buffer, got %q", b.String())
	}
}

// TestBytes_MixedOperations 混合使用 Write/WriteString/WriteByte 后验证结果
func Test_Bytes_MixedOperations(t *testing.T) {
	t.Parallel()
	var b Bytes

	b.Write([]byte("he"))
	b.WriteString("llo")
	b.WriteByte(' ')
	b.Write([]byte("wo"))
	b.WriteString("rld")
	b.WriteByte('!')

	if b.String() != "hello world!" {
		t.Fatalf("混合写入后应为 'hello world!', got %q", b.String())
	}
	if b.Len() != 12 {
		t.Fatalf("总长度应为 12, got %d", b.Len())
	}
}

// TestBufferPool_MaxCap_Boundary 验证 BufferPool 的 maxCap 边界行为
func Test_BufferPool_MaxCap_Boundary(t *testing.T) {
	t.Parallel()

	t.Run("超过maxCap丢弃", func(t *testing.T) {
		t.Parallel()
		bp := NewBufferPool()
		bp.SetMaxCap(64)
		// 获取 buffer 并写入大量数据使其容量超过 maxCap
		b := bp.Get()
		b.Write(make([]byte, 128))
		prevCap := b.Cap()
		bp.Put(b)
		// 再次获取, 由于之前 Put 时容量超过 maxCap, 应得到新 buffer
		b2 := bp.Get()
		// 新 buffer 容量应远小于 prevCap(新创建的 buffer, 不是那个被丢弃的)
		if b2.Cap() >= prevCap {
			t.Fatalf("被丢弃的大 buffer 不应被复用, prevCap=%d, got Cap=%d", prevCap, b2.Cap())
		}
	})

	t.Run("等于maxCap保留", func(t *testing.T) {
		t.Parallel()
		bp := NewBufferPool()
		bp.SetMaxCap(1024)
		b := bp.Get()
		b.Write(make([]byte, 512))
		// 容量应 <= maxCap, buffer 应被放回池中
		bp.Put(b)
	})

	t.Run("SetMaxCap下限修正为16", func(t *testing.T) {
		t.Parallel()
		bp := NewBufferPool()
		// 传入 <= 16 的值应被修正为 16
		bp.SetMaxCap(1)
		b := bp.Get()
		b.Write(make([]byte, 32))
		// 容量 > 16, 应被丢弃
		bp.Put(b)
	})
}

// TestBytePool_InitCap_Boundary 验证 BytePool 的 initCap 边界行为
func Test_BytePool_InitCap_Boundary(t *testing.T) {
	t.Parallel()

	t.Run("Get返回buffer容量满足initCap", func(t *testing.T) {
		t.Parallel()
		bp := NewSimpleBytesPool()
		bp.SetInitCap(128)
		b := bp.Get()
		if b.Cap() < 128 {
			t.Fatalf("Get 返回 buffer 容量应 >= initCap(128), got %d", b.Cap())
		}
	})

	t.Run("Put时容量超过maxCap丢弃", func(t *testing.T) {
		t.Parallel()
		bp := NewSimpleBytesPool()
		bp.SetMaxCap(64)
		bp.SetInitCap(16)
		b := bp.Get()
		// 扩容使其超过 maxCap
		b.Write(make([]byte, 128))
		bp.Put(b)
	})

	t.Run("SetInitCap下限修正为16", func(t *testing.T) {
		t.Parallel()
		bp := NewSimpleBytesPool()
		bp.SetInitCap(1)
		b := bp.Get()
		// initCap 被修正为 16, 所以容量至少为 16
		if b.Cap() < 16 {
			t.Fatalf("initCap 被修正为 16, 容量应 >= 16, got %d", b.Cap())
		}
	})

	t.Run("Put时容量等于maxCap保留", func(t *testing.T) {
		t.Parallel()
		bp := NewSimpleBytesPool()
		bp.SetMaxCap(1024)
		bp.SetInitCap(16)
		b := bp.Get()
		b.Write(make([]byte, 512))
		// 容量应 <= maxCap, buffer 应被放回池中
		bp.Put(b)
	})
}
