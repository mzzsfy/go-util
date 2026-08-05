package script

import "testing"

// ========== 数组高级操作测试 ==========
// 覆盖数组创建、索引、切片、赋值、遍历、拼接、push、函数交互、实用模式等全部细节

// ========== 1. 数组创建 ==========

// Test_Array_Create 空数组和基础创建
func Test_Array_Create(t *testing.T) {
	t.Run("空数组len为0", func(t *testing.T) {
		runIntTest(t, `len([])`, 0)
	})

	t.Run("单元素数组", func(t *testing.T) {
		runIntTest(t, `[42][0]`, 42)
	})

	t.Run("多元素数组首元素", func(t *testing.T) {
		runIntTest(t, `[1, 2, 3][0]`, 1)
	})

	t.Run("多元素数组尾元素", func(t *testing.T) {
		runIntTest(t, `[1, 2, 3][2]`, 3)
	})

	t.Run("多元素数组len", func(t *testing.T) {
		runIntTest(t, `len([10, 20, 30, 40, 50])`, 5)
	})
}

// Test_Array_Create_MixedTypes 混合类型数组
func Test_Array_Create_MixedTypes(t *testing.T) {
	t.Run("int元素", func(t *testing.T) {
		runIntTest(t, `[1, "a", true, nil, 3.14][0]`, 1)
	})

	t.Run("string元素", func(t *testing.T) {
		runStringTest(t, `[1, "a", true, nil, 3.14][1]`, "a")
	})

	t.Run("bool元素", func(t *testing.T) {
		runBoolTest(t, `[1, "a", true, nil, 3.14][2]`, true)
	})

	t.Run("nil元素", func(t *testing.T) {
		result := runScript(t, `[1, "a", true, nil, 3.14][3]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("float元素", func(t *testing.T) {
		runFloatTest(t, `[1, "a", true, nil, 3.14][4]`, 3.14)
	})
}

// Test_Array_Create_Nested 嵌套数组
func Test_Array_Create_Nested(t *testing.T) {
	t.Run("两层嵌套访问", func(t *testing.T) {
		runIntTest(t, `[[1, 2], [3, 4]][0][1]`, 2)
	})

	t.Run("两层嵌套第二行", func(t *testing.T) {
		runIntTest(t, `[[1, 2], [3, 4]][1][0]`, 3)
	})

	t.Run("三层嵌套访问", func(t *testing.T) {
		runIntTest(t, `[[[1, 2], [3, 4]], [[5, 6], [7, 8]]][1][0][1]`, 6)
	})

	t.Run("嵌套数组外层len", func(t *testing.T) {
		runIntTest(t, `len([[1, 2], [3, 4], [5, 6]])`, 3)
	})

	t.Run("嵌套数组内层len", func(t *testing.T) {
		runIntTest(t, `len([[1, 2, 3], [4, 5]][0])`, 3)
	})
}

// Test_Array_Create_WithMap 数组中包含Map
func Test_Array_Create_WithMap(t *testing.T) {
	t.Run("数组Map混合访问", func(t *testing.T) {
		runIntTest(t, `[{"a": 1}, {"b": 2}][0]["a"]`, 1)
	})

	t.Run("数组Map第二个元素", func(t *testing.T) {
		runIntTest(t, `[{"a": 1}, {"b": 2}][1]["b"]`, 2)
	})

	t.Run("Map数组混合访问", func(t *testing.T) {
		runIntTest(t, `{"key": [10, 20]}["key"][1]`, 20)
	})
}

// Test_Array_Create_WithExpression 数组中包含表达式
func Test_Array_Create_WithExpression(t *testing.T) {
	t.Run("加法表达式元素", func(t *testing.T) {
		runIntTest(t, `[1+1, 2+2, 3+3][2]`, 6)
	})

	t.Run("乘法表达式元素", func(t *testing.T) {
		runIntTest(t, `[2*3, 4*5, 6*7][1]`, 20)
	})

	t.Run("混合运算表达式", func(t *testing.T) {
		runIntTest(t, `[10-3, 20/4, 7%3][0]`, 7)
	})
}

// Test_Array_Create_WithVariable 数组引用变量
func Test_Array_Create_WithVariable(t *testing.T) {
	t.Run("变量作为元素", func(t *testing.T) {
		runIntTest(t, `
			x := 42
			arr := [x, x+1]
			arr[1]
		`, 43)
	})

	t.Run("多个变量作为元素", func(t *testing.T) {
		runIntTest(t, `
			a := 10
			b := 20
			c := 30
			arr := [a, b, c]
			arr[2]
		`, 30)
	})

	t.Run("变量运算后作为元素", func(t *testing.T) {
		runIntTest(t, `
			base := 5
			arr := [base*2, base*3, base*4]
			arr[1]
		`, 15)
	})
}

// ========== 2. 数组索引访问 ==========

// Test_Array_Index_Basic 基础索引访问
func Test_Array_Index_Basic(t *testing.T) {
	t.Run("首元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [100, 200, 300]
			arr[0]
		`, 100)
	})

	t.Run("中间元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [100, 200, 300]
			arr[1]
		`, 200)
	})

	t.Run("尾元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [100, 200, 300]
			arr[len(arr)-1]
		`, 300)
	})
}

