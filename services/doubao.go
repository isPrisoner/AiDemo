package services

import (
	"AiDemo/config"
	"AiDemo/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var History []models.Message

func init() {
	// 初始化对话历史
	History = append(History, models.Message{Role: "system", Content: "你是一个乐于助人的AI助手"})
}

func CallDoubao(messages []models.Message) (string, error) {
	url := "https://ark.cn-beijing.volces.com/api/v3/chat/completions"

	body := models.RequestBody{
		Model:    "ep-20250811150312-h4mvh", // 你的模型 ID
		Messages: messages,
	}
	jsonData, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	respBody, _ := io.ReadAll(resp.Body)

	var response models.ResponseBody
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", err
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("API返回空结果")
}
