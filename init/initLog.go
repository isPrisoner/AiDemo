package init

import (
	"AiDemo/utils"
	"os"
	"path/filepath"
)

// InitLog 初始化日志系统
func InitLog() {
	// 创建日志目录
	logDir := "./logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		utils.Error("创建日志目录失败: %v", err)
		return
	}

	// 设置日志文件路径
	logFile := filepath.Join(logDir, "app.log")
	err = utils.SetLogFile(logFile)
	if err != nil {
		utils.Error("设置日志文件失败: %v", err)
	}

	// 设置日志级别（可根据环境变量或配置文件调整）
	utils.SetLevel(utils.INFO)
	utils.Info("日志系统初始化完成")
}

// CloseLog 关闭日志系统
func CloseLog() {
	utils.Info("正在关闭日志系统...")
	utils.Close()
}