// Test_Array_Index_Nested 嵌套索引访问
func Test_Array_Index_Nested(t *testing.T) {
	t.Run("嵌套数组索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [[10, 20], [30, 40]]
			arr[0][1]
		`, 20)
	})

	t.Run("三层嵌套索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [[[1, 2], [3, 4]], [[5, 6], [7, 8]]]
			arr[1][1][0]
		`, 7)
	})

	t.Run("数组Map混合索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [{"x": 1, "y": 2}, {"x": 3, "y": 4}]
			arr[1]["x"]
		`, 3)
	})

	t.Run("Map数组混合索引", func(t *testing.T) {
		runIntTest(t, `
			m := {"data": [100, 200, 300]}
			m["data"][2]
		`, 300)
	})
}

// Test_Array_Index_Dynamic 动态索引
func Test_Array_Index_Dynamic(t *testing.T) {
	t.Run("表达式索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30, 40]
			arr[1+1]
		`, 30)
	})

	t.Run("变量索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30, 40]
			i := 2
			arr[i]
		`, 30)
	})

	t.Run("运算表达式索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30, 40, 50]
			n := 1
			arr[n*2]
		`, 30)
	})
}

// ========== 3. 数组索引越界 ==========

// Test_Array_Index_OutOfBounds 索引越界返回nil
func Test_Array_Index_OutOfBounds(t *testing.T) {
	t.Run("超大索引返回nil", func(t *testing.T) {
		result := runScript(t, `[1, 2, 3][100]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("刚好越界返回nil", func(t *testing.T) {
		result := runScript(t, `[1, 2, 3][3]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})
}

// Test_Array_Index_Negative 负索引返回nil
func Test_Array_Index_Negative(t *testing.T) {
	t.Run("负一索引返回nil", func(t *testing.T) {
		result := runScript(t, `[1, 2, 3][-1]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("大负数索引返回nil", func(t *testing.T) {
		result := runScript(t, `[1, 2, 3][-100]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})
}

// Test_Array_Index_EmptyArray 空数组访问返回nil
func Test_Array_Index_EmptyArray(t *testing.T) {
	t.Run("空数组索引零", func(t *testing.T) {
		result := runScript(t, `[][0]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})

	t.Run("空数组任意索引", func(t *testing.T) {
		result := runScript(t, `[][5]`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})
}

// ========== 4. 数组切片 ==========

// Test_Array_Slice_Basic 基础切片
func Test_Array_Slice_Basic(t *testing.T) {
	t.Run("正常切片", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sub := arr[1:3]
			sub[0]
		`, 2)
	})

	t.Run("切片第二个元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sub := arr[1:4]
			sub[2]
		`, 4)
	})

	t.Run("切片长度", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5, 6]
			sub := arr[1:5]
			len(sub)
		`, 4)
	})
}

