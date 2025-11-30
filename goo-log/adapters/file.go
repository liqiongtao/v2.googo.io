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
	dir           string        // 日志目录
	fileName      string        // 文件名模板（支持日期格式）
	maxSize       int64         // 最大文件大小（字节）
	retainDays    int           // 保留天数
	useJSON       bool          // 是否使用 JSON 格式
	currentFile   *os.File      // 当前文件
	currentSize   int64         // 当前文件大小
	currentDate   string        // 当前日期
	currentPath   string        // 当前文件路径（基础文件名）
	mu            sync.Mutex    // 互斥锁（保护文件操作）
	writeChan     chan []byte   // 写入通道（异步缓冲）
	buffer        []byte        // 批量写入缓冲区
	bufferSize    int           // 缓冲区大小（字节）
	flushInterval time.Duration // 刷新间隔
	compressChan  chan string   // 压缩任务通道
	stopChan      chan struct{} // 停止信号
	wg            sync.WaitGroup
}

// FileConfig 文件适配器配置
type FileConfig struct {
	Dir           string        // 日志目录，默认 "logs"
	FileName      string        // 文件名模板，默认 "2006-01-02.log"（会格式化为当前日期，如：2024-01-15.log）
	MaxSize       int64         // 最大文件大小（字节），默认 200MB
	RetainDays    int           // 保留天数，默认 30
	UseJSON       bool          // 是否使用 JSON 格式，默认 true
	BufferSize    int           // 缓冲区大小（字节），默认 64KB
	FlushInterval time.Duration // 刷新间隔，默认 100ms
	ChannelSize   int           // 写入通道缓冲区大小，默认 5000
}

// NewFileAdapter 创建文件适配器
func NewFileAdapter(config ...FileConfig) (*FileAdapter, error) {
	cfg := FileConfig{
		Dir:        "logs",
		FileName:   "2006-01-02.log",
		MaxSize:    200 * 1024 * 1024, // 200MB
		RetainDays: 30,
		UseJSON:    true,
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

	// 设置默认值
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 64 * 1024 // 64KB
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 100 * time.Millisecond
	}
	if cfg.ChannelSize <= 0 {
		cfg.ChannelSize = 5000
	}

	adapter := &FileAdapter{
		dir:           cfg.Dir,
		fileName:      cfg.FileName,
		maxSize:       cfg.MaxSize,
		retainDays:    cfg.RetainDays,
		useJSON:       cfg.UseJSON,
		writeChan:     make(chan []byte, cfg.ChannelSize),
		buffer:        make([]byte, 0, cfg.BufferSize),
		bufferSize:    cfg.BufferSize,
		flushInterval: cfg.FlushInterval,
		compressChan:  make(chan string, 100),
		stopChan:      make(chan struct{}),
	}

	// 创建日志目录
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 启动写入协程（异步缓冲写入）
	adapter.wg.Add(1)
	go adapter.writeWorker()

	// 启动压缩协程
	adapter.wg.Add(1)
	go adapter.compressWorker()

	// 启动清理协程
	adapter.wg.Add(1)
	go adapter.cleanupWorker()

	return adapter, nil
}

// Write 写入日志（异步缓冲写入）
func (f *FileAdapter) Write(msg *goolog.Message) {
	// 快速序列化并发送到 channel，不阻塞
	var data []byte
	if f.useJSON {
		data = msg.JSON()
	} else {
		data = []byte(msg.Text())
	}
	data = append(data, '\n')

	// 非阻塞发送，如果 channel 满了则丢弃（避免阻塞调用者）
	select {
	case f.writeChan <- data:
		// 成功发送
	default:
		// channel 满了，可以选择记录警告或丢弃
		// 这里选择静默丢弃，避免阻塞
		fmt.Fprintf(os.Stderr, "[goo-log] 写入通道已满，丢弃日志\n")
	}
}

// writeWorker 异步写入工作协程
func (f *FileAdapter) writeWorker() {
	defer f.wg.Done()

	flushTicker := time.NewTicker(f.flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case <-f.stopChan:
			// 关闭时，处理剩余数据
			f.flushRemaining()
			return
		case data := <-f.writeChan:
			// 添加到缓冲区
			f.buffer = append(f.buffer, data...)
			// 如果缓冲区达到大小，立即刷新
			if len(f.buffer) >= f.bufferSize {
				f.flush()
			}
		case <-flushTicker.C:
			// 定期刷新
			f.flush()
		}
	}
}

