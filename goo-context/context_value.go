package goocontext

import (
	"strconv"
)

func (c *Context) Value(key any) any {
	return c.Context.Value(key)
}

func (c *Context) ValueString(key any) string {
	if v, ok := c.Value(key).(int64); ok {
		return strconv.FormatInt(v, 10)
	}
	if v, ok := c.Value(key).(int32); ok {
		return strconv.FormatInt(int64(v), 10)
	}
	if v, ok := c.Value(key).(int); ok {
		return strconv.FormatInt(int64(v), 10)
	}
	if v, ok := c.Value(key).(float64); ok {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	if v, ok := c.Value(key).(float32); ok {
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	}
	if v, ok := c.Value(key).(string); ok {
		return v
	}
	if v, ok := c.Value(key).(bool); ok {
		if v {
			return "true"
		}
		return "false"
	}
	return ""
}

func (c *Context) ValueInt64(key any) int64 {
	if v, ok := c.Value(key).(int64); ok {
		return v
	}
	if v, ok := c.Value(key).(int32); ok {
		return int64(v)
	}
	if v, ok := c.Value(key).(int); ok {
		return int64(v)
	}
	if v, ok := c.Value(key).(float64); ok {
		return int64(v)
	}
	if v, ok := c.Value(key).(float32); ok {
		return int64(v)
	}
	if v, ok := c.Value(key).(string); ok {
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	}
	if v, ok := c.Value(key).(bool); ok {
		if v {
			return 1
		}
		return 0
	}
	return 0
}

func (c *Context) ValueInt32(key any) int32 {
	if v, ok := c.Value(key).(int64); ok {
		return int32(v)
	}
	if v, ok := c.Value(key).(int32); ok {
		return v
	}
	if v, ok := c.Value(key).(int); ok {
		return int32(v)
	}
	if v, ok := c.Value(key).(float64); ok {
		return int32(v)
	}
	if v, ok := c.Value(key).(float32); ok {
		return int32(v)
	}
	if v, ok := c.Value(key).(string); ok {
		n, _ := strconv.ParseInt(v, 10, 64)
		return int32(n)
	}
	if v, ok := c.Value(key).(bool); ok {
		if v {
			return 1
		}
		return 0
	}
	return 0
}

func (c *Context) ValueInt(key any) int {
	if v, ok := c.Value(key).(int64); ok {
		return int(v)
	}
	if v, ok := c.Value(key).(int32); ok {
		return int(v)
	}
	if v, ok := c.Value(key).(int); ok {
		return v
	}
	if v, ok := c.Value(key).(float64); ok {
		return int(v)
	}
	if v, ok := c.Value(key).(float32); ok {
		return int(v)
	}
	if v, ok := c.Value(key).(string); ok {
		n, _ := strconv.ParseInt(v, 10, 64)
		return int(n)
	}
	if v, ok := c.Value(key).(bool); ok {
		if v {
			return 1
		}
		return 0
	}
	return 0
}

func (c *Context) ValueFloat64(key any) float64 {
	if v, ok := c.Value(key).(int64); ok {
		return float64(v)
	}
	if v, ok := c.Value(key).(int32); ok {
		return float64(v)
	}
	if v, ok := c.Value(key).(int); ok {
		return float64(v)
	}
	if v, ok := c.Value(key).(float64); ok {
		return v
	}
	if v, ok := c.Value(key).(float32); ok {
		return float64(v)
	}
	if v, ok := c.Value(key).(string); ok {
		n, _ := strconv.ParseFloat(v, 64)
		return n
	}
	if v, ok := c.Value(key).(bool); ok {
		if v {
			return 1
		}
		return 0
	}
	return 0
}

func (c *Context) ValueFloat32(key any) float32 {
	if v, ok := c.Value(key).(int64); ok {
		return float32(v)
	}
	if v, ok := c.Value(key).(int32); ok {
		return float32(v)
	}
	if v, ok := c.Value(key).(int); ok {
		return float32(v)
	}
	if v, ok := c.Value(key).(float64); ok {
		return float32(v)
	}
	if v, ok := c.Value(key).(float32); ok {
		return v
	}
	if v, ok := c.Value(key).(string); ok {
		n, _ := strconv.ParseFloat(v, 64)
		return float32(n)
	}
	if v, ok := c.Value(key).(bool); ok {
		if v {
			return 1
		}
		return 0
	}
	return 0
}

func (c *Context) ValueBool(key any) bool {
	if v, ok := c.Value(key).(bool); ok {
		return v
	}
	if v, ok := c.Value(key).(int64); ok {
		if v > 0 {
			return true
		}
		return false
	}
	if v, ok := c.Value(key).(int32); ok {
		if v > 0 {
			return true
		}
		return false
	}
	if v, ok := c.Value(key).(int); ok {
		if v > 0 {
			return true
		}
		return false
	}
	if v, ok := c.Value(key).(float64); ok {
		if v > 0 {
			return true
		}
		return false
	}
	if v, ok := c.Value(key).(float32); ok {
		if v > 0 {
			return true
		}
		return false
	}
	if v, ok := c.Value(key).(string); ok {
		if v == "" {
			return true
		}
		return false
	}
	return false
}

func (c *Context) ValueArrayString(key any) []string {
	if v, ok := c.Value(key).([]string); ok {
		return v
	}
	return []string{}
}