// Test_Array_Slice_Boundary 切片边界
func Test_Array_Slice_Boundary(t *testing.T) {
	t.Run("从头切片", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sub := arr[0:2]
			sub[1]
		`, 2)
	})

	t.Run("到尾切片", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sub := arr[2:len(arr)]
			sub[len(sub)-1]
		`, 5)
	})

	t.Run("全切片", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sub := arr[0:len(arr)]
			len(sub)
		`, 3)
	})
}

// Test_Array_Slice_Special 特殊切片
func Test_Array_Slice_Special(t *testing.T) {
	t.Run("空切片", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			sub := arr[1:1]
			len(sub)
		`, 0)
	})

	t.Run("嵌套切片后索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sub := arr[1:4]
			sub[1]
		`, 3)
	})

	t.Run("多层切片", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5, 6, 7, 8]
			s1 := arr[2:7]
			s2 := s1[1:4]
			s2[0]
		`, 4)
	})

	t.Run("负索引切片规范化", func(t *testing.T) {
		// 引擎行为: 负索引切片从末尾倒数, -1+5=4, [4:3]为空
		runIntTest(t, `len([1, 2, 3, 4, 5][-1:3])`, 0)
	})
}

// ========== 5. 数组赋值 ==========

// Test_Array_Assign_Basic 基础数组赋值
func Test_Array_Assign_Basic(t *testing.T) {
	t.Run("赋值已有索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr[0] = 99
			arr[0]
		`, 99)
	})

	t.Run("赋值中间索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr[1] = 50
			arr[1]
		`, 50)
	})

	t.Run("赋值后其他元素不变", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr[1] = 99
			arr[0] + arr[2]
		`, 4)
	})

	t.Run("多次赋值", func(t *testing.T) {
		runIntTest(t, `
			arr := [0, 0, 0]
			arr[0] = 10
			arr[1] = 20
			arr[2] = 30
			arr[0] + arr[1] + arr[2]
		`, 60)
	})
}

// Test_Array_Assign_Nested 嵌套数组赋值
func Test_Array_Assign_Nested(t *testing.T) {
	t.Run("嵌套数组赋值", func(t *testing.T) {
		runIntTest(t, `
			arr := [[1, 2], [3, 4]]
			arr[0][1] = 99
			arr[0][1]
		`, 99)
	})

	t.Run("嵌套数组赋值不影响同行", func(t *testing.T) {
		runIntTest(t, `
			arr := [[1, 2], [3, 4]]
			arr[0][1] = 99
			arr[0][0]
		`, 1)
	})

	t.Run("数组Map混合赋值", func(t *testing.T) {
		runIntTest(t, `
			arr := [{"k": 1}]
			arr[0]["k"] = 99
			arr[0]["k"]
		`, 99)
	})

	t.Run("Map数组混合赋值", func(t *testing.T) {
		runIntTest(t, `
			m := {"key": [1, 2, 3]}
			m["key"][0] = 99
			m["key"][0]
		`, 99)
	})
}

// Test_Array_Assign_OutOfBounds 赋值越界为运行时错误
func Test_Array_Assign_OutOfBounds(t *testing.T) {
	t.Run("赋值超出长度报错", func(t *testing.T) {
		runRuntimeErrorTest(t, `
			arr := [1, 2, 3]
			arr[5] = 99
		`)
	})

	t.Run("赋值负索引报错", func(t *testing.T) {
		runRuntimeErrorTest(t, `
			arr := [1, 2, 3]
			arr[-1] = 99
		`)
	})
}

// ========== 6. 数组长度 ==========

// Test_Array_Len 数组长度测试
func Test_Array_Len(t *testing.T) {
	t.Run("空数组len为0", func(t *testing.T) {
		runIntTest(t, `len([])`, 0)
	})

	t.Run("多元素数组len", func(t *testing.T) {
		runIntTest(t, `len([1, 2, 3, 4, 5])`, 5)
	})

	t.Run("嵌套数组外层len", func(t *testing.T) {
		runIntTest(t, `len([[1, 2], [3, 4], [5, 6]])`, 3)
	})

	t.Run("push后len变化", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2]
			arr = push(arr, 3)
			arr = push(arr, 4)
			len(arr)
		`, 4)
	})
}

// ========== 7. 数组遍历 ==========

