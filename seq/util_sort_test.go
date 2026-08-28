package seq

import (
	"math/rand"
	"sort"
	"testing"
)

// assertSorted 断言与标准库排序结果一致
func assertSorted(t *testing.T, r []int) {
	t.Helper()
	expect := make([]int, len(r))
	copy(expect, r)
	sort.Slice(expect, func(i, j int) bool { return expect[i] < expect[j] })
	for i := range r {
		if r[i] != expect[i] {
			t.Fatalf("排序结果与sort.Slice不一致,位置%d: %v != %v", i, r[i:], expect[i:])
		}
	}
}

// 场景: 高重复随机数据排序正确性
// Given 值域远小于长度的随机切片
// When 多轮随机规模排序
// Then 结果与sort.Slice一致
func Test_SortSlice_HighDuplicate(t *testing.T) {
	t.Parallel()
	rd := rand.New(rand.NewSource(1))
	for i := 0; i < 2000; i++ {
		n := rd.Intn(64)
		r := make([]int, n)
		for j := range r {
			r[j] = rd.Intn(5)
		}
		sortSlice(r, LessT[int])
		assertSorted(t, r)
	}
}

// 场景: 全等元素排序
// Given 100个相同元素
// When 排序
// Then 长度不变且全部相等
func Test_SortSlice_AllEqual(t *testing.T) {
	t.Parallel()
	r := make([]int, 100)
	sortSlice(r, LessT[int])
	for _, v := range r {
		if v != 0 {
			t.Fatalf("全等元素排序后应全为0,实际:%v", r)
		}
	}
}

// 场景: 逆序大数组排序
// Given 100k逆序元素
// When 排序
// Then 结果与sort.Slice一致
func Test_SortSlice_Reversed(t *testing.T) {
	t.Parallel()
	r := make([]int, 100*1000)
	for i := range r {
		r[i] = len(r) - i
	}
	sortSlice(r, LessT[int])
	assertSorted(t, r)
}

// 场景: 跨插入排序阈值边界的规模
// Given 长度为阈值两侧的乱序切片
// When 多轮排序
// Then 结果与sort.Slice一致
func Test_SortSlice_ThresholdBoundary(t *testing.T) {
	t.Parallel()
	rd := rand.New(rand.NewSource(2))
	for round := 0; round < 100; round++ {
		for _, n := range []int{11, 12, 13, 14} {
			r := make([]int, n)
			for j := range r {
				r[j] = rd.Intn(100)
			}
			sortSlice(r, LessT[int])
			assertSorted(t, r)
		}
	}
}

// 场景: 已有序数据排序
// Given 递增切片
// When 排序
// Then 结果不变
func Test_SortSlice_AlreadySorted(t *testing.T) {
	t.Parallel()
	r := make([]int, 1000)
	for i := range r {
		r[i] = i
	}
	sortSlice(r, LessT[int])
	assertSorted(t, r)
}
