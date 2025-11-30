package adapters

import (
	"fmt"
	"os"

	goolog "v2.googo.io/goo-log"
)

// ConsoleAdapter 控制台适配器
type ConsoleAdapter struct {
	useJSON bool // 是否使用 JSON 格式
}

// NewConsoleAdapter 创建控制台适配器
func NewConsoleAdapter(useJSON ...bool) *ConsoleAdapter {
	json := false
	if len(useJSON) > 0 {
		json = useJSON[0]
	}
	return &ConsoleAdapter{
		useJSON: json,
	}
}

// Write 写入日志
func (c *ConsoleAdapter) Write(msg *goolog.Message) {
	var output string
	if c.useJSON {
		output = string(msg.JSON()) + "\n"
	} else {
		output = msg.Text() + "\n"
	}
	fmt.Fprint(os.Stdout, output)
}