// Test_Array_Range_SingleVar 单变量range遍历
func Test_Array_Range_SingleVar(t *testing.T) {
	t.Run("单变量遍历求和", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for v := range arr {
				sum = sum + v
			}
			sum
		`, 60)
	})

	t.Run("单变量遍历计数", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			count := 0
			for v := range arr {
				count = count + 1
			}
			count
		`, 5)
	})

	t.Run("单变量遍历偶数累加", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5, 6]
			total := 0
			for v := range arr {
				if v % 2 == 0 {
					total = total + v
				}
			}
			total
		`, 12)
	})
}

// Test_Array_Range_TwoVars 双变量range遍历
func Test_Array_Range_TwoVars(t *testing.T) {
	t.Run("获取索引和值", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for i, v := range arr {
				sum = sum + i*100 + v
			}
			sum
		`, 360)
	})

	t.Run("只取索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for i, v := range arr {
				sum = sum + i
			}
			sum
		`, 3)
	})

	t.Run("使用索引和值构建新值", func(t *testing.T) {
		// arr = [3, 1, 4, 1, 5]: 0*3+1*1+2*4+3*1+4*5 = 0+1+8+3+20 = 32
		runIntTest(t, `
			arr := [3, 1, 4, 1, 5]
			result := 0
			for i, v := range arr {
				result = result + i * v
			}
			result
		`, 32)
	})
}

// Test_Array_Range_Modify 遍历时修改元素
func Test_Array_Range_Modify(t *testing.T) {
	t.Run("索引循环修改元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			for i := 0; i < len(arr); i = i + 1 {
				arr[i] = arr[i] * 2
			}
			arr[2]
		`, 6)
	})

	t.Run("索引循环修改后求和", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			for i := 0; i < len(arr); i = i + 1 {
				arr[i] = arr[i] + 10
			}
			arr[0] + arr[1] + arr[2]
		`, 36)
	})
}

// Test_Array_Range_Nested 遍历嵌套数组
func Test_Array_Range_Nested(t *testing.T) {
	t.Run("索引循环遍历二维数组", func(t *testing.T) {
		runIntTest(t, `
			matrix := [[1, 2], [3, 4]]
			total := 0
			for i := 0; i < len(matrix); i = i + 1 {
				row := matrix[i]
				for j := 0; j < len(row); j = j + 1 {
					total = total + row[j]
				}
			}
			total
		`, 10)
	})

	t.Run("二维数组每行求和", func(t *testing.T) {
		runIntTest(t, `
			matrix := [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
			rowSum := 0
			for i := 0; i < len(matrix); i = i + 1 {
				row := matrix[i]
				for j := 0; j < len(row); j = j + 1 {
					rowSum = rowSum + row[j]
				}
			}
			rowSum
		`, 45)
	})
}

// Test_Array_Range_BreakContinue break和continue
func Test_Array_Range_BreakContinue(t *testing.T) {
	t.Run("break中断遍历", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sum := 0
			for v := range arr {
				if v == 3 {
					break
				}
				sum = sum + v
			}
			sum
		`, 3)
	})

	t.Run("continue跳过元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sum := 0
			for v := range arr {
				if v == 3 {
					continue
				}
				sum = sum + v
			}
			sum
		`, 12)
	})

	t.Run("break在双变量range中", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			hit := 0
			for i, v := range arr {
				if v == 3 {
					break
				}
				hit = hit + 1
			}
			hit
		`, 2)
	})
}

// ========== 8. 数组拼接 ==========

// Test_Array_Concat 数组拼接
func Test_Array_Concat(t *testing.T) {
	t.Run("拼接两个数组", func(t *testing.T) {
		runIntTest(t, `
			a := [1, 2]
			b := [3, 4]
			c := a + b
			len(c)
		`, 4)
	})

	t.Run("拼接后访问元素", func(t *testing.T) {
		runIntTest(t, `
			a := [1, 2]
			b := [3, 4]
			c := a + b
			c[2]
		`, 3)
	})

	t.Run("拼接多个数组", func(t *testing.T) {
		runIntTest(t, `
			a := [1, 2]
			b := [3, 4]
			c := [5, 6]
			result := a + b + c
			len(result)
		`, 6)
	})

	t.Run("空数组拼接", func(t *testing.T) {
		runIntTest(t, `len([] + [])`, 0)
	})

	t.Run("与非空数组拼接", func(t *testing.T) {
		runIntTest(t, `len([] + [1, 2, 3])`, 3)
	})

	t.Run("字面量直接拼接", func(t *testing.T) {
		runIntTest(t, `([1, 2] + [3, 4])[0]`, 1)
	})
}

// ========== 9. push操作 ==========

// Test_Array_Push push基本操作
func Test_Array_Push(t *testing.T) {
	t.Run("push单个元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr2 := push(arr, 4)
			arr2[3]
		`, 4)
	})

	t.Run("push后len增加", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2]
			arr = push(arr, 3)
			arr = push(arr, 4)
			len(arr)
		`, 4)
	})
}

