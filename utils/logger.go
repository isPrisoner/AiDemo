package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
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

// Logger 日志记录器结构体
type Logger struct {
	level      int
	logFile    *os.File
	logger     *log.Logger
	showCaller bool
}

var defaultLogger *Logger

// 初始化默认日志记录器
func init() {
	defaultLogger = NewLogger(INFO, "", true)
}

// NewLogger 创建新的日志记录器
func NewLogger(level int, logFilePath string, showCaller bool) *Logger {
	var writer io.Writer = os.Stdout
	var logFile *os.File
	var err error

	// 如果指定了日志文件路径，同时写入文件和标准输出
	if logFilePath != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(logFilePath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			os.MkdirAll(logDir, 0755)
		}

		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			writer = io.MultiWriter(os.Stdout, logFile)
		} else {
			fmt.Printf("无法打开日志文件: %v，将只输出到控制台\n", err)
		}
	}

	logger := log.New(writer, "", 0)

	return &Logger{
		level:      level,
		logFile:    logFile,
		logger:     logger,
		showCaller: showCaller,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level int) {
	if level >= DEBUG && level <= FATAL {
		l.level = level
	}
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() int {
	return l.level
}

// SetLevel 设置默认日志记录器的日志级别
func SetLevel(level int) {
	defaultLogger.SetLevel(level)
}

// SetLogFile 设置日志文件
func SetLogFile(logFilePath string) error {
	if defaultLogger.logFile != nil {
		defaultLogger.logFile.Close()
	}

	logger := NewLogger(defaultLogger.level, logFilePath, defaultLogger.showCaller)
	if logger.logFile == nil && logFilePath != "" {
		return fmt.Errorf("无法打开日志文件: %s", logFilePath)
	}

	defaultLogger = logger
	return nil
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

	now := time.Now().Format("2006-01-02 15:04:05.000")
	var msg string

	if format == "" {
		msg = fmt.Sprint(args...)
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	var callerInfo string
	if l.showCaller {
		callerInfo = getCaller(3) // 跳过自身调用栈
		callerInfo = " [" + callerInfo + "]"
	}

	l.logger.Printf("[%s] %s%s %s", levelNames[level], now, callerInfo, msg)

	// 如果是致命错误，程序退出
	if level == FATAL {
		os.Exit(1)
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

// 默认日志记录器的便捷方法
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
	if l.logFile != nil {
		l.logFile.Close()
	}
}

// Close 关闭默认日志记录器
func Close() {
	defaultLogger.Close()
}
