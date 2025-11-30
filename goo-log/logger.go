package goolog

import (
	"sync"

	goocontext "v2.googo.io/goo-context"
)

type Logger struct {
	hooks      []func(msg *Message)
	adapter    Adapter
	level      Level
	traceLevel Level // 追踪级别，达到此级别及以上时自动添加追踪信息
	mu         sync.Mutex
	entryPool  sync.Pool
}

func New() *Logger {
	return &Logger{
		level:      DEBUG,
		traceLevel: WARN, // 默认 WARN 级别及以上自动添加追踪
	}
}

// 获取或创建一个Entry
func (l *Logger) newEntry() *Entry {
	entry, ok := l.entryPool.Get().(*Entry)
	if ok {
		// 重置 entry
		entry.Tags = entry.Tags[:0]
		entry.Data = entry.Data[:0]
		entry.Trace = entry.Trace[:0]
		entry.msg = nil
		return entry
	}
	return NewEntry(l)
}

func (l *Logger) releaseEntry(entry *Entry) {
	l.entryPool.Put(entry)
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetTraceLevel 设置追踪级别
func (l *Logger) SetTraceLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.traceLevel = level
}

// SetAdapter 设置适配器
func (l *Logger) SetAdapter(adapter Adapter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.adapter = adapter
}

// AddHook 添加钩子函数
func (l *Logger) AddHook(fn func(msg *Message)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, fn)
}

func (l *Logger) WithTag(tags ...any) *Entry {
	return l.newEntry().WithTag(tags...)
}

func (l *Logger) WithField(field string, value any) *Entry {
	return l.newEntry().WithField(field, value)
}

func (l *Logger) WithFieldF(field string, format string, args ...any) *Entry {
	return l.newEntry().WithFieldF(field, format, args...)
}

func (l *Logger) WithContext(ctx *goocontext.Context) *Entry {
	return l.newEntry().WithContext(ctx)
}

func (l *Logger) WithTrace() *Entry {
	return l.newEntry().WithTrace()
}

func (l *Logger) Debug(v ...any) {
	if l.level <= DEBUG {
		l.newEntry().Debug(v...)
	}
}

func (l *Logger) DebugF(format string, v ...any) {
	if l.level <= DEBUG {
		l.newEntry().DebugF(format, v...)
	}
}

func (l *Logger) Info(v ...any) {
	if l.level <= INFO {
		l.newEntry().Info(v...)
	}
}

func (l *Logger) InfoF(format string, v ...any) {
	if l.level <= INFO {
		l.newEntry().InfoF(format, v...)
	}
}

func (l *Logger) Warn(v ...any) {
	if l.level <= WARN {
		l.newEntry().Warn(v...)
	}
}

func (l *Logger) WarnF(format string, v ...any) {
	if l.level <= WARN {
		l.newEntry().WarnF(format, v...)
	}
}

func (l *Logger) Error(v ...any) {
	if l.level <= ERROR {
		l.newEntry().Error(v...)
	}
}

func (l *Logger) ErrorF(format string, v ...any) {
	if l.level <= ERROR {
		l.newEntry().ErrorF(format, v...)
	}
}

func (l *Logger) Panic(v ...any) {
	if l.level <= PANIC {
		l.newEntry().Panic(v...)
	}
}

func (l *Logger) PanicF(format string, v ...any) {
	if l.level <= PANIC {
		l.newEntry().PanicF(format, v...)
	}
}

func (l *Logger) Fatal(v ...any) {
	if l.level <= FATAL {
		l.newEntry().Fatal(v...)
	}
}

func (l *Logger) FatalF(format string, v ...any) {
	if l.level <= FATAL {
		l.newEntry().FatalF(format, v...)
	}
}
