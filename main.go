package main

import (
	"AiDemo/config"
	"AiDemo/handlers"
	initPkg "AiDemo/init"
	"AiDemo/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志系统
	if err := initPkg.InitLog(); err != nil {
		// 日志系统初始化失败，使用标准日志记录错误并退出
		log.Fatalf("日志系统初始化失败: %v", err)
	}
	defer initPkg.CloseLog()

	// 加载配置
	utils.Info("正在加载配置...")
	err := config.LoadEnv()
	if err != nil {
		utils.Fatal("加载配置失败: %v", err)
		return
	}
	utils.Info("配置加载完成")

	// 创建 Gin 引擎
	gin.SetMode(gin.ReleaseMode) // 生产环境可设置为 ReleaseMode
	r := gin.Default()

	// 静态文件（前端页面）
	r.Static("/web", "./web")
	utils.Info("静态文件路由已配置")

	// 默认首页跳转
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/web/index.html")
	})

	// 注册路由
	r.POST("/chat", handlers.ChatHandler)
	r.GET("/get-prompt", handlers.GetPromptHandler)
	r.POST("/set-prompt", handlers.SetPromptHandler)
	utils.Info("API路由已注册")

	// 启动提示
	utils.Info("🚀 服务已启动，请在浏览器访问: http://localhost:8080")

	// 启动服务
	err = r.Run(":8080")
	if err != nil {
		utils.Fatal("服务启动失败: %v", err)
		return
	}
}
