package script

import (
	"testing"
)

// ========== 数据结构深度测试 ==========
// 参照 goja/otto (JavaScript引擎) 和 gopher-lua (table操作) 的测试模式
// 覆盖数组/Map/字符串的深度操作、混合数据结构、控制流结合等场景

// ========== 1. 数组深度操作 ==========

// Test_DS_Array_CreateLarge 大数组创建和访问
func Test_DS_Array_CreateLarge(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"100元素首尾访问", `
			arr := []
			for i := 0; i < 100; i = i + 1 {
				arr = push(arr, i)
			}
			arr[0] + arr[99]
		`, 99},
		{"大数组中间元素", `
			arr := []
			for i := 0; i < 50; i = i + 1 {
				arr = push(arr, i * 2)
			}
			arr[25]
		`, 50},
		{"大数组长度", `
			arr := []
			for i := 0; i < 100; i = i + 1 {
				arr = push(arr, i)
			}
			len(arr)
		`, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_DS_Array_StoreArray 数组存储到数组中（嵌套数组赋值）
func Test_DS_Array_StoreArray(t *testing.T) {
	t.Run("嵌套数组字面量访问", func(t *testing.T) {
		runIntTest(t, `[[1, 2], [3, 4]][0][1]`, 2)
	})

	t.Run("嵌套数组赋值后访问", func(t *testing.T) {
		runIntTest(t, `
			arr := [[10, 20], [30, 40]]
			arr[1][0]
		`, 30)
	})

	t.Run("嵌套数组修改内部值", func(t *testing.T) {
		runIntTest(t, `
			arr := [[1, 2], [3, 4]]
			arr[0][1] = 99
			arr[0][1]
		`, 99)
	})

	t.Run("数组索引赋值为数组", func(t *testing.T) {
		runIntTest(t, `
			matrix := [[0, 0], [0, 0]]
			matrix[0] = [1, 2]
			matrix[0][1]
		`, 2)
	})
}

// Test_DS_Array_StoreMap 数组中存储Map
func Test_DS_Array_StoreMap(t *testing.T) {
	t.Run("数组字面量含Map", func(t *testing.T) {
		runIntTest(t, `[{"x": 1}, {"y": 2}][0]["x"]`, 1)
	})

	t.Run("数组中Map的值修改", func(t *testing.T) {
		runIntTest(t, `
			arr := [{"name": "before", "val": 1}]
			arr[0]["val"] = 42
			arr[0]["val"]
		`, 42)
	})

	t.Run("遍历数组中Map的字段累加", func(t *testing.T) {
		runIntTest(t, `
			arr := [{"v": 10}, {"v": 20}, {"v": 30}]
			total := 0
			for item := range arr {
				total = total + item["v"]
			}
			total
		`, 60)
	})
}

// Test_DS_Array_SliceThenIndex 数组切片后再索引
func Test_DS_Array_SliceThenIndex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"切片后索引0", `arr := [1, 2, 3, 4, 5]
			sub := arr[1:4]
			sub[0]`, 2},
		{"切片后索引末尾", `arr := [10, 20, 30, 40, 50]
			sub := arr[2:5]
			sub[2]`, 50},
		{"切片长度", `arr := [1, 2, 3, 4, 5, 6]
			sub := arr[1:5]
			len(sub)`, 4},
		{"嵌套切片", `arr := [1, 2, 3, 4, 5, 6, 7, 8]
			s1 := arr[2:7]
			s2 := s1[1:4]
			s2[0]`, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_DS_Array_PushSemantics push返回新数组语义
func Test_DS_Array_PushSemantics(t *testing.T) {
	t.Run("push到空数组", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			arr2 := push(arr, 42)
			arr2[0]
		`, 42)
	})

	t.Run("push不修改原数组", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr2 := push(arr, 4)
			len(arr)
		`, 3)
	})

	t.Run("push链式构建", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			arr = push(arr, 1)
			arr = push(arr, 2)
			arr = push(arr, 3)
			len(arr)
		`, 3)
	})

	t.Run("push多个元素累加", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			total := 0
			for i := 0; i < 10; i = i + 1 {
				arr = push(arr, i)
				total = total + 1
			}
			total
		`, 10)
	})

	t.Run("push后访问末尾元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2]
			arr = push(arr, 3)
			arr = push(arr, 4)
			arr[3]
		`, 4)
	})
}

// Test_DS_Array_LengthBoundary 数组长度边界
func Test_DS_Array_LengthBoundary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"空数组长度", `len([])`, 0},
		{"单元素数组长度", `len([42])`, 1},
		{"大数组长度", `len([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])`, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}

	t.Run("越界访问返回nil", func(t *testing.T) {
		result := runScript(t, `[1, 2, 3][10]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})
}

