//go:build json_parser

package config

import "encoding/json"

func init() {
    Parser["json"] = func(data []byte) map[string]any {
        r := make(map[string]any)
        err := json.Unmarshal(data, &r)
        if err != nil {
            return nil
        }
        return r
    }
}
