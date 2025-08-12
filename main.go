package main

import (
	"AiDemo/config"
	"AiDemo/handlers"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// 加载配置
	config.LoadEnv()

	// 创建 Gin 引擎
	r := gin.Default()

	// 静态文件（前端页面）
	r.Static("/web", "./web")

	// 默认首页跳转
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/web/index.html")
	})

	// 注册路由
	r.POST("/chat", handlers.ChatHandler)

	// 启动提示
	fmt.Println("🚀 服务已启动，请在浏览器访问: http://localhost:8080")

	// 启动服务
	err := r.Run(":8080")
	if err != nil {
		return
	}
}