// Test_DS_Array_ForRange 数组元素遍历
func Test_DS_Array_ForRange(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"range求和", `
			arr := [10, 20, 30, 40, 50]
			sum := 0
			for v := range arr {
				sum = sum + v
			}
			sum
		`, 150},
		{"range计数", `
			arr := [1, 2, 3, 4, 5, 6, 7]
			count := 0
			for v := range arr {
				count = count + 1
			}
			count
		`, 7},
		{"range偶数累加", `
			arr := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
			total := 0
			for v := range arr {
				if v % 2 == 0 {
					total = total + v
				}
			}
			total
		`, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_DS_Array_StoreFuncResult 数组中存储函数调用结果
func Test_DS_Array_StoreFuncResult(t *testing.T) {
	t.Run("函数返回值存入数组", func(t *testing.T) {
		runIntTest(t, `
			fn double(x) {
				return x * 2
			}
			arr := [double(1), double(2), double(3)]
			arr[2]
		`, 6)
	})

	t.Run("函数返回数组并索引", func(t *testing.T) {
		runIntTest(t, `
			fn makeArr() {
				return [100, 200, 300]
			}
			makeArr()[1]
		`, 200)
	})

	t.Run("函数操作数组元素", func(t *testing.T) {
		runIntTest(t, `
			fn sum(a) {
				total := 0
				for v := range a {
					total = total + v
				}
				return total
			}
			sum([1, 2, 3, 4, 5])
		`, 15)
	})
}

// ========== 2. Map深度操作 ==========

// Test_DS_Map_NestedMap Map嵌套Map（多层）
func Test_DS_Map_NestedMap(t *testing.T) {
	t.Run("两层嵌套访问", func(t *testing.T) {
		runIntTest(t, `{"outer": {"inner": 42}}["outer"]["inner"]`, 42)
	})

	t.Run("三层嵌套访问", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": {"b": {"c": 99}}}
			m["a"]["b"]["c"]
		`, 99)
	})

	t.Run("嵌套Map赋值", func(t *testing.T) {
		runIntTest(t, `
			m := {"outer": {"val": 1}}
			m["outer"]["val"] = 777
			m["outer"]["val"]
		`, 777)
	})

	t.Run("深嵌套Map构建", func(t *testing.T) {
		runIntTest(t, `
			m := {"level1": {"level2": {"level3": {"level4": 42}}}}
			m["level1"]["level2"]["level3"]["level4"]
		`, 42)
	})
}

// Test_DS_Map_StoreArray Map中存储数组
func Test_DS_Map_StoreArray(t *testing.T) {
	t.Run("Map值为数组", func(t *testing.T) {
		runIntTest(t, `
			m := {"arr": [1, 2, 3]}
			m["arr"][1]
		`, 2)
	})

	t.Run("Map数组值修改", func(t *testing.T) {
		runIntTest(t, `
			m := {"data": [10, 20, 30]}
			m["data"][1] = 99
			m["data"][1]
		`, 99)
	})

	t.Run("Map多数组长度求和", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": [1, 2], "b": [3, 4, 5], "c": [6]}
			len(m["a"]) + len(m["b"]) + len(m["c"])
		`, 6)
	})
}