// Test_Array_Push_NoModify push不修改原数组
func Test_Array_Push_NoModify(t *testing.T) {
	t.Run("原数组len不变", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr2 := push(arr, 4)
			len(arr)
		`, 3)
	})

	t.Run("新数组len正确", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			arr2 := push(arr, 4)
			len(arr2)
		`, 4)
	})

	t.Run("原数组内容不变", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			arr2 := push(arr, 40)
			arr[0] + arr[1] + arr[2]
		`, 60)
	})
}

// Test_Array_Push_EdgeCases push边界情况
func Test_Array_Push_EdgeCases(t *testing.T) {
	t.Run("push到空数组", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			arr = push(arr, 42)
			arr[0]
		`, 42)
	})

	t.Run("连续push构建数组", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			arr = push(arr, 1)
			arr = push(arr, 2)
			arr = push(arr, 3)
			arr[0] + arr[1] + arr[2]
		`, 6)
	})

	t.Run("嵌套push构建数组", func(t *testing.T) {
		runIntTest(t, `
			arr := push(push(push([], 1), 2), 3)
			arr[2]
		`, 3)
	})

	t.Run("push不同类型元素", func(t *testing.T) {
		runIntTest(t, `
			arr := push(push([], 1), "hello")
			len(arr)
		`, 2)
	})
}

// ========== 10. 数组与函数 ==========

// Test_Array_Func_Param 数组作为函数参数
func Test_Array_Func_Param(t *testing.T) {
	t.Run("函数读取数组元素", func(t *testing.T) {
		runIntTest(t, `
			fn first(a) {
				return a[0]
			}
			first([100, 200, 300])
		`, 100)
	})

	t.Run("函数内修改数组元素", func(t *testing.T) {
		runIntTest(t, `
			fn modify(a) {
				a[0] = 99
				return a[0]
			}
			arr := [1, 2, 3]
			modify(arr)
		`, 99)
	})
}

// Test_Array_Func_Return 函数返回数组
func Test_Array_Func_Return(t *testing.T) {
	t.Run("函数返回字面量数组", func(t *testing.T) {
		runIntTest(t, `
			fn makeArr() {
				return [10, 20, 30]
			}
			makeArr()[1]
		`, 20)
	})

	t.Run("函数返回push构建的数组", func(t *testing.T) {
		runIntTest(t, `
			fn build(n) {
				result := []
				for i := 0; i < n; i = i + 1 {
					result = push(result, i * i)
				}
				return result
			}
			arr := build(5)
			arr[3]
		`, 9)
	})

	t.Run("函数返回数组后访问", func(t *testing.T) {
		runIntTest(t, `
			fn range123() {
				return [1, 2, 3]
			}
			arr := range123()
			arr[2]
		`, 3)
	})
}

// Test_Array_Func_Push 函数内push操作
func Test_Array_Func_Push(t *testing.T) {
	t.Run("函数内push返回新数组", func(t *testing.T) {
		runIntTest(t, `
			fn addOne(a) {
				return push(a, 99)
			}
			arr := [1, 2]
			result := addOne(arr)
			result[2]
		`, 99)
	})

	t.Run("函数内连续push", func(t *testing.T) {
		runIntTest(t, `
			fn appendThree(a) {
				a = push(a, 1)
				a = push(a, 2)
				a = push(a, 3)
				return a
			}
			arr := [0]
			result := appendThree(arr)
			len(result)
		`, 4)
	})
}

// ========== 11. 数组实用模式 ==========

// Test_Array_Pattern_Aggregate 聚合操作
func Test_Array_Pattern_Aggregate(t *testing.T) {
	t.Run("查找最大值", func(t *testing.T) {
		runIntTest(t, `
			arr := [3, 7, 2, 9, 1, 5]
			max := arr[0]
			for v := range arr {
				if v > max {
					max = v
				}
			}
			max
		`, 9)
	})

	t.Run("查找最小值", func(t *testing.T) {
		runIntTest(t, `
			arr := [5, 3, 8, 1, 9, 2]
			min := arr[0]
			for v := range arr {
				if v < min {
					min = v
				}
			}
			min
		`, 1)
	})

	t.Run("计算总和", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sum := 0
			for v := range arr {
				sum = sum + v
			}
			sum
		`, 15)
	})

	t.Run("计算平均值", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for v := range arr {
				sum = sum + v
			}
			sum / len(arr)
		`, 20)
	})
}

// Test_Array_Pattern_FilterMap 过滤和映射
func Test_Array_Pattern_FilterMap(t *testing.T) {
	t.Run("过滤偶数", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5, 6]
			result := []
			for v := range arr {
				if v % 2 == 0 {
					result = push(result, v)
				}
			}
			len(result)
		`, 3)
	})

	t.Run("映射为平方", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4]
			result := []
			for v := range arr {
				result = push(result, v * v)
			}
			result[3]
		`, 16)
	})
}

// Test_Array_Pattern_Reverse 反转数组
func Test_Array_Pattern_Reverse(t *testing.T) {
	t.Run("反转后首元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			reversed := []
			for i := 0; i < len(arr); i = i + 1 {
				reversed = push(reversed, arr[len(arr)-1-i])
			}
			reversed[0]
		`, 5)
	})

	t.Run("反转后len一致", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			reversed := []
			for i := 0; i < len(arr); i = i + 1 {
				reversed = push(reversed, arr[len(arr)-1-i])
			}
			len(reversed)
		`, 5)
	})

	t.Run("反转后中间元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			reversed := []
			for i := 0; i < len(arr); i = i + 1 {
				reversed = push(reversed, arr[len(arr)-1-i])
			}
			reversed[2]
		`, 3)
	})
}

