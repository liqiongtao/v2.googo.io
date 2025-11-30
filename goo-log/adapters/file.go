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
		c := config[0]
		if c.Dir != "" {
			cfg.Dir = c.Dir
		}
		if c.FileName != "" {
			cfg.FileName = c.FileName
		}
		if c.MaxSize > 0 {
			cfg.MaxSize = c.MaxSize
		}
		if c.RetainDays > 0 {
			cfg.RetainDays = c.RetainDays
		}
		cfg.UseJSON = c.UseJSON
		cfg.BufferSize = c.BufferSize
		cfg.FlushInterval = c.FlushInterval
		cfg.ChannelSize = c.ChannelSize
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
		fmt.Fprintf(os.Stderr, "[goo-log] 写入通道已满，丢弃日志: %s\n", string(data))
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
			// 批量接收数据，减少加锁次数
			f.mu.Lock()
			f.buffer = append(f.buffer, data...)
			bufferLen := len(f.buffer)
			f.mu.Unlock()

			// 如果缓冲区达到大小，立即刷新
			if bufferLen >= f.bufferSize {
				f.flush()
			} else {
				// 尝试批量接收更多数据（非阻塞）
				batchDone := false
				for i := 0; i < 10 && !batchDone; i++ { // 最多批量接收 10 条
					select {
					case moreData := <-f.writeChan:
						f.mu.Lock()
						f.buffer = append(f.buffer, moreData...)
						bufferLen = len(f.buffer)
						f.mu.Unlock()
						if bufferLen >= f.bufferSize {
							f.flush()
							batchDone = true
						}
					default:
						// 没有更多数据，退出批量接收
						batchDone = true
					}
				}
			}
		case <-flushTicker.C:
			// 定期刷新
			f.flush()
		}
	}
}

// flush 刷新缓冲区到文件
func (f *FileAdapter) flush() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 检查缓冲区是否为空
	if len(f.buffer) == 0 {
		return
	}

	f.flushLocked()
}

// flushLocked 刷新缓冲区到文件（内部方法，调用前必须持有 mu 锁）
func (f *FileAdapter) flushLocked() {
	// 复制缓冲区数据
	bufferData := make([]byte, len(f.buffer))
	copy(bufferData, f.buffer)
	f.buffer = f.buffer[:0] // 清空缓冲区（保留容量）

	// 检查是否需要切换文件
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	needRotate := f.currentFile == nil || f.currentDate != dateStr

	// 如果文件过大，需要切换文件
	if !needRotate && f.currentFile != nil {
		if f.currentSize+int64(len(bufferData)) >= f.maxSize {
			needRotate = true
		}
	}

	if needRotate {
		if err := f.rotateFileLocked(dateStr); err != nil {
			fmt.Fprintf(os.Stderr, "[goo-log] 文件切换失败: %v\n", err)
			return
		}
	}

	// 确保文件已打开
	if f.currentFile == nil {
		fmt.Fprintf(os.Stderr, "[goo-log] 当前文件未打开\n")
		return
	}

	// 写入数据
	n, err := f.currentFile.Write(bufferData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[goo-log] 写入日志失败: %v\n", err)
		// 尝试重新打开文件
		if err := f.rotateFileLocked(dateStr); err != nil {
			fmt.Fprintf(os.Stderr, "[goo-log] 重新打开文件失败: %v\n", err)
		} else {
			// 重新打开成功，尝试再次写入
			if f.currentFile != nil {
				if n, err := f.currentFile.Write(bufferData); err == nil {
					f.currentSize += int64(n)
				}
			}
		}
		return
	}

	f.currentSize += int64(n)
}

// flushRemaining 刷新剩余数据（关闭时调用）
func (f *FileAdapter) flushRemaining() {
	// 批量处理 channel 中剩余的数据
	f.mu.Lock()

	// 批量接收所有剩余数据
	for {
		select {
		case data := <-f.writeChan:
			f.buffer = append(f.buffer, data...)
		default:
			// channel 已空，退出循环
			goto flush
		}
	}

flush:
	// 如果缓冲区有数据，使用公共刷新逻辑（注意：此时已持有锁）
	if len(f.buffer) > 0 {
		f.flushLocked()
	}
	f.mu.Unlock()
}

