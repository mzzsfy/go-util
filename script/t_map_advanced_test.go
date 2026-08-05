package script

import "testing"

// ========== Map高级操作测试 ==========
// 完整测试Map的创建、访问、赋值、长度、遍历、删除、比较等操作

// ========== 1. Map创建和初始化 ==========

// Test_Map_Create 空Map及各种初始化
func Test_Map_Create(t *testing.T) {
	t.Run("空Map字面量", func(t *testing.T) {
		runIntTest(t, `len({})`, 0)
	})

	t.Run("空Map赋值给变量", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			len(m)
		`, 0)
	})

	t.Run("单元素Map", func(t *testing.T) {
		runIntTest(t, `{"a": 1}["a"]`, 1)
	})

	t.Run("多元素Map", func(t *testing.T) {
		runIntTest(t, `{"a": 1, "b": 2, "c": 3}["b"]`, 2)
	})

	t.Run("Map多元素无逗号结尾", func(t *testing.T) {
		runIntTest(t, `len({"a": 1, "b": 2})`, 2)
	})

	t.Run("Map只含一个key值对赋值后访问", func(t *testing.T) {
		runIntTest(t, `
			m := {"single": 42}
			m["single"]
		`, 42)
	})
}

// Test_Map_Create_Nested 嵌套Map创建
func Test_Map_Create_Nested(t *testing.T) {
	t.Run("两层嵌套Map", func(t *testing.T) {
		runIntTest(t, `{"outer": {"inner": 1}}["outer"]["inner"]`, 1)
	})

	t.Run("三层嵌套Map", func(t *testing.T) {
		runIntTest(t, `{"a": {"b": {"c": 42}}}["a"]["b"]["c"]`, 42)
	})

	t.Run("四层嵌套Map构建", func(t *testing.T) {
		runIntTest(t, `
			m := {"l1": {"l2": {"l3": {"l4": 99}}}}
			m["l1"]["l2"]["l3"]["l4"]
		`, 99)
	})

	t.Run("同级多key嵌套Map", func(t *testing.T) {
		runIntTest(t, `
			m := {
				"first": {"a": 1, "b": 2},
				"second": {"a": 3, "b": 4}
			}
			m["first"]["b"] + m["second"]["a"]
		`, 5)
	})
}

// Test_Map_Create_MixedValues Map中值类型多样性
func Test_Map_Create_MixedValues(t *testing.T) {
	t.Run("int值", func(t *testing.T) {
		runIntTest(t, `{"k": 42}["k"]`, 42)
	})

	t.Run("string值", func(t *testing.T) {
		runStringTest(t, `{"k": "hello"}["k"]`, "hello")
	})

	t.Run("bool值", func(t *testing.T) {
		runBoolTest(t, `{"k": true}["k"]`, true)
	})

	t.Run("float值", func(t *testing.T) {
		runFloatTest(t, `{"k": 3.14}["k"]`, 3.14)
	})

	t.Run("nil值", func(t *testing.T) {
		result := runScript(t, `{"k": nil}["k"]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("数组值", func(t *testing.T) {
		runIntTest(t, `{"arr": [1, 2, 3]}["arr"][2]`, 3)
	})

	t.Run("Map值嵌套Map", func(t *testing.T) {
		runIntTest(t, `{"m": {"x": 7}}["m"]["x"]`, 7)
	})

	t.Run("混合类型Map", func(t *testing.T) {
		runStringTest(t, `
			m := {"i": 42, "s": "hello", "b": true, "f": 1.5}
			m["s"]
		`, "hello")
	})
}

// Test_Map_Create_ArrayInteraction Map与数组交互
func Test_Map_Create_ArrayInteraction(t *testing.T) {
	t.Run("Map中包含数组", func(t *testing.T) {
		runIntTest(t, `
			m := {"arr": [1, 2, 3]}
			m["arr"][0] + m["arr"][1] + m["arr"][2]
		`, 6)
	})

	t.Run("数组中包含Map", func(t *testing.T) {
		runIntTest(t, `[{"a": 1}, {"b": 2}][1]["b"]`, 2)
	})

	t.Run("Map数组Map三层混合", func(t *testing.T) {
		runIntTest(t, `
			data := {"items": [{"val": 10}, {"val": 20}]}
			data["items"][1]["val"]
		`, 20)
	})

	t.Run("数组Map数组三层混合", func(t *testing.T) {
		runIntTest(t, `
			data := [{"nums": [100, 200]}, {"nums": [300, 400]}]
			data[1]["nums"][0]
		`, 300)
	})
}

// ========== 2. Map访问 ==========

// Test_Map_Access 各种Map访问方式
func Test_Map_Access(t *testing.T) {
	t.Run("正常key访问", func(t *testing.T) {
		runIntTest(t, `
			m := {"name": "test", "age": 25}
			m["age"]
		`, 25)
	})

	t.Run("不存在的key返回nil", func(t *testing.T) {
		result := runScript(t, `
			m := {"a": 1}
			m["nonexistent"]
		`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("空Map访问返回nil", func(t *testing.T) {
		result := runScript(t, `{}["any"]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("数字字符串key", func(t *testing.T) {
		runIntTest(t, `
			m := {"123": 99}
			m["123"]
		`, 99)
	})

	t.Run("空字符串key", func(t *testing.T) {
		runIntTest(t, `
			m := {"": 42}
			m[""]
		`, 42)
	})

	t.Run("长字符串key", func(t *testing.T) {
		runIntTest(t, `
			m := {"a_very_long_key_name": 7}
			m["a_very_long_key_name"]
		`, 7)
	})

	t.Run("变量作为key访问", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 10, "y": 20}
			key := "y"
			m[key]
		`, 20)
	})

	t.Run("表达式拼接作为key", func(t *testing.T) {
		runStringTest(t, `
			m := {"user_42": "data"}
			prefix := "user"
			m[prefix + "_42"]
		`, "data")
	})
}

// Test_Map_Access_Nested 嵌套Map访问
func Test_Map_Access_Nested(t *testing.T) {
	t.Run("嵌套Map访问", func(t *testing.T) {
		runIntTest(t, `
			m := {"outer": {"inner": 42}}
			m["outer"]["inner"]
		`, 42)
	})

	t.Run("Map数组混合访问", func(t *testing.T) {
		runIntTest(t, `
			m := {"arr": [10, 20, 30]}
			m["arr"][2]
		`, 30)
	})

	t.Run("数组Map混合访问", func(t *testing.T) {
		runIntTest(t, `
			arr := [{"x": 1}, {"x": 2}, {"x": 3}]
			arr[2]["x"]
		`, 3)
	})

	t.Run("三层混合Map数组Map访问", func(t *testing.T) {
		runIntTest(t, `
			data := {"list": [{"val": 77}]}
			data["list"][0]["val"]
		`, 77)
	})

	t.Run("不存在的嵌套key返回nil", func(t *testing.T) {
		result := runScript(t, `
			m := {"a": {"b": 1}}
			m["a"]["nonexist"]
		`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})
}

// ========== 3. Map赋值 ==========

// Test_Map_Assign Map赋值操作
func Test_Map_Assign(t *testing.T) {
	t.Run("新key赋值", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1}
			m["b"] = 2
			m["b"]
		`, 2)
	})

	t.Run("已有key覆盖", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 10}
			m["x"] = 20
			m["x"]
		`, 20)
	})

	t.Run("赋值后立即读取", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			m["k"] = 99
			m["k"]
		`, 99)
	})

	t.Run("多次赋值同一个key", func(t *testing.T) {
		runIntTest(t, `
			m := {"v": 1}
			m["v"] = 2
			m["v"] = 3
			m["v"] = 4
			m["v"]
		`, 4)
	})

	t.Run("赋值为新Map", func(t *testing.T) {
		runIntTest(t, `
			m := {"old": 1}
			m["new"] = {"inner": 5}
			m["new"]["inner"]
		`, 5)
	})

	t.Run("赋值为新数组", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			m["arr"] = [1, 2, 3]
			m["arr"][1]
		`, 2)
	})

	t.Run("动态添加多个key", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			m["a"] = 1
			m["b"] = 2
			m["c"] = 3
			len(m)
		`, 3)
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

// Test_Map_Assign_Nested 嵌套Map赋值
func Test_Map_Assign_Nested(t *testing.T) {
	t.Run("嵌套Map赋值", func(t *testing.T) {
		runIntTest(t, `
			m := {"outer": {"inner": 1}}
			m["outer"]["inner"] = 5
			m["outer"]["inner"]
		`, 5)
	})

	t.Run("Map中数组元素赋值", func(t *testing.T) {
		runIntTest(t, `
			m := {"arr": [1, 2, 3]}
			m["arr"][0] = 99
			m["arr"][0]
		`, 99)
	})

	t.Run("Map中数组多元素赋值", func(t *testing.T) {
		runIntTest(t, `
			m := {"data": [10, 20, 30]}
			m["data"][0] = 100
			m["data"][2] = 300
			m["data"][0] + m["data"][2]
		`, 400)
	})

	t.Run("三层嵌套赋值", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": {"b": {"c": 1}}}
			m["a"]["b"]["c"] = 777
			m["a"]["b"]["c"]
		`, 777)
	})

	t.Run("数组Map嵌套赋值", func(t *testing.T) {
		runIntTest(t, `
			arr := [{"val": 1}, {"val": 2}]
			arr[1]["val"] = 42
			arr[1]["val"]
		`, 42)
	})
}

