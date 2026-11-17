# 配置

工作模式类似spring配置工具类

标准使用方式:  
1. 读取文件并解析为map
2. 将所有配置使用`MergeMultiAndTilingMap`扁平化并合并为单个map(按 `ParseConfigs` 的 `Order` 降序结果传入,越靠前优先级越高;`MergeMultiAndTilingMap` 本身不读 `Order`)
3. 单map使用`UntilingMap`转换为正常结构的map(返回 map 与 error,键路径冲突时需处理)
4. 使用`ValueFromPath(cfg,"路径")`或`GetByPath(cfg,"路径")`获取配置

参考 [config_test.go](config_test.go)

```go
// 片段(置于 package main/init):使用方需自行 import config、embed、fmt、os 与 yaml 库,并准备 resource/*.yml 资源文件
var (
    cfg map[string]any
    //go:embed resource/*.yml
    configFiles embed.FS
)

func init() {
    profile := "dev"
    if env := os.Getenv("xxx.profile"); env != "" {
        profile = env
    }
    config.Parser["yaml"] = func(data []byte) map[string]any {
        r := make(map[string]any)
        if err := yaml.Unmarshal(data, r); err != nil {
            return nil
        }
        return r
    }
    applicationFile, _ := configFiles.ReadFile("resource/application.yml")
    profileFile, _ := configFiles.ReadFile("resource/application-" + profile + ".yml")
    files := []*config.File{
        {Data: applicationFile, Name: "application.yml", Order: -1},
        {Data: profileFile, Name: "application-" + profile + ".yml", Order: 1},
    }
    envMap := config.EnvMap()
    configs := config.ParseConfigs(files)
    tilingMap := config.MergeMultiAndTilingMap(append([]map[string]any{envMap}, configs...)...)
    resolveMap := config.ResolveMap(tilingMap)
    resolved, err := config.UntilingMap(resolveMap)
    if err != nil {
        // 键路径冲突:同一命名空间下某键既是叶子又是其他键的前缀,常由带点的环境变量与配置文件同域引起
        panic(err)
    }
    cfg = resolved
}

func main() {
    fmt.Println("读取配置文件的值为", config.GetByPath(cfg, "middlewares.rabbitmq.url"))
}
```

```yaml
# application.yml 类似spring配置文件写法,理论支持嵌套,未测试
middlewares:
  rabbitmq:
    host: ${RABBITMQ_HOST:rabbitmq}
    port: ${RABBITMQ_PORT:5672}
    username: ${RABBITMQ_USERNAME:admin}
    password: ${RABBITMQ_PASSWORD:password}
    vhost: ${RABBITMQ_VHOST:test_vhost}
    url: amqp://${middlewares.rabbitmq.username}:${middlewares.rabbitmq.password}@${middlewares.rabbitmq.host}:${middlewares.rabbitmq.port}/${middlewares.rabbitmq.vhost}
```

## API

### 值访问

`ValueFromPath(m any, path string) Value` - 按路径获取Value,提供类型安全访问

`ValueFrom(a any) Value` - 从any值创建Value

`GetByPathAny(obj any, path string) any` - 类似GetByPath,接受任意类型

`Value` 接口提供类型安全的值访问,支持默认值回退:

- `IsNil() bool`
- `Any() any` / `AnyD(defaultValue any) any`
- `String() string` / `StringD(defaultValue string) string`
- `Int() int` / `IntD(defaultValue int) int`
- `Float() float64` / `FloatD(defaultValue float64) float64`
- `Bool() bool` / `BoolD(defaultValue bool) bool`
- `Child(name string) Value`

### 其他工具函数

- `UntilingMap(m map[string]any) (map[string]any, error)`: 单层map转多层map,`TilingMap`的逆操作(仅 map 结构互逆;切片与数组经 `TilingMap` 会展开为以十进制索引为键的 map,非 string 键会转成字符串,键名含 `.` 的 string 键与嵌套路径同形、往返会被拆分且不报错,还原后得到的是 map 而非原类型;nil 叶子与空容器会被 `TilingMap` 丢弃,无法还原);键与键路径冲突(如同时存在 `a` 与 `a.b`)时返回错误
- `TilingMap(m map[string]any) map[string]any`: 将多层级map平铺为单层级map
- `MergeAndTilingMap(highOrder, lowOrder map[string]any) map[string]any` / `MergeMultiAndTilingMap(maps ...map[string]any) map[string]any`: 平铺后合并,前者 highOrder 覆盖 lowOrder,后者参数越靠前优先级越高(不读 `Order`)
- `ResolveMap(data map[string]any) map[string]any`: 解析 `${key:default}` 占位符,仅处理单层 map,嵌套结构需先经 `TilingMap` 平铺
- `ParseConfig(file *File) map[string]any` / `ParseConfigs(files []*File) []map[string]any`: 前者解析单个文件;后者按 `Order` 降序解析多个文件并**就地重排传入切片**,返回未合并的 map 切片(只排序不合并;按该顺序传给 `MergeMultiAndTilingMap` 时高 `Order` 优先)
- `ParseConfigs2Map(files ...*File) (map[string]any, error)`: 解析多个配置文件并合并为单个嵌套map,内部完成平铺、合并、占位符解析
- `GetByPath(m map[string]any, path string) any` / `GetByPathSlice(m []any, path string) any`: 按点分路径取值,空路径分别返回整个 map/切片
- `EnvMap() map[string]any`: 获取map格式的环境变量

## JSON 解析器

内置的 json 解析器位于 [parser_json.go](parser_json.go),文件首行带有 `//go:build json_parser` 构建约束:

- 默认构建(不带该 tag)**不编译**此文件,`Parser["json"]` 未注册;此时解析 `.json` 不会报错而是被静默忽略:`ParseConfig` 返回 `nil`,`ParseConfigs2Map` 得到空配置
- 启用方式: 构建或测试时添加 `-tags json_parser`,例如 `go build -tags json_parser ./...`,注册后即可直接解析 json 配置文件
- 解析错误不会上抛:启用 tag 后 `json.Unmarshal` 失败同样返回 `nil`,配置文件被静默忽略
- yaml 等其他格式无内置解析器,需按上文示例自行向 `Parser` 注册
