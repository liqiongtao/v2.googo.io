package goocontext

import (
	"strconv"
	"strings"
)

// Value 从上下文中获取指定key的值（泛型版本）
// 如果值不存在或类型不匹配，返回类型T的零值
func Value[T any](c *Context, key string) T {
	var zero T
	if c == nil || c.Context == nil {
		return zero
	}
	v := c.Context.Value(key)
	if v == nil {
		return zero
	}
	if t, ok := v.(T); ok {
		return t
	}
	return zero
}

// ValueAny 从上下文中获取指定key的值
// 返回原始值，需要调用者进行类型断言
func ValueAny(c *Context, key string) any {
	if c == nil || c.Context == nil {
		return nil
	}
	return c.Context.Value(key)
}

// ValueString 从上下文中获取字符串类型的值
// 支持自动转换：int, int32, int64, float32, float64, bool
func ValueString(c *Context, key string) string {
	v := ValueAny(c, key)
	if v == nil {
		return ""
	}

	// 直接是 string 类型
	if s, ok := v.(string); ok {
		return s
	}

	// 尝试从其他类型转换
	switch val := v.(type) {
	case int:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	}

	return ""
}

// ValueInt 从上下文中获取int类型的值
// 支持自动转换：string, int32, int64, float32, float64, bool
func ValueInt(c *Context, key string) int {
	v := ValueAny(c, key)
	if v == nil {
		return 0
	}

	// 直接是 int 类型
	if i, ok := v.(int); ok {
		return i
	}

	// 尝试从其他类型转换
	switch val := v.(type) {
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	case int32:
		return int(val)
	case int64:
		return int(val)
	case float32:
		return int(val)
	case float64:
		return int(val)
	case bool:
		if val {
			return 1
		}
		return 0
	}

	return 0
}

// ValueInt32 从上下文中获取int32类型的值
// 支持自动转换：string, int, int64, float32, float64, bool
func ValueInt32(c *Context, key string) int32 {
	v := ValueAny(c, key)
	if v == nil {
		return 0
	}

	// 直接是 int32 类型
	if i, ok := v.(int32); ok {
		return i
	}

	// 尝试从其他类型转换
	switch val := v.(type) {
	case string:
		if i, err := strconv.ParseInt(val, 10, 32); err == nil {
			return int32(i)
		}
	case int:
		return int32(val)
	case int64:
		return int32(val)
	case float32:
		return int32(val)
	case float64:
		return int32(val)
	case bool:
		if val {
			return 1
		}
		return 0
	}

	return 0
}

// ValueInt64 从上下文中获取int64类型的值
// 支持自动转换：string, int, int32, float32, float64, bool
func ValueInt64(c *Context, key string) int64 {
	v := ValueAny(c, key)
	if v == nil {
		return 0
	}

	// 直接是 int64 类型
	if i, ok := v.(int64); ok {
		return i
	}

	// 尝试从其他类型转换
	switch val := v.(type) {
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float32:
		return int64(val)
	case float64:
		return int64(val)
	case bool:
		if val {
			return 1
		}
		return 0
	}

	return 0
}

// ValueFloat32 从上下文中获取float32类型的值
// 支持自动转换：string, int, int32, int64, float64, bool
func ValueFloat32(c *Context, key string) float32 {
	v := ValueAny(c, key)
	if v == nil {
		return 0
	}

	// 直接是 float32 类型
	if f, ok := v.(float32); ok {
		return f
	}

	// 尝试从其他类型转换
	switch val := v.(type) {
	case string:
		if f, err := strconv.ParseFloat(val, 32); err == nil {
			return float32(f)
		}
	case int:
		return float32(val)
	case int32:
		return float32(val)
	case int64:
		return float32(val)
	case float64:
		return float32(val)
	case bool:
		if val {
			return 1
		}
		return 0
	}

	return 0
}

// ValueFloat64 从上下文中获取float64类型的值
// 支持自动转换：string, int, int32, int64, float32, bool
func ValueFloat64(c *Context, key string) float64 {
	v := ValueAny(c, key)
	if v == nil {
		return 0
	}

	// 直接是 float64 类型
	if f, ok := v.(float64); ok {
		return f
	}

	// 尝试从其他类型转换
	switch val := v.(type) {
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case bool:
		if val {
			return 1
		}
		return 0
	}

	return 0
}

// ValueBool 从上下文中获取bool类型的值
// 支持自动转换：string, int, int32, int64, float32, float64
// 对于 string 类型：0, nil, null, false, "" 等返回 false，其他返回 true
func ValueBool(c *Context, key string) bool {
	v := ValueAny(c, key)
	if v == nil {
		return false
	}

	// 直接是 bool 类型
	if b, ok := v.(bool); ok {
		return b
	}

	// 尝试从其他类型转换
	switch val := v.(type) {
	case string:
		s := strings.ToLower(strings.TrimSpace(val))
		// 0, nil, null, false, "" 等返回 false
		if s == "" || s == "0" || s == "nil" || s == "null" || s == "false" || s == "no" || s == "off" {
			return false
		}
		return true
	case int:
		return val != 0
	case int32:
		return val != 0
	case int64:
		return val != 0
	case float32:
		return val != 0
	case float64:
		return val != 0
	}

	return false
}
