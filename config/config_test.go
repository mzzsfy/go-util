package config

import (
    "encoding/json"
    "testing"
)

func Test_Parse(t *testing.T) {
    var testConfig = map[string]any{
        "runtime": map[string]any{
            "workId": 1,
        },
        "test": "test",
        "test2": map[string]any{
            "test":      "test",
            "testBool":  true,
            "testInt":   1,
            "testFloat": 1.1,
            "testArr": []any{
                1, 2, "3", 4.0,
            },
            "testMap": map[string]any{
                "test":  "test",
                "test1": 1,
                "test2": 1.1,
                "test3": true,
            },
        },
    }
    var configStr []byte

    configStr, _ = json.Marshal(&testConfig)
    //t.Log(string(configStr))
    Parser["json"] = func(data []byte) map[string]any {
        r := make(map[string]any)
        err := json.Unmarshal(data, &r)
        if err != nil {
            return nil
        }
        return r
    }
    //t.Log(envMap)

    //以上为数据准备
    parseConfig := ParseConfig(&File{Data: configStr, Name: "testConfig.json"})
    m := MergeMultiAndTilingMap(EnvMap(), parseConfig)
    if len(m) <= len(testConfig) {
        t.Error("未成功合并环境变量")
    }
    //t.Log("m", m)
    resolveMap := ResolveMap(m)
    //t.Log("resolveMap", resolveMap)
    res := UntilingMap(resolveMap)
    t.Log("res", res)
    if Item("Path").ValueString(res) == "" && Item("PATH").ValueString(res) == "" {
        t.Error("path 未成功解析")
    }
    if Item("runtime").ValueAny(res) == nil {
        t.Error("runtime 未成功解析")
    }
    if Item("runtime.workId").ValueInt(res) != 1 {
        t.Error("runtime.workId 未成功解析")
    }
    if Item("test2.test").ValueString(res) != "test" {
        t.Error("test2.test 未成功解析")
    }
    if Item("test2.testBool").ValueBool(res) != true {
        t.Error("test2.testBool 未成功解析")
    }
    if Item("test2.testInt").ValueInt(res) != 1 {
        t.Error("test2.testInt 未成功解析")
    }
    if Item("test2.testFloat").ValueFloat(res) != 1.1 {
        t.Error("test2.testFloat 未成功解析")
    }
    dataItem := NewDataItem(res, "")
    if dataItem.ValueChild("Path").ValueString() == "" && dataItem.ValueChild("PATH").ValueString() == "" {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("runtime").ValueAny() == nil {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("runtime").ValueChild("workId").ValueInt() != 1 {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("runtime.workId").ValueInt() != 1 {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2").ValueChild("test").ValueString() != "test" {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.test").ValueString() != "test" {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.testBool").ValueBool() != true {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.testInt").ValueInt() != 1 {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.testFloat").ValueFloat() != 1.1 {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.testArr").ValueChild("0").ValueInt() != 1 {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.testArr").ValueChild("1").ValueInt() != 2 {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.testArr").ValueChild("1").ValueString() != "2" {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.testArr").ValueChild("2").ValueString() != "3" {
        t.Error("dataItem.ValueChild 未成功解析")
    }
    if dataItem.ValueChild("test2.testArr").ValueChild("3").ValueFloat() != 4.0 {
        t.Error("dataItem.ValueChild 未成功解析")
    }
}
