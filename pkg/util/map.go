package util

import (
	"fmt"
	"reflect"
)

// ToAnyValue 将map[K]V转换为map[K]any
func ToAnyValue[K comparable, V any](m map[K]V) map[K]any {
	n := make(map[K]any, len(m))
	for k, v := range m {
		n[k] = v
	}
	return n
}

// TransformKey 将map[K1]V转换为map[K2]V
func TransformKey[K1, K2 comparable, V any](m map[K1]V, f func(K1) K2) map[K2]V {
	n := make(map[K2]V, len(m))
	for k1, v := range m {
		n[f(k1)] = v
	}
	return n
}

// 辅助函数：分割路径
func splitPath(path string) []string {
	var result []string
	start := 0
	for i, char := range path {
		if char == '.' {
			if i > start {
				result = append(result, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		result = append(result, path[start:])
	}
	return result
}

// GetFromNestedMapWithPath 使用路径表达式获取嵌套数据
// 路径格式: "parent.child.grandchild"
func GetFromNestedMapWithPath(data map[string]interface{}, path string) (interface{}, bool) {
	if data == nil || path == "" {
		return nil, false
	}
	current := data
	keys := splitPath(path)
	for i, key := range keys {
		val, exists := current[key]
		if !exists {
			return nil, false
		}
		// 如果是最后一个key，直接返回值
		if i == len(keys)-1 {
			return val, true
		}
		// 否则继续深入
		if nextMap, ok := val.(map[string]interface{}); ok {
			current = nextMap
		} else {
			// 路径中间遇到非map类型
			return nil, false
		}
	}
	return nil, false
}

// GetStringFromNestedMap 获取字符串类型数据
func GetStringFromNestedMap(data map[string]interface{}, key string) (string, bool) {
	val, exists := GetFromNestedMapWithPath(data, key)
	if !exists {
		return "", false
	}
	if str, ok := val.(string); ok {
		return str, true
	}
	// 尝试转换
	return fmt.Sprintf("%v", val), true
}

// GetIntFromNestedMap 获取整数类型数据
func GetIntFromNestedMap(data map[string]interface{}, key string) (int, bool) {
	val, exists := GetFromNestedMapWithPath(data, key)
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetFloatFromNestedMap 获取浮点数类型数据
func GetFloatFromNestedMap(data map[string]interface{}, key string) (float64, bool) {
	val, exists := GetFromNestedMapWithPath(data, key)
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// GetBoolFromNestedMap 获取布尔类型数据
func GetBoolFromNestedMap(data map[string]interface{}, key string) (bool, bool) {
	val, exists := GetFromNestedMapWithPath(data, key)
	if !exists {
		return false, false
	}
	switch v := val.(type) {
	case bool:
		return v, true
	case string:
		return v == "true" || v == "1", true
	case int:
		return v != 0, true
	default:
		return false, false
	}
}

// GetSliceFromNestedMap 获取切片类型数据
func GetSliceFromNestedMap(data map[string]interface{}, key string) ([]interface{}, bool) {
	val, exists := GetFromNestedMapWithPath(data, key)
	if !exists {
		return nil, false
	}
	if slice, ok := val.([]interface{}); ok {
		return slice, true
	}
	// 尝试反射转换
	valRef := reflect.ValueOf(val)
	if valRef.Kind() == reflect.Slice || valRef.Kind() == reflect.Array {
		result := make([]interface{}, valRef.Len())
		for i := 0; i < valRef.Len(); i++ {
			result[i] = valRef.Index(i).Interface()
		}
		return result, true
	}

	return nil, false
}

// GetMapFromNestedMap 获取map类型数据
func GetMapFromNestedMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	val, exists := GetFromNestedMapWithPath(data, key)
	if !exists {
		return nil, false
	}
	if m, ok := val.(map[string]interface{}); ok {
		return m, true
	}
	return nil, false
}

// GetInt64 从map中获取int64类型的值
func GetInt64(data map[string]interface{}, key string) (int64, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int64:
		return v, true
	default:
		return 0, false
	}
}
