package main

import (
	"encoding/json"
	"log"
	"time"
)

// Hub 维护活跃的客户端集合并向客户端广播消息
type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 客户端用户名映射
	usernames map[string]*Client

	// 广播消息到所有客户端
	broadcast chan []byte

	// 私聊消息
	private chan *PrivateMessage

	// 注册新客户端
	register chan *Client

	// 注销客户端
	unregister chan *Client
}

// PrivateMessage 私聊消息结构
type PrivateMessage struct {
	to      string // 目标用户名
	message []byte // 消息内容
}

// newHub 创建新的Hub实例
func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		private:    make(chan *PrivateMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		usernames:  make(map[string]*Client),
	}
}

// run 启动Hub，处理所有注册、注销和消息广播
func (h *Hub) run() {
	log.Println("🚀 WebSocket Hub 已启动")

	for {
		select {
		case client := <-h.register:
			// 注册新客户端
			h.clients[client] = true
			h.usernames[client.username] = client

			log.Printf("✅ 用户 [%s] 加入聊天室 (在线人数: %d)", client.username, len(h.clients))

			// 发送欢迎消息给新用户
			welcomeMsg := Message{
				Type:      "system",
				Content:   "欢迎加入聊天室！",
				Timestamp: time.Now().Format("15:04:05"),
			}
			msgBytes, _ := json.Marshal(welcomeMsg)
			client.send <- msgBytes

			// 广播系统消息：有新用户加入
			joinMsg := Message{
				Type:      "system",
				Content:   client.username + " 加入了聊天室",
				Timestamp: time.Now().Format("15:04:05"),
			}
			joinBytes, _ := json.Marshal(joinMsg)
			h.broadcastToAll(joinBytes)

			// 广播更新后的用户列表
			h.broadcastUserList()

		case client := <-h.unregister:
			// 注销客户端
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.usernames, client.username)
				close(client.send)

				log.Printf("👋 用户 [%s] 离开聊天室 (在线人数: %d)", client.username, len(h.clients))

				// 广播系统消息：用户离开
				leaveMsg := Message{
					Type:      "system",
					Content:   client.username + " 离开了聊天室",
					Timestamp: time.Now().Format("15:04:05"),
				}
				leaveBytes, _ := json.Marshal(leaveMsg)
				h.broadcastToAll(leaveBytes)

				// 广播更新后的用户列表
				h.broadcastUserList()
			}

		case message := <-h.broadcast:
			// 广播消息到所有客户端
			h.broadcastToAll(message)

		case pm := <-h.private:
			// 发送私聊消息
			if targetClient, ok := h.usernames[pm.to]; ok {
				select {
				case targetClient.send <- pm.message:
					log.Printf("📨 私聊消息已发送到 [%s]", pm.to)
				default:
					// 发送失败，关闭连接
					close(targetClient.send)
					delete(h.clients, targetClient)
					delete(h.usernames, targetClient.username)
					log.Printf("⚠️  无法发送消息到 [%s]，连接已关闭", pm.to)
				}
			} else {
				log.Printf("⚠️  目标用户 [%s] 不在线", pm.to)
			}
		}
	}
}

// broadcastToAll 向所有在线客户端广播消息
func (h *Hub) broadcastToAll(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
			// 消息发送成功
		default:
			// 发送失败，关闭连接并移除客户端
			close(client.send)
			delete(h.clients, client)
			delete(h.usernames, client.username)
			log.Printf("⚠️  客户端 [%s] 连接异常，已移除", client.username)
		}
	}
}

// broadcastUserList 广播在线用户列表
func (h *Hub) broadcastUserList() {
	// 收集所有在线用户名
	users := make([]string, 0, len(h.clients))
	for client := range h.clients {
		users = append(users, client.username)
	}

	// 构建用户列表消息
	userListMsg := Message{
		Type:      "userlist",
		Users:     users,
		Timestamp: time.Now().Format("15:04:05"),
		Data: map[string]interface{}{
			"count": len(users),
		},
	}

	msgBytes, _ := json.Marshal(userListMsg)
	h.broadcastToAll(msgBytes)

	log.Printf("📊 已广播用户列表 (共 %d 人)", len(users))
}

// GetOnlineCount 获取在线用户数量
func (h *Hub) GetOnlineCount() int {
	return len(h.clients)
}