// ========== 4. Map长度 ==========

// Test_Map_Length Map长度操作
func Test_Map_Length(t *testing.T) {
	t.Run("空Map长度", func(t *testing.T) {
		runIntTest(t, `len({})`, 0)
	})

	t.Run("单元素Map长度", func(t *testing.T) {
		runIntTest(t, `len({"k": "v"})`, 1)
	})

	t.Run("多元素Map长度", func(t *testing.T) {
		runIntTest(t, `len({"a": 1, "b": 2, "c": 3, "d": 4, "e": 5})`, 5)
	})

	t.Run("变量Map长度", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 1, "y": 2}
			len(m)
		`, 2)
	})

	t.Run("添加后len增加", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1}
			m["b"] = 2
			m["c"] = 3
			len(m)
		`, 3)
	})

	t.Run("删除后len减少", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			delete(m, "b")
			len(m)
		`, 2)
	})

	t.Run("覆盖已有key不影响len", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2}
			m["a"] = 99
			m["b"] = 88
			len(m)
		`, 2)
	})
}

// ========== 5. Map遍历 ==========

// Test_Map_Iterate_SingleVar 单变量遍历 for k := range m
func Test_Map_Iterate_SingleVar(t *testing.T) {
	t.Run("Map字面量遍历计数", func(t *testing.T) {
		runIntTest(t, `
			count := 0
			for k := range {"a": 1, "b": 2, "c": 3} {
				count = count + 1
			}
			count
		`, 3)
	})

	t.Run("Map变量遍历计数", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			count := 0
			for k := range m {
				count = count + 1
			}
			count
		`, 3)
	})

	t.Run("隐式range遍历计数", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			count := 0
			for k := m {
				count = count + 1
			}
			count
		`, 3)
	})

	t.Run("空Map遍历不执行循环体", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			hit := 0
			for k := range m {
				hit = 1
			}
			hit
		`, 0)
	})

	t.Run("单元素Map遍历", func(t *testing.T) {
		runIntTest(t, `
			m := {"only": 42}
			count := 0
			for k := range m {
				count = count + 1
			}
			count
		`, 1)
	})

	t.Run("大Map遍历计数", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8}
			count := 0
			for k := range m {
				count = count + 1
			}
			count
		`, 8)
	})
}

