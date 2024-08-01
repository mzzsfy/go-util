package logger

import (
    "context"
    "fmt"
    "github.com/mzzsfy/go-util/concurrent"
    "github.com/mzzsfy/go-util/helper"
    "math/rand"
    "strings"
    "sync"
    "testing"
    "time"
)

func Test_Logger_1(t *testing.T) {
    info := PrintYearInfo
    defer func() {
        PrintYearInfo = info
    }()
    Logger("test.test").I("test")
    PrintYearInfo = 1
    Logger("test.test").I("test")
    PrintYearInfo = 2
    Logger("test.test").I("test")
    PrintYearInfo = 3
    Logger("test.test").I("test")
    PrintYearInfo = 1
    log1 := Logger("test.test")
    log1.I("test")
    length := showNameMaxLength
    defer SetLogNameMaxLength(length)
    SetLogNameMaxLength(6)
    log1.I("short test")
}

func Test_Logger_11(t *testing.T) {
    log := Logger("test.test", WithSetContext(context.Background()))
    log.UnUse()
    log.I("test")
}

func Test_Logger_2(t *testing.T) {
    Logger("test.test.test.test.test.test.test.test.test.test.test").I("test")
    Logger("test").I("test")
    Logger("test.test").I("test")
    Logger("test.test.test").I("test")
    Logger("test.test.test.test").I("test")
    Logger("test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test.test.test.test.test").I("test")
    Logger("你好.你好").I("test")
    Logger("你好.你好.你好.你好").I("test")
    Logger("你好.你好.你好.你好.你好.你好.你好").I("test")
    Logger("你好.你好.你好.你好.你好.你好.你好.你好.你好").I("test")
}
func Test_Logger_3(t *testing.T) {
    Logger("test").I("test")
    Logger("test", WithTag('@')).I("test")
    Logger("test").I("test")
    PrintBlankTag = true
    Logger("test", WithTag('#')).I("test")
    Logger("test").I("test")
    Logger("test", WithTag('*')).I("test")
    Logger("test").I("test")
    PrintBlankTag = false
    Logger("test").I("test")
}

type Opt func(f *pluginF)

func PluginWithBeforeWrite(fn func(Level, Buffer, Log, Plugin)) Opt {
    return func(f *pluginF) {
        if fn == nil {
            panic("fn 不能为nil")
        }
        f.beforeWriteF = fn
    }
}

func NewPlugin(name string, opts ...Opt) Plugin {
    p := &pluginF{
        name:         name,
        beforeWriteF: func(Level, Buffer, Log, Plugin) {},
    }
    for _, o := range opts {
        o(p)
    }
    return p
}

type pluginF struct {
    name         string
    beforeWriteF func(Level, Buffer, Log, Plugin)
}

func (t *pluginF) BeforeWrite(l Level, s Buffer, log Log, plugin Plugin) {
    t.beforeWriteF(l, s, log, plugin)
}

var _ Plugin = (*pluginF)(nil)

func Test_Logger_Plugin(t *testing.T) {
    log := Logger("test.test.test.test")
    plog := log.WithPlugin(NewPlugin("", PluginWithBeforeWrite(func(_ Level, s Buffer, _ Log, _ Plugin) {
        fmt.Println(s.String() + "-plog")
    })))
    pplog := plog.WithPlugin(NewPlugin("", PluginWithBeforeWrite(func(_ Level, s Buffer, _ Log, _ Plugin) {
        fmt.Println(s.String() + "-pplog")
    })))
    log1 := log.WithPlugin(NewPlugin("", PluginWithBeforeWrite(func(_ Level, s Buffer, _ Log, _ Plugin) {
        fmt.Println(s.String() + "-log1")
        s.Reset()
    })))
    log2 := log.WithPlugin(NewPlugin("", PluginWithBeforeWrite(func(_ Level, s Buffer, _ Log, _ Plugin) {
        s.WriteString("-log2")
    })))
    log.I("test1")
    plog.I("test1")
    pplog.I("test1")
    log1.I("test1")
    log2.I("test1")
    log.I("test2")
    plog.I("test2")
    pplog.I("test2")
    log1.I("test2")
    log2.I("test2")

    level := WarnLevel
    log.SetLevel(&level)

    Logger("test.test.test").I("test3")
    log.I("test3")
    plog.I("test3")
    pplog.I("test3")
    log1.I("test3")
    log2.I("test3")
    Logger("test.test.test").I("test3")
    log.SetLevel(nil)
}

type l interface {
    Log(args ...any)
}
type w struct {
    l
}

func (w w) Write(p []byte) (n int, err error) {
    w.l.Log(string(p))
    return len(p), nil
}

func TestLogger1(t *testing.T) {
    //f, _ := os.Create("test.pprof")
    //pprof.StartCPUProfile(f)
    //defer pprof.StopCPUProfile()
    info := PrintYearInfo
    defer func() {
        PrintYearInfo = info
    }()
    PrintYearInfo = 0
    compressedLogName := CompressedLogName
    CompressedLogName = true
    target := DefaultWriterTarget()
    SetDefaultWriterTarget(helper.AsyncConsole())
    //SetDefaultWriterTarget(w{t})
    defer func() {
        CompressedLogName = compressedLogName
        SetDefaultWriterTarget(target)
    }()
    var names []string
    l := 100000
    for i := 0; i < l; i++ {
        s := strings.Builder{}
        nameLevels := rand.Intn(8) + 1
        for l := 0; l < nameLevels; l++ {
            nameLen := rand.Intn(5) + 2
            for i := 0; i < nameLen; i++ {
                s.WriteRune('a' + rune(rand.Intn(26)))
            }
            if l != nameLevels-1 {
                s.WriteRune('.')
            }
        }
        names = append(names, s.String())
    }
    wg := sync.WaitGroup{}
    adder := concurrent.Int64Adder{}
    plugin1 := NewPlugin("1", func(f *pluginF) {
        f.beforeWriteF = func(l Level, s Buffer, log Log, plugin Plugin) {
            adder.IncrementSimple()
        }
    })
    plugins := GlobalPlugins()
    AddGlobalPlugin(plugin1)
    defer func() {
        CleanGlobalPlugin()
        AddGlobalPlugin(plugins...)
    }()
    start := time.Now()
    for _, name := range names {
        wg.Add(1)
        name := name
        go func() {
            defer wg.Done()
            for i := 0; i < 100; i++ {
                i := i
                Logger(name) /*.WithPlugin(plugin1)*/ .T("test", i).UnUse()
                Logger(name) /*.WithPlugin(plugin1)*/ .D("test", i).UnUse()
                Logger(name) /*.WithPlugin(plugin1)*/ .I("test", i).UnUse()
                Logger(name) /*.WithPlugin(plugin1)*/ .W("test", i).UnUse()
            }
        }()
    }
    wg.Wait()
    duration := time.Since(start)
    t.Log("结束", adder.Sum(), duration, float64(adder.Sum())/duration.Seconds())
    time.Sleep(100 * time.Millisecond)
    t.Log("结束", adder.Sum(), duration, float64(adder.Sum())/duration.Seconds())
}

func Test_Logger_111(t *testing.T) {
    Logger("test").I("-{}-", 123, "fsadf", 100)
    Logger("test").I("{}--{}--{}--")
    Logger("test").I("--{}--{}--{}")
    Logger("test").I("{}--{}--{}--{}")
    Logger("test").I("{}--{}--{}--{}", 1, 2)
    Logger("test").I("--{}{}}{--", 1)
    Logger("test").I("--{{}}}--", 123)
}
