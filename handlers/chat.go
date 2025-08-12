package handlers

import (
	"AiDemo/models"
	"AiDemo/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

var mu sync.Mutex // 锁，防止并发修改历史

func ChatHandler(c *gin.Context) {
	var req struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	mu.Lock()
	services.History = append(services.History, models.Message{Role: "user", Content: req.Message})
	mu.Unlock()

	respText, err := services.CallDoubao(services.History)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	services.History = append(services.History, models.Message{Role: "assistant", Content: respText})
	mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"reply": respText})
}
