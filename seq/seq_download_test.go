package seq

import (
    "math/rand"
    "testing"
    "time"
)

//强制顺序的文件下载
func TestFileDownload(t *testing.T) {
    preTest(t)
    it := IteratorInt(1)
    From(func(t func(str string)) {
        for {
            t("")
        }
    }).Take(33).MapBiSerialNumber(1).OnEachAfter(func(i int, a string) {
        t.Logf("开始下载第%d个文件2\n", i)
    }).OnEachN(4, func(k int, v string) {
        t.Logf("有4个文件开始下载\n")
    }).OnEach(func(i int, s string) {
        t.Logf("开始下载第%d个文件1\n", i)
    }).MapVParallel(func(i int, s string) any {
        t.Logf("实际开始下载第%d个文件\n", i)
        time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
        //下载文件,返回[]byte
        t.Logf("实际下载完成第%d个文件:%s\n", i, s)
        return []byte(s)
    },
        //设置为按照顺序下载 1弱排序 2强排序
        2,
        //并发数
        10,
        //强制转换
    ).OnEachN(5, func(k int, v any) {
        t.Logf("有5个文件下载完成\n")
    }).OnBefore(1, func(i int, a any) {
        t.Logf("第一个文件下载完成,%d\n", i)
    }).OnEach(func(i int, a any) {
        t.Logf("下载完成,%d\n", i)
    }).OnLast(func(i *int, a *any) {
        t.Logf("所有文件下载完成,%d\n", *i)
    }).ForEach(func(i int, a any) {
        t.Logf("完成,%d\n", i)
        e, _ := it()
        if e != i {
            t.Fail()
        }
    })
}

//非强制顺序的文件下载
func TestFileDownload1(t *testing.T) {
    preTest(t)
    it := IteratorInt(1)
    From(func(t func(str string)) {
        for {
            t("")
        }
    }).Take(44).MapBiSerialNumber(1).OnEachAfter(func(i int, a string) {
        t.Logf("开始下载第%d个文件2\n", i)
    }).OnEachN(5, func(k int, v string) {
        t.Logf("有5个文件开始下载\n")
    }).OnEach(func(i int, s string) {
        t.Logf("开始下载第%d个文件1\n", i)
    }).Parallel(10).MapV(func(i int, s string) any {
        t.Logf("实际开始下载第%d个文件\n", i)
        time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
        //下载文件,返回[]byte
        t.Logf("实际下载完成第%d个文件:%s\n", i, s)
        return []byte(s)
    },
    ).OnBefore(1, func(i int, a any) {
        t.Logf("第一个文件下载完成,%d\n", i)
    }).OnEachN(3, func(k int, v any) {
        t.Logf("有3个文件下载完成\n")
    }).OnEach(func(i int, a any) {
        t.Logf("下载完成,%d\n", i)
    }).OnLast(func(i *int, a *any) {
        t.Logf("所有文件下载完成,%d\n", *i)
    }).ForEach(func(i int, a any) {
        t.Logf("完成,%d\n", i)
        e, _ := it()
        if e != i {
            t.Logf("预期的非强一致顺序,预计id%d,实际%d", e, i)
        }
    })
}
