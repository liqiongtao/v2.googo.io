package goolog

import "sync"

type Logger struct {
	hooks     []func(msg *Message)
	adapter   Adapter
	level     Level
	mu        sync.Mutex
	entryPool sync.Pool
}

func New() *Logger {
	return &Logger{}
}

// 获取或创建一个Entry
func (l *Logger) newEntry() *Entry {
	entry, ok := l.entryPool.Get().(*Entry)
	if ok {
		return entry
	}
	return NewEntry(l)
}

func (l *Logger) releaseEntry(entry *Entry) {
	l.entryPool.Put(entry)
}

func (l *Logger) WithTag(tags ...string) *Entry {
	return l.newEntry().WithTag(tags...)
}

func (l *Logger) WithField(field string, value interface{}) *Entry {
	return l.newEntry().WithField(field, value)
}

func (l *Logger) WithTrace() *Entry {
	return l.newEntry().WithTrace()
}

func (l *Logger) Debug(v ...interface{}) {
	l.newEntry().Debug(v...)
}

func (l *Logger) DebugF(format string, v ...interface{}) {
	l.newEntry().DebugF(format, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.newEntry().Info(v...)
}

func (l *Logger) InfoF(format string, v ...interface{}) {
	l.newEntry().InfoF(format, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.newEntry().Warn(v...)
}

func (l *Logger) WarnF(format string, v ...interface{}) {
	l.newEntry().WarnF(format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.newEntry().Error(v...)
}

func (l *Logger) ErrorF(format string, v ...interface{}) {
	l.newEntry().ErrorF(format, v...)
}

func (l *Logger) Panic(v ...interface{}) {
	l.newEntry().Panic(v...)
}

func (l *Logger) PanicF(format string, v ...interface{}) {
	l.newEntry().PanicF(format, v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.newEntry().Fatal(v...)
}

func (l *Logger) FatalF(format string, v ...interface{}) {
	l.newEntry().FatalF(format, v...)
}