// rotateFile 切换文件（Lumberjack 风格）
// 参考 Lumberjack 实现：始终写入基础文件，达到大小后原子重命名
// 注意：调用此函数前必须持有 mu 锁
func (f *FileAdapter) rotateFileLocked(dateStr string) error {
	// 关闭旧文件
	if f.currentFile != nil {
		f.currentFile.Close()
		f.currentFile = nil
	}

	// 生成基础文件名（无索引）
	baseFileName := f.generateBaseFileName(dateStr)
	baseFilePath := filepath.Join(f.dir, baseFileName)

	// 检查基础文件是否存在，如果存在则重命名为索引文件
	if _, err := os.Stat(baseFilePath); err == nil {
		f.renameToIndexFile(baseFilePath, dateStr)
	}

	// 更新当前日期
	if f.currentDate != dateStr {
		f.currentDate = dateStr
	}

	// 创建新的基础文件（始终使用基础文件名）
	file, err := os.OpenFile(baseFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %w", err)
	}

	f.currentFile = file
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
	date := f.parseDate(dateStr)
	return date.Format(f.fileName)
}

// parseDate 解析日期字符串
func (f *FileAdapter) parseDate(dateStr string) time.Time {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Now()
	}
	return date
}

// getBaseName 获取基础文件名（不含扩展名）
func (f *FileAdapter) getBaseName(dateStr string) (string, string) {
	date := f.parseDate(dateStr)
	baseFileName := date.Format(f.fileName)
	ext := filepath.Ext(baseFileName)
	baseName := baseFileName[:len(baseFileName)-len(ext)]
	return baseName, ext
}

// findMaxIndex 找到当前日期的最大索引文件索引
// 返回最大索引，如果没有索引文件则返回 -1
func (f *FileAdapter) findMaxIndex(dateStr string) int {
	baseName, ext := f.getBaseName(dateStr)
	baseFileName := baseName + ext

	// 读取目录，查找所有匹配的索引文件
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return -1
	}

	maxIndex := -1
	prefix := baseName + "."
	suffix := ext

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()

		// 跳过基础文件本身
		if fileName == baseFileName {
			continue
		}

		// 检查是否是索引文件：baseName.index.ext
		if strings.HasPrefix(fileName, prefix) && strings.HasSuffix(fileName, suffix) {
			// 提取索引部分
			indexStr := fileName[len(prefix) : len(fileName)-len(suffix)]
			var index int
			if n, _ := fmt.Sscanf(indexStr, "%d", &index); n == 1 && index > 0 {
				if index > maxIndex {
					maxIndex = index
				}
			}
		}
	}

	return maxIndex
}

// generateFileNameWithIndex 生成指定索引的文件名
func (f *FileAdapter) generateFileNameWithIndex(dateStr string, index int) string {
	baseName, ext := f.getBaseName(dateStr)
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
	f.walkLogFiles(func(filePath string, fileName string, info os.FileInfo) {
		// 只处理未压缩的日志文件
		if !strings.HasSuffix(fileName, ".gz") && info.ModTime().Before(cutoffDate) {
			select {
			case f.compressChan <- filePath:
			default:
			}
		}
	})
}

// walkLogFiles 遍历日志文件（公共函数）
func (f *FileAdapter) walkLogFiles(fn func(filePath, fileName string, info os.FileInfo)) {
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		// 只处理日志文件
		if !strings.HasSuffix(fileName, ".log") && !strings.HasSuffix(fileName, ".log.gz") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(f.dir, fileName)
		fn(filePath, fileName, info)
	}
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
	f.walkLogFiles(func(filePath string, fileName string, info os.FileInfo) {
		// 只处理已压缩的日志文件
		if strings.HasSuffix(fileName, ".log.gz") && info.ModTime().Before(cutoffDate) {
			os.Remove(filePath)
		}
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
