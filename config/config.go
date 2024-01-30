package config

import (
    "fmt"
    "reflect"
    "sort"
    "strconv"
    "strings"
)

var (
    // Parser 需要手动注册解析工具
    Parser = map[string]func(data []byte) map[string]any{
        //"yaml": func(data []byte) map[string]any {
        //    r := make(map[string]any)
        //    err := yaml.Unmarshal(data, r)
        //    if err != nil {
        //        return nil
        //    }
        //    return r
        //},
    }
    FileTypeHandler = []func(string) string{
        func(name string) string {
            if strings.HasSuffix(name, ".json") {
                return "json"
            } else if strings.HasSuffix(name, ".toml") {
                return "toml"
            } else if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
                return "yaml"
            }
            return ""
        },
    }
)

type File struct {
    Data  []byte
    Name  string
    Order int
}

func ParseConfigs2Map(files ...*File) (map[string]any, error) {
    ds := ParseConfigs(files)
    l := len(ds)
    for i := 1; i < l; i++ {
        ds[i] = MergeAndTilingMap(ds[i-1], ds[i])
    }
    return UntilingMap(ResolveMap(ds[l-1])), nil
}
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

func ParseConfig(file *File) map[string]any {
    f := Parser[FileType(file.Name)]
    if f == nil {
        return nil
    }
    return f(file.Data)
}

func FileType(name string) string {
    for _, f := range FileTypeHandler {
        s := f(name)
        if s != "" {
            return s
        }
    }
    panic("无法解析文件类型:" + name)
}

// ResolveMap 只支持单层map,处理${xxx:default}占位符,无法处理的保留原始值
func ResolveMap(data map[string]any) map[string]any {
    for k, v := range data {
        vv := reflect.Indirect(reflect.ValueOf(v))
        if vv.Kind() == reflect.String {
            s := vv.Interface().(string)
            if strings.Contains(s, "${") {
                data[k] = resolveStringFromMap(s, data)
            }
        }
    }
    return data
}
func resolveStringFromMap(s string, m map[string]any) any {
    startIndex := strings.LastIndex(s, "${")
    var df string
    var pre string
    var aft string
    var test = s
    if startIndex >= 0 {
        endIndex := startIndex + strings.Index(s[startIndex:], "}")
        if endIndex < 0 {
            goto x
        }
        pre = s[:startIndex]
        aft = s[endIndex+1:]
        test = s[startIndex+2 : endIndex]
        before, after, found := strings.Cut(test, ":")
        if found {
            test = before
            df = after
        }
        if strings.Contains(test, "${") {
            test = fmt.Sprint(resolveStringFromMap(test, m))
        }
    }
x:
    r := m[test]
    if r == nil || r == "" {
        r = df
    }
    var r1 string
    if s1, ok := r.(string); ok {
        r1 = s1
    } else {
        r1 = fmt.Sprint(r)
    }
    if pre != "" || aft != "" {
        r1 = pre + r1 + aft
    }
    if strings.ContainsAny(r1, "${") {
        return resolveStringFromMap(r1, m)
    }
    return r1
}

// MergeAndTilingMap 合并多个map的属性
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
func MergeMultiAndTilingMap(maps ...map[string]any) map[string]any {
    l := len(maps)
    for i := 1; i < l; i++ {
        maps[i] = MergeAndTilingMap(maps[i-1], maps[i])
    }
    return maps[len(maps)-1]
}

// TilingMap 将多层级map平铺为单层级的map
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
func GetByPathSlice(m []any, path string) any {
    return getByPath(reflect.ValueOf(m), path)
}
func GetByPathAny(obj any, path string) any {
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
    default:
        //非map和list的直接合并
        r[prefix] = vv.Interface()
    }
}

// UntilingMap 单层map转多层map
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