// Test_Array_Pattern_FindIndex 查找元素位置
func Test_Array_Pattern_FindIndex(t *testing.T) {
	t.Run("查找存在的元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30, 40]
			found := -1
			for i := 0; i < len(arr); i = i + 1 {
				if arr[i] == 30 {
					found = i
					break
				}
			}
			found
		`, 2)
	})

	t.Run("查找不存在的元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30, 40]
			found := -1
			for i := 0; i < len(arr); i = i + 1 {
				if arr[i] == 99 {
					found = i
					break
				}
			}
			found
		`, -1)
	})
}

// ========== 12. 数组比较 ==========

// Test_Array_Compare 数组相等比较
func Test_Array_Compare(t *testing.T) {
	t.Run("相同数组相等", func(t *testing.T) {
		runBoolTest(t, `[1, 2, 3] == [1, 2, 3]`, true)
	})

	t.Run("不同数组不相等", func(t *testing.T) {
		runBoolTest(t, `[1, 2, 3] == [1, 2, 4]`, false)
	})

	t.Run("空数组比较", func(t *testing.T) {
		runBoolTest(t, `[] == []`, true)
	})

	t.Run("不同长度不相等", func(t *testing.T) {
		runBoolTest(t, `[1, 2] == [1, 2, 3]`, false)
	})

	t.Run("嵌套数组相等", func(t *testing.T) {
		runBoolTest(t, `[[1, 2], [3, 4]] == [[1, 2], [3, 4]]`, true)
	})

	t.Run("嵌套数组不等", func(t *testing.T) {
		runBoolTest(t, `[[1, 2], [3, 4]] == [[1, 2], [3, 5]]`, false)
	})

	t.Run("不等比较", func(t *testing.T) {
		runBoolTest(t, `[1, 2] != [3, 4]`, true)
	})
}

// ========== 13. 数组与字符串交互 ==========

// Test_Array_String 字符串数组操作
func Test_Array_String(t *testing.T) {
	t.Run("字符串数组访问", func(t *testing.T) {
		runStringTest(t, `["alpha", "beta", "gamma"][1]`, "beta")
	})

	t.Run("字符串数组len", func(t *testing.T) {
		runIntTest(t, `len(["a", "b", "c", "d"])`, 4)
	})

	t.Run("字符串数组遍历拼接", func(t *testing.T) {
		runStringTest(t, `
			arr := ["x", "y", "z"]
			result := ""
			for v := range arr {
				result = result + v
			}
			result
		`, "xyz")
	})

	t.Run("数组转字符串", func(t *testing.T) {
		runStringTest(t, `string([1, 2, 3])`, "[1, 2, 3]")
	})
}

// ========== 14. 更多遍历模式 ==========

