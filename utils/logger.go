package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 日志级别常量
const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	FATAL
)

// 日志级别名称
var levelNames = []string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"FATAL",
}

// LogFormat 日志格式类型
const (
	TextFormat = iota // 文本格式
	JsonFormat        // JSON格式
)

// 默认缓冲区大小
const (
	DefaultBufferSize = 1000 // 默认日志缓冲区大小
	FlushInterval     = 3    // 默认刷新间隔（秒）
)

// LogEntry 结构化日志条目
type LogEntry struct {
	Level     string                 `json:"level"`
	Timestamp string                 `json:"timestamp"`
	Message   string                 `json:"message"`
	Caller    string                 `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// logMessage 内部日志消息结构
type logMessage struct {
	level     int
	format    string
	args      []interface{}
	timestamp time.Time
	caller    string
	fields    map[string]interface{}
}

// Logger 日志记录器结构体
type Logger struct {
	level          int
	logFile        *os.File
	logger         *log.Logger
	showCaller     bool
	logDir         string
	baseFileName   string
	currentLogFile string
	rotateEnabled  bool
	lastRotateTime time.Time
	mutex          sync.Mutex
	format         int // 日志格式

	// 异步日志相关
	asyncEnabled  bool
	logChan       chan *logMessage
	waitGroup     sync.WaitGroup
	flushInterval time.Duration
	bufferSize    int
	stopChan      chan struct{}
}

var defaultLogger *Logger

// 初始化默认日志记录器
func init() {
	defaultLogger = NewLogger(INFO, "", true)
	if defaultLogger == nil {
		// 如果创建失败，使用基本配置创建一个简单的记录器
		defaultLogger = &Logger{
			level:      INFO,
			logger:     log.New(os.Stdout, "", 0),
			showCaller: true,
			format:     TextFormat,
		}
	}
}

// NewLogger 创建新的日志记录器
func NewLogger(level int, logFilePath string, showCaller bool) *Logger {
	var writer io.Writer = os.Stdout
	var logFile *os.File
	var err error
	var logDir, baseFileName, currentLogFile string

	// 如果指定了日志文件路径，同时写入文件和标准输出
	if logFilePath != "" {
		// 确保日志目录存在
		logDir = filepath.Dir(logFilePath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			err := os.MkdirAll(logDir, 0755)
			if err != nil {
				fmt.Printf("创建日志目录失败: %v\n", err)
				return nil
			}
		}

		baseFileName = filepath.Base(logFilePath)
		ext := filepath.Ext(baseFileName)
		baseFileName = strings.TrimSuffix(baseFileName, ext) + ext

		currentLogFile = logFilePath
		logFile, err = os.OpenFile(currentLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			writer = io.MultiWriter(os.Stdout, logFile)
		} else {
			fmt.Printf("无法打开日志文件: %v，将只输出到控制台\n", err)
		}
	}

	logger := log.New(writer, "", 0)

	return &Logger{
		level:          level,
		logFile:        logFile,
		logger:         logger,
		showCaller:     showCaller,
		logDir:         logDir,
		baseFileName:   baseFileName,
		currentLogFile: currentLogFile,
		rotateEnabled:  false,
		lastRotateTime: time.Now(),
		mutex:          sync.Mutex{},
		format:         TextFormat, // 默认为文本格式
		asyncEnabled:   false,
		bufferSize:     DefaultBufferSize,
		flushInterval:  FlushInterval * time.Second,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level int) {
	if level >= DEBUG && level <= FATAL {
		l.level = level
	}
}

// SetLevel 设置默认日志记录器的日志级别
func SetLevel(level int) {
	defaultLogger.SetLevel(level)
}

// SetFormat 设置日志格式
func (l *Logger) SetFormat(format int) {
	if format == TextFormat || format == JsonFormat {
		l.format = format
	}
}

// SetFormat 设置默认日志记录器的日志格式
func SetFormat(format int) {
	defaultLogger.SetFormat(format)
}

// EnableAsync 启用异步日志
func (l *Logger) EnableAsync(bufferSize int, flushInterval time.Duration) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.asyncEnabled {
		return // 已经启用
	}

	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}

	if flushInterval <= 0 {
		flushInterval = FlushInterval * time.Second
	}

	l.asyncEnabled = true
	l.bufferSize = bufferSize
	l.flushInterval = flushInterval
	l.logChan = make(chan *logMessage, bufferSize)
	l.stopChan = make(chan struct{})

	// 启动后台处理goroutine
	l.waitGroup.Add(1)
	go l.processLogs()
}

// EnableAsync 启用默认日志记录器的异步日志
func EnableAsync(bufferSize int, flushInterval time.Duration) {
	defaultLogger.EnableAsync(bufferSize, flushInterval)
}

// DisableAsync 禁用异步日志
func (l *Logger) DisableAsync() {
	l.mutex.Lock()

	if !l.asyncEnabled {
		l.mutex.Unlock()
		return // 已经禁用
	}

	l.asyncEnabled = false
	close(l.stopChan) // 通知处理goroutine停止
	l.mutex.Unlock()

	// 等待所有日志处理完成
	l.waitGroup.Wait()
}

// DisableAsync 禁用默认日志记录器的异步日志
func DisableAsync() {
	defaultLogger.DisableAsync()
}

// Flush 刷新异步日志缓冲区
func (l *Logger) Flush() {
	if !l.asyncEnabled {
		return
	}

	// 创建一个刷新通知通道
	flushChan := make(chan struct{})

	// 发送刷新消息
	select {
	case l.logChan <- &logMessage{
		level: -1, // 特殊级别表示刷新操作
		args:  []interface{}{flushChan},
	}:
		// 成功发送到通道
		// 等待刷新完成
		<-flushChan
	default:
		// 通道已满，无法发送刷新消息
		// 这种情况下不等待，直接返回
	}
}

// Flush 刷新默认日志记录器的异步日志缓冲区
func Flush() {
	defaultLogger.Flush()
}

// processLogs 处理异步日志的goroutine
func (l *Logger) processLogs() {
	defer l.waitGroup.Done()

	ticker := time.NewTicker(l.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-l.logChan:
			if !ok {
				return // 通道已关闭
			}

			// 处理刷新操作
			if msg.level == -1 && len(msg.args) > 0 {
				if flushChan, ok := msg.args[0].(chan struct{}); ok {
					close(flushChan) // 通知刷新完成
				}
				continue
			}

			// 正常日志处理
			l.writeLog(msg)

		case <-ticker.C:
			// 定期检查是否需要轮转
			l.checkAndRotate()

		case <-l.stopChan:
			// 处理剩余日志
			l.drainLogChannel()
			return
		}
	}
}

// drainLogChannel 处理通道中剩余的日志
func (l *Logger) drainLogChannel() {
	for {
		select {
		case msg, ok := <-l.logChan:
			if !ok {
				return
			}
			l.writeLog(msg)
		default:
			// 通道已清空
			close(l.logChan)
			return
		}
	}
}

// checkAndRotate 检查并执行日志轮转
func (l *Logger) checkAndRotate() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.checkRotate() {
		l.rotate()
	}
}

// SetLogFile 设置日志文件
func SetLogFile(logFilePath string) error {
	if defaultLogger.logFile != nil {
		err := defaultLogger.logFile.Close()
		if err != nil {
			return fmt.Errorf("关闭当前日志文件失败: %w", err)
		}
	}

	logger := NewLogger(defaultLogger.level, logFilePath, defaultLogger.showCaller)
	if logger == nil {
		return fmt.Errorf("创建新的日志记录器失败")
	}

	if logger.logFile == nil && logFilePath != "" {
		return fmt.Errorf("无法打开日志文件: %s", logFilePath)
	}

	// 保留原有的异步设置
	wasAsync := defaultLogger.asyncEnabled
	asyncBufferSize := defaultLogger.bufferSize
	asyncFlushInterval := defaultLogger.flushInterval

	// 如果原来是异步的，先禁用
	if wasAsync {
		defaultLogger.DisableAsync()
	}

	defaultLogger = logger

	// 如果原来是异步的，重新启用
	if wasAsync {
		defaultLogger.EnableAsync(asyncBufferSize, asyncFlushInterval)
	}

	return nil
}

// EnableRotate 启用日志轮转
func (l *Logger) EnableRotate() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.rotateEnabled = true
	l.lastRotateTime = time.Now()

	// 立即执行一次轮转，确保使用正确的文件名
	if l.logFile != nil {
		l.rotate()
	}
}

// EnableRotate 启用默认日志记录器的日志轮转
func EnableRotate() {
	defaultLogger.EnableRotate()
}

// DisableRotate 禁用日志轮转
func (l *Logger) DisableRotate() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.rotateEnabled = false
}

// DisableRotate 禁用默认日志记录器的日志轮转
func DisableRotate() {
	defaultLogger.DisableRotate()
}

// 生成轮转后的日志文件名
func (l *Logger) getRotatedFileName() string {
	if l.baseFileName == "" {
		return ""
	}

	now := time.Now()
	ext := filepath.Ext(l.baseFileName)
	baseName := strings.TrimSuffix(l.baseFileName, ext)

	// 按天格式化日期
	timeFormat := "2006-01-02"

	return filepath.Join(l.logDir, fmt.Sprintf("%s.%s%s", baseName, now.Format(timeFormat), ext))
}

// 检查是否需要轮转
func (l *Logger) checkRotate() bool {
	if !l.rotateEnabled || l.logFile == nil {
		return false
	}

	now := time.Now()

	// 检查日期是否变化（按天轮转）
	return now.Day() != l.lastRotateTime.Day() ||
		now.Month() != l.lastRotateTime.Month() || now.Year() != l.lastRotateTime.Year()
}

// 执行日志轮转
func (l *Logger) rotate() {
	var err error

	if l.logFile == nil {
		return
	}

	// 关闭当前日志文件
	err = l.logFile.Close()
	if err != nil {
		fmt.Printf("关闭日志文件失败: %v\n", err)
		return
	}

	// 生成新的日志文件名
	newLogFile := l.getRotatedFileName()
	if newLogFile == "" {
		return
	}

	// 打开新的日志文件
	l.logFile, err = os.OpenFile(newLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("轮转日志文件失败: %v，将只输出到控制台\n", err)
		l.logFile = nil
		l.logger = log.New(os.Stdout, "", 0)
	} else {
		l.logger = log.New(io.MultiWriter(os.Stdout, l.logFile), "", 0)
		l.currentLogFile = newLogFile
	}

	// 更新最后轮转时间
	l.lastRotateTime = time.Now()
}

// getCaller 获取调用者信息
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	// 只显示文件名，不显示完整路径
	short := filepath.Base(file)
	return fmt.Sprintf("%s:%d", short, line)
}

// log 记录日志的内部方法
func (l *Logger) log(level int, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// 获取调用者信息
	var callerInfo string
	if l.showCaller {
		callerInfo = getCaller(3) // 跳过自身调用栈
	}

	// 如果启用异步，将日志消息发送到通道
	if l.asyncEnabled && level != FATAL { // 致命错误仍然同步处理
		select {
		case l.logChan <- &logMessage{
			level:     level,
			format:    format,
			args:      args,
			timestamp: time.Now(),
			caller:    callerInfo,
		}:
			// 成功发送到通道
		default:
			// 通道已满，回退到同步写入
			l.writeLogSync(level, format, callerInfo, args...)
		}

		// 如果是致命错误，等待日志写入完成后退出
		if level == FATAL {
			l.Flush()
			os.Exit(1)
		}
	} else {
		// 同步写入日志
		l.writeLogSync(level, format, callerInfo, args...)

		// 如果是致命错误，程序退出
		if level == FATAL {
			os.Exit(1)
		}
	}
}

// writeLogSync 同步写入日志
func (l *Logger) writeLogSync(level int, format, callerInfo string, args ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 检查是否需要轮转
	if l.checkRotate() {
		l.rotate()
	}

	msg := &logMessage{
		level:     level,
		format:    format,
		args:      args,
		timestamp: time.Now(),
		caller:    callerInfo,
	}

	l.writeLog(msg)
}

// writeLog 实际写入日志的方法
func (l *Logger) writeLog(msg *logMessage) {
	if l.logger == nil {
		// 如果logger为空，使用标准输出
		fmt.Printf("[%s] %s %s\n", levelNames[msg.level], msg.timestamp.Format("2006-01-02 15:04:05.000"), msg.format)
		return
	}

	timestamp := msg.timestamp.Format("2006-01-02 15:04:05.000")
	var content string

	if msg.format == "" {
		content = fmt.Sprint(msg.args...)
	} else {
		content = fmt.Sprintf(msg.format, msg.args...)
	}

	// 检查最后一个参数是否为字段映射
	var fields map[string]interface{}
	if len(msg.args) > 0 {
		if f, ok := msg.args[len(msg.args)-1].(map[string]interface{}); ok && msg.format != "" {
			fields = f
		}
	}

	// 根据格式输出日志
	if l.format == JsonFormat {
		entry := LogEntry{
			Level:     levelNames[msg.level],
			Timestamp: timestamp,
			Message:   content,
			Fields:    fields,
		}

		if l.showCaller && msg.caller != "" {
			entry.Caller = msg.caller
		}

		jsonData, err := json.Marshal(entry)
		if err != nil {
			// 如果JSON序列化失败，回退到文本格式
			l.logger.Print(fmt.Sprintf("[%s] %s [%s] %s", levelNames[msg.level], timestamp, msg.caller, content))
		} else {
			l.logger.Print(string(jsonData))
		}
	} else {
		// 文本格式
		callerStr := ""
		if l.showCaller && msg.caller != "" {
			callerStr = " [" + msg.caller + "]"
		}
		l.logger.Printf("[%s] %s%s %s", levelNames[msg.level], timestamp, callerStr, content)
	}
}

// Debug 调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warning 警告级别日志
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(WARNING, format, args...)
}

// Error 错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal 致命错误级别日志，记录后程序退出
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// Debug 默认日志记录器的便捷方法
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warning(format string, args ...interface{}) {
	defaultLogger.Warning(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// Close 关闭日志文件
func (l *Logger) Close() {
	// 如果启用了异步日志，先禁用它
	if l.asyncEnabled {
		l.DisableAsync()
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.logFile != nil {
		err := l.logFile.Close()
		if err != nil {
			fmt.Printf("关闭日志文件失败: %v\n", err)
			return
		}
		l.logFile = nil
	}
}

// Close 关闭默认日志记录器
func Close() {
	defaultLogger.Close()
}
