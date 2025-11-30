package goolog

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type Message struct {
	Level   Level
	Message []any
	Time    time.Time
	Entry   *Entry
}

func (msg *Message) JSON() []byte {
	data := map[string]any{}

	if l := len(msg.Entry.Data); l > 0 {
		for _, i := range msg.Entry.Data {
			data[i.Field] = i.Value
		}
	}

	{
		data["log_level"] = LevelText[msg.Level]
		data["log_datetime"] = msg.Time.Format("2006-01-02 15:04:05")

		if l := len(msg.Entry.Tags); l > 0 {
			data["log_tags"] = msg.Entry.Tags
		}

		if l := len(msg.Message); l > 0 {
			var arr []string
			for _, i := range msg.Message {
				arr = append(arr, fmt.Sprint(i))
			}
			data["log_message"] = arr
		}

		if l := len(msg.Entry.Trace); l > 0 {
			data["log_trace"] = msg.Entry.Trace
		}
	}

	buf, err := json.Marshal(&data)
	if err != nil {
		log.Println("[goo-log][message2json]", data, err)
		return []byte{}
	}

	return buf
}

// Text 返回控制台格式的文本
func (msg *Message) Text() string {
	var buf strings.Builder

	// 时间
	buf.WriteString(msg.Time.Format("2006-01-02 15:04:05"))
	buf.WriteString(" ")

	// 级别（带颜色）
	levelText := LevelText[msg.Level]
	if color := Color(msg.Level); color != nil {
		levelText = color(levelText)
	}
	buf.WriteString(fmt.Sprintf("[%s]", levelText))
	buf.WriteString(" ")

	// 标签
	if len(msg.Entry.Tags) > 0 {
		buf.WriteString(fmt.Sprintf("[%s] ", strings.Join(msg.Entry.Tags, ",")))
	}

	// 消息
	if len(msg.Message) > 0 {
		for i, m := range msg.Message {
			if i > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(fmt.Sprint(m))
		}
	}

	// 字段
	if len(msg.Entry.Data) > 0 {
		buf.WriteString(" | ")
		var fields []string
		for _, field := range msg.Entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", field.Field, field.Value))
		}
		buf.WriteString(strings.Join(fields, " "))
	}

	// 追踪信息
	if len(msg.Entry.Trace) > 0 {
		buf.WriteString("\n")
		buf.WriteString("Trace: ")
		buf.WriteString(strings.Join(msg.Entry.Trace, " -> "))
	}

	return buf.String()
}
