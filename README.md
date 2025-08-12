# AiDemo - 豆包AI聊天应用

一个基于Go语言和Gin框架开发的AI聊天应用，使用豆包AI API实现智能对话功能。

## 项目架构

```
AiDemo/
  ├── config/          # 配置管理
  ├── handlers/        # HTTP请求处理器
  ├── init/            # 初始化和环境变量配置
  ├── models/          # 数据模型
  ├── services/        # 业务逻辑服务
  ├── utils/           # 工具类（日志系统等）
  └── web/             # 前端页面
```

## 主要功能

- 基于会话ID的多用户聊天
- 完整的日志记录系统（支持按天轮转）
- 基于环境的配置管理
- RESTful API接口
- 跨域资源共享支持
- 统一的错误处理

## 技术栈

- Go 1.24.4
- Gin Web框架
- 自定义日志系统
- 豆包AI API

## 快速开始

### 环境要求

- Go 1.24+
- 豆包AI API密钥

### 安装与运行

1. 克隆项目

```bash
git clone https://github.com/isPrisoner/AiDemo.git
cd AiDemo
```

2. 安装依赖

```bash
go mod tidy
```

3. 配置环境

编辑 `init/init.env` 文件，设置您的API密钥：

```
DOUBAO_API_KEY=YOUR_API_KEY
```

4. 运行应用

```bash
go run main.go
```

应用将在 http://localhost:8080 上启动。

## API接口

### 聊天接口

**POST /chat**

请求体:
```json
{
  "message": "你好，AI",
  "session_id": "optional-session-id"
}
```

响应:
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "reply": "你好！我是AI助手，有什么可以帮助你的？",
    "session_id": "session-id"
  }
}
```

## 日志系统

本项目使用自定义日志系统，支持多级别日志记录、按天轮转、结构化日志和异步写入功能。

### 日志级别

系统支持以下日志级别（从低到高）：
- `DEBUG`：调试信息，用于开发过程中的详细跟踪
- `INFO`：普通信息，记录应用的正常运行状态
- `WARNING`：警告信息，表示可能的问题但不影响正常运行
- `ERROR`：错误信息，表示发生了错误但应用可以继续运行
- `FATAL`：致命错误，记录后程序会自动退出

### 记录日志

```go
import "AiDemo/utils"

// 记录不同级别的日志
utils.Debug("这是一条调试信息: %s", "详细数据")
utils.Info("这是一条普通信息")
utils.Warning("这是一条警告信息: %v", err)
utils.Error("这是一条错误信息: %v", err)
utils.Fatal("这是一条致命错误信息，记录后程序会退出") // 会导致程序退出
```

### 设置日志级别

默认日志级别为 `INFO`，可以通过以下方式修改：

```go
// 设置为DEBUG级别，记录所有日志
utils.SetLevel(utils.DEBUG)

// 设置为ERROR级别，只记录ERROR和FATAL级别的日志
utils.SetLevel(utils.ERROR)
```

### 日志轮转功能

系统支持按天自动轮转日志文件，每天零点会创建新的日志文件，文件名格式为：`app.2023-05-20.log`

#### 启用日志轮转

日志轮转功能已在 `init/initLog.go` 中配置好：

```go
// 初始化日志系统
func InitLog() {
    // ...
    utils.EnableRotate() // 启用按天轮转
    // ...
}
```

#### 禁用日志轮转

如果在特定场景下需要禁用日志轮转，可以调用：

```go
// 禁用日志轮转
init.DisableLogRotate()
```

### 结构化日志

系统支持结构化日志输出，可以输出为JSON格式，便于日志分析和处理。

#### 启用JSON格式

```go
// 设置为JSON格式输出
utils.SetFormat(utils.JSON_FORMAT)

// 设置回文本格式
utils.SetFormat(utils.TEXT_FORMAT)
```

#### 使用带字段的日志

```go
// 创建带有字段的日志
logger := utils.WithFields(map[string]interface{}{
    "user_id": 12345,
    "action": "login",
    "ip": "192.168.1.1",
})

// 记录带有字段的日志
logger.Info("用户登录")

// 输出的JSON格式如下：
// {"level":"INFO","timestamp":"2023-05-20 15:04:05.123","message":"用户登录","fields":{"action":"login","ip":"192.168.1.1","user_id":12345}}
```

### 异步日志

系统支持异步日志写入，可以提高应用性能，避免日志写入阻塞主线程。

#### 启用异步日志

```go
// 启用异步日志，使用默认缓冲区大小和刷新间隔
utils.EnableAsync(0, 0)

// 自定义缓冲区大小和刷新间隔
utils.EnableAsync(5000, 5*time.Second)
```

#### 刷新和关闭

```go
// 手动刷新日志缓冲区
utils.Flush()

// 禁用异步日志（会等待所有日志写入完成）
utils.DisableAsync()

// 关闭日志系统（会自动刷新并关闭）
utils.Close()
```

#### 异步日志注意事项

1. 致命错误（FATAL）日志会强制同步写入，确保在程序退出前记录
2. 当缓冲区已满时，会自动回退到同步写入模式
3. 应用退出前应调用`utils.Close()`确保所有日志都被写入

### 自定义日志记录器

如果需要创建独立的日志记录器，可以使用：

```go
// 创建自定义日志记录器
logger := utils.NewLogger(utils.DEBUG, "./logs/custom.log", true)

// 使用自定义记录器记录日志
logger.Debug("自定义记录器的调试信息")
logger.Info("自定义记录器的普通信息")

// 关闭自定义记录器
defer logger.Close()
```

### 日志系统注意事项

1. 日志文件会自动创建，但需要确保应用有权限写入指定目录
2. 在应用退出前应调用 `utils.Close()` 关闭日志文件
3. 日志轮转发生在写日志时检查，如果长时间没有日志写入，可能不会立即轮转
4. 当前实现不会自动清理或压缩旧日志文件，需要外部工具管理
5. 使用异步日志时，应确保在应用退出前调用 `utils.Close()` 或 `utils.Flush()`

## 许可证

本项目采用 MIT 许可证 - 详细信息请查看 [LICENSE](LICENSE) 文件。 