// 全局变量
let systemPrompt = ""; // 当前系统提示词
let waitingForAIResponse = false; // 是否正在等待AI响应

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', function () {
    // 获取系统提示词
    fetch("/get-prompt")
        .then(response => response.json())
        .then(data => {
            systemPrompt = data.prompt;
            document.getElementById("system-prompt").value = systemPrompt;
        })
        .catch(error => {
            console.error("获取系统提示词失败:", error);
        });
});

// 发送消息
function sendMessage() {
    // 如果正在等待AI响应，则不处理新消息
    if (waitingForAIResponse) return;

    const inputElement = document.getElementById("input");
    const message = inputElement.value.trim();

    // 检查消息是否为空
    if (!message) return;

    // 清空输入框
    inputElement.value = "";

    // 添加用户消息到聊天框
    const userMessageElement = document.createElement("div");
    userMessageElement.className = "message user";
    userMessageElement.textContent = "你: " + message;
    document.getElementById("chat-box").appendChild(userMessageElement);

    // 添加AI正在输入的消息
    const aiMessageElement = document.createElement("div");
    aiMessageElement.className = "message ai typing";
    aiMessageElement.textContent = "AI: 正在输入...";
    document.getElementById("chat-box").appendChild(aiMessageElement);

    // 标记正在等待AI响应
    waitingForAIResponse = true;

    // 滚动到底部
    scrollToBottom();

    // 发送请求到服务器
    fetch("/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            message: message,
            systemPrompt: document.getElementById("system-prompt").value
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error("服务器响应错误: " + response.status);
            }
            return response.json();
        })
        .then(data => {
            // 移除"正在输入..."文本
            aiMessageElement.textContent = "AI: ";
            aiMessageElement.classList.remove("typing");

            // 逐字显示AI回复
            typeText(aiMessageElement, data.reply || "出错了，请稍后再试");
        })
        .catch(error => {
            console.error("发送消息出错:", error);
            aiMessageElement.textContent = "AI: 出错了，请稍后再试";
            aiMessageElement.classList.remove("typing");
            waitingForAIResponse = false;
        });
}

// 逐字显示文本
function typeText(element, text) {
    let index = 0;
    const prefix = "AI: ";

    function type() {
        if (index < text.length) {
            element.textContent = prefix + text.substring(0, index + 1);
            index++;
            scrollToBottom();
            setTimeout(type, 30);
        } else {
            // 输入完成
            waitingForAIResponse = false;
        }
    }

    type();
}

// 滚动到底部
function scrollToBottom() {
    const chatBox = document.getElementById("chat-box");
    chatBox.scrollTop = chatBox.scrollHeight;
}

// 添加系统消息
function addSystemMessage(text) {
    const messageElement = document.createElement("div");
    messageElement.className = "message ai system-message";
    messageElement.textContent = "系统: " + text;
    document.getElementById("chat-box").appendChild(messageElement);
    scrollToBottom();
}

// 打开提示词设置对话框
function openPromptModal() {
    const modal = document.getElementById("prompt-modal");
    document.getElementById("system-prompt").value = systemPrompt;
    modal.style.display = "flex";
}

// 关闭提示词设置对话框
function closePromptModal() {
    document.getElementById("prompt-modal").style.display = "none";
}

// 保存提示词设置
function savePrompt() {
    const newPrompt = document.getElementById("system-prompt").value.trim();

    // 如果提示词为空，不处理
    if (!newPrompt) {
        alert("提示词不能为空");
        return;
    }

    // 如果提示词没有变化，直接关闭对话框
    if (newPrompt === systemPrompt) {
        closePromptModal();
        return;
    }

    // 保存新的提示词
    systemPrompt = newPrompt;

    // 发送到服务器
    fetch("/set-prompt", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: newPrompt })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error("服务器响应错误: " + response.status);
            }
            closePromptModal();
            addSystemMessage("提示词已更新，将在下一次对话中生效");
        })
        .catch(error => {
            console.error("保存提示词失败:", error);
            alert("保存提示词失败，请稍后再试");
        });
}