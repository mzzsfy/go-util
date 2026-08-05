package script

import "testing"

// ========== range 双变量遍历测试 ==========

// Test_RangeTwoVars 数组双变量range: for i, v := range arr
func Test_RangeTwoVars(t *testing.T) {
	t.Run("获取index和value", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for i, v := range arr {
				sum = sum + i*100 + v
			}
			sum
		`, 360) // 0+10 + 100+20 + 200+30 = 360
	})

	t.Run("index从零开始", func(t *testing.T) {
		runIntTest(t, `
			arr := [100, 200, 300]
			firstIdx := -1
			firstVal := -1
			for i, v := range arr {
				firstIdx = i
				firstVal = v
				break
			}
			firstIdx*1000 + firstVal
		`, 100) // i=0, v=100 -> 0*1000+100=100
	})

	t.Run("求和", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for i, v := range arr {
				sum = sum + v
			}
			sum
		`, 60)
	})

	t.Run("使用index累加", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for i, v := range arr {
				sum = sum + i
			}
			sum
		`, 3) // 0+1+2=3
	})

	t.Run("break正常工作", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			hit := 0
			for i, v := range arr {
				if v == 3 { break }
				hit = hit + 1
			}
			hit
		`, 2)
	})

	t.Run("continue正常工作", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3, 4, 5]
			sum := 0
			for i, v := range arr {
				if i == 1 { continue }
				sum = sum + v
			}
			sum
		`, 13) // 1+3+4+5=13, 跳过i=1时的v=2
	})

	t.Run("空数组不执行循环体", func(t *testing.T) {
		runIntTest(t, `
			arr := []
			hit := 0
			for i, v := range arr {
				hit = 1
			}
			hit
		`, 0)
	})

	t.Run("单元素数组", func(t *testing.T) {
		runIntTest(t, `
			arr := [42]
			idx := -1
			val := -1
			for i, v := range arr {
				idx = i
				val = v
			}
			idx*100 + val
		`, 42) // 0*100+42=42
	})

	t.Run("range关键字显式形式", func(t *testing.T) {
		runIntTest(t, `
			arr := [5, 10, 15]
			sum := 0
			for i, v := range arr {
				sum = sum + i + v
			}
			sum
		`, 33) // (0+5)+(1+10)+(2+15) = 5+11+17 = 33
	})

	t.Run("带括号形式", func(t *testing.T) {
		runIntTest(t, `
			arr := [1, 2, 3]
			sum := 0
			for (i, v := range arr) {
				sum = sum + i*10 + v
			}
			sum
		`, 36) // (0+1)+(10+2)+(20+3) = 1+12+23 = 36
	})
}

// Test_RangeSingleVarBackwardCompat 单变量range向后兼容
func Test_RangeSingleVarBackwardCompat(t *testing.T) {
	t.Run("单变量range仍正常工作", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for v := range arr {
				sum = sum + v
			}
			sum
		`, 60)
	})

	t.Run("隐式range仍正常工作", func(t *testing.T) {
		runIntTest(t, `
			arr := [10, 20, 30]
			sum := 0
			for v := arr {
				sum = sum + v
			}
			sum
		`, 60)
	})
}