// Test_DS_Map_KeyOverride Map key覆盖
func Test_DS_Map_KeyOverride(t *testing.T) {
	t.Run("字面量重复key后者覆盖前者", func(t *testing.T) {
		// 引擎行为: 重复key保留最后一个值(与主流语言一致)
		runIntTest(t, `
			m := {"k": 1, "k": 2, "k": 3}
			m["k"]
		`, 3)
	})

	t.Run("赋值覆盖", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 10}
			m["x"] = 20
			m["x"]
		`, 20)
	})

	t.Run("多次覆盖", func(t *testing.T) {
		runIntTest(t, `
			m := {"v": 1}
			m["v"] = 2
			m["v"] = 3
			m["v"] = 4
			m["v"]
		`, 4)
	})
}

// Test_DS_Map_DeleteThenAccess Map delete后访问返回nil
func Test_DS_Map_DeleteThenAccess(t *testing.T) {
	t.Run("delete后访问返回nil", func(t *testing.T) {
		result := runScript(t, `
			m := {"a": 1, "b": 2}
			delete(m, "a")
			m["a"]
		`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("delete后len减少", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			delete(m, "b")
			len(m)
		`, 2)
	})

	t.Run("delete后其他key不变", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			delete(m, "b")
			m["a"] + m["c"]
		`, 4)
	})
}

// Test_DS_Map_Iterate Map遍历（for k := m 形式）
func Test_DS_Map_Iterate(t *testing.T) {
	t.Run("Map遍历计数", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			count := 0
			for k := m {
				count = count + 1
			}
			count
		`, 3)
	})

	t.Run("空Map遍历", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			count := 0
			for k := m {
				count = count + 1
			}
			count
		`, 0)
	})

	t.Run("大Map遍历计数", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8}
			count := 0
			for k := m {
				count = count + 1
			}
			count
		`, 8)
	})
}

// Test_DS_Map_MixedTypes Map中存储不同类型值
func Test_DS_Map_MixedTypes(t *testing.T) {
	t.Run("混合类型Map", func(t *testing.T) {
		runStringTest(t, `
			m := {"i": 42, "s": "hello", "b": true}
			m["s"]
		`, "hello")
	})

	t.Run("混合类型typeof", func(t *testing.T) {
		runStringTest(t, `
			m := {"v": 42}
			typeof(m["v"])
		`, "int")
	})

	t.Run("混合类型访问int", func(t *testing.T) {
		runIntTest(t, `
			m := {"name": "test", "age": 25, "active": true}
			m["age"]
		`, 25)
	})

	t.Run("混合类型bool", func(t *testing.T) {
		runBoolTest(t, `
			m := {"name": "test", "active": true}
			m["active"]
		`, true)
	})
}

// Test_DS_Map_LengthBoundary Map长度边界
func Test_DS_Map_LengthBoundary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"空Map", `len({})`, 0},
		{"单key", `len({"k": "v"})`, 1},
		{"多key", `len({"a": 1, "b": 2, "c": 3, "d": 4, "e": 5})`, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_DS_Map_DynamicAddKey 动态添加key到Map
func Test_DS_Map_DynamicAddKey(t *testing.T) {
	t.Run("动态添加单个key", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			m["new"] = 42
			m["new"]
		`, 42)
	})

	t.Run("动态添加后len增加", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1}
			m["b"] = 2
			m["c"] = 3
			len(m)
		`, 3)
	})

	t.Run("动态添加覆盖已有key", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 1}
			m["x"] = 100
			m["x"]
		`, 100)
	})

	t.Run("循环中动态添加key", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			for i := 0; i < 5; i = i + 1 {
				key := string(i)
				m[key] = i
			}
			len(m)
		`, 5)
	})
}

// ========== 3. 字符串操作 ==========

