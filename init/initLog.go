package init

import (
	"AiDemo/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// InitLog 初始化日志系统
func InitLog() error {
	// 创建日志目录
	logDir := "./logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 设置日志文件路径
	logFile := filepath.Join(logDir, "app.log")
	err = utils.SetLogFile(logFile)
	if err != nil {
		utils.Error("设置日志文件失败: %v", err)
		// 即使设置文件失败，也继续使用控制台输出
	}

	// 启用日志轮转（按天轮转）
	utils.EnableRotate()

	// 设置日志级别（可根据环境变量或配置文件调整）
	utils.SetLevel(utils.INFO)

	// 启用异步日志写入（缓冲区大小为1000，刷新间隔为3秒）
	utils.EnableAsync(1000, 3*time.Second)

	// 设置日志格式（默认为文本格式，可选JSON格式）
	// utils.SetFormat(utils.JsonFormat) // 取消注释启用JSON格式

	utils.Info("日志系统初始化完成，已启用按天轮转和异步写入")
	return nil
}

// CloseLog 关闭日志系统
func CloseLog() {
	utils.Info("正在关闭日志系统...")
	utils.Close()
}
