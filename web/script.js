// 简化后的前端逻辑（支持角色选择）
let waitingForAIResponse = false;

function sendMessage() {
    if (waitingForAIResponse) return;

    const inputElement = document.getElementById("input");
    const roleSelect = document.getElementById("role-select");
    const message = inputElement.value.trim();
    const role = roleSelect ? roleSelect.value : "general";
    if (!message) return;

    inputElement.value = "";

    // 用户消息
    const userEl = document.createElement("div");
    userEl.className = "message user";
    userEl.textContent = "你: " + message;
    document.getElementById("chat-box").appendChild(userEl);

    // AI 占位
    const aiEl = document.createElement("div");
    aiEl.className = "message ai typing";
    aiEl.textContent = "AI: 正在输入...";
    document.getElementById("chat-box").appendChild(aiEl);
    waitingForAIResponse = true;
    scrollToBottom();

    fetch("/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message, role })
    })
        .then(res => { if (!res.ok) throw new Error("HTTP " + res.status); return res.json(); })
        .then(data => {
            aiEl.textContent = "AI: ";
            aiEl.classList.remove("typing");
            typeText(aiEl, data.reply || "出错了，请稍后再试");
        })
        .catch(err => {
            console.error(err);
            aiEl.textContent = "AI: 出错了，请稍后再试";
            aiEl.classList.remove("typing");
            waitingForAIResponse = false;
        });
}

function typeText(element, text) {
    let i = 0;
    const prefix = "AI: ";
    (function type() {
        if (i < text.length) {
            element.textContent = prefix + text.substring(0, i + 1);
            i++;
            scrollToBottom();
            setTimeout(type, 30);
        } else {
            waitingForAIResponse = false;
        }
    })();
}

function scrollToBottom() {
    const chatBox = document.getElementById("chat-box");
    chatBox.scrollTop = chatBox.scrollHeight;
}