// Test_String_Deep_Index 字符串索引返回字符
func Test_String_Deep_Index(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"首字符", `"hello"[0]`, "h"},
		{"中间字符", `"hello"[2]`, "l"},
		{"末尾字符", `"hello"[4]`, "o"},
		{"单字符串索引", `"x"[0]`, "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runStringTest(t, tt.input, tt.expected)
		})
	}

	t.Run("越界索引返回nil", func(t *testing.T) {
		result := runScript(t, `"hi"[10]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})
}

// Test_String_Deep_Slice 字符串切片各种范围
func Test_String_Deep_Slice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"前缀切片", `"hello world"[0:5]`, "hello"},
		{"后缀切片", `"hello world"[6:11]`, "world"},
		{"中间切片", `"hello world"[3:8]`, "lo wo"},
		{"单字符切片", `"abc"[1:2]`, "b"},
		{"全长度切片", `"test"[0:4]`, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runStringTest(t, tt.input, tt.expected)
		})
	}
}

// Test_String_Deep_Concat 字符串拼接链
func Test_String_Deep_Concat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"两字符串拼接", `"hello" + " world"`, "hello world"},
		{"三字符串拼接", `"a" + "b" + "c"`, "abc"},
		{"变量拼接", `
			prefix := "Hello"
			suffix := "World"
			prefix + ", " + suffix + "!"
		`, "Hello, World!"},
		{"字符串与数字拼接", `"count: " + 42`, "count: 42"},
		{"字符串与布尔拼接", `"flag=" + true`, "flag=true"},
		{"空字符串拼接", `"" + "data"`, "data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runStringTest(t, tt.input, tt.expected)
		})
	}
}

// Test_String_Deep_Empty 空字符串操作
func Test_String_Deep_Empty(t *testing.T) {
	t.Run("空字符串长度", func(t *testing.T) {
		runIntTest(t, `len("")`, 0)
	})

	t.Run("空字符串索引返回nil", func(t *testing.T) {
		result := runScript(t, `""[0]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("空字符串拼接", func(t *testing.T) {
		runStringTest(t, `"" + "" + "x"`, "x")
	})

	t.Run("空字符串比较", func(t *testing.T) {
		runBoolTest(t, `"" == ""`, true)
	})
}

// Test_String_Deep_Long 长字符串操作
func Test_String_Deep_Long(t *testing.T) {
	t.Run("长字符串长度", func(t *testing.T) {
		runIntTest(t, `len("abcdefghijklmnopqrstuvwxyz")`, 26)
	})

	t.Run("长字符串切片", func(t *testing.T) {
		runStringTest(t, `"abcdefghijklmnopqrstuvwxyz"[10:20]`, "klmnopqrst")
	})

	t.Run("长字符串索引末尾", func(t *testing.T) {
		runStringTest(t, `"abcdefghijklmnopqrstuvwxyz"[25]`, "z")
	})

	t.Run("循环构建长字符串", func(t *testing.T) {
		runIntTest(t, `
			s := ""
			for i := 0; i < 10; i = i + 1 {
				s = s + "ab"
			}
			len(s)
		`, 20)
	})
}

// Test_String_Deep_Compare 字符串比较
func Test_String_Deep_Compare(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"相等", `"abc" == "abc"`, true},
		{"不等", `"abc" != "abd"`, true},
		{"小于", `"abc" < "abd"`, true},
		{"大于", `"xyz" > "abc"`, true},
		{"空串等于空串", `"" == ""`, true},
		{"不同长度不等", `"ab" != "abc"`, true},
		{"大于等于", `"abc" >= "abc"`, true},
		{"小于等于", `"abc" <= "abd"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

// Test_String_Deep_AsMapKey 字符串作为Map key
func Test_String_Deep_AsMapKey(t *testing.T) {
	t.Run("静态字符串key", func(t *testing.T) {
		runIntTest(t, `
			m := {"apple": 1, "banana": 2, "cherry": 3}
			m["banana"]
		`, 2)
	})

	t.Run("动态字符串key", func(t *testing.T) {
		runIntTest(t, `
			key := "dynamic"
			m := {key: 100}
			m[key]
		`, 100)
	})

	t.Run("字符串拼接作为key", func(t *testing.T) {
		runStringTest(t, `
			prefix := "user"
			m := {"user_42": "data"}
			m[prefix + "_42"]
		`, "data")
	})
}

// Test_String_Deep_InArray 字符串在数组中
func Test_String_Deep_InArray(t *testing.T) {
	t.Run("字符串数组访问", func(t *testing.T) {
		runStringTest(t, `["alpha", "beta", "gamma"][1]`, "beta")
	})

	t.Run("字符串数组长度", func(t *testing.T) {
		runIntTest(t, `len(["a", "b", "c", "d"])`, 4)
	})

	t.Run("字符串数组range", func(t *testing.T) {
		runStringTest(t, `
			arr := ["x", "y", "z"]
			result := ""
			for v := range arr {
				result = result + v
			}
			result
		`, "xyz")
	})
}

// Test_String_Deep_Convert 字符串与数字的转换
func Test_String_Deep_Convert(t *testing.T) {
	t.Run("数字转字符串", func(t *testing.T) {
		runStringTest(t, `string(42)`, "42")
	})

	t.Run("字符串转整数", func(t *testing.T) {
		runIntTest(t, `int("123")`, 123)
	})

	t.Run("负数字符串转整数", func(t *testing.T) {
		runIntTest(t, `int("-456")`, -456)
	})

	t.Run("字符串转浮点", func(t *testing.T) {
		runFloatTest(t, `float("3.14")`, 3.14)
	})

	t.Run("整数转字符串再拼接", func(t *testing.T) {
		runStringTest(t, `"val:" + string(100)`, "val:100")
	})

	t.Run("字符串转整数后运算", func(t *testing.T) {
		runIntTest(t, `int("50") + int("50")`, 100)
	})
}

// ========== 4. 混合数据结构 ==========

// Test_Mixed_Data_NestedArray3 三层数组嵌套
func Test_Mixed_Data_NestedArray3(t *testing.T) {
	t.Run("3层数组字面量访问", func(t *testing.T) {
		runIntTest(t, `[[[1, 2], [3, 4]], [[5, 6], [7, 8]]][1][0][1]`, 6)
	})

	t.Run("3层数组赋值", func(t *testing.T) {
		runIntTest(t, `
			deep := [[[0, 0], [0, 0]], [[0, 0], [0, 0]]]
			deep[1][1][0] = 42
			deep[1][1][0]
		`, 42)
	})

	t.Run("3层数组range遍历", func(t *testing.T) {
		// 用索引循环遍历二维数组
		runIntTest(t, `
			matrix := [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
			total := 0
			for i := 0; i < len(matrix); i = i + 1 {
				row := matrix[i]
				for j := 0; j < len(row); j = j + 1 {
					total = total + row[j]
				}
			}
			total
		`, 45)
	})
}

// Test_Mixed_Data_NestedMap3 三层Map嵌套
func Test_Mixed_Data_NestedMap3(t *testing.T) {
	t.Run("3层Map访问", func(t *testing.T) {
		runIntTest(t, `{"a": {"b": {"c": 42}}}["a"]["b"]["c"]`, 42)
	})

	t.Run("3层Map修改", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": {"y": {"z": 1}}}
			m["x"]["y"]["z"] = 999
			m["x"]["y"]["z"]
		`, 999)
	})

	t.Run("3层Map不同路径", func(t *testing.T) {
		runIntTest(t, `
			m := {
				"first": {"a": 1, "b": 2},
				"second": {"a": 3, "b": 4}
			}
			m["first"]["b"] + m["second"]["a"]
		`, 5)
	})
}

