// WebSocket客户端应用
class ChatApp {
    constructor() {
        this.ws = null;
        this.username = '';
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 3000;
        this.sentCount = 0;
        this.receivedCount = 0;
        this.connectionStartTime = null;
        this.connectionTimeInterval = null;

        // 获取DOM元素
        this.elements = {
            status: document.getElementById('status'),
            onlineCount: document.getElementById('onlineCount'),
            messages: document.getElementById('messages'),
            userList: document.getElementById('userList'),
            usernameInput: document.getElementById('usernameInput'),
            connectBtn: document.getElementById('connectBtn'),
            disconnectBtn: document.getElementById('disconnectBtn'),
            messageInput: document.getElementById('messageInput'),
            sendBtn: document.getElementById('sendBtn'),
            connectionControls: document.getElementById('connectionControls'),
            messageControls: document.getElementById('messageControls'),
            privateModeCheck: document.getElementById('privateModeCheck'),
            privateTarget: document.getElementById('privateTarget'),
            sentCount: document.getElementById('sentCount'),
            receivedCount: document.getElementById('receivedCount'),
            connectionTime: document.getElementById('connectionTime')
        };

        this.init();
    }

    // 初始化应用
    init() {
        // 绑定事件
        this.elements.connectBtn.addEventListener('click', () => this.connect());
        this.elements.disconnectBtn.addEventListener('click', () => this.disconnect());
        this.elements.sendBtn.addEventListener('click', () => this.sendMessage());

        // 用户名输入框回车连接
        this.elements.usernameInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.connect();
            }
        });

        // 消息输入框快捷键
        this.elements.messageInput.addEventListener('keypress', (e) => {
            if (e.shiftKey && e.key === 'Enter') {
                e.preventDefault();
                this.sendMessage();
            }
        });

        this.elements.messageInput.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.elements.messageInput.value = '';
            }
        });

        // 私聊模式切换
        this.elements.privateModeCheck.addEventListener('change', (e) => {
            this.elements.privateTarget.classList.toggle('hidden', !e.target.checked);
        });

        console.log('📱 聊天应用初始化完成');
    }

    // 连接WebSocket
    connect() {
        const username = this.elements.usernameInput.value.trim();
        if (!username) {
            alert('请输入用户名');
            return;
        }

        this.username = username;
        this.updateStatus('connecting', '连接中...');

        try {
            // 创建WebSocket连接
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws?username=${encodeURIComponent(username)}`;
            
            console.log('🔌 正在连接:', wsUrl);
            this.ws = new WebSocket(wsUrl);

            // 连接打开
            this.ws.onopen = () => this.onOpen();

            // 接收消息
            this.ws.onmessage = (event) => this.onMessage(event);

            // 连接关闭
            this.ws.onclose = () => this.onClose();

            // 连接错误
            this.ws.onerror = (error) => this.onError(error);

        } catch (error) {
            console.error('❌ 连接失败:', error);
            this.updateStatus('disconnected', '连接失败');
            alert('连接失败: ' + error.message);
        }
    }

    // 断开连接
    disconnect() {
        if (this.ws) {
            console.log('👋 主动断开连接');
            this.reconnectAttempts = this.maxReconnectAttempts; // 阻止自动重连
            this.ws.close();
        }
    }

    // 连接打开回调
    onOpen() {
        console.log('✅ WebSocket连接已建立');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.updateStatus('connected', '已连接');

        // 切换UI
        this.elements.connectionControls.classList.add('hidden');
        this.elements.messageControls.classList.remove('hidden');
        this.elements.messageInput.disabled = false;
        this.elements.sendBtn.disabled = false;

        // 清空欢迎消息
        this.elements.messages.innerHTML = '';

        // 记录连接时间
        this.connectionStartTime = Date.now();
        this.startConnectionTimer();

        // 添加系统消息
        this.addSystemMessage('✅ 已成功连接到服务器');
    }

    // 接收消息回调
    onMessage(event) {
        try {
            const message = JSON.parse(event.data);
            console.log('📨 收到消息:', message);

            this.receivedCount++;
            this.elements.receivedCount.textContent = this.receivedCount;

            switch (message.type) {
                case 'chat':
                    this.addChatMessage(message);
                    break;
                case 'private':
                    this.addPrivateMessage(message);
                    break;
                case 'system':
                    this.addSystemMessage(message.content);
                    break;
                case 'userlist':
                    this.updateUserList(message.users);
                    this.elements.onlineCount.textContent = `在线: ${message.data.count}`;
                    break;
                default:
                    console.warn('⚠️  未知消息类型:', message.type);
            }

        } catch (error) {
            console.error('❌ 解析消息失败:', error);
        }
    }

    // 连接关闭回调
    onClose() {
        console.log('🔌 WebSocket连接已关闭');
        this.isConnected = false;
        this.updateStatus('disconnected', '已断开');

        this.elements.messageInput.disabled = true;
        this.elements.sendBtn.disabled = true;

        this.stopConnectionTimer();

        // 尝试重连
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = this.reconnectDelay * this.reconnectAttempts;
            console.log(`🔄 ${delay/1000}秒后尝试第${this.reconnectAttempts}次重连...`);
            
            this.addSystemMessage(`⚠️  连接已断开，${delay/1000}秒后自动重连...`);
            
            setTimeout(() => {
                if (!this.isConnected) {
                    this.connect();
                }
            }, delay);
        } else {
            this.addSystemMessage('❌ 连接已断开，请手动重新连接');
            this.elements.connectionControls.classList.remove('hidden');
            this.elements.messageControls.classList.add('hidden');
        }
    }

    // 连接错误回调
    onError(error) {
        console.error('❌ WebSocket错误:', error);
        this.addSystemMessage('❌ 连接错误，请检查网络');
    }

    // 发送消息
    sendMessage() {
        const content = this.elements.messageInput.value.trim();
        if (!content) {
            return;
        }

        if (!this.isConnected) {
            alert('未连接到服务器');
            return;
        }

        // 构建消息对象
        const message = {
            type: 'chat',
            content: content,
            timestamp: new Date().toLocaleTimeString()
        };

        // 检查是否私聊模式
        if (this.elements.privateModeCheck.checked) {
            const target = this.elements.privateTarget.value;
            if (!target) {
                alert('请选择私聊对象');
                return;
            }
            message.type = 'private';
            message.to = target;
        }

        try {
            // 发送消息
            this.ws.send(JSON.stringify(message));
            console.log('📤 发送消息:', message);

            // 更新统计
            this.sentCount++;
            this.elements.sentCount.textContent = this.sentCount;

            // 显示自己发送的消息
            if (message.type === 'chat') {
                this.addChatMessage({
                    username: this.username,
                    content: content,
                    timestamp: message.timestamp
                }, true);
            } else {
                this.addPrivateMessage({
                    username: '我',
                    content: content,
                    to: message.to,
                    timestamp: message.timestamp
                }, true);
            }

            // 清空输入框
            this.elements.messageInput.value = '';

        } catch (error) {
            console.error('❌ 发送消息失败:', error);
            alert('发送失败: ' + error.message);
        }
    }

    // 添加聊天消息
    addChatMessage(message, isOwn = false) {
        const messageEl = document.createElement('div');
        messageEl.className = `message chat ${isOwn ? 'own' : ''}`;
        
        messageEl.innerHTML = `
            <div class="message-header">
                <span class="message-username">${message.username}</span>
                <span class="message-time">${message.timestamp}</span>
            </div>
            <div class="message-content">${this.escapeHtml(message.content)}</div>
        `;

        this.elements.messages.appendChild(messageEl);
        this.scrollToBottom();
    }

    // 添加私聊消息
    addPrivateMessage(message, isOwn = false) {
        const messageEl = document.createElement('div');
        messageEl.className = 'message private';
        
        const direction = isOwn ? `发送给 ${message.to}` : `来自 ${message.username}`;
        
        messageEl.innerHTML = `
            <div class="message-header">
                <span class="message-username">
                    ${isOwn ? '我' : message.username}
                    <span class="private-tag">私聊</span>
                </span>
                <span class="message-time">${message.timestamp}</span>
            </div>
            <div class="message-content">
                <small style="opacity: 0.7">${direction}</small><br>
                ${this.escapeHtml(message.content)}
            </div>
        `;

        this.elements.messages.appendChild(messageEl);
        this.scrollToBottom();
    }

    // 添加系统消息
    addSystemMessage(content) {
        const messageEl = document.createElement('div');
        messageEl.className = 'message system';
        messageEl.textContent = content;

        this.elements.messages.appendChild(messageEl);
        this.scrollToBottom();
    }

    // 更新用户列表
    updateUserList(users) {
        console.log('👥 更新用户列表:', users);
        
        this.elements.userList.innerHTML = '';
        
        if (!users || users.length === 0) {
            const li = document.createElement('li');
            li.className = 'no-users';
            li.textContent = '暂无用户在线';
            this.elements.userList.appendChild(li);
            return;
        }

        // 更新私聊目标下拉框
        this.elements.privateTarget.innerHTML = '<option value="">选择私聊对象...</option>';

        users.forEach(user => {
            // 添加到用户列表
            const li = document.createElement('li');
            li.textContent = user;
            if (user === this.username) {
                li.classList.add('current-user');
                li.textContent += ' (我)';
            }
            this.elements.userList.appendChild(li);

            // 添加到私聊下拉框（排除自己）
            if (user !== this.username) {
                const option = document.createElement('option');
                option.value = user;
                option.textContent = user;
                this.elements.privateTarget.appendChild(option);
            }
        });

        // 启用私聊选择器
        this.elements.privateTarget.disabled = false;
    }

    // 更新连接状态
    updateStatus(status, text) {
        const statusDot = this.elements.status.querySelector('.status-dot');
        statusDot.className = `status-dot ${status}`;
        
        const statusText = this.elements.status.childNodes[1];
        if (statusText) {
            statusText.textContent = text;
        }
    }

    // 开始连接时间计时
    startConnectionTimer() {
        this.connectionTimeInterval = setInterval(() => {
            if (this.connectionStartTime) {
                const elapsed = Math.floor((Date.now() - this.connectionStartTime) / 1000);
                const minutes = Math.floor(elapsed / 60);
                const seconds = elapsed % 60;
                this.elements.connectionTime.textContent = 
                    `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
            }
        }, 1000);
    }

    // 停止连接时间计时
    stopConnectionTimer() {
        if (this.connectionTimeInterval) {
            clearInterval(this.connectionTimeInterval);
            this.connectionTimeInterval = null;
        }
    }

    // 滚动到底部
    scrollToBottom() {
        this.elements.messages.scrollTop = this.elements.messages.scrollHeight;
    }

    // HTML转义
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    console.log('🚀 页面加载完成，初始化聊天应用...');
    window.chatApp = new ChatApp();
});
