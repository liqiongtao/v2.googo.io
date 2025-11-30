package goocontext

import (
	"context"
	"testing"
	"time"
)

func TestAppName(t *testing.T) {
	ctx := WithAppName(context.Background(), "myApp")
	t.Logf("AppName: %s", AppName(ctx))
}

func TestTraceId(t *testing.T) {
	ctx := WithTraceId(context.Background())
	traceId := TraceId(ctx)
	t.Logf("TraceId: %s", traceId)

	// 测试自定义traceId
	customId := "custom-trace-id-123"
	ctx2 := WithTraceId(context.Background(), customId)
	t.Logf("TraceId: %s", TraceId(ctx2))
}

func TestWithValue(t *testing.T) {
	ctx := WithAppName(context.Background(), "myApp")
	t.Logf("AppName: %s", AppName(ctx))

	ctx = ctx.WithValue("name", "hnatao")
	t.Logf("Name: %s", ValueString(ctx, "name"))

	ctx = ctx.WithValue("age", "18")
	t.Logf("Age: %d", ValueInt(ctx, "age"))
}

func TestWithCancel(t *testing.T) {
	ctx := Default(context.Background())
	ctx, cancel := ctx.WithCancel()

	done := make(chan bool)
	go func() {
		<-ctx.Done()
		done <- true
	}()

	// 取消上下文
	cancel()

	select {
	case <-done:
		t.Log("Context cancelled")
	case <-time.After(1 * time.Second):
		t.Error("Context should be cancelled")
	}
}

func TestWithTimeout(t *testing.T) {
	ctx := Default(context.Background())
	ctx, cancel := ctx.WithTimeout(100 * time.Millisecond)
	defer cancel()

	done := make(chan bool)
	go func() {
		<-ctx.Done()
		done <- true
	}()

	select {
	case <-done:
		t.Log("Context timeout")
	case <-time.After(200 * time.Millisecond):
		t.Error("Context should timeout")
	}
}

func TestValueTypes(t *testing.T) {
	ctx := Default(context.Background())

	ctx = ctx.WithValue("string", "test")
	t.Logf("String: %s", ValueString(ctx, "string"))

	ctx = ctx.WithValue("str_int", "100")
	t.Logf("String: %s", ValueString(ctx, "str_int"))

	ctx = ctx.WithValue("int", 42)
	t.Logf("Int: %d", ValueInt(ctx, "int"))

	ctx = ctx.WithValue("int64", int64(100))
	t.Logf("Int64: %d", ValueInt64(ctx, "int64"))

	ctx = ctx.WithValue("float64", 3.14)
	t.Logf("Float64: %f", ValueFloat64(ctx, "float64"))

	ctx = ctx.WithValue("bool", true)
	t.Logf("Bool: %t", ValueBool(ctx, "bool"))
}

// TestValueInt64AutoConvert 测试 ValueInt64 的自动类型转换
func TestValueInt64AutoConvert(t *testing.T) {
	ctx := Default(context.Background())

	// 测试从 string 转换
	ctx = ctx.WithValue("str_int", "12345")
	t.Logf("String to Int64: %d", ValueInt64(ctx, "str_int"))

	// 测试从 int 转换
	ctx = ctx.WithValue("int_val", 100)
	t.Logf("Int to Int64: %d", ValueInt64(ctx, "int_val"))

	// 测试从 int32 转换
	ctx = ctx.WithValue("int32_val", int32(200))
	t.Logf("Int32 to Int64: %d", ValueInt64(ctx, "int32_val"))

	// 测试从 float32 转换
	ctx = ctx.WithValue("float32_val", float32(300.5))
	t.Logf("Float32 to Int64: %d", ValueInt64(ctx, "float32_val"))

	// 测试从 float64 转换
	ctx = ctx.WithValue("float64_val", 400.7)
	t.Logf("Float64 to Int64: %d", ValueInt64(ctx, "float64_val"))

	// 测试从 bool 转换
	ctx = ctx.WithValue("bool_true", true)
	t.Logf("Bool to Int64: %d", ValueInt64(ctx, "bool_true"))
}
