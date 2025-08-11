package main

import (
	"bufio"                    // 读取终端输入
	"bytes"                    // 处理字节数据
	"encoding/json"            // JSON 编解码
	"fmt"                      // 输出
	"github.com/joho/godotenv" // 读取 .env 文件
	"io"                       // 读写数据流
	"net/http"                 // 发送 HTTP 请求
	"os"                       // 访问系统文件、环境变量
	"strings"                  // 字符串处理
)

// Message 消息结构体（每条对话消息）
type Message struct {
	Role    string `json:"role"`    // 消息角色：system / user / assistant
	Content string `json:"content"` // 消息内容
}

// RequestBody 请求体结构
type RequestBody struct {
	Model    string    `json:"model"`    // 使用的模型名称
	Messages []Message `json:"messages"` // 对话消息历史
}

// Choice API 返回结果中的选项
type Choice struct {
	Message Message `json:"message"`
}

// ResponseBody API 响应体
type ResponseBody struct {
	Choices []Choice `json:"choices"`
}

func main() {
	// 1. 加载 .env 文件（读取 API Key）
	// 如果 .env 文件在其他目录，比如 init/.env，可以改成 godotenv.Load("init/.env")
	err := godotenv.Load("init/init.env")
	if err != nil {
		fmt.Println("加载.env文件失败:", err)
		return
	}

	// 从环境变量中获取 API Key
	apiKey := os.Getenv("DOUBAO_API_KEY")
	if apiKey == "" {
		fmt.Println("请先在.env文件中设置 DOUBAO_API_KEY")
		return
	}

	// 2. 初始化对话历史
	var history []Message
	history = append(history, Message{"system", "你是一个乐于助人的AI助手"})

	// 3. 读取用户输入
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n你: ")
		userInput, _ := reader.ReadString('\n')  // 读取一行输入
		userInput = strings.TrimSpace(userInput) // 去掉空格和换行

		// 输入 exit 则退出程序
		if userInput == "exit" {
			break
		}

		// 用户消息加入对话历史
		history = append(history, Message{"user", userInput})

		// 4. 调用豆包 API 获取回复
		respText, err := callDoubao(apiKey, history)
		if err != nil {
			fmt.Println("调用出错:", err)
			continue
		}

		// 输出 AI 回复
		fmt.Println("AI:", respText)

		// AI 回复加入历史
		history = append(history, Message{"assistant", respText})
	}
}

// 调用豆包 API 的函数
func callDoubao(apiKey string, messages []Message) (string, error) {
	url := "https://ark.cn-beijing.volces.com/api/v3/chat/completions"

	// 创建请求体
	body := RequestBody{
		Model:    "ep-20250811150312-h4mvh",
		Messages: messages,
	}
	jsonData, _ := json.Marshal(body) // 序列化成 JSON

	// 创建 HTTP 请求
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey) // 认证信息

	// 发送请求
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

	// 读取响应体
	respBody, _ := io.ReadAll(resp.Body)

	// 解析 JSON
	var response ResponseBody
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return "", err
	}

	// 返回第一条 AI 回复
	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("API返回空结果")
}
