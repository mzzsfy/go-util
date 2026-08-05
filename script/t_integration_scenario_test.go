package script

import (
	"strings"
	"testing"
)

// ========== 复杂集成场景测试 ==========
// 模拟真实使用场景, 测试数据处理、算法实现、控制流、函数组合等综合行为

// replaceScriptTarget 替换脚本中的TARGET占位符为整数字面量
func replaceScriptTarget(script string, val int) string {
	return strings.ReplaceAll(script, "TARGET", itoa(val))
}

// replaceScriptWord 替换脚本中的指定单词
func replaceScriptWord(script, old, new string) string {
	return strings.ReplaceAll(script, old, new)
}

// itoa 简易整数转字符串, 避免引入strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n = n / 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

// ========== 数据处理场景 ==========

// Test_IntegrationScenario_ArraySum 数组求和
func Test_IntegrationScenario_ArraySum(t *testing.T) {
	t.Run("for-range累加", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30, 40, 50]
			sum := 0
			for v := range arr {
				sum = sum + v
			}
			sum
		`, 150)
	})
	t.Run("索引循环累加", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
			sum := 0
			for i := 0; i < len(arr); i = i + 1 {
				sum = sum + arr[i]
			}
			sum
		`, 55)
	})
	t.Run("空数组求和", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			sum := 0
			for v := range arr {
				sum = sum + v
			}
			sum
		`, 0)
	})
}

// Test_IntegrationScenario_ArrayAverage 数组平均值
func Test_IntegrationScenario_ArrayAverage(t *testing.T) {
	t.Run("整数平均", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30, 40, 50, 60]
			total := 0
			for v := range arr {
				total = total + v
			}
			total / len(arr)
		`, 35)
	})
	t.Run("函数封装平均值", func(t *testing.T) {
		runIntTest(t, `
			fn average(arr) {
				total := 0
				count := 0
				for v := range arr {
					total = total + v
					count = count + 1
				}
				if count == 0 {
					return 0
				}
				return total / count
			}
			average([100, 200, 300])
		`, 200)
	})
}

// Test_IntegrationScenario_ArrayMaxMin 数组最大最小值查找
func Test_IntegrationScenario_ArrayMaxMin(t *testing.T) {
	t.Run("最大值", func(t *testing.T) {
		runIntTest(t, `
			arr := [3, 7, 2, 9, 1, 8, 4]
			maxVal := arr[0]
			for i := 1; i < len(arr); i = i + 1 {
				if arr[i] > maxVal {
					maxVal = arr[i]
				}
			}
			maxVal
		`, 9)
	})
	t.Run("最小值", func(t *testing.T) {
		runIntTest(t, `
			arr := [3, 7, 2, 9, 1, 8, 4]
			minVal := arr[0]
			for i := 1; i < len(arr); i = i + 1 {
				if arr[i] < minVal {
					minVal = arr[i]
				}
			}
			minVal
		`, 1)
	})
	t.Run("最大值与最小值之和", func(t *testing.T) {
		runIntTest(t, `
			arr := [15, 22, 8, 41, 3, 19]
			maxVal := arr[0]
			minVal := arr[0]
			for i := 1; i < len(arr); i = i + 1 {
				if arr[i] > maxVal {
					maxVal = arr[i]
				}
				if arr[i] < minVal {
					minVal = arr[i]
				}
			}
			maxVal + minVal
		`, 44)
	})
}

