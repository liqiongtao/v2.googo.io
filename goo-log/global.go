package goolog

import goocontext "v2.googo.io/goo-context"

var defaultLogger = New()

// SetLevel 设置默认日志级别
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// SetTraceLevel 设置默认追踪级别
func SetTraceLevel(level Level) {
	defaultLogger.SetTraceLevel(level)
}

// SetAdapter 设置默认适配器
func SetAdapter(adapter Adapter) {
	defaultLogger.SetAdapter(adapter)
}

// AddHook 添加默认钩子函数
func AddHook(fn func(msg *Message)) {
	defaultLogger.AddHook(fn)
}

// WithTag 使用默认日志器创建带标签的 Entry
func WithTag(tags ...any) *Entry {
	return defaultLogger.WithTag(tags...)
}

// WithField 使用默认日志器创建带字段的 Entry
func WithField(field string, value any) *Entry {
	return defaultLogger.WithField(field, value)
}

// WithFieldF 使用默认日志器创建带格式化字段的 Entry
func WithFieldF(field string, format string, args ...any) *Entry {
	return defaultLogger.WithFieldF(field, format, args...)
}

// WithContext 使用默认日志器从上下文创建 Entry
func WithContext(ctx *goocontext.Context) *Entry {
	return defaultLogger.WithContext(ctx)
}

// WithTrace 使用默认日志器创建带追踪信息的 Entry
func WithTrace() *Entry {
	return defaultLogger.WithTrace()
}

// Debug 使用默认日志器记录 DEBUG 级别日志
func Debug(v ...any) {
	defaultLogger.Debug(v...)
}

// DebugF 使用默认日志器记录 DEBUG 级别日志（格式化）
func DebugF(format string, v ...any) {
	defaultLogger.DebugF(format, v...)
}

// Info 使用默认日志器记录 INFO 级别日志
func Info(v ...any) {
	defaultLogger.Info(v...)
}

// InfoF 使用默认日志器记录 INFO 级别日志（格式化）
func InfoF(format string, v ...any) {
	defaultLogger.InfoF(format, v...)
}

// Warn 使用默认日志器记录 WARN 级别日志
func Warn(v ...any) {
	defaultLogger.Warn(v...)
}

// WarnF 使用默认日志器记录 WARN 级别日志（格式化）
func WarnF(format string, v ...any) {
	defaultLogger.WarnF(format, v...)
}

// Error 使用默认日志器记录 ERROR 级别日志
func Error(v ...any) {
	defaultLogger.Error(v...)
}

// ErrorF 使用默认日志器记录 ERROR 级别日志（格式化）
func ErrorF(format string, v ...any) {
	defaultLogger.ErrorF(format, v...)
}

// Panic 使用默认日志器记录 PANIC 级别日志
func Panic(v ...any) {
	defaultLogger.Panic(v...)
}

// PanicF 使用默认日志器记录 PANIC 级别日志（格式化）
func PanicF(format string, v ...any) {
	defaultLogger.PanicF(format, v...)
}

// Fatal 使用默认日志器记录 FATAL 级别日志
func Fatal(v ...any) {
	defaultLogger.Fatal(v...)
}

// FatalF 使用默认日志器记录 FATAL 级别日志（格式化）
func FatalF(format string, v ...any) {
	defaultLogger.FatalF(format, v...)
}

// Default 获取默认日志器
func Default() *Logger {
	return defaultLogger
}
