# 配置

工作模式类似spring配置工具类

标准使用方式:  
1. 读取文件并解析为map
2. 将所有配置使用`MergeMultiAndTilingMap`扁平化并合并为单个map
3. 单map使用`UntilingMap`转换为正常结构的map
4. 使用`ValueFromPath(config,"路径")`或`GetByPath(config,"路径")`获取配置

参考 [config_test.go](config_test.go)

```go
package main

var (
    config map[string]any
    //go:embed resource/*.yml
    configFiles embed.FS
)

func init() {
    profile := "dev"
    env := os.Getenv("xxx.profile")
    if env != "" {
        profile = env
    }
    Parser["yaml"] = func(data []byte) map[string]any {
        r := make(map[string]any)
        err := yaml.Unmarshal(data, r)
        if err != nil {
            return nil
        }
        return r
    }
    println("当前环境: ", env)
    var files []*File
    envMap := make(map[string]any)
    for _, e := range os.Environ() {
        k, v, _ := strings.Cut(e, "=")
        envMap[strings.TrimSpace(k)] = strings.TrimSpace(v)
    }
    applicationFile, _ := configFiles.ReadFile("resource/application.yml")
    files = append(files, &File{
        Data:  applicationFile,
        Name:  "application.yml",
        Order: -1,
    })
    profileFileName := "application-" + profile + ".yml"
    envFile, _ := configFiles.ReadFile("resource/" + profileFileName)
    files = append(files, &File{
        Data:  envFile,
        Name:  "application-" + profile + ".yml",
        Order: 1,
    })
    println("已加载环境变量文件: ", profileFileName)
    configs := ParseConfigs(files)
    tilingMap := MergeMultiAndTilingMap(append([]map[string]any{envMap}, configs...)...)
    resolveMap := ResolveMap(tilingMap)
    config = UntilingMap(resolveMap)
}
func main() {
    println("读取配置文件的值为",GetByPath(config, "middlewares.rabbitmq.url"))
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

- `ParseConfigs2Map(files ...*File) (map[string]any, error)`: 解析多个配置文件并合并为单个嵌套map,内部完成平铺、合并、占位符解析
- `TilingMap(m map[string]any) map[string]any`: 将多层级map平铺为单层级map
- `EnvMap() map[string]any`: 获取map格式的环境变量