// Test_IntegrationScenario_ArrayFilter 数组过滤
func Test_IntegrationScenario_ArrayFilter(t *testing.T) {
	t.Run("筛选偶数", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
			result := []
			for v := range arr {
				if v % 2 == 0 {
					result = push(result, v)
				}
			}
			len(result)
		`, 5)
	})
	t.Run("筛选大于阈值", func(t *testing.T) {
		runIntTest(t, `
			arr := [12, 5, 28, 3, 47, 19, 8]
			threshold := 10
			result := []
			for v := range arr {
				if v > threshold {
					result = push(result, v)
				}
			}
			result[0] + result[1] + result[2] + result[3]
		`, 106)
	})
}

// Test_IntegrationScenario_ArrayMapTransform 数组映射变换
func Test_IntegrationScenario_ArrayMapTransform(t *testing.T) {
	t.Run("每个元素乘以2", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			result := []
			for v := range arr {
				result = push(result, v * 2)
			}
			result[4]
		`, 10)
	})
	t.Run("平方变换", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			result := []
			for v := range arr {
				result = push(result, v * v)
			}
			result[0] + result[1] + result[2] + result[3] + result[4]
		`, 55)
	})
}

// Test_IntegrationScenario_ArrayReverse 数组反转
func Test_IntegrationScenario_ArrayReverse(t *testing.T) {
	t.Run("反转5元素数组", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			n := len(arr)
			result := []
			for i := 0; i < n; i = i + 1 {
				result = push(result, arr[n - 1 - i])
			}
			result[0]
		`, 5)
	})
	t.Run("反转后求和验证", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30, 40]
			n := len(arr)
			reversed := []
			for i := 0; i < n; i = i + 1 {
				reversed = push(reversed, arr[n - 1 - i])
			}
			reversed[0] + reversed[3]
		`, 50)
	})
}

// Test_IntegrationScenario_2DArrayTraversal 二维数组遍历
func Test_IntegrationScenario_2DArrayTraversal(t *testing.T) {
	t.Run("矩阵所有元素求和", func(t *testing.T) {
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
	t.Run("矩阵对角线之和", func(t *testing.T) {
		runIntTest(t, `
			matrix := [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
			diagSum := 0
			for i := 0; i < len(matrix); i = i + 1 {
				diagSum = diagSum + matrix[i][i]
			}
			diagSum
		`, 15)
	})
}

// ========== Map 操作场景 ==========

// Test_IntegrationScenario_MapCounter 计数器模式
func Test_IntegrationScenario_MapCounter(t *testing.T) {
	t.Run("统计字符出现次数", func(t *testing.T) {
		runIntTest(t, `
			s := "aabbccc"
			counts := {}
			for i := 0; i < len(s); i = i + 1 {
				ch := s[i]
				cur := counts[ch]
				if cur == nil {
					counts[ch] = 1
				} else {
					counts[ch] = cur + 1
				}
			}
			counts["c"]
		`, 3)
	})
	t.Run("统计后查询多个key", func(t *testing.T) {
		runIntTest(t, `
			s := "abracadabra"
			counts := {}
			for i := 0; i < len(s); i = i + 1 {
				ch := s[i]
				cur := counts[ch]
				if cur == nil {
					counts[ch] = 1
				} else {
					counts[ch] = cur + 1
				}
			}
			counts["a"] * 100 + counts["b"] * 10 + counts["r"]
		`, 522)
	})
}

// Test_IntegrationScenario_MapLookupTable 查找表模式
func Test_IntegrationScenario_MapLookupTable(t *testing.T) {
	t.Run("罗马数字查找表", func(t *testing.T) {
		runIntTest(t, `
			roman := {"I": 1, "V": 5, "X": 10, "L": 50, "C": 100}
			roman["X"] + roman["V"]
		`, 15)
	})
	t.Run("单词到数字映射", func(t *testing.T) {
		runIntTest(t, `
			lookup := {"one": 1, "two": 2, "three": 3, "four": 4, "five": 5}
			lookup["three"] * lookup["two"]
		`, 6)
	})
}

// Test_IntegrationScenario_MapGroupAggregation 分组聚合
func Test_IntegrationScenario_MapGroupAggregation(t *testing.T) {
	t.Run("按部门分组计数", func(t *testing.T) {
		runIntTest(t, `
			people := [
				{"dept": "eng", "name": "A"},
				{"dept": "eng", "name": "B"},
				{"dept": "sales", "name": "C"},
				{"dept": "eng", "name": "D"},
				{"dept": "sales", "name": "E"}
			]
			engCount := 0
			salesCount := 0
			for p := range people {
				if p["dept"] == "eng" {
					engCount = engCount + 1
				} else {
					salesCount = salesCount + 1
				}
			}
			engCount * 10 + salesCount
		`, 32)
	})
	t.Run("按分数段统计人数", func(t *testing.T) {
		runIntTest(t, `
			scores := [95, 82, 76, 61, 88, 55, 91, 73]
			high := 0
			mid := 0
			low := 0
			for s := range scores {
				if s >= 80 {
					high = high + 1
				} else if s >= 60 {
					mid = mid + 1
				} else {
					low = low + 1
				}
			}
			high * 100 + mid * 10 + low
		`, 431)
	})
}

// Test_IntegrationScenario_MapTraversalSum Map遍历累加
func Test_IntegrationScenario_MapTraversalSum(t *testing.T) {
	t.Run("遍历Map值求和", func(t *testing.T) {
		runIntTest(t, `
			m := {"a": 10, "b": 20, "c": 30}
			total := 0
			for k := m {
				total = total + m[k]
			}
			total
		`, 60)
	})
	t.Run("遍历Map计数", func(t *testing.T) {
		runIntTest(t, `
			m := {"x": 1, "y": 2, "z": 3, "w": 4, "v": 5}
			count := 0
			for k := m {
				count = count + 1
			}
			count
		`, 5)
	})
}

// Test_IntegrationScenario_NestedMapAccess 嵌套Map访问
func Test_IntegrationScenario_NestedMapAccess(t *testing.T) {
	t.Run("三层嵌套Map访问", func(t *testing.T) {
		runIntTest(t, `
			data := {"server": {"db": {"port": 3306}}}
			data["server"]["db"]["port"]
		`, 3306)
	})
	t.Run("嵌套Map含数组访问", func(t *testing.T) {
		runStringTest(t, `
			config := {"users": [{"name": "Alice", "role": "admin"}, {"name": "Bob", "role": "user"}]}
			config["users"][0]["name"]
		`, "Alice")
	})
}

// ========== 算法实现场景 ==========

// Test_IntegrationScenario_BubbleSort 冒泡排序
func Test_IntegrationScenario_BubbleSort(t *testing.T) {
	sortScript := `
		fn bubbleSort(a) {
			n := len(a)
			for i := 0; i < n; i = i + 1 {
				for j := 0; j < n - 1 - i; j = j + 1 {
					if a[j] > a[j + 1] {
						tmp := a[j]
						a[j] = a[j + 1]
						a[j + 1] = tmp
					}
				}
			}
			return a
		}
	`
	t.Run("排序后首尾元素", func(t *testing.T) {
		runIntTest(t, sortScript+`
			result := bubbleSort([5, 2, 8, 1, 9, 3])
			result[0] + result[5]
		`, 10)
	})
	t.Run("排序验证中间元素", func(t *testing.T) {
		runIntTest(t, sortScript+`
			sorted := bubbleSort([42, 17, 88, 3, 56, 29, 71])
			sorted[3]
		`, 42)
	})
}

// Test_IntegrationScenario_LinearSearch 线性搜索
func Test_IntegrationScenario_LinearSearch(t *testing.T) {
	t.Run("查找存在的元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [15, 22, 8, 41, 3, 19]
			target := 41
			index := -1
			for i := 0; i < len(arr); i = i + 1 {
				if arr[i] == target {
					index = i
					break
				}
			}
			index
		`, 3)
	})
	t.Run("查找不存在的元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [15, 22, 8, 41, 3, 19]
			target := 99
			index := -1
			for i := 0; i < len(arr); i = i + 1 {
				if arr[i] == target {
					index = i
					break
				}
			}
			index
		`, -1)
	})
}

// Test_IntegrationScenario_BinarySearch 二分搜索
func Test_IntegrationScenario_BinarySearch(t *testing.T) {
	binarySearchScript := `
		fn binarySearch(a, target) {
			left := 0
			right := len(a) - 1
			for left <= right {
				mid := (left + right) / 2
				if a[mid] == target {
					return mid
				}
				if a[mid] < target {
					left = mid + 1
				} else {
					right = mid - 1
				}
			}
			return -1
		}
		binarySearch([1, 3, 5, 7, 9, 11, 13], TARGET)
	`
	t.Run("查找中间元素", func(t *testing.T) {
		runIntTest(t, replaceScriptTarget(binarySearchScript, 7), 3)
	})
	t.Run("查找首元素", func(t *testing.T) {
		runIntTest(t, replaceScriptTarget(binarySearchScript, 1), 0)
	})
	t.Run("查找不存在的元素", func(t *testing.T) {
		runIntTest(t, replaceScriptTarget(binarySearchScript, 6), -1)
	})
}

// Test_IntegrationScenario_FibonacciRecursive 斐波那契递归
func Test_IntegrationScenario_FibonacciRecursive(t *testing.T) {
	fibScript := `
		fn fib(n) {
			if n <= 1 {
				return n
			}
			return fib(n - 1) + fib(n - 2)
		}
		fib(NUM)
	`
	t.Run("fib10", func(t *testing.T) {
		runIntTest(t, replaceScriptWord(fibScript, "NUM", "10"), 55)
	})
	t.Run("fib15", func(t *testing.T) {
		runIntTest(t, replaceScriptWord(fibScript, "NUM", "15"), 610)
	})
}

// Test_IntegrationScenario_FibonacciIterative 斐波那契迭代
func Test_IntegrationScenario_FibonacciIterative(t *testing.T) {
	t.Run("迭代fib10", func(t *testing.T) {
		runIntTest(t, `
			fn fibIter(n) {
				if n <= 1 {
					return n
				}
				a := 0
				b := 1
				for i := 2; i <= n; i = i + 1 {
					c := a + b
					a = b
					b = c
				}
				return b
			}
			fibIter(10)
		`, 55)
	})
	t.Run("迭代fib20", func(t *testing.T) {
		runIntTest(t, `
			fn fibIter(n) {
				if n <= 1 {
					return n
				}
				a := 0
				b := 1
				for i := 2; i <= n; i = i + 1 {
					c := a + b
					a = b
					b = c
				}
				return b
			}
			fibIter(20)
		`, 6765)
	})
}

// Test_IntegrationScenario_Factorial 阶乘
func Test_IntegrationScenario_Factorial(t *testing.T) {
	t.Run("递归阶乘", func(t *testing.T) {
		runIntTest(t, `
			fn factorial(n) {
				if n <= 1 {
					return 1
				}
				return n * factorial(n - 1)
			}
			factorial(5)
		`, 120)
	})
	t.Run("迭代阶乘", func(t *testing.T) {
		runIntTest(t, `
			fn factIter(n) {
				result := 1
				for i := 1; i <= n; i = i + 1 {
					result = result * i
				}
				return result
			}
			factIter(6)
		`, 720)
	})
}

// Test_IntegrationScenario_GCD 最大公约数
func Test_IntegrationScenario_GCD(t *testing.T) {
	t.Run("辗转相除法", func(t *testing.T) {
		runIntTest(t, `
			fn gcd(a, b) {
				for b != 0 {
					tmp := b
					b = a % b
					a = tmp
				}
				return a
			}
			gcd(48, 18)
		`, 6)
	})
	t.Run("递归GCD", func(t *testing.T) {
		runIntTest(t, `
			fn gcd(a, b) {
				if b == 0 {
					return a
				}
				return gcd(b, a % b)
			}
			gcd(100, 60)
		`, 20)
	})
}

// Test_IntegrationScenario_PrimeCheck 素数判定
func Test_IntegrationScenario_PrimeCheck(t *testing.T) {
	primeScript := `
		fn isPrime(n) {
			if n < 2 {
				return 0
			}
			if n == 2 {
				return 1
			}
			if n % 2 == 0 {
				return 0
			}
			i := 3
			for i * i <= n {
				if n % i == 0 {
					return 0
				}
				i = i + 2
			}
			return 1
		}
		isPrime(NUMBER)
	`
	t.Run("素数返回1", func(t *testing.T) {
		runIntTest(t, replaceScriptWord(primeScript, "NUMBER", "17"), 1)
	})
	t.Run("合数返回0", func(t *testing.T) {
		runIntTest(t, replaceScriptWord(primeScript, "NUMBER", "15"), 0)
	})
	t.Run("统计范围内素数个数", func(t *testing.T) {
		runIntTest(t, `
			fn isPrime(n) {
				if n < 2 {
					return 0
				}
				if n == 2 {
					return 1
				}
				if n % 2 == 0 {
					return 0
				}
				i := 3
				for i * i <= n {
					if n % i == 0 {
						return 0
					}
					i = i + 2
				}
				return 1
			}
			count := 0
			for n := 2; n <= 20; n = n + 1 {
				if isPrime(n) == 1 {
					count = count + 1
				}
			}
			count
		`, 8)
	})
}

// ========== 字符串处理场景 ==========

// Test_IntegrationScenario_StringConcatBuild 字符串拼接构建
func Test_IntegrationScenario_StringConcatBuild(t *testing.T) {
	t.Run("循环拼接数字", func(t *testing.T) {
		runStringTest(t, `
			result := ""
			for i := 1; i <= 5; i = i + 1 {
				result = result + string(i)
			}
			result
		`, "12345")
	})
	t.Run("条件拼接", func(t *testing.T) {
		runStringTest(t, `
			arr := [1, 2, 3, 4, 5]
			result := ""
			for v := range arr {
				if v % 2 == 0 {
					result = result + "E"
				} else {
					result = result + "O"
				}
			}
			result
		`, "OEOEO")
	})
}

// Test_IntegrationScenario_StringLengthCalc 字符串长度计算
func Test_IntegrationScenario_StringLengthCalc(t *testing.T) {
	t.Run("拼接后长度", func(t *testing.T) {
		runIntTest(t, `
			s := ""
			for i := 0; i < 10; i = i + 1 {
				s = s + "ab"
			}
			len(s)
		`, 20)
	})
	t.Run("数组中字符串长度求和", func(t *testing.T) {
		runIntTest(t, `
			arr := ["hello", "world", "foo", "bar"]
			total := 0
			for s := range arr {
				total = total + len(s)
			}
			total
		`, 16)
	})
}

// Test_IntegrationScenario_StringArrayJoin 字符串数组处理
func Test_IntegrationScenario_StringArrayJoin(t *testing.T) {
	t.Run("字符串数组拼接", func(t *testing.T) {
		runStringTest(t, `
			arr := ["Hello", " ", "World", "!"]
			result := ""
			for s := range arr {
				result = result + s
			}
			result
		`, "Hello World!")
	})
	t.Run("逗号分隔拼接", func(t *testing.T) {
		runStringTest(t, `
			arr := ["apple", "banana", "cherry"]
			result := ""
			for i := 0; i < len(arr); i = i + 1 {
				if i > 0 {
					result = result + ","
				}
				result = result + arr[i]
			}
			result
		`, "apple,banana,cherry")
	})
}

// Test_IntegrationScenario_StringLookupTable 字符串作为查找表
func Test_IntegrationScenario_StringLookupTable(t *testing.T) {
	t.Run("字符索引查表", func(t *testing.T) {
		runStringTest(t, `
			table := "0123456789ABCDEF"
			table[10]
		`, "A")
	})
	t.Run("字符串长度作验证", func(t *testing.T) {
		runIntTest(t, `
			digits := "0123456789"
			len(digits)
		`, 10)
	})
}

// ========== 控制流复杂场景 ==========

// Test_IntegrationScenario_MultiIfElseDecision 多重if-else if决策树
func Test_IntegrationScenario_MultiIfElseDecision(t *testing.T) {
	gradeScript := `
		score := SCORE
		grade := "F"
		if score >= 90 {
			grade = "A"
		} else if score >= 80 {
			grade = "B"
		} else if score >= 70 {
			grade = "C"
		} else if score >= 60 {
			grade = "D"
		} else {
			grade = "F"
		}
		grade
	`
	t.Run("A级", func(t *testing.T) {
		runStringTest(t, replaceScriptWord(gradeScript, "SCORE", "95"), "A")
	})
	t.Run("B级", func(t *testing.T) {
		runStringTest(t, replaceScriptWord(gradeScript, "SCORE", "85"), "B")
	})
	t.Run("C级", func(t *testing.T) {
		runStringTest(t, replaceScriptWord(gradeScript, "SCORE", "75"), "C")
	})
	t.Run("D级", func(t *testing.T) {
		runStringTest(t, replaceScriptWord(gradeScript, "SCORE", "65"), "D")
	})
	t.Run("F级", func(t *testing.T) {
		runStringTest(t, replaceScriptWord(gradeScript, "SCORE", "55"), "F")
	})
}

// Test_IntegrationScenario_NestedLoopMatrix 嵌套循环矩阵运算
func Test_IntegrationScenario_NestedLoopMatrix(t *testing.T) {
	t.Run("矩阵乘法之积的和", func(t *testing.T) {
		runIntTest(t, `
			a := [[1, 2], [3, 4]]
			b := [[5, 6], [7, 8]]
			total := 0
			for i := 0; i < 2; i = i + 1 {
				for j := 0; j < 2; j = j + 1 {
					for k := 0; k < 2; k = k + 1 {
						total = total + a[i][k] * b[k][j]
					}
				}
			}
			total
		`, 134)
	})
}

// Test_IntegrationScenario_LoopConditionalBreak 循环内条件break
func Test_IntegrationScenario_LoopConditionalBreak(t *testing.T) {
	t.Run("找到第一个满足条件的元素", func(t *testing.T) {
		runIntTest(t, `
			arr := [4, 7, 2, 9, 15, 3, 22]
			found := 0
			for v := range arr {
				if v > 10 {
					found = v
					break
				}
			}
			found
		`, 15)
	})
}

// Test_IntegrationScenario_LoopConditionalContinue 循环内条件continue
func Test_IntegrationScenario_LoopConditionalContinue(t *testing.T) {
	t.Run("跳过3和5的倍数求和", func(t *testing.T) {
		runIntTest(t, `
			sum := 0
			for i := 1; i <= 10; i = i + 1 {
				if i % 3 == 0 {
					continue
				}
				if i % 5 == 0 {
					continue
				}
				sum = sum + i
			}
			sum
		`, 22)
	})
}

// Test_IntegrationScenario_ComplexBreakContinue 复杂break/continue嵌套
func Test_IntegrationScenario_ComplexBreakContinue(t *testing.T) {
	t.Run("嵌套循环break只跳出内层", func(t *testing.T) {
		runIntTest(t, `
			total := 0
			for i := 0; i < 3; i = i + 1 {
				for j := 0; j < 5; j = j + 1 {
					if j == 2 {
						break
					}
					total = total + 1
				}
			}
			total
		`, 6)
	})
	t.Run("嵌套循环continue", func(t *testing.T) {
		runIntTest(t, `
			total := 0
			for i := 0; i < 3; i = i + 1 {
				for j := 0; j < 3; j = j + 1 {
					if j == 1 {
						continue
					}
					total = total + 1
				}
			}
			total
		`, 6)
	})
}

// ========== 函数组合场景 ==========

// Test_IntegrationScenario_FunctionCallChain 函数调用链
func Test_IntegrationScenario_FunctionCallChain(t *testing.T) {
	t.Run("A调用B调用C", func(t *testing.T) {
		runIntTest(t, `
			fn base(x) {
				return x * 3
			}
			fn middle(x) {
				return base(x) + 1
			}
			fn top(x) {
				return middle(x) + 1
			}
			top(10)
		`, 32)
	})
}

// Test_IntegrationScenario_RecursionConditionLoop 递归+条件+循环组合
func Test_IntegrationScenario_RecursionConditionLoop(t *testing.T) {
	t.Run("递归累加数组", func(t *testing.T) {
		runIntTest(t, `
			fn sumArr(arr, n) {
				if n <= 0 {
					return 0
				}
				return arr[n - 1] + sumArr(arr, n - 1)
			}
			sumArr([1, 2, 3, 4, 5], 5)
		`, 15)
	})
}

// Test_IntegrationScenario_FunctionReturnArrayIterate 函数返回数组后遍历
func Test_IntegrationScenario_FunctionReturnArrayIterate(t *testing.T) {
	t.Run("生成数组并求和", func(t *testing.T) {
		runIntTest(t, `
			fn genRange(n) {
				arr := []
				for i := 0; i < n; i = i + 1 {
					arr = push(arr, i * 2)
				}
				return arr
			}
			result := genRange(5)
			total := 0
			for v := range result {
				total = total + v
			}
			total
		`, 20)
	})
}

// Test_IntegrationScenario_FunctionDataProcessor 函数作为数据处理器
func Test_IntegrationScenario_FunctionDataProcessor(t *testing.T) {
	t.Run("过滤+变换+求和", func(t *testing.T) {
		runIntTest(t, `
			fn isEven(n) {
				return n % 2 == 0
			}
			fn square(n) {
				return n * n
			}
			data := [1, 2, 3, 4, 5, 6, 7, 8]
			total := 0
			for v := range data {
				if isEven(v) {
					total = total + square(v)
				}
			}
			total
		`, 120)
	})
}

// Test_IntegrationScenario_MultiFunctionCollaboration 多函数协作完成复杂任务
func Test_IntegrationScenario_MultiFunctionCollaboration(t *testing.T) {
	t.Run("多函数协作求统计量", func(t *testing.T) {
		runIntTest(t, `
			fn sumArr(arr) {
				s := 0
				for v := range arr {
					s = s + v
				}
				return s
			}
			fn mean(arr) {
				return sumArr(arr) / len(arr)
			}
			fn variance(arr) {
				m := mean(arr)
				total := 0
				for v := range arr {
					d := v - m
					total = total + d * d
				}
				return total / len(arr)
			}
			variance([2, 4, 6, 8])
		`, 5)
	})
}

// ========== 外部函数集成场景 ==========

// Test_IntegrationScenario_ExternalFuncDataProcessing Go函数处理脚本数据
func Test_IntegrationScenario_ExternalFuncDataProcessing(t *testing.T) {
	t.Run("脚本传数组给Go求和", func(t *testing.T) {
		result := runScriptWithFunc(t, `
			#fn sumArr(arr)=>int
			data := [10, 20, 30, 40, 50]
			sumArr(data)
		`, "sumArr", func(arr []int) int {
			total := 0
			for _, v := range arr {
				total += v
			}
			return total
		})
		assertInt(t, result, 150)
	})
}

// Test_IntegrationScenario_ExternalFuncArrayProcessing 脚本调用Go函数处理数组
func Test_IntegrationScenario_ExternalFuncArrayProcessing(t *testing.T) {
	t.Run("Go函数返回数组脚本处理", func(t *testing.T) {
		result := runScriptWithFunc(t, `
			#fn genNums()=>arr
			nums := genNums()
			total := 0
			for v := range nums {
				total = total + v
			}
			total
		`, "genNums", func() []int {
			return []int{5, 10, 15, 20}
		})
		assertInt(t, result, 50)
	})
}

// Test_IntegrationScenario_ExternalFuncRoundTrip 脚本传数据到Go再取回
func Test_IntegrationScenario_ExternalFuncRoundTrip(t *testing.T) {
	t.Run("数据往返Go处理", func(t *testing.T) {
		result := runScriptWithFunc(t, `
			#fn double(int)=>int
			result := []
			for i := 1; i <= 5; i = i + 1 {
				result = push(result, double(i))
			}
			result[0] + result[1] + result[2] + result[3] + result[4]
		`, "double", func(x int) int {
			return x * 2
		})
		assertInt(t, result, 30)
	})
}

// Test_IntegrationScenario_ExternalFuncOrchestration 外部函数实现核心逻辑脚本编排
func Test_IntegrationScenario_ExternalFuncOrchestration(t *testing.T) {
	t.Run("多外部函数编排", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("max", func(a, b int) int {
			if a > b {
				return a
			}
			return b
		})
		ctx.BindFunc("min", func(a, b int) int {
			if a < b {
				return a
			}
			return b
		})
		compiled, err := NewParser().Compile(`
			#fn max(int, int)=>int
			#fn min(int, int)=>int
			a := 10
			b := 20
			c := 15
			maxVal := max(a, b)
			maxVal = max(maxVal, c)
			minVal := min(a, b)
			minVal = min(minVal, c)
			maxVal + minVal
		`)
		assertNoError(t, err)
		result, err := NewEngine().Run(ctx, compiled)
		assertNoError(t, err)
		assertInt(t, result, 30)
	})
}

// ========== 状态管理场景 ==========

// Test_IntegrationScenario_AccumulatorPattern 累加器模式
func Test_IntegrationScenario_AccumulatorPattern(t *testing.T) {
	t.Run("多轮累加", func(t *testing.T) {
		runIntTest(t, `
			accumulator := 0
			for i := 1; i <= 100; i = i + 1 {
				accumulator = accumulator + i
			}
			accumulator
		`, 5050)
	})
}

// Test_IntegrationScenario_StateMachine 状态机模拟
func Test_IntegrationScenario_StateMachine(t *testing.T) {
	t.Run("事件驱动状态转换", func(t *testing.T) {
		runStringTest(t, `
			state := "idle"
			events := ["start", "work", "done", "start", "stop"]
			for e := range events {
				if state == "idle" {
					if e == "start" {
						state = "running"
					}
				} else if state == "running" {
					if e == "work" {
						state = "processing"
					}
					if e == "stop" {
						state = "idle"
					}
				} else if state == "processing" {
					if e == "done" {
						state = "running"
					}
				}
			}
			state
		`, "idle")
	})
}

// Test_IntegrationScenario_CounterIncrement 计数器递增
func Test_IntegrationScenario_CounterIncrement(t *testing.T) {
	t.Run("条件计数器", func(t *testing.T) {
		runIntTest(t, `
			data := [3, 7, 2, 8, 5, 9, 1, 6]
			above := 0
			below := 0
			threshold := 5
			for v := range data {
				if v > threshold {
					above = above + 1
				} else {
					below = below + 1
				}
			}
			above * 10 + below
		`, 44)
	})
}

// Test_IntegrationScenario_ConditionalStateSwitch 条件状态切换
func Test_IntegrationScenario_ConditionalStateSwitch(t *testing.T) {
	t.Run("开关切换模拟", func(t *testing.T) {
		runIntTest(t, `
			state := 0
			toggles := [1, 1, 1, 1, 1]
			for t := range toggles {
				if t == 1 {
					if state == 0 {
						state = 1
					} else {
						state = 0
					}
				}
			}
			state
		`, 1)
	})
}

// ========== 实用程序场景 ==========

// Test_IntegrationScenario_FizzBuzz FizzBuzz
func Test_IntegrationScenario_FizzBuzz(t *testing.T) {
	t.Run("前15项FizzBuzz", func(t *testing.T) {
		runStringTest(t, `
			result := ""
			for i := 1; i <= 15; i = i + 1 {
				if i % 15 == 0 {
					result = result + "FizzBuzz "
				} else if i % 3 == 0 {
					result = result + "Fizz "
				} else if i % 5 == 0 {
					result = result + "Buzz "
				} else {
					result = result + string(i) + " "
				}
			}
			result
		`, "1 2 Fizz 4 Buzz Fizz 7 8 Fizz Buzz 11 Fizz 13 14 FizzBuzz ")
	})
}

// Test_IntegrationScenario_MultiplicationTable 九九乘法表计算
func Test_IntegrationScenario_MultiplicationTable(t *testing.T) {
	t.Run("九九表总和", func(t *testing.T) {
		runIntTest(t, `
			total := 0
			for i := 1; i <= 9; i = i + 1 {
				for j := 1; j <= 9; j = j + 1 {
					total = total + i * j
				}
			}
			total
		`, 2025)
	})
}

// Test_IntegrationScenario_ArrayToString 数组转字符串
func Test_IntegrationScenario_ArrayToString(t *testing.T) {
	t.Run("整数数组转逗号分隔字符串", func(t *testing.T) {
		runStringTest(t, `
			arr := [1, 2, 3, 4, 5]
			result := ""
			for i := 0; i < len(arr); i = i + 1 {
				if i > 0 {
					result = result + ","
				}
				result = result + string(arr[i])
			}
			result
		`, "1,2,3,4,5")
	})
}

// Test_IntegrationScenario_SimpleCalculator 简易计算器
func Test_IntegrationScenario_SimpleCalculator(t *testing.T) {
	calcScript := `
		fn calc(op, a, b) {
			if op == "add" {
				return a + b
			}
			if op == "sub" {
				return a - b
			}
			if op == "mul" {
				return a * b
			}
			if op == "div" {
				return a / b
			}
			return 0
		}
		calc(OP, A, B)
	`
	t.Run("加法", func(t *testing.T) {
		s := strings.ReplaceAll(calcScript, "OP", `"add"`)
		s = strings.ReplaceAll(s, "A", "10")
		s = strings.ReplaceAll(s, "B", "20")
		runIntTest(t, s, 30)
	})
	t.Run("乘法", func(t *testing.T) {
		s := strings.ReplaceAll(calcScript, "OP", `"mul"`)
		s = strings.ReplaceAll(s, "A", "6")
		s = strings.ReplaceAll(s, "B", "7")
		runIntTest(t, s, 42)
	})
}

// Test_IntegrationScenario_DataValidator 数据验证器
func Test_IntegrationScenario_DataValidator(t *testing.T) {
	t.Run("有效数据", func(t *testing.T) {
		runIntTest(t, `
			fn validate(data) {
				if len(data) == 0 {
					return 0
				}
				age := data["age"]
				name := data["name"]
				if name == "" {
					return 0
				}
				if age < 0 {
					return 0
				}
				if age > 150 {
					return 0
				}
				return 1
			}
			validate({"name": "Alice", "age": 30})
		`, 1)
	})
	t.Run("无效年龄", func(t *testing.T) {
		runIntTest(t, `
			fn validate(data) {
				if len(data) == 0 {
					return 0
				}
				age := data["age"]
				if age < 0 {
					return 0
				}
				if age > 150 {
					return 0
				}
				return 1
			}
			validate({"name": "Bob", "age": -5})
		`, 0)
	})
}

// ========== 边界集成 ==========

// Test_IntegrationScenario_EmptyArrayProcessing 空数组处理流程
func Test_IntegrationScenario_EmptyArrayProcessing(t *testing.T) {
	t.Run("空数组安全求和", func(t *testing.T) {
		runIntTest(t, `
			fn safeSum(arr) {
				if len(arr) == 0 {
					return 0
				}
				total := 0
				for v := range arr {
					total = total + v
				}
				return total
			}
			safeSum([])
		`, 0)
	})
}

// Test_IntegrationScenario_EmptyMapProcessing 空 Map 处理流程
func Test_IntegrationScenario_EmptyMapProcessing(t *testing.T) {
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
}

// Test_IntegrationScenario_LargeDataLoop 大数据量循环
func Test_IntegrationScenario_LargeDataLoop(t *testing.T) {
	t.Run("1000次循环累加", func(t *testing.T) {
		runIntTest(t, `
			total := 0
			for i := 0; i < 1000; i = i + 1 {
				total = total + i
			}
			total
		`, 499500)
	})
}

// Test_IntegrationScenario_DeepRecursion 深层递归
func Test_IntegrationScenario_DeepRecursion(t *testing.T) {
	t.Run("递归深度50", func(t *testing.T) {
		runIntTest(t, `
			fn sumTo(n) {
				if n <= 0 {
					return 0
				}
				return n + sumTo(n - 1)
			}
			sumTo(50)
		`, 1275)
	})
	t.Run("递归深度100", func(t *testing.T) {
		runIntTest(t, `
			fn sumTo(n) {
				if n <= 0 {
					return 0
				}
				return n + sumTo(n - 1)
			}
			sumTo(100)
		`, 5050)
	})
}

// Test_IntegrationScenario_DeepNestedDataAccess 多层嵌套数据结构访问
func Test_IntegrationScenario_DeepNestedDataAccess(t *testing.T) {
	t.Run("4层嵌套Map访问", func(t *testing.T) {
		runIntTest(t, `
			data := {"l1": {"l2": {"l3": {"l4": 42}}}}
			data["l1"]["l2"]["l3"]["l4"]
		`, 42)
	})
	t.Run("数组Map交替嵌套", func(t *testing.T) {
		runStringTest(t, `
			data := [{"items": [{"name": "deep"}]}]
			data[0]["items"][0]["name"]
		`, "deep")
	})
}