// Test_Map_Iterate_TwoVars 双变量遍历 for k, v := range m
func Test_Map_Iterate_TwoVars(t *testing.T) {
	t.Run("Map字面量双变量遍历求和", func(t *testing.T) {
		runIntTest(t, `
			sum := 0
			for k, v := range {"a": 1, "b": 2, "c": 3} {
				sum = sum + v
			}
			sum
		`, 6)
	})

	t.Run("Map变量双变量遍历求和", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 10, "b": 20, "c": 30}
			total := 0
			for k, v := range m {
				total = total + v
			}
			total
		`, 60)
	})

	t.Run("双变量遍历收集key计数", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 1, "y": 2, "z": 3}
			count := 0
			for k, v := range m {
				count = count + 1
			}
			count
		`, 3)
	})

	t.Run("双变量遍历break", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			count := 0
			for k, v := range m {
				count = count + 1
				break
			}
			count
		`, 1)
	})

	t.Run("双变量遍历continue", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			count := 0
			for k, v := range m {
				if v == 2 { continue }
				count = count + 1
			}
			count
		`, 2)
	})

	t.Run("双变量遍历空Map", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			hit := 0
			for k, v := range m {
				hit = 1
			}
			hit
		`, 0)
	})

	t.Run("双变量遍历累加float", func(t *testing.T) {
		runFloatTest(t, `
			m := {"a": 1.5, "b": 2.5}
			total := 0.0
			for k, v := range m {
				total = total + v
			}
			total
		`, 4.0)
	})

	t.Run("双变量遍历值做条件", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 5, "b": 10, "c": 15, "d": 20}
			count := 0
			for k, v := range m {
				if v > 8 {
					count = count + 1
				}
			}
			count
		`, 3)
	})
}

