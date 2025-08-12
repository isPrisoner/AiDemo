async function sendMessage() {
    const input = document.getElementById("input");
    const chatBox = document.getElementById("chat-box");
    const message = input.value.trim();
    if (!message) return;

    // 添加用户消息
    addMessage("user", message);
    input.value = "";

    // 添加 AI 占位消息
    const aiId = addMessage("ai", "正在输入...");

    try {
        const res = await fetch("/chat", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ message })
        });

        const data = await res.json();
        const reply = data.reply || "出错了";

        // 逐字打字效果
        typeWriter(aiId, reply);
    } catch (error) {
        document.getElementById(aiId).textContent = "AI: 出错了";
    }

    chatBox.scrollTop = chatBox.scrollHeight;
}

function addMessage(role, text) {
    const chatBox = document.getElementById("chat-box");
    const div = document.createElement("div");
    div.className = `message ${role}`;
    div.textContent = (role === "user" ? "你: " : "AI: ") + text;
    const id = "msg-" + Date.now();
    div.id = id;
    chatBox.appendChild(div);
    chatBox.scrollTop = chatBox.scrollHeight;
    return id;
}

// AI 逐字打字效果
function typeWriter(elementId, text) {
    const el = document.getElementById(elementId);
    el.textContent = "AI: ";
    let i = 0;
    const interval = setInterval(() => {
        el.textContent += text.charAt(i);
        i++;
        if (i >= text.length) {
            clearInterval(interval);
        }
        document.getElementById("chat-box").scrollTop = document.getElementById("chat-box").scrollHeight;
    }, 30); // 每30ms输出一个字符
}