// Test_Mixed_Data_ArrayMapArray 数组中嵌套Map再嵌套数组
func Test_Mixed_Data_ArrayMapArray(t *testing.T) {
	t.Run("数组-Map-数组访问", func(t *testing.T) {
		runIntTest(t, `
			data := [{"nums": [10, 20, 30]}, {"nums": [40, 50, 60]}]
			data[1]["nums"][2]
		`, 60)
	})

	t.Run("数组-Map-数组修改", func(t *testing.T) {
		runIntTest(t, `
			data := [{"items": [1, 2, 3]}]
			data[0]["items"][1] = 99
			data[0]["items"][1]
		`, 99)
	})

	t.Run("遍历数组中Map的数组", func(t *testing.T) {
		// 单层遍历: 从数组中取Map字段值累加
		runIntTest(t, `
			data := [{"vals": [1, 2]}, {"vals": [3, 4]}]
			len(data[0]["vals"]) + len(data[1]["vals"])
		`, 4)
	})
}

// Test_Mixed_Data_ComplexJSON 复杂JSON式数据结构构建和访问
func Test_Mixed_Data_ComplexJSON(t *testing.T) {
	t.Run("用户档案结构", func(t *testing.T) {
		runStringTest(t, `
			user := {
				"name": "Alice",
				"age": 30,
				"tags": ["admin", "user"],
				"meta": {"created": "2024-01-01", "level": 5}
			}
			user["name"]
		`, "Alice")
	})

	t.Run("复杂结构数值访问", func(t *testing.T) {
		runIntTest(t, `
			config := {
				"server": {
					"port": 8080,
					"hosts": ["localhost", "127.0.0.1"]
				},
				"debug": true
			}
			config["server"]["port"]
		`, 8080)
	})

	t.Run("复杂结构数组访问", func(t *testing.T) {
		runStringTest(t, `
			config := {
				"server": {
					"port": 8080,
					"hosts": ["localhost", "127.0.0.1"]
				}
			}
			config["server"]["hosts"][0]
		`, "localhost")
	})

	t.Run("复杂结构嵌套修改", func(t *testing.T) {
		runIntTest(t, `
			data := {
				"list": [{"val": 1}, {"val": 2}]
			}
			data["list"][1]["val"] = 42
			data["list"][1]["val"]
		`, 42)
	})
}

