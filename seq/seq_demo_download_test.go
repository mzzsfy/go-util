package seq

import (
    "math/rand"
    "testing"
    "time"
)

//非强制顺序的文件下载
func TestFileDownload1(t *testing.T) {
    preTest(t)
    it := IteratorInt(1)
    BiMapV(MapBiSerialNumber(From(func(t func(str string)) {
        for {
            t("")
        }
    }).Take(44), 1).OnEach(func(i int, a string) {
        t.Logf("开始下载第%d个文件\n", i)
    }).OnEach(func(i int, s string) {
        t.Logf("开始下载第%d个文件1\n", i)
    }).Parallel(10), func(i int, s string) any {
        t.Logf("实际开始下载第%d个文件\n", i)
        time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
        //下载文件,返回[]byte
        t.Logf("实际下载完成第%d个文件:%s\n", i, s)
        return []byte(s)
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
            t.Logf("预期的非强一致顺序,预计id%d,实际%d", e, i)
        }
    })
}

//强制顺序的文件下载,尽量保证并发的情况下,快速下载完成
func TestFileDownload2(t *testing.T) {
    preTest(t)
    it := IteratorInt(1)
    MapBiSerialNumber(From(func(t func(str string)) {
        for {
            t("")
        }
    }).Take(13), 1).OnEach(func(i int, a string) {
        t.Logf("开始下载第%d个文件\n", i)
    }).MapVParallel(func(i int, s string) any {
        t.Logf("实际开始下载第%d个文件\n", i)
        time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
        //下载文件,返回[]byte
        t.Logf("实际下载完成第%d个文件:%s\n", i, s)
        return []byte(s)
    },
        //设置为按照顺序下载 1弱排序 2强排序
        2,
        //并发数
        5,
        //强制转换
    ).OnBefore(1, func(i int, a any) {
        t.Logf("第一个文件下载完成,%d\n", i)
    }).OnEach(func(i int, a any) {
        t.Logf("下载完成,%d\n", i)
    }).OnLast(func(i *int, a *any) {
        t.Logf("所有文件下载完成,%d\n", *i)
    }).ForEach(func(i int, a any) {
        t.Logf("开始本地费时逻辑,%d\n", i)
        time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
        t.Logf("完成,%d\n", i)
        e, _ := it()
        if e != i {
            t.Fail()
        }
    })
}

//强制顺序的文件下载,并且本地有费时操作,同时保证内存消耗尽量少
func TestFileDownload3(t *testing.T) {
    preTest(t)
    it := IteratorInt(1)
    MapBiSerialNumber(From(func(t func(str string)) {
        for {
            t("")
        }
    }).Take(8), 1).OnEach(func(i int, a string) {
        t.Logf("开始下载第%d个文件\n", i)
    }).OnEach(func(i int, s string) {
        t.Logf("开始下载第%d个文件1\n", i)
    }).MapVParallel(func(i int, s string) any {
        t.Logf("实际开始下载第%d个文件\n", i)
        time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
        //下载文件,返回[]byte
        t.Logf("实际下载完成第%d个文件:%s\n", i, s)
        return []byte(s)
    },
        //设置为按照顺序下载 1弱排序 2强排序 3 耦合式强排序
        3,
        //并发数
        3,
        //强制转换
    ).OnBefore(1, func(i int, a any) {
        t.Logf("第一个文件下载完成,%d\n", i)
    }).OnEach(func(i int, a any) {
        t.Logf("下载完成,%d\n", i)
    }).OnLast(func(i *int, a *any) {
        t.Logf("所有文件下载完成,%d\n", *i)
    }).ForEach(func(i int, a any) {
        t.Logf("开始本地费时逻辑,%d\n", i)
        time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
        t.Logf("完成,%d\n", i)
        e, _ := it()
        if e != i {
            t.Fail()
        }
    })
}