// flush 刷新缓冲区到文件
func (f *FileAdapter) flush() {
	if len(f.buffer) == 0 {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// 检查是否需要切换文件
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// 如果日期变化，必须切换文件
	needRotate := f.currentFile == nil || f.currentDate != dateStr

	// 如果文件过大，需要切换文件
	if !needRotate && f.currentFile != nil {
		// 检查当前文件大小（需要重新获取，因为可能被其他进程修改）
		if info, err := f.currentFile.Stat(); err == nil {
			actualSize := info.Size()
			if actualSize >= f.maxSize {
				needRotate = true
			} else {
				// 更新当前文件大小
				f.currentSize = actualSize
			}
		} else {
			// 文件状态获取失败，可能需要重新打开
			needRotate = true
		}
	} else if !needRotate && f.currentFile == nil {
		// 文件未打开，需要初始化
		needRotate = true
	}

	if needRotate {
		if err := f.rotateFile(dateStr); err != nil {
			fmt.Fprintf(os.Stderr, "[goo-log] 文件切换失败: %v\n", err)
			// 清空缓冲区，避免重复写入
			f.buffer = f.buffer[:0]
			return
		}
	}

	// 确保文件已打开
	if f.currentFile == nil {
		fmt.Fprintf(os.Stderr, "[goo-log] 当前文件未打开\n")
		f.buffer = f.buffer[:0]
		return
	}

	// 写入缓冲区数据
	n, err := f.currentFile.Write(f.buffer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[goo-log] 写入日志失败: %v\n", err)
		// 清空缓冲区
		f.buffer = f.buffer[:0]
		// 尝试重新打开文件
		if err := f.rotateFile(dateStr); err != nil {
			fmt.Fprintf(os.Stderr, "[goo-log] 重新打开文件失败: %v\n", err)
		}
		return
	}

	f.currentSize += int64(n)

	// 写入后再次检查文件大小，如果超过限制，下次刷新时切换文件
	if f.currentSize >= f.maxSize {
		// 标记需要切换，但不立即切换（避免频繁切换）
		// 下次 flush 时会自动切换
	}

	// 清空缓冲区（保留容量）
	f.buffer = f.buffer[:0]
}

// flushRemaining 刷新剩余数据（关闭时调用）
func (f *FileAdapter) flushRemaining() {
	// 处理 channel 中剩余的数据
	for {
		select {
		case data := <-f.writeChan:
			f.buffer = append(f.buffer, data...)
		default:
			// channel 已空，刷新缓冲区
			f.flush()
			return
		}
	}
}

// rotateFile 切换文件（Lumberjack 风格）
// 参考 Lumberjack 实现：始终写入基础文件，达到大小后原子重命名
func (f *FileAdapter) rotateFile(dateStr string) error {
	// 关闭旧文件
	if f.currentFile != nil {
		f.currentFile.Close()
		f.currentFile = nil
	}

	// 生成基础文件名（无索引）
	baseFileName := f.generateBaseFileName(dateStr)
	baseFilePath := filepath.Join(f.dir, baseFileName)

	// 如果日期变化，重置并处理旧文件
	if f.currentDate != dateStr {
		// 如果存在旧日期的基础文件，重命名为索引文件
		if _, err := os.Stat(baseFilePath); err == nil {
			f.renameToIndexFile(baseFilePath, dateStr)
		}
		f.currentDate = dateStr
	} else {
		// 文件过大，需要轮转
		// 如果基础文件存在，重命名为索引文件（Lumberjack 方式）
		if _, err := os.Stat(baseFilePath); err == nil {
			f.renameToIndexFile(baseFilePath, dateStr)
		}
	}

	// 创建新的基础文件（始终使用基础文件名）
	file, err := os.OpenFile(baseFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %w", err)
	}

	f.currentFile = file
	f.currentPath = baseFilePath
	f.currentSize = 0
	return nil
}

// renameToIndexFile 将基础文件重命名为索引文件（原子操作）
func (f *FileAdapter) renameToIndexFile(baseFilePath, dateStr string) {
	// 找到最大索引
	maxIndex := f.findMaxIndex(dateStr)
	nextIndex := maxIndex + 1
	if maxIndex < 0 {
		nextIndex = 1
	}

	// 原子重命名：基础文件 → 索引文件
	indexFileName := f.generateFileNameWithIndex(dateStr, nextIndex)
	indexFilePath := filepath.Join(f.dir, indexFileName)
	if err := os.Rename(baseFilePath, indexFilePath); err != nil {
		// 重命名失败，删除旧文件（避免阻塞）
		os.Remove(baseFilePath)
	}
}

// generateBaseFileName 生成基础文件名（无索引）
func (f *FileAdapter) generateBaseFileName(dateStr string) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}
	return date.Format(f.fileName)
}

// findMaxIndex 找到当前日期的最大索引文件索引
// 返回最大索引，如果没有索引文件则返回 -1
func (f *FileAdapter) findMaxIndex(dateStr string) int {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}
	baseFileName := date.Format(f.fileName)
	ext := filepath.Ext(baseFileName)
	baseName := baseFileName[:len(baseFileName)-len(ext)]

	maxIndex := -1
	// 从索引1开始查找，找到最大的索引
	for i := 1; ; i++ {
		fileName := fmt.Sprintf("%s.%d%s", baseName, i, ext)
		filePath := filepath.Join(f.dir, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		maxIndex = i
	}

	return maxIndex
}

// generateFileNameWithIndex 生成指定索引的文件名
func (f *FileAdapter) generateFileNameWithIndex(dateStr string, index int) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}
	baseFileName := date.Format(f.fileName)
	ext := filepath.Ext(baseFileName)
	baseName := baseFileName[:len(baseFileName)-len(ext)]
	return fmt.Sprintf("%s.%d%s", baseName, index, ext)
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
	// 关闭停止信号，触发 writeWorker 退出
	close(f.stopChan)

	// 等待所有协程退出（包括 writeWorker 处理剩余数据）
	f.wg.Wait()

	f.mu.Lock()
	defer f.mu.Unlock()

	// 关闭当前文件
	if f.currentFile != nil {
		return f.currentFile.Close()
	}
	return nil
}