// Test_Map_Iterate_CollectValues 遍历时累加和收集
func Test_Map_Iterate_CollectValues(t *testing.T) {
	t.Run("遍历累加所有值", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 100, "b": 200, "c": 300}
			sum := 0
			for k, v := range m {
				sum = sum + v
			}
			sum
		`, 600)
	})

	t.Run("遍历统计值大于阈值的数量", func(t *testing.T) {
		runIntTest(t, `
			scores := {"a": 85, "b": 60, "c": 95, "d": 40}
			pass := 0
			for k, v := range scores {
				if v >= 60 {
					pass = pass + 1
				}
			}
			pass
		`, 3)
	})

	t.Run("遍历字符串拼接长度", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": "a", "y": "b", "z": "c"}
			result := ""
			for k, v := range m {
				result = result + v
			}
			len(result)
		`, 3)
	})
}

// ========== 6. Map删除 ==========

// Test_Map_Delete Map删除操作
func Test_Map_Delete(t *testing.T) {
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

	t.Run("delete不存在的key不报错", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1}
			delete(m, "nonexistent")
			len(m)
		`, 1)
	})

	t.Run("delete空Map的key不报错", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			delete(m, "x")
			len(m)
		`, 0)
	})

	t.Run("delete所有key后Map为空", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2, "c": 3}
			delete(m, "a")
			delete(m, "b")
			delete(m, "c")
			len(m)
		`, 0)
	})

	t.Run("delete后重新添加", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1}
			delete(m, "a")
			m["a"] = 2
			m["a"]
		`, 2)
	})

	t.Run("delete后重新添加len恢复", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2}
			delete(m, "a")
			m["a"] = 10
			len(m)
		`, 2)
	})

	t.Run("多次delete同一个key", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 1, "b": 2}
			delete(m, "a")
			delete(m, "a")
			delete(m, "a")
			len(m)
		`, 1)
	})
}

// ========== 7. Map比较 ==========

// Test_Map_Equality Map相等性
func Test_Map_Equality(t *testing.T) {
	t.Run("相同Map相等", func(t *testing.T) {
		runBoolTest(t, `{"a": 1, "b": 2} == {"a": 1, "b": 2}`, true)
	})

	t.Run("key顺序不同Map相等", func(t *testing.T) {
		runBoolTest(t, `{"a": 1, "b": 2} == {"b": 2, "a": 1}`, true)
	})

	t.Run("值不同Map不等", func(t *testing.T) {
		runBoolTest(t, `{"a": 1} == {"a": 2}`, false)
	})

	t.Run("key不同Map不等", func(t *testing.T) {
		runBoolTest(t, `{"a": 1} == {"b": 1}`, false)
	})

	t.Run("长度不同Map不等", func(t *testing.T) {
		runBoolTest(t, `{"a": 1} == {"a": 1, "b": 2}`, false)
	})

	t.Run("空Map相等", func(t *testing.T) {
		runBoolTest(t, `{} == {}`, true)
	})

	t.Run("嵌套Map相等", func(t *testing.T) {
		runBoolTest(t, `{"a": {"b": 1}} == {"a": {"b": 1}}`, true)
	})

	t.Run("嵌套Map不等", func(t *testing.T) {
		runBoolTest(t, `{"a": {"b": 1}} == {"a": {"b": 2}}`, false)
	})

	t.Run("Map不等于非Map类型", func(t *testing.T) {
		runBoolTest(t, `{"a": 1} == 1`, false)
	})

	t.Run("Map不等于数组", func(t *testing.T) {
		runBoolTest(t, `{} == []`, false)
	})

	t.Run("Map不等运算符", func(t *testing.T) {
		runBoolTest(t, `{"a": 1} != {"a": 2}`, true)
	})
}

