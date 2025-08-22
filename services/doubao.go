package services

import (
	"AiDemo/config"
	"AiDemo/models"
	"AiDemo/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func CallDoubao(messages []models.Message) (string, error) {
	url := "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
	utils.Debug("准备调用API: %s", url)

	body := models.RequestBody{
		Model:    "ep-20250811150312-h4mvh", // 你的模型 ID
		Messages: messages,
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		utils.Error("请求体序列化失败: %v", err)
		return "", err
	}

	utils.Debug("API请求体: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Error("创建HTTP请求失败: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	utils.Debug("HTTP请求头已设置")

	client := &http.Client{}
	utils.Info("发送API请求...")
	resp, err := client.Do(req)
	if err != nil {
		utils.Error("HTTP请求失败: %v", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			utils.Warning("关闭响应体失败: %v", err)
		}
	}(resp.Body)

	utils.Info("API响应状态码: %d", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Error("读取响应体失败: %v", err)
		return "", err
	}

	utils.Debug("API原始响应: %s", string(respBody))

	var response models.ResponseBody
	if err := json.Unmarshal(respBody, &response); err != nil {
		utils.Error("解析响应JSON失败: %v", err)
		return "", err
	}

	if len(response.Choices) > 0 {
		content := response.Choices[0].Message.Content
		utils.Info("API调用成功，返回内容长度: %d", len(content))
		return content, nil
	}

	utils.Error("API返回空结果")
	return "", fmt.Errorf("API返回空结果")
}
