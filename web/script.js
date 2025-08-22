// 简化后的前端逻辑（支持角色选择 + 会话ID持久化）
let waitingForAIResponse = false;
let sessionId = getOrCreateSessionId();

function getOrCreateSessionId() {
    try {
        const k = 'ai_session_id';
        let v = localStorage.getItem(k);
        if (!v) {
            v = crypto.randomUUID ? crypto.randomUUID() : (Date.now().toString(36) + Math.random().toString(36).slice(2));
            localStorage.setItem(k, v);
        }
        return v;
    } catch (e) {
        return Date.now().toString(36) + Math.random().toString(36).slice(2);
    }
}

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
        body: JSON.stringify({ message, role, session_id: sessionId })
    })
        .then(res => { if (!res.ok) throw new Error("HTTP " + res.status); return res.json(); })
        .then(data => {
            aiEl.textContent = "AI: ";
            aiEl.classList.remove("typing");
            if (data && data.session_id) {
                // 以服务端返回为准（兼容后端生成）
                sessionId = data.session_id;
                try { localStorage.setItem('ai_session_id', sessionId); } catch (e) { }
            }
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