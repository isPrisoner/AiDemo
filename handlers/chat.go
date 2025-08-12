package handlers

import (
	"AiDemo/models"
	"AiDemo/services"
	"AiDemo/utils"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var mu sync.Mutex // 锁，防止并发修改历史

func ChatHandler(c *gin.Context) {
	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Warning("请求参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	utils.Info("收到用户消息: %s", req.Message)

	mu.Lock()
	services.History = append(services.History, models.Message{Role: "user", Content: req.Message})
	mu.Unlock()

	utils.Debug("开始调用AI服务...")
	respText, err := services.CallDoubao(services.History)
	if err != nil {
		utils.Error("AI服务调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	utils.Debug("AI服务响应成功，长度: %d", len(respText))

	mu.Lock()
	services.History = append(services.History, models.Message{Role: "assistant", Content: respText})
	mu.Unlock()

	utils.Info("返回AI回复给用户")
	c.JSON(http.StatusOK, gin.H{"reply": respText})
}
