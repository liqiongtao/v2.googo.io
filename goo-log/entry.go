package goolog

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	goocontext "v2.googo.io/goo-context"
)

type Entry struct {
	Tags  []string
	Data  []DataField
	Trace []string
	msg   *Message
	l     *Logger
}

type DataField struct {
	Field string
	Value any
}

func NewEntry(l *Logger) *Entry {
	return &Entry{
		Tags:  make([]string, 0, 4),
		Data:  make([]DataField, 0, 4),
		Trace: make([]string, 0, 8),
		l:     l,
	}
}

func (entry *Entry) WithTag(tags ...any) *Entry {
	if len(tags) > 0 {
		for _, tag := range tags {
			entry.Tags = append(entry.Tags, fmt.Sprint(tag))
		}
	}
	return entry
}

func (entry *Entry) WithField(field string, value any) *Entry {
	entry.Data = append(entry.Data, DataField{Field: field, Value: value})
	return entry
}

func (entry *Entry) WithFieldF(field string, format string, args ...any) *Entry {
	value := fmt.Sprintf(format, args...)
	entry.Data = append(entry.Data, DataField{Field: field, Value: value})
	return entry
}

func (entry *Entry) WithContext(ctx *goocontext.Context) *Entry {
	if ctx != nil {
		appName := goocontext.AppName(ctx)
		if appName != "" {
			entry.Data = append(entry.Data, DataField{Field: "app-name", Value: appName})
		}
		traceId := goocontext.TraceId(ctx)
		if traceId != "" {
			entry.Data = append(entry.Data, DataField{Field: "trace-id", Value: traceId})
		}
	}
	return entry
}

func (entry *Entry) WithTrace() *Entry {
	entry.Trace = entry.trace()
	return entry
}

func (entry *Entry) Debug(v ...any) {
	entry.output(DEBUG, v...)
}

func (entry *Entry) DebugF(format string, v ...any) {
	entry.output(DEBUG, fmt.Sprintf(format, v...))
}

func (entry *Entry) Info(v ...any) {
	entry.output(INFO, v...)
}

func (entry *Entry) InfoF(format string, v ...any) {
	entry.output(INFO, fmt.Sprintf(format, v...))
}

func (entry *Entry) Warn(v ...any) {
	entry.output(WARN, v...)
}

func (entry *Entry) WarnF(format string, v ...any) {
	entry.output(WARN, fmt.Sprintf(format, v...))
}

func (entry *Entry) Error(v ...any) {
	entry.output(ERROR, v...)
}

func (entry *Entry) ErrorF(format string, v ...any) {
	entry.output(ERROR, fmt.Sprintf(format, v...))
}

func (entry *Entry) Panic(v ...any) {
	entry.output(PANIC, v...)
}

func (entry *Entry) PanicF(format string, v ...any) {
	entry.output(PANIC, fmt.Sprintf(format, v...))
}

func (entry *Entry) Fatal(v ...any) {
	entry.output(FATAL, v...)
	os.Exit(1)
}

func (entry *Entry) FatalF(format string, v ...any) {
	entry.output(FATAL, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (entry *Entry) output(level Level, v ...any) {
	entry.msg = &Message{
		Level:   level,
		Message: v,
		Time:    time.Now(),
		Entry:   entry,
	}

	// 如果级别达到设置的追踪级别，或者已经手动设置了追踪信息，则添加追踪
	if level >= entry.l.traceLevel || len(entry.Trace) > 0 {
		if len(entry.Trace) == 0 {
			entry.WithTrace()
		}
	}

	for _, fn := range entry.l.hooks {
		go entry.hookHandler(fn)
	}

	if entry.l.adapter != nil {
		entry.l.adapter.Write(entry.msg)
	}
}

func (entry *Entry) hookHandler(fn func(msg *Message)) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	fn(entry.msg)
}

// runtime.Caller 仅能获取非 goroutine 的信息
func (entry *Entry) trace() (arr []string) {
	arr = []string{}

	for i := 3; i < 16; i++ {
		_, file, line, _ := runtime.Caller(i)
		if file == "" {
			continue
		}
		if strings.Contains(file, ".pb.go") ||
			strings.Contains(file, "runtime/") ||
			(!strings.Contains(file, "googo.io") &&
				(strings.Contains(file, "src/") || strings.Contains(file, "pkg/mod/") || strings.Contains(file, "vendor/"))) ||
			strings.Contains(file, "goo-log") {
			continue
		}
		arr = append(arr, fmt.Sprintf("%s %dL", entry.prettyFile(file), line))
	}

	return
}

func (entry *Entry) prettyFile(file string) string {
	var (
		index  int
		index2 int
	)

	if index = strings.LastIndex(file, "src/test/"); index >= 0 {
		return file[index+9:]
	}
	if index = strings.LastIndex(file, "src/"); index >= 0 {
		return file[index+4:]
	}
	if index = strings.LastIndex(file, "pkg/mod/"); index >= 0 {
		return file[index+8:]
	}
	if index = strings.LastIndex(file, "vendor/"); index >= 0 {
		return file[index+7:]
	}

	if index = strings.LastIndex(file, "/"); index < 0 {
		return file
	}
	if index2 = strings.LastIndex(file[:index], "/"); index2 < 0 {
		return file[index+1:]
	}
	return file[index2+1:]
}