// Test_Mixed_Data_Traverse 复杂结构遍历
func Test_Mixed_Data_Traverse(t *testing.T) {
	t.Run("遍历Map中数组长度求和", func(t *testing.T) {
		// 通过字段名直接访问Map中的数组值
		runIntTest(t, `
			m := {"a": [1, 2, 3], "b": [4, 5], "c": [6, 7, 8, 9]}
			len(m["a"]) + len(m["b"]) + len(m["c"])
		`, 9)
	})

	t.Run("数组中Map字段聚合", func(t *testing.T) {
		runIntTest(t, `
			people := [
				{"name": "A", "score": 85},
				{"name": "B", "score": 90},
				{"name": "C", "score": 78}
			]
			total := 0
			for p := range people {
				total = total + p["score"]
			}
			total
		`, 253)
	})
}

// ========== 5. 数据结构与控制流结合 ==========

// Test_DS_Control_BuildArrayInLoop 在for循环中构建数组
func Test_DS_Control_BuildArrayInLoop(t *testing.T) {
	t.Run("循环push构建数组", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			for i := 0; i < 10; i = i + 1 {
				arr = push(arr, i * i)
			}
			arr[5]
		`, 25)
	})

	t.Run("条件push构建数组", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			for i := 0; i < 20; i = i + 1 {
				if i % 3 == 0 {
					arr = push(arr, i)
				}
			}
			len(arr)
		`, 7)
	})

	t.Run("循环构建嵌套数组", func(t *testing.T) {
		runIntTest(t, `
			matrix := []
			for i := 0; i < 3; i = i + 1 {
				row := []
				for j := 0; j < 3; j = j + 1 {
					row = push(row, i * 3 + j)
				}
				matrix = push(matrix, row)
			}
			matrix[2][2]
		`, 8)
	})
}

