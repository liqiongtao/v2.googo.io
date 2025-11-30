package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	goolog "v2.googo.io/goo-log"
)

// ESAdapter Elasticsearch 适配器
type ESAdapter struct {
	url      string       // ES 地址
	index    string       // 索引名称
	client   *http.Client // HTTP 客户端
	useAsync bool         // 是否异步写入
}

// ESConfig ES 适配器配置
type ESConfig struct {
	URL      string // ES 地址，例如 "http://localhost:9200"
	Index    string // 索引名称，默认 "goolog"
	UseAsync bool   // 是否异步写入，默认 true
}

// NewESAdapter 创建 ES 适配器
func NewESAdapter(config ESConfig) *ESAdapter {
	if config.Index == "" {
		config.Index = "goolog"
	}
	if config.UseAsync {
		// 默认异步
		config.UseAsync = true
	}

	return &ESAdapter{
		url:      config.URL,
		index:    config.Index,
		client:   &http.Client{Timeout: 5 * time.Second},
		useAsync: config.UseAsync,
	}
}

// Write 写入日志到 ES
func (e *ESAdapter) Write(msg *goolog.Message) {
	if e.useAsync {
		go e.writeToES(msg)
	} else {
		e.writeToES(msg)
	}
}

// writeToES 实际写入 ES
func (e *ESAdapter) writeToES(msg *goolog.Message) {
	// 构建文档
	doc := e.buildDocument(msg)
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return
	}

	// 构建 URL（使用日期作为索引后缀）
	dateStr := msg.Time.Format("2006-01-02")
	url := fmt.Sprintf("%s/%s-%s/_doc", e.url, e.index, dateStr)

	// 发送请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(docBytes))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// buildDocument 构建 ES 文档
func (e *ESAdapter) buildDocument(msg *goolog.Message) map[string]any {
	doc := map[string]any{
		"@timestamp": msg.Time.Format(time.RFC3339),
		"level":      goolog.LevelText[msg.Level],
		"message":    fmt.Sprint(msg.Message...),
	}

	// 添加标签
	if len(msg.Entry.Tags) > 0 {
		doc["tags"] = msg.Entry.Tags
	}

	// 添加字段
	if len(msg.Entry.Data) > 0 {
		for _, field := range msg.Entry.Data {
			doc[field.Field] = field.Value
		}
	}

	// 添加追踪信息
	if len(msg.Entry.Trace) > 0 {
		doc["trace"] = msg.Entry.Trace
	}

	return doc
}
