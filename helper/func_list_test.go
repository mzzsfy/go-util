package helper_test

import (
    "github.com/mzzsfy/go-util/helper"
    "math/rand"
    "testing"
    "time"
)

func TestFuncList(t *testing.T) {
    rand.Seed(time.Now().UnixMilli())
    list := &helper.FuncList{}
    list.AddFunc(func(c *helper.FuncListContext) {
        if c.CallNumber() == 0 {
            t.Log("第一次执行,跳过")
            return
        }
        if c.Value("1") != nil {
            t.Log("已有数据,跳过执行", c.Value("1"))
            return
        }
        i := rand.Intn(10)
        t.Log("1", i)
        if i > 5 {
            c.WithValue("1", i)
        } else {
            t.Log("数据小于5,跳过")
        }
    })
    list.AddFunc(func(c *helper.FuncListContext) {
        if c.CallNumber() > 0 {
            return
        }
        t.Log("2被执行了")
    })
    list.AddFunc(func(c *helper.FuncListContext) {
        t.Log("3被执行了")
        if c.CallNumber() > 2 {
            return
        }
        t.Log("后退3步")
        c.RedirectStep(-3)
    })
    list.AddFunc(func(c *helper.FuncListContext) {
        t.Log("4被执行了")
        c.RedirectStep(100)
    })
    list.AddFunc(func(c *helper.FuncListContext) {
        t.Error("5不会被执行")
    })
    list.Complete()
}

func TestFuncList1(t *testing.T) {
    rand.Seed(time.Now().UnixMilli())
    list := &helper.FuncListNamed{}
    list.AddFunc("1", func(c *helper.FuncListNamedContext) {
        if c.CallNumber() == 0 {
            t.Log("第一次执行,跳过")
            return
        }
        if c.Value("1") != nil {
            t.Log("已有数据,跳过执行", c.Value("1"))
            return
        }
        i := rand.Intn(10)
        t.Log("1", i)
        if i > 5 {
            c.WithValue("1", i)
        } else {
            t.Log("数据小于5,跳过")
        }
    })
    list.AddFunc("2", func(c *helper.FuncListNamedContext) {
        if c.CallNumber() > 0 {
            return
        }
        t.Log("2被执行了")
    })
    list.AddFunc("3", func(c *helper.FuncListNamedContext) {
        t.Log("3被执行了")
        if c.CallNumber() > 2 {
            return
        }
        t.Log("后退到1")
        c.RedirectStep("1")
    })
    list.AddFunc("4", func(c *helper.FuncListNamedContext) {
        t.Log("4被执行了")
        c.RedirectStepStop()
    })
    list.AddFunc("5", func(c *helper.FuncListNamedContext) {
        t.Error("5不会被执行")
    })
    list.Complete()
}
