package adapters

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	goolog "v2.googo.io/goo-log"
)

// FileAdapter 文件适配器
type FileAdapter struct {
	dir          string        // 日志目录
	fileName     string        // 文件名模板（支持日期格式）
	maxSize      int64         // 最大文件大小（字节）
	retainDays   int           // 保留天数
	useJSON      bool          // 是否使用 JSON 格式
	currentFile  *os.File      // 当前文件
	currentSize  int64         // 当前文件大小
	currentDate  string        // 当前日期
	fileIndex    int           // 当前文件索引
	mu           sync.Mutex    // 互斥锁
	compressChan chan string   // 压缩任务通道
	stopChan     chan struct{} // 停止信号
	wg           sync.WaitGroup
}

// FileConfig 文件适配器配置
type FileConfig struct {
	Dir        string // 日志目录，默认 "logs"
	FileName   string // 文件名模板，默认 "2006-01-02.log"（会格式化为当前日期，如：2024-01-15.log）
	MaxSize    int64  // 最大文件大小（字节），默认 200MB
	RetainDays int    // 保留天数，默认 30
	UseJSON    bool   // 是否使用 JSON 格式，默认 false
}

// NewFileAdapter 创建文件适配器
func NewFileAdapter(config ...FileConfig) (*FileAdapter, error) {
	cfg := FileConfig{
		Dir:        "logs",
		FileName:   "2006-01-02.log",
		MaxSize:    200 * 1024 * 1024, // 200MB
		RetainDays: 30,
		UseJSON:    false,
	}
	if len(config) > 0 {
		if config[0].Dir != "" {
			cfg.Dir = config[0].Dir
		}
		if config[0].FileName != "" {
			cfg.FileName = config[0].FileName
		}
		if config[0].MaxSize > 0 {
			cfg.MaxSize = config[0].MaxSize
		}
		if config[0].RetainDays > 0 {
			cfg.RetainDays = config[0].RetainDays
		}
		cfg.UseJSON = config[0].UseJSON
	}

	adapter := &FileAdapter{
		dir:          cfg.Dir,
		fileName:     cfg.FileName,
		maxSize:      cfg.MaxSize,
		retainDays:   cfg.RetainDays,
		useJSON:      cfg.UseJSON,
		compressChan: make(chan string, 100),
		stopChan:     make(chan struct{}),
	}

	// 创建日志目录
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 启动压缩协程
	adapter.wg.Add(1)
	go adapter.compressWorker()

	// 启动清理协程
	adapter.wg.Add(1)
	go adapter.cleanupWorker()

	return adapter, nil
}

// Write 写入日志
func (f *FileAdapter) Write(msg *goolog.Message) {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// 检查是否需要切换文件（日期变化或文件过大）
	if f.currentFile == nil || f.currentDate != dateStr || f.currentSize >= f.maxSize {
		if err := f.rotateFile(dateStr); err != nil {
			fmt.Fprintf(os.Stderr, "[goo-log] 文件切换失败: %v\n", err)
			return
		}
	}

	// 写入日志
	var data []byte
	if f.useJSON {
		data = msg.JSON()
	} else {
		data = []byte(msg.Text())
	}
	data = append(data, '\n')

	n, err := f.currentFile.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[goo-log] 写入日志失败: %v\n", err)
		return
	}

	f.currentSize += int64(n)
}

// rotateFile 切换文件
func (f *FileAdapter) rotateFile(dateStr string) error {
	// 关闭旧文件
	if f.currentFile != nil {
		f.currentFile.Close()
	}

	// 如果日期变化，重置索引
	if f.currentDate != dateStr {
		f.fileIndex = 0
		f.currentDate = dateStr
	} else {
		// 文件过大，增加索引
		f.fileIndex++
	}

	// 生成文件名并找到可用的文件
	var filePath string
	var file *os.File
	var err error

	for {
		fileName := f.generateFileName(dateStr)
		filePath = filepath.Join(f.dir, fileName)

		// 尝试打开文件
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %w", err)
		}

		// 获取文件信息
		info, err := file.Stat()
		if err != nil {
			file.Close()
			return fmt.Errorf("获取文件信息失败: %w", err)
		}

		// 如果文件大小小于最大限制，使用这个文件
		if info.Size() < f.maxSize {
			f.currentFile = file
			f.currentSize = info.Size()
			return nil
		}

		// 文件太大，关闭并尝试下一个索引
		file.Close()
		f.fileIndex++
	}
}

// generateFileName 生成文件名
func (f *FileAdapter) generateFileName(dateStr string) string {
	// 解析日期
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// 如果解析失败，使用默认格式
		date = time.Now()
	}

	// 使用时间格式化生成文件名
	fileName := date.Format(f.fileName)

	if f.fileIndex > 0 {
		// 添加索引后缀
		ext := filepath.Ext(fileName)
		name := fileName[:len(fileName)-len(ext)]
		fileName = fmt.Sprintf("%s.%d%s", name, f.fileIndex, ext)
	}

	return fileName
}

// compressWorker 压缩工作协程
func (f *FileAdapter) compressWorker() {
	defer f.wg.Done()

	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-f.stopChan:
			return
		case filePath := <-f.compressChan:
			f.compressFile(filePath)
		case <-ticker.C:
			f.compressOldFiles()
		}
	}
}

// compressFile 压缩单个文件
func (f *FileAdapter) compressFile(filePath string) {
	// 检查文件是否已压缩
	if strings.HasSuffix(filePath, ".gz") {
		return
	}

	// 打开源文件
	srcFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	// 创建压缩文件
	dstFile, err := os.Create(filePath + ".gz")
	if err != nil {
		return
	}
	defer dstFile.Close()

	// 压缩
	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		os.Remove(filePath + ".gz")
		return
	}

	// 删除原文件
	os.Remove(filePath)
}

// compressOldFiles 压缩旧文件
func (f *FileAdapter) compressOldFiles() {
	cutoffDate := time.Now().AddDate(0, 0, -f.retainDays)

	filepath.Walk(f.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// 只处理日志文件
		if info.IsDir() || (!strings.HasSuffix(path, ".log") && !strings.HasSuffix(path, ".log.gz")) {
			return nil
		}

		// 检查文件修改时间
		if info.ModTime().Before(cutoffDate) {
			// 如果未压缩，发送到压缩通道
			if !strings.HasSuffix(path, ".gz") {
				select {
				case f.compressChan <- path:
				default:
				}
			}
		}

		return nil
	})
}

// cleanupWorker 清理工作协程
func (f *FileAdapter) cleanupWorker() {
	defer f.wg.Done()

	ticker := time.NewTicker(24 * time.Hour) // 每天检查一次
	defer ticker.Stop()

	// 立即执行一次
	f.cleanupOldFiles()

	for {
		select {
		case <-f.stopChan:
			return
		case <-ticker.C:
			f.cleanupOldFiles()
		}
	}
}

// cleanupOldFiles 清理过期文件
func (f *FileAdapter) cleanupOldFiles() {
	cutoffDate := time.Now().AddDate(0, 0, -f.retainDays)

	filepath.Walk(f.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// 只处理已压缩的日志文件
		if info.IsDir() || !strings.HasSuffix(path, ".log.gz") {
			return nil
		}

		// 检查文件修改时间
		if info.ModTime().Before(cutoffDate) {
			os.Remove(path)
		}

		return nil
	})
}

// Close 关闭适配器
func (f *FileAdapter) Close() error {
	close(f.stopChan)
	f.wg.Wait()

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.currentFile != nil {
		return f.currentFile.Close()
	}
	return nil
}