// Test_DS_Control_OperateMapInLoop 在for循环中操作Map
func Test_DS_Control_OperateMapInLoop(t *testing.T) {
	t.Run("循环中构建Map", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			for i := 0; i < 5; i = i + 1 {
				key := string(i)
				m[key] = i * 10
			}
			m[string(3)]
		`, 30)
	})

	t.Run("循环中累加Map值", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 0, "b": 0, "c": 0}
			for i := 0; i < 10; i = i + 1 {
				if i % 3 == 0 {
					m["a"] = m["a"] + i
				}
			}
			m["a"]
		`, 18)
	})

	t.Run("循环中动态添加多key", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			for i := 0; i < 10; i = i + 1 {
				key := "key_" + string(i)
				m[key] = i
			}
			len(m)
		`, 10)
	})
}

// Test_DS_Control_ConditionAccess 条件判断中访问数据结构
func Test_DS_Control_ConditionAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"条件取数组元素", `
			arr := [10, 20, 30, 40, 50]
			result := 0
			if len(arr) > 3 {
				result = arr[3]
			}
			result
		`, 40},
		{"条件取Map值", `
			m := {"x": 1, "y": 2}
			result := 0
			if m["x"] > 0 {
				result = m["y"]
			}
			result
		`, 2},
		{"else分支取数组", `
			arr := [1, 2, 3]
			idx := 5
			result := 0
			if idx < len(arr) {
				result = arr[idx]
			} else {
				result = arr[0]
			}
			result
		`, 1},
		{"嵌套条件Map访问", `
			m := {"a": {"val": 10}, "b": {"val": 20}}
			key := "b"
			result := 0
			if key == "a" {
				result = m["a"]["val"]
			} else {
				result = m["b"]["val"]
			}
			result
		`, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_DS_Control_FuncParamPassing 函数参数传递数据结构
func Test_DS_Control_FuncParamPassing(t *testing.T) {
	t.Run("函数接收数组", func(t *testing.T) {
		runIntTest(t, `
			fn sum(arr) {
				total := 0
				for v := range arr {
					total = total + v
				}
				return total
			}
			sum([1, 2, 3, 4, 5])
		`, 15)
	})

	t.Run("函数接收Map", func(t *testing.T) {
		runIntTest(t, `
			fn getValue(m) {
				return m["key"]
			}
			getValue({"key": 42})
		`, 42)
	})

	t.Run("函数接收字符串", func(t *testing.T) {
		runStringTest(t, `
			fn process(s) {
				return s + "!"
			}
			process("hello")
		`, "hello!")
	})

	t.Run("函数修改传入的Map", func(t *testing.T) {
		runIntTest(t, `
			fn setVal(m) {
				m["new"] = 99
				return m["new"]
			}
			setVal({"old": 1})
		`, 99)
	})
}

// Test_DS_Control_FuncReturnDS 函数返回数据结构
func Test_DS_Control_FuncReturnDS(t *testing.T) {
	t.Run("函数返回数组", func(t *testing.T) {
		// range是保留关键字, 使用genRange作为函数名
		runIntTest(t, `
			fn genRange(n) {
				arr := []
				for i := 0; i < n; i = i + 1 {
					arr = push(arr, i)
				}
				return arr
			}
			result := genRange(5)
			result[4]
		`, 4)
	})

	t.Run("函数返回Map", func(t *testing.T) {
		runStringTest(t, `
			fn makeUser(name) {
				return {"name": name, "active": true}
			}
			user := makeUser("Bob")
			user["name"]
		`, "Bob")
	})

	t.Run("函数返回嵌套结构", func(t *testing.T) {
		runIntTest(t, `
			fn makeConfig() {
				return {
					"limits": {"max": 100, "min": 0},
					"tags": ["a", "b"]
				}
			}
			cfg := makeConfig()
			cfg["limits"]["max"]
		`, 100)
	})

	t.Run("函数返回数组后索引", func(t *testing.T) {
		runIntTest(t, `
			fn reverse(arr) {
				result := []
				n := len(arr)
				for i := 0; i < n; i = i + 1 {
					result = push(result, arr[n - 1 - i])
				}
				return result
			}
			reverse([1, 2, 3, 4])[0]
		`, 4)
	})
}

// ========== 6. push/delete操作语义 ==========

// Test_DS_PushDelete_PushToEmpty push到空数组
func Test_DS_PushDelete_PushToEmpty(t *testing.T) {
	t.Run("空数组push一个元素", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			arr2 := push(arr, 1)
			len(arr2)
		`, 1)
	})

	t.Run("空数组push后访问", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			arr = push(arr, 42)
			arr[0]
		`, 42)
	})

	t.Run("push到空数组返回非nil", func(t *testing.T) {
		result := runScript(t, `
			arr := []
			push(arr, "item")
		`)
		if result.IsNil() {
			t.Error("push到空数组不应返回nil")
		}
	})
}

// Test_DS_PushDelete_PushToExisting push到已有数组
func Test_DS_PushDelete_PushToExisting(t *testing.T) {
	t.Run("追加到末尾", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr2 := push(arr, 4)
			arr2[3]
		`, 4)
	})

	t.Run("追加后长度正确", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr = push(arr, 4)
			arr = push(arr, 5)
			len(arr)
		`, 5)
	})

	t.Run("原数组完整保留", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			arr2 := push(arr, 40)
			arr[0] + arr[1] + arr[2]
		`, 60)
	})
}

