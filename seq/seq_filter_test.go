package seq

import (
    "testing"
)

func TestDistinctByKey(t *testing.T) {
    t.Parallel()
    // 测试基本去重
    seq := FromSlice([]int{1, 2, 2, 3, 3, 3, 4, 5, 5})
    result := DistinctByKey(seq, func(e int) int { return e }).ToSlice()
    if len(result) != 5 {
        t.Fatalf("expected 5 elements, got %d", len(result))
    }
    expected := []int{1, 2, 3, 4, 5}
    for i, v := range result {
        if v != expected[i] {
            t.Fatalf("expected %d at index %d, got %d", expected[i], i, v)
        }
    }
}

func TestDistinctByKey_CustomKey(t *testing.T) {
    t.Parallel()
    // 测试自定义key函数
    type User struct {
        ID   int
        Name string
    }
    users := FromSlice([]User{
        {ID: 1, Name: "a"},
        {ID: 2, Name: "b"},
        {ID: 1, Name: "c"}, // 重复ID
        {ID: 3, Name: "d"},
    })
    result := DistinctByKey(users, func(u User) int { return u.ID }).ToSlice()
    if len(result) != 3 {
        t.Fatalf("expected 3 elements, got %d", len(result))
    }
    // 验证保留首次出现的元素
    if result[0].ID != 1 || result[0].Name != "a" {
        t.Fatalf("expected first user {1, a}, got %+v", result[0])
    }
    if result[1].ID != 2 || result[1].Name != "b" {
        t.Fatalf("expected second user {2, b}, got %+v", result[1])
    }
}

func TestDistinctByKey_Empty(t *testing.T) {
    t.Parallel()
    // 测试空序列
    seq := FromSlice([]int{})
    result := DistinctByKey(seq, func(e int) int { return e }).ToSlice()
    if len(result) != 0 {
        t.Fatalf("expected 0 elements, got %d", len(result))
    }
}

func TestDistinctComparable(t *testing.T) {
    t.Parallel()
    // 测试comparable类型直接去重
    seq := FromSlice([]string{"a", "b", "a", "c", "b", "d"})
    result := DistinctComparable(seq).ToSlice()
    if len(result) != 4 {
        t.Fatalf("expected 4 elements, got %d", len(result))
    }
    expected := []string{"a", "b", "c", "d"}
    for i, v := range result {
        if v != expected[i] {
            t.Fatalf("expected %s at index %d, got %s", expected[i], i, v)
        }
    }
}

func TestDistinctComparable_Int(t *testing.T) {
    t.Parallel()
    // 测试int类型
    seq := FromSlice([]int{5, 3, 5, 1, 3, 2, 1})
    result := DistinctComparable(seq).ToSlice()
    if len(result) != 4 {
        t.Fatalf("expected 4 elements, got %d", len(result))
    }
    // 验证顺序保持
    expected := []int{5, 3, 1, 2}
    for i, v := range result {
        if v != expected[i] {
            t.Fatalf("expected %d at index %d, got %d", expected[i], i, v)
        }
    }
}

// TestTakeWhile_Empty 测试TakeWhile空序列
func TestTakeWhile_Empty(t *testing.T) {
    t.Parallel()
    // 测试空序列
    seq := FromSlice([]int{})
    result := seq.TakeWhile(func(e int) bool { return e < 10 }).ToSlice()
    if len(result) != 0 {
        t.Fatalf("空序列结果应为空, 实际长度 %d", len(result))
    }
}

// TestTakeWhile_FirstElementFails 测试TakeWhile第一个元素就不满足条件
func TestTakeWhile_FirstElementFails(t *testing.T) {
    t.Parallel()
    // 第一个元素就不满足条件
    seq := FromSlice([]int{5, 1, 2, 3, 4})
    result := seq.TakeWhile(func(e int) bool { return e < 5 }).ToSlice()
    if len(result) != 0 {
        t.Fatalf("第一个元素不满足条件结果应为空, 实际长度 %d", len(result))
    }
}

// TestTakeWhile_AllElementsSatisfy 测试TakeWhile所有元素都满足条件
func TestTakeWhile_AllElementsSatisfy(t *testing.T) {
    t.Parallel()
    // 所有元素都满足条件
    seq := FromSlice([]int{1, 2, 3, 4, 5})
    result := seq.TakeWhile(func(e int) bool { return e < 10 }).ToSlice()
    if len(result) != 5 {
        t.Fatalf("所有元素满足条件应返回全部, 期望长度 5, 实际 %d", len(result))
    }
    expected := []int{1, 2, 3, 4, 5}
    for i, v := range result {
        if v != expected[i] {
            t.Fatalf("索引 %d 期望 %d, 实际 %d", i, expected[i], v)
        }
    }
}

// TestTakeWhile_PartialMatch 测试TakeWhile部分匹配
func TestTakeWhile_PartialMatch(t *testing.T) {
    t.Parallel()
    // 部分元素满足条件
    seq := FromSlice([]int{1, 2, 3, 10, 4, 5})
    result := seq.TakeWhile(func(e int) bool { return e < 10 }).ToSlice()
    if len(result) != 3 {
        t.Fatalf("部分匹配应返回前3个元素, 实际长度 %d", len(result))
    }
    expected := []int{1, 2, 3}
    for i, v := range result {
        if v != expected[i] {
            t.Fatalf("索引 %d 期望 %d, 实际 %d", i, expected[i], v)
        }
    }
}

// TestDistinctByKey_AllSame 测试DistinctByKey所有元素相同
func TestDistinctByKey_AllSame(t *testing.T) {
    t.Parallel()
    // 所有元素key相同
    seq := FromSlice([]int{5, 5, 5, 5, 5})
    result := DistinctByKey(seq, func(e int) int { return e }).ToSlice()
    if len(result) != 1 {
        t.Fatalf("所有元素相同应只保留一个, 实际长度 %d", len(result))
    }
    if result[0] != 5 {
        t.Fatalf("期望值 5, 实际 %d", result[0])
    }
}

// TestDistinctByKey_AllSameKey 测试DistinctByKey所有元素key相同但值不同
func TestDistinctByKey_AllSameKey(t *testing.T) {
    t.Parallel()
    type Item struct {
        ID    int
        Value string
    }
    // 所有元素key相同但值不同
    seq := FromSlice([]Item{
        {ID: 1, Value: "a"},
        {ID: 1, Value: "b"},
        {ID: 1, Value: "c"},
    })
    result := DistinctByKey(seq, func(e Item) int { return e.ID }).ToSlice()
    if len(result) != 1 {
        t.Fatalf("所有key相同应只保留第一个, 实际长度 %d", len(result))
    }
    // 验证保留的是第一个元素
    if result[0].Value != "a" {
        t.Fatalf("期望保留第一个元素 Value='a', 实际 Value='%s'", result[0].Value)
    }
}

// TestDistinctByKey_NilKey 测试DistinctByKey使用nil或零值作为key
func TestDistinctByKey_NilKey(t *testing.T) {
    t.Parallel()
    // 测试零值key
    seq := FromSlice([]int{0, 0, 0, 1, 1, 2})
    result := DistinctByKey(seq, func(e int) int { return e }).ToSlice()
    if len(result) != 3 {
        t.Fatalf("期望去重后3个元素, 实际 %d", len(result))
    }
    expected := []int{0, 1, 2}
    for i, v := range result {
        if v != expected[i] {
            t.Fatalf("索引 %d 期望 %d, 实际 %d", i, expected[i], v)
        }
    }
}
