//go:build yaml_parser

package config

import (
    "gopkg.in/yaml.v3"
)

func init() {
    Parser["yaml"] = func(data []byte) map[string]any {
        r := make(map[string]any)
        err := yaml.Unmarshal(data, r)
        if err != nil {
            return nil
        }
        return r
    }
    Parser["yml"] = func(data []byte) map[string]any {
        r := make(map[string]any)
        err := yaml.Unmarshal(data, r)
        if err != nil {
            return nil
        }
        return r
    }
}