// Test_Array_Range_Empty 空数组遍历
func Test_Array_Range_Empty(t *testing.T) {
	t.Run("空数组range不执行循环体", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			hit := 0
			for v := range arr {
				hit = 1
			}
			hit
		`, 0)
	})

	t.Run("空数组双变量range不执行", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			hit := 0
			for i, v := range arr {
				hit = 1
			}
			hit
		`, 0)
	})
}

// Test_Array_Range_SingleElement 单元素数组遍历
func Test_Array_Range_SingleElement(t *testing.T) {
	t.Run("单元素遍历获取值", func(t *testing.T) {
		runIntTest(t, `
			arr := [42]
			result := 0
			for v := range arr {
				result = v
			}
			result
		`, 42)
	})

	t.Run("单元素双变量获取索引", func(t *testing.T) {
		runIntTest(t, `
			arr := [42]
			idx := -1
			for i, v := range arr {
				idx = i
			}
			idx
		`, 0)
	})
}

// Test_Array_BuildInLoop 循环构建数组
func Test_Array_BuildInLoop(t *testing.T) {
	t.Run("循环push构建数组", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			for i := 0; i < 5; i = i + 1 {
				arr = push(arr, i * i)
			}
			arr[4]
		`, 16)
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

// ========== 15. 更多实用模式 ==========

// Test_Array_Pattern_Count 计数模式
func Test_Array_Pattern_Count(t *testing.T) {
	t.Run("计数满足条件的元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5, 6, 7, 8]
			count := 0
			for v := range arr {
				if v > 4 {
					count = count + 1
				}
			}
			count
		`, 4)
	})

	t.Run("计数偶数", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
			count := 0
			for v := range arr {
				if v % 2 == 0 {
					count = count + 1
				}
			}
			count
		`, 5)
	})
}

// Test_Array_Pattern_Contains 存在性检查
func Test_Array_Pattern_Contains(t *testing.T) {
	t.Run("存在检查返回true", func(t *testing.T) {
		runBoolTest(t, `
			arr := [1, 3, 5, 7, 9]
			found := false
			for v := range arr {
				if v == 5 {
					found = true
					break
				}
			}
			found
		`, true)
	})

	t.Run("不存在检查返回false", func(t *testing.T) {
		runBoolTest(t, `
			arr := [1, 3, 5, 7, 9]
			found := false
			for v := range arr {
				if v == 4 {
					found = true
					break
				}
			}
			found
		`, false)
	})
}

// Test_Array_Concat_PreservesOriginal 拼接不修改原数组
func Test_Array_Concat_PreservesOriginal(t *testing.T) {
	t.Run("拼接后原数组len不变", func(t *testing.T) {
		runIntTest(t, `
			a := [1, 2]
			b := [3, 4]
			c := a + b
			len(a)
		`, 2)
	})

	t.Run("拼接后原数组内容不变", func(t *testing.T) {
		runIntTest(t, `
			a := [10, 20]
			b := [30, 40]
			c := a + b
			a[0] + a[1]
		`, 30)
	})
}

// Test_Array_Slice_OutOfBounds 越界切片行为
func Test_Array_Slice_OutOfBounds(t *testing.T) {
	t.Run("end超出长度返回有效部分", func(t *testing.T) {
		// 引擎行为: 超出范围的end被截断到数组长度
		runIntTest(t, `
			arr := [1, 2, 3]
			sub := arr[0:10]
			len(sub)
		`, 3)
	})

	t.Run("start超出长度返回空", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			sub := arr[5:10]
			len(sub)
		`, 0)
	})
}

// Test_Array_Push_MixedTypes push不同类型元素
func Test_Array_Push_MixedTypes(t *testing.T) {
	t.Run("push字符串到int数组", func(t *testing.T) {
		runStringTest(t, `
			arr := [1, 2]
			arr = push(arr, "hello")
			arr[2]
		`, "hello")
	})

	t.Run("push布尔值", func(t *testing.T) {
		runBoolTest(t, `
			arr := [1]
			arr = push(arr, true)
			arr[1]
		`, true)
	})

	t.Run("push数组(嵌套)", func(t *testing.T) {
		runIntTest(t, `
			arr := [1]
			arr = push(arr, [2, 3])
			arr[1][1]
		`, 3)
	})

	t.Run("push nil", func(t *testing.T) {
		result := runScript(t, `
			arr := [1]
			arr = push(arr, nil)
			arr[1]
		`)
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})
}
