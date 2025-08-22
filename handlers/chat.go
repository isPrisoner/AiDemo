package handlers

import (
	"AiDemo/models"
	"AiDemo/services"
	"AiDemo/utils"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 角色 -> 系统提示词
var roleSystemPrompts = map[string]string{
	"general":    "你是一个专业、友善且简洁的中文AI助理。要求：1) 理解用户真实意图，优先给出可执行答案；2) 回答清晰分点，必要时给示例；3) 不编造事实，未知则说明并给出获取方法；4) 默认使用简体中文；5) 保持礼貌且不啰嗦。",
	"coder":      "你是资深全栈工程师与代码审阅者。要求：1) 以问题为导向，提供可运行代码与关键说明；2) 代码风格清晰、命名规范、错误处理完善；3) 指出潜在边界条件与复杂度；4) 能根据上下文给出重构建议；5) 输出中避免无意义的客套。默认中文回答。",
	"translator": "你是专业中英互译员。要求：1) 优先保证语义准确，其次流畅自然；2) 根据语境选择直译或意译；3) 保留专有名词与技术术语；4) 提供1-2种可选表达以供选择；5) 如用户未说明目标语言，优先中译英。",
	"pm":         "你是资深产品经理。要求：1) 澄清目标、用户、场景与约束；2) 以列表与结构化表达需求；3) 补充验收标准与关键KPI；4) 提供里程碑与风险缓解建议；5) 如问题含糊，先反问澄清。",
	"scholar":    "你是学术写作与研究助手。要求：1) 用严谨学术语气组织内容；2) 先给提纲再展开；3) 引入必要定义、公式或参考路径；4) 强调方法、数据与限制；5) 避免臆测，必要时提示需查证。默认中文。",
}

func ChatHandler(c *gin.Context) {
	var req struct {
		Message   string `json:"message"`
		Role      string `json:"role"`
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Warning("请求参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	role := req.Role
	if role == "" {
		role = "general"
	}
	sysPrompt, ok := roleSystemPrompts[role]
	if !ok {
		sysPrompt = roleSystemPrompts["general"]
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = genSessionID()
	}

	utils.Info("收到用户消息: %s (role=%s, session=%s)", req.Message, role, sessionID)

	// 初始化会话（若不存在）
	if !services.HasSession(sessionID) {
		services.ResetSession(sessionID, sysPrompt)
	}
	// 追加用户消息
	services.AppendMessage(sessionID, models.Message{Role: "user", Content: req.Message})

	// 调用AI服务
	utils.Debug("开始调用AI服务...")
	respText, err := services.CallDoubao(services.GetHistory(sessionID))
	if err != nil {
		utils.Error("AI服务调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	utils.Debug("AI服务响应成功，长度: %d", len(respText))

	// 记录助手回复
	services.AppendMessage(sessionID, models.Message{Role: "assistant", Content: respText})

	utils.Info("返回AI回复给用户")
	c.JSON(http.StatusOK, gin.H{"reply": respText, "session_id": sessionID})
}

func genSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "sess"
	}
	return hex.EncodeToString(b)
}
