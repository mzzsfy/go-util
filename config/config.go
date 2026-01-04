package config

import (
    "fmt"
    "os"
    "reflect"
    "sort"
    "strconv"
    "strings"
)

var (
    // Parser 需要手动注册解析工具
    Parser          = map[string]func(data []byte) map[string]any{}
    FileTypeHandler = []func(string) string{
        func(name string) string {
            if strings.HasSuffix(name, ".yml") {
                return "yaml"
            }
            i := strings.LastIndexByte(name, '.')
            if i < 0 {
                return ""
            }
            return name[i+1:]
        },
    }
)

type File struct {
    Data  []byte
    Name  string
    Order int
}

// ParseConfigs2Map 解析多个配置文件为并合并为单个map
func ParseConfigs2Map(files ...*File) (map[string]any, error) {
    ds := ParseConfigs(files)
    l := len(ds)
    if l == 0 {
        return nil, fmt.Errorf("no config files provided")
    }

    // 先将所有配置平铺，然后合并，再解析占位符
    tiledMaps := make([]map[string]any, l)
    for i, m := range ds {
        tiledMaps[i] = TilingMap(m)
    }

    // 合并平铺后的配置
    for i := 1; i < l; i++ {
        // 合并平铺后的map
        for k, v := range tiledMaps[i-1] {
            if _, exists := tiledMaps[i][k]; !exists {
                tiledMaps[i][k] = v
            }
        }
    }

    // 在最终平铺的结构上解析占位符
    resolved := ResolveMap(tiledMaps[l-1])

    // 转换回嵌套结构
    return UntilingMap(resolved), nil
}

// ParseConfigs 解析多个配置文件
func ParseConfigs(files []*File) []map[string]any {
    sort.Slice(files, func(i, j int) bool {
        return files[i].Order > files[j].Order
    })
    r := make([]map[string]any, 0, len(files))
    for _, item := range files {
        r = append(r, ParseConfig(item))
    }
    return r
}

// ParseConfig 解析单个配置文件
func ParseConfig(file *File) map[string]any {
    f := Parser[fileType(file.Name)]
    if f == nil {
        return nil
    }
    return f(file.Data)
}

func fileType(name string) string {
    for _, f := range FileTypeHandler {
        s := f(name)
        if s != "" {
            return s
        }
    }
    panic("无法解析文件类型:" + name)
}

// ResolveMap 处理${xxx:default}占位符,无法处理的保留原始值,只支持单层结构的map,使用 TilingMap 转换为单层结构
func ResolveMap(data map[string]any) map[string]any {
    // 需要多次解析，直到没有更多变化
    maxIterations := 10
    for iteration := 0; iteration < maxIterations; iteration++ {
        changed := false
        for k, v := range data {
            vv := reflect.Indirect(reflect.ValueOf(v))
            if vv.Kind() == reflect.String {
                s := vv.Interface().(string)
                if strings.Contains(s, "${") {
                    resolvedValue := resolveStringFromMap(s, data)
                    // 检查是否实际发生了变化
                    if resolvedValue != s {
                        data[k] = resolvedValue
                        changed = true
                    }
                }
            }
        }
        // 如果没有变化，可以提前退出
        if !changed {
            break
        }
    }
    return data
}
func resolveStringFromMap(s string, m map[string]any) any {
    // 处理嵌套的 ${} 占位符
    // 使用计数器防止无限循环
    maxIterations := 50
    iteration := 0

    for {
        if iteration >= maxIterations {
            // 防止无限循环
            break
        }

        startIndex := strings.Index(s, "${")
        if startIndex == -1 {
            break
        }

        endIndex := startIndex + strings.Index(s[startIndex:], "}")
        if endIndex < 0 {
            break // 没有结束括号
        }

        // 提取占位符内容
        placeholder := s[startIndex+2 : endIndex]
        before, after, found := strings.Cut(placeholder, ":")

        var value any
        if found {
            // 有默认值
            // 避免递归引用：如果键的值是包含自身的占位符，则使用默认值
            currentValue, exists := m[before]
            if exists && strings.Contains(fmt.Sprintf("%v", currentValue), "${"+before) {
                // 如果是递归引用，使用默认值
                value = after
            } else {
                value = m[before]
                if value == nil || value == "" {
                    value = after
                }
            }
        } else {
            // 没有默认值
            value = m[before]
        }

        if value == nil {
            value = ""
        }

        // 将值转换为字符串
        var valueStr string
        if s1, ok := value.(string); ok {
            valueStr = s1
        } else {
            valueStr = fmt.Sprint(value)
        }

        // 替换占位符
        s = s[:startIndex] + valueStr + s[endIndex+1:]
        iteration++
    }

    return s
}

// getOriginalKey 用于检测递归引用
func getOriginalKey(originalS, currentS, key string) string {
    // 简单检查是否是递归引用
    // 如果 currentS 中包含 "${key:" 或 "${key}" 形式的引用，则认为是递归引用
    if strings.Contains(currentS, "${"+key+":") || strings.Contains(currentS, "${"+key+"}") {
        return key
    }
    return ""
}

// MergeAndTilingMap 平铺并合并2个map
func MergeAndTilingMap(highOrder, lowOrder map[string]any) map[string]any {
    hm := TilingMap(highOrder)
    lm := TilingMap(lowOrder)
    r := make(map[string]any, len(hm))
    for k, v := range lm {
        r[k] = v
    }
    for k, v := range hm {
        r[k] = v
    }
    return r
}

// MergeMultiAndTilingMap 平铺并合并多个map
func MergeMultiAndTilingMap(maps ...map[string]any) map[string]any {
    l := len(maps)
    for i := 1; i < l; i++ {
        maps[i] = MergeAndTilingMap(maps[i-1], maps[i])
    }
    return maps[len(maps)-1]
}

