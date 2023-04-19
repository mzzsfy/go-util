package seq

import "testing"

var global int

func doNothing(a int) int {
    return a
}

func Benchmark_For(b *testing.B) {
    var v int
    for i := 0; i < b.N; i++ {
        //去除函数调用开销,不去除 Benchmark_R_For
        v = doNothing(i)
    }
    global = v
}

func Benchmark_SeqEach(b *testing.B) {
    From(func(t func(int)) {
        for i := 0; i < b.N; i++ {
            t(i)
        }
    }).Complete()
}
func Benchmark_R_SeqRang(b *testing.B) {
    FromIntSeq(b.N-1, 0).Complete()
}

func Benchmark_SeqRang(b *testing.B) {
    FromIntSeq(0, b.N-1).Complete()
}

func Benchmark_SeqTack(b *testing.B) {
    FromIntSeq().Take(b.N).Complete()
}

func Benchmark_R_For(b *testing.B) {
    for i := b.N; i > 0; i-- {
    }
}

func Benchmark_R_SeqEach(b *testing.B) {
    From(func(t func(int)) {
        for i := b.N; i > 0; i-- {
            t(i)
        }
    }).Complete()
}

//func Benchmark_Cain_Buff_1(b *testing.B) {
//    c := make(chan int, 1)
//    go func() {
//        for range c {
//        }
//    }()
//    for i := 0; i < b.N; i++ {
//        c <- i
//    }
//    close(c)
//}
func Benchmark_Cain_Buff_128(b *testing.B) {
    c := make(chan int, 128)
    go func() {
        for range c {
        }
    }()
    for i := 0; i < b.N; i++ {
        c <- i
    }
    close(c)
}

//func Benchmark_Cain(b *testing.B) {
//    c := make(chan int)
//    go func() {
//        for range c {
//        }
//    }()
//    for i := 0; i < b.N; i++ {
//        c <- i
//    }
//    close(c)
//}

//func Benchmark_Cain_Buff_N(b *testing.B) {
//    c := make(chan int, b.N)
//    go func() {
//        for range c {
//        }
//    }()
//    for i := 0; i < b.N; i++ {
//        c <- i
//    }
//    close(c)
//}

//func Benchmark_R_Cain(b *testing.B) {
//    c := make(chan int)
//    go func() {
//        for range c {
//        }
//    }()
//    for i := b.N; i > 0; i-- {
//        c <- i
//    }
//    close(c)
//}

//func Benchmark_R_Cain_Buff_1(b *testing.B) {
//    c := make(chan int, 1)
//    go func() {
//        for range c {
//        }
//    }()
//    for i := b.N; i > 0; i-- {
//        c <- i
//    }
//    close(c)
//}

func Benchmark_R_Cain_Buff_128(b *testing.B) {
    c := make(chan int, 128)
    go func() {
        for range c {
        }
    }()
    for i := b.N; i > 0; i-- {
        c <- i
    }
    close(c)
}

//func Benchmark_R_Cain_Buff_N(b *testing.B) {
//    c := make(chan int, b.N)
//    go func() {
//        for range c {
//        }
//    }()
//    for i := b.N; i > 0; i-- {
//        c <- i
//    }
//    close(c)
//}

//内存消耗与原生一致,性能与原生调用性能差距已经不大了,显著优于市面上其他类似库
//非常感谢 https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw 如此优秀的分享

//go test -v -run=Benchmark.+ -count 3 -benchmem -gcflags="-l" -bench=. -benchmem ./seq  
//goos: windows
//goarch: amd64
//cpu: Intel(R) Core(TM) i5-8500 CPU
