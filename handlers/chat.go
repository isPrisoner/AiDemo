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
		Message      string `json:"message"`
		SystemPrompt string `json:"systemPrompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Warning("请求参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	utils.Info("收到用户消息: %s", req.Message)

	// 如果提供了自定义提示词，使用它替换系统提示词
	if req.SystemPrompt != "" {
		utils.Info("使用自定义提示词")
		mu.Lock()
		// 保存旧的历史记录
		oldHistory := services.History
		// 创建新的历史记录，使用自定义提示词
		newHistory := []models.Message{
			{Role: "system", Content: req.SystemPrompt},
		}
		// 添加用户消息到新的历史记录
		newHistory = append(newHistory, models.Message{Role: "user", Content: req.Message})
		// 临时替换历史记录
		services.History = newHistory
		mu.Unlock()

		utils.Debug("开始调用AI服务...")
		respText, err := services.CallDoubao(services.History)
		if err != nil {
			utils.Error("AI服务调用失败: %v", err)
			// 恢复原始历史记录
			mu.Lock()
			services.History = oldHistory
			mu.Unlock()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.Debug("AI服务响应成功，长度: %d", len(respText))

		// 更新系统提示词
		mu.Lock()
		// 保存新的历史记录，包含系统提示词、用户消息和AI回复
		services.History = []models.Message{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.Message},
			{Role: "assistant", Content: respText},
		}
		mu.Unlock()

		utils.Info("返回AI回复给用户")
		c.JSON(http.StatusOK, gin.H{"reply": respText})
		return
	}

	// 使用默认提示词
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

// GetPromptHandler 获取当前系统提示词
func GetPromptHandler(c *gin.Context) {
	mu.Lock()
	var systemPrompt string
	for _, msg := range services.History {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			break
		}
	}
	mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"prompt": systemPrompt})
}

// SetPromptHandler 设置系统提示词
func SetPromptHandler(c *gin.Context) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Warning("请求参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "提示词不能为空"})
		return
	}

	utils.Info("设置新的系统提示词")
	mu.Lock()
	// 找到并更新系统提示词
	systemPromptUpdated := false
	for i, msg := range services.History {
		if msg.Role == "system" {
			services.History[i].Content = req.Prompt
			systemPromptUpdated = true
			break
		}
	}

	// 如果没有找到系统提示词，添加一个
	if !systemPromptUpdated {
		// 创建新的历史记录
		newHistory := []models.Message{
			{Role: "system", Content: req.Prompt},
		}
		// 添加旧的历史记录（除了可能存在的系统提示词）
		for _, msg := range services.History {
			if msg.Role != "system" {
				newHistory = append(newHistory, msg)
			}
		}
		services.History = newHistory
	}
	mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"success": true})
}