// ========== 8. Map作为函数参数 ==========

// Test_Map_FuncParam Map作为函数参数和返回值
func Test_Map_FuncParam(t *testing.T) {
	t.Run("传递Map给函数", func(t *testing.T) {
		runIntTest(t, `
			fn getValue(m) {
				return m["key"]
			}
			getValue({"key": 42})
		`, 42)
	})

	t.Run("函数内修改Map", func(t *testing.T) {
		runIntTest(t, `
			fn setVal(m) {
				m["new"] = 99
				return m["new"]
			}
			setVal({"old": 1})
		`, 99)
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

	t.Run("函数返回嵌套Map", func(t *testing.T) {
		runIntTest(t, `
			fn makeConfig() {
				return {"db": {"port": 3306}}
			}
			cfg := makeConfig()
			cfg["db"]["port"]
		`, 3306)
	})

	t.Run("函数返回Map后访问", func(t *testing.T) {
		runIntTest(t, `
			fn create() {
				return {"a": 1, "b": 2, "c": 3}
			}
			m := create()
			m["b"]
		`, 2)
	})

	t.Run("函数接收Map并计算len", func(t *testing.T) {
		runIntTest(t, `
			fn countKeys(m) {
				return len(m)
			}
			countKeys({"a": 1, "b": 2, "c": 3})
		`, 3)
	})

	t.Run("函数接收Map遍历", func(t *testing.T) {
		runIntTest(t, `
			fn sumValues(m) {
				total := 0
				for k, v := range m {
					total = total + v
				}
				return total
			}
			sumValues({"a": 10, "b": 20, "c": 30})
		`, 60)
	})
}

// ========== 9. Map复杂操作 ==========

// Test_Map_Complex 复杂Map操作模式
func Test_Map_Complex(t *testing.T) {
	t.Run("计数器模式", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 0, "b": 0, "c": 0}
			arr := ["a", "b", "a", "c", "a", "b"]
			for v := range arr {
				m[v] = m[v] + 1
			}
			m["a"]
		`, 3)
	})

	t.Run("计数器模式验证全部", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 0, "b": 0, "c": 0}
			arr := ["a", "b", "a", "c", "a", "b"]
			for v := range arr {
				m[v] = m[v] + 1
			}
			m["a"] * 100 + m["b"] * 10 + m["c"]
		`, 321)
	})

	t.Run("Map字段累加", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 0, "y": 0}
			for i := 0; i < 10; i = i + 1 {
				if i % 2 == 0 {
					m["x"] = m["x"] + i
				} else {
					m["y"] = m["y"] + i
				}
			}
			m["x"] + m["y"]
		`, 45)
	})

	t.Run("条件Map访问", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": {"val": 10}, "b": {"val": 20}}
			key := "b"
			result := 0
			if key == "a" {
				result = m["a"]["val"]
			} else {
				result = m["b"]["val"]
			}
			result
		`, 20)
	})

	t.Run("Map构建查找表", func(t *testing.T) {
		runIntTest(t, `
			lookup := {"one": 1, "two": 2, "three": 3}
			lookup["two"] + lookup["three"]
		`, 5)
	})

	t.Run("Map嵌套数组遍历", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": [1, 2, 3], "b": [4, 5, 6]}
			total := 0
			for k := range m {
				arr := m[k]
				for i := 0; i < len(arr); i = i + 1 {
					total = total + arr[i]
				}
			}
			total
		`, 21)
	})

	t.Run("数组Map聚合", func(t *testing.T) {
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

	t.Run("Map构建并查询", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			for i := 0; i < 5; i = i + 1 {
				key := "key_" + string(i)
				m[key] = i * i
			}
			m["key_3"]
		`, 9)
	})
}

// ========== 10. Map与绑定值交互 ==========

