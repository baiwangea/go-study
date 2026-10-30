package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 客户端写入等待时间
	writeWait = 10 * time.Second

	// 客户端读取超时时间
	pongWait = 60 * time.Second

	// Ping周期（必须小于pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client 表示一个WebSocket客户端连接
type Client struct {
	hub      *Hub              // 所属的Hub管理器
	conn     *websocket.Conn   // WebSocket连接
	send     chan []byte       // 发送消息的通道
	username string            // 用户名
	id       string            // 唯一标识
}

// Message 消息结构
type Message struct {
	Type      string      `json:"type"`                // 消息类型：chat/system/userlist/private
	Username  string      `json:"username,omitempty"`  // 发送者用户名
	Content   string      `json:"content,omitempty"`   // 消息内容
	Timestamp string      `json:"timestamp"`           // 时间戳
	Users     []string    `json:"users,omitempty"`     // 在线用户列表
	To        string      `json:"to,omitempty"`        // 私聊目标用户
	Data      interface{} `json:"data,omitempty"`      // 额外数据
}

// readPump 从WebSocket连接读取消息并发送到hub
// 为每个连接运行在单独的goroutine中
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		log.Printf("🔌 客户端 [%s] 断开连接", c.username)
	}()

	// 设置读取限制
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket错误: %v", err)
			}
			break
		}

		// 清理消息内容
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// 解析消息
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("⚠️  解析消息失败: %v", err)
			continue
		}

		// 添加时间戳和用户名
		msg.Timestamp = time.Now().Format("15:04:05")
		msg.Username = c.username

		// 根据消息类型处理
		switch msg.Type {
		case "chat":
			// 广播聊天消息
			log.Printf("💬 [%s]: %s", c.username, msg.Content)
			processedMsg, _ := json.Marshal(msg)
			c.hub.broadcast <- processedMsg

		case "private":
			// 私聊消息
			log.Printf("🔒 [%s] -> [%s]: %s", c.username, msg.To, msg.Content)
			processedMsg, _ := json.Marshal(msg)
			c.hub.private <- &PrivateMessage{
				to:      msg.To,
				message: processedMsg,
			}

		default:
			log.Printf("⚠️  未知消息类型: %s", msg.Type)
		}
	}
}

// writePump 将消息从hub发送到WebSocket连接
// 为每个连接运行在单独的goroutine中
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将队列中的其他消息也一起发送
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// 发送ping保持连接
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