// TilingMap 将多层级map平铺为单层级的map
// 例如:
//  map[string]any{
//      "test": "test",
//      "test2": map[string]any{"s":"v", "i":1 }
//     }
// 会转换为
//  map[string]any{
//      "test": "test",
//      "test2.s": "v",
//      "test2.i": i,
//     }
func TilingMap(m map[string]any) map[string]any {
    r := make(map[string]any, len(m))
    for k, v := range m {
        tilingMap(k, reflect.ValueOf(v), r)
    }
    return r
}

// GetByPath 从嵌套map中获取按路径获取值
func GetByPath(m map[string]any, path string) any {
    return getByPath(reflect.ValueOf(m), path)
}

// GetByPathSlice 类似 GetByPath
func GetByPathSlice(m []any, path string) any {
    return getByPath(reflect.ValueOf(m), path)
}

// GetByPathAny 类似 GetByPath,任意类型的包装方法
func GetByPathAny(obj any, path string) any {
    if path == "" {
        return obj
    }
    return getByPath(reflect.ValueOf(obj), path)
}
func getByPath(vv reflect.Value, path string) any {
    vv = reflect.Indirect(vv)
    if vv.Kind() == reflect.Interface {
        vv = vv.Elem()
    }
    if path[0] == '[' {
        _, path = FindFirstKey(path)
    }
    key, nextKey := FindFirstKey(path)
    switch vv.Kind() {
    case reflect.Map:
        for _, v := range vv.MapKeys() {
            if v.String() == key {
                value := vv.MapIndex(v)
                if nextKey == "" {
                    return value.Interface()
                }
                return getByPath(value, nextKey)
            }
        }
        return nil
    case reflect.Slice, reflect.Array:
        i, err := strconv.Atoi(key)
        if err != nil {
            panic("不支持的路径")
        }
        value := vv.Index(i)
        if value.Kind() != reflect.Invalid {
            return getByPath(value, nextKey)
        }
        return nil
    default:
        panic("不支持的路径")
    }
}

func tilingMap(prefix string, vv reflect.Value, r map[string]any) {
    vv = reflect.Indirect(vv)
    if vv.Kind() == reflect.Interface {
        vv = vv.Elem()
    }
    switch vv.Kind() {
    case reflect.Array, reflect.Slice:
        l := vv.Len()
        for i := 0; i < l; i++ {
            tilingMap(prefix+"."+fmt.Sprint(i), vv.Index(i), r)
        }
    case reflect.Map:
        for _, v := range vv.MapKeys() {
            index := vv.MapIndex(v)
            //fmt.Printf("map->%+v\n", reflect.Indirect(index).Interface())
            tilingMap(prefix+"."+fmt.Sprint(v), index, r)
        }
    case reflect.Invalid:
        //nil
    default:
        //非map和list的直接合并
        r[prefix] = vv.Interface()
    }
}

// UntilingMap 单层map转多层map, TilingMap 的逆操作
func UntilingMap(m map[string]any) map[string]any {
    r := make(map[string]any, len(m))
    for k, v := range m {
        //fmt.Println(k)
        untilingMap("", k, reflect.ValueOf(v), r)
    }
    return r
}
func untilingMap(lastKey, oldKey string, vv reflect.Value, r map[string]any) {
    vv = reflect.Indirect(vv)
    if vv.Kind() == reflect.Interface {
        vv = vv.Elem()
    }
    key, next := FindFirstKey(oldKey)
    if next != "" {
        var m map[string]any
        if _, ok := r[key]; !ok {
            m = make(map[string]any)
            r[key] = m
        } else {
            m = r[key].(map[string]any)
        }
        untilingMap(key, next, vv, m)
        return
    }
    //if strings.HasPrefix(testPrefix, "[") {
    //    if i := strings.IndexAny(key, "[."); i >= 0 {
    //        next = key[i:]
    //        key = key[:i-1]
    //    } else {
    //        key = key[:len(key)-1]
    //    }
    //    var l []any
    //    if _, ok := r[lastKey]; !ok {
    //        l = []any{}
    //        r[lastKey] = l
    //    } else {
    //        l = r[lastKey].([]any)
    //    }
    //    index := ToInt(key)
    //    //确保长度足够
    //    if len(l) < index+1 {
    //        l = append(l, make([]any, index-len(l)+1)...)
    //        r[lastKey] = l
    //    }
    //    if next == "" {
    //        l[index] = vv.Interface()
    //        return
    //    }
    //    if strings.HasPrefix(next, "[") {
    //        panic("不支持连续数组结构")
    //    }
    //    var m map[string]any
    //    old := l[index]
    //    if old == nil {
    //        m = make(map[string]any)
    //        r[lastKey] = m
    //    } else {
    //        m = old.(map[string]any)
    //    }
    //    untilingMap(key, next, vv, m)
    //    return
    //}
    r[key] = vv.Interface()
}

func FindFirstKey(path string) (key, nextKey string) {
    if i := strings.IndexAny(path, "[."); i >= 0 {
        next := path[i:]
        path = path[:i]
        if strings.HasPrefix(next, "[") {
            idx := strings.Index(next, "]")
            next = next[:idx] + next[idx+1:]
        }
        next = next[1:]
        return path, next
    }
    return path, ""
}

// EnvMap 获取map格式的环境变量
func EnvMap() map[string]any {
    envMap := make(map[string]any)
    for _, e := range os.Environ() {
        k, v, _ := strings.Cut(e, "=")
        envMap[strings.TrimSpace(k)] = strings.TrimSpace(v)
    }
    return envMap
}
