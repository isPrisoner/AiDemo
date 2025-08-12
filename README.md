# AiDemo - 豆包AI聊天应用

一个基于Go语言和Gin框架开发的AI聊天应用，使用豆包AI API实现智能对话功能。

## 项目架构

```
AiDemo/
  ├── config/          # 配置管理
  ├── handlers/        # HTTP请求处理器
  ├── init/            # 环境变量配置
  ├── models/          # 数据模型
  ├── services/        # 业务逻辑服务
  └── web/             # 前端页面
```

## 主要功能

- 基于会话ID的多用户聊天
- 完整的日志记录系统
- 基于环境的配置管理
- RESTful API接口
- 跨域资源共享支持
- 统一的错误处理

## 技术栈

- Go 1.24.4
- Gin Web框架
- Zap日志库
- Viper配置管理
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

编辑 `init/initApi.env` 文件，设置您的API密钥：

```
DOUBAO_API_KEY=YOUR_API_KEY
```

4. 运行应用

```bash
go run main.go
```

应用将在 http://localhost:8081 上启动。

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

## 许可证

本项目采用 MIT 许可证 - 详细信息请查看 [LICENSE](LICENSE) 文件。 