// Test_Map_BindValue Map与外部绑定值交互
func Test_Map_BindValue(t *testing.T) {
	t.Run("从绑定值获取Map", func(t *testing.T) {
		result := runScriptWithBinds(t, `
			m :=>map getBindValue("data")
			m["x"]
		`, map[string]interface{}{
			"data": map[string]interface{}{"x": 42},
		})
		if result.Int() != 42 {
			t.Errorf("期望 42, 得到 %d", result.Int())
		}
	})

	t.Run("从绑定值获取Map长度", func(t *testing.T) {
		result := runScriptWithBinds(t, `
			m :=>map getBindValue("data")
			len(m)
		`, map[string]interface{}{
			"data": map[string]interface{}{"a": 1, "b": 2, "c": 3},
		})
		if result.Int() != 3 {
			t.Errorf("期望 3, 得到 %d", result.Int())
		}
	})

	t.Run("从绑定值获取Map遍历", func(t *testing.T) {
		result := runScriptWithBinds(t, `
			m :=>map getBindValue("data")
			count := 0
			for k := range m {
				count = count + 1
			}
			count
		`, map[string]interface{}{
			"data": map[string]interface{}{"a": 1, "b": 2},
		})
		if result.Int() != 2 {
			t.Errorf("期望 2, 得到 %d", result.Int())
		}
	})

	t.Run("从绑定值获取嵌套Map", func(t *testing.T) {
		result := runScriptWithBinds(t, `
			m :=>map getBindValue("data")
			m["config"]["port"]
		`, map[string]interface{}{
			"data": map[string]interface{}{
				"config": map[string]interface{}{"port": 8080},
			},
		})
		if result.Int() != 8080 {
			t.Errorf("期望 8080, 得到 %d", result.Int())
		}
	})
}

// ========== 11. Map边界和错误场景 ==========

// Test_Map_EdgeCases Map边界情况
func Test_Map_EdgeCases(t *testing.T) {
	t.Run("Map字面量重复key后者覆盖", func(t *testing.T) {
		runIntTest(t, `
			m := {"k": 1, "k": 2, "k": 3}
			m["k"]
		`, 3)
	})

	t.Run("delete返回nil", func(t *testing.T) {
		result := runScript(t, `
			m := {"a": 1}
			delete(m, "a")
		`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("Map赋值为字符串", func(t *testing.T) {
		runStringTest(t, `
			m := {"key": "value"}
			m["key"]
		`, "value")
	})

	t.Run("Map多层嵌套后delete", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": {"b": {"c": 1, "d": 2}}}
			delete(m["a"]["b"], "c")
			m["a"]["b"]["d"]
		`, 2)
	})

	t.Run("Map多层嵌套delete后len", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": {"b": {"c": 1, "d": 2}}}
			delete(m["a"]["b"], "c")
			len(m["a"]["b"])
		`, 1)
	})

	t.Run("动态key构建Map", func(t *testing.T) {
		runIntTest(t, `
			m := {}
			prefix := "item"
			for i := 0; i < 3; i = i + 1 {
				key := prefix + "_" + string(i)
				m[key] = i
			}
			m["item_2"]
		`, 2)
	})

	t.Run("Map作为表达式的一部分", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 5}
			m["a"] + 10
		`, 15)
	})

	t.Run("Map值参与运算", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 3, "y": 4}
			m["x"] * m["y"]
		`, 12)
	})
}

// ========== 12. Map字面量key形式 ==========

// Test_Map_KeyForms Map中key的不同形式
func Test_Map_KeyForms(t *testing.T) {
	t.Run("字符串字面量key", func(t *testing.T) {
		runIntTest(t, `{"key": 1}["key"]`, 1)
	})

	t.Run("标识符key(动态求值)", func(t *testing.T) {
		runIntTest(t, `
			key := "dynamic"
			m := {key: 100}
			m[key]
		`, 100)
	})

	t.Run("标识符key求值为字符串", func(t *testing.T) {
		runStringTest(t, `
			name := "status"
			m := {name: "ok"}
			m["status"]
		`, "ok")
	})

	t.Run("数字字符串key访问", func(t *testing.T) {
		runStringTest(t, `
			m := {"0": "zero", "1": "one"}
			m["1"]
		`, "one")
	})

	t.Run("特殊字符key", func(t *testing.T) {
		runIntTest(t, `
			m := {"a.b.c": 42}
			m["a.b.c"]
		`, 42)
	})
}
