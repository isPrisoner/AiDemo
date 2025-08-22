package services

import (
	"AiDemo/models"
	"sync"
)

// 会话历史管理（内存版）
var (
	sessionHistories = make(map[string][]models.Message)
	sessionsMu       sync.RWMutex
)

// ResetSession 用系统提示词重置/初始化指定session的历史
func ResetSession(sessionID string, systemPrompt string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	sessionHistories[sessionID] = []models.Message{{Role: "system", Content: systemPrompt}}
}

// AppendMessage 向指定session追加一条消息
func AppendMessage(sessionID string, msg models.Message) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	sessionHistories[sessionID] = append(sessionHistories[sessionID], msg)
	trimHistoryIfTooLong(sessionID)
}

// GetHistory 返回指定session的全部历史（拷贝）
func GetHistory(sessionID string) []models.Message {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	h := sessionHistories[sessionID]
	// 返回一个拷贝，避免外部修改内部切片
	copied := make([]models.Message, len(h))
	copy(copied, h)
	return copied
}

// HasSession 判断是否已有该session的历史
func HasSession(sessionID string) bool {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	_, ok := sessionHistories[sessionID]
	return ok
}

// 简单长度裁剪，避免历史无限增长（按消息条数裁剪，保留system）
func trimHistoryIfTooLong(sessionID string) {
	const maxMessages = 30 // 包含 user/assistant，不含system约束；可按需调整
	h := sessionHistories[sessionID]
	if len(h) <= maxMessages+1 { // +1 预留给system
		return
	}
	// 保留第0条system，从尾部开始保留最近的maxMessages条
	start := len(h) - maxMessages
	trimmed := append([]models.Message{h[0]}, h[start:]...)
	sessionHistories[sessionID] = trimmed
}
