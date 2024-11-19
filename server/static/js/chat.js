let ws;
const messagesDiv = document.getElementById('messages');
const messageInput = document.getElementById('message-input');
const onlineCount = document.getElementById('online-count');

function connect() {
    // 使用一个固定的名字作为浏览器端的标识
    ws = new WebSocket(`ws://${window.location.host}/chat/browser-viewer`);
    
    ws.onmessage = function(e) {
        const msg = JSON.parse(e.data);
        const messageDiv = document.createElement('div');
        
        switch (msg.type) {
            case 0: // 普通消息
                messageDiv.className = 'message';
                messageDiv.textContent = `${msg.from}: ${msg.content}`;
                break;
            case 1: // 系统消息
                messageDiv.className = 'message system-message';
                messageDiv.textContent = msg.content;
                
                // 更新在线人数
                if (msg.content.startsWith('当前在线用户:')) {
                    const users = msg.content.split(':')[1].trim().split(',');
                    onlineCount.textContent = `在线人数: ${users.length}`;
                }
                break;
        }
        
        messagesDiv.appendChild(messageDiv);
        messagesDiv.scrollTop = messagesDiv.scrollHeight;
    };
    
    ws.onclose = function() {
        console.log('连接断开，尝试重新连接...');
        setTimeout(connect, 1000);
    };

    ws.onerror = function(err) {
        console.error('WebSocket错误:', err);
    };
}

function sendMessage() {
    const message = messageInput.value.trim();
    if (message && ws) {
        ws.send(message);
        messageInput.value = '';
    }
}

messageInput.addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        sendMessage();
    }
});

// 页面加载完成后自动连接
document.addEventListener('DOMContentLoaded', function() {
    connect();
}); 