// Test_DS_PushDelete_PushChain push链式调用
func Test_DS_PushDelete_PushChain(t *testing.T) {
	t.Run("链式push构建数组", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			arr = push(arr, 1)
			arr = push(arr, 2)
			arr = push(arr, 3)
			arr = push(arr, 4)
			arr = push(arr, 5)
			len(arr)
		`, 5)
	})

	t.Run("链式push求和", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			arr = push(arr, 10)
			arr = push(arr, 20)
			arr = push(arr, 30)
			total := 0
			for v := range arr {
				total = total + v
			}
			total
		`, 60)
	})

	t.Run("push不能直接链式调用", func(t *testing.T) {
		// push返回新数组，不是链式方法调用
		// push(push(push([], 1), 2), 3) 形式可以
		runIntTest(t, `
			arr := push(push(push([], 1), 2), 3)
			len(arr)
		`, 3)
	})
}

// Test_DS_PushDelete_DeleteExisting delete已存在的key
func Test_DS_PushDelete_DeleteExisting(t *testing.T) {
	t.Run("删除单个key", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			delete(m, "b")
			len(m)
		`, 2)
	})

	t.Run("删除后访问返回nil", func(t *testing.T) {
		result := runScript(t, `
			m := {"x": 42}
			delete(m, "x")
			m["x"]
		`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("删除后其他key不受影响", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			delete(m, "b")
			m["a"] + m["c"]
		`, 4)
	})

	t.Run("delete返回nil", func(t *testing.T) {
		result := runScript(t, `
			m := {"a": 1}
			delete(m, "a")
		`)
		if !result.IsNil() {
			t.Errorf("期望 delete 返回 nil, 得到 %v", result)
		}
	})
}

// Test_DS_PushDelete_DeleteNonExist delete不存在的key
func Test_DS_PushDelete_DeleteNonExist(t *testing.T) {
	t.Run("删除不存在的key不报错", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2}
			delete(m, "nonexistent")
			len(m)
		`, 2)
	})

	t.Run("删除空Map的key", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			delete(m, "x")
			len(m)
		`, 0)
	})

	t.Run("多次删除不存在的key", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1}
			delete(m, "x")
			delete(m, "y")
			delete(m, "z")
			len(m)
		`, 1)
	})
}

// Test_DS_PushDelete_DeleteLenChange delete后Map长度变化
func Test_DS_PushDelete_DeleteLenChange(t *testing.T) {
	t.Run("删除一个key后长度减一", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3, "d": 4}
			delete(m, "c")
			len(m)
		`, 3)
	})

	t.Run("逐个删除直到空", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			delete(m, "a")
			delete(m, "b")
			delete(m, "c")
			len(m)
		`, 0)
	})

	t.Run("删除后重新添加长度恢复", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2}
			delete(m, "a")
			m["a"] = 10
			len(m)
		`, 2)
	})

	t.Run("快速排序场景中的push", func(t *testing.T) {
		runIntTest(t, `
			arr := [3, 1, 4, 1, 5, 9, 2, 6]
			pivot := 3
			less := []
			greater := []
			for v := range arr {
				if v < pivot {
					less = push(less, v)
				} else {
					greater = push(greater, v)
				}
			}
			len(less) + len(greater)
		`, 8)
	})
}
