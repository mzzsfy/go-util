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
    //t.Log("res", res)
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
    if dataItem.Child("Path").String() == "" && dataItem.Child("PATH").String() == "" {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("runtime").Any() == nil {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("runtime").Child("workId").Int() != 1 {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("runtime.workId").Int() != 1 {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2").Child("test").String() != "test" {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.test").String() != "test" {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.testBool").Bool() != true {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.testInt").Int() != 1 {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.testFloat").Float() != 1.1 {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.testArr").Child("0").Int() != 1 {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.testArr").Child("1").Int() != 2 {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.testArr").Child("1").String() != "2" {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.testArr").Child("2").String() != "3" {
        t.Error("dataItem.Child 未成功解析")
    }
    if dataItem.Child("test2.testArr").Child("3").Float() != 4.0 {
        t.Error("dataItem.Child 未成功解析")
    }
}
