package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket升级器配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源（生产环境需要配置具体域名）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub实例
var hub *Hub

// serveWs 处理WebSocket连接请求
func serveWs(w http.ResponseWriter, r *http.Request) {
	// 获取用户名参数
	username := r.URL.Query().Get("username")
	if username == "" {
		username = fmt.Sprintf("游客%d", time.Now().Unix()%10000)
	}

	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket升级失败: %v", err)
		return
	}

	// 创建新客户端
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		username: username,
		id:       fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	// 注册客户端
	client.hub.register <- client

	// 启动读写goroutine
	go client.writePump()
	go client.readPump()

	log.Printf("🎉 新连接建立: %s [%s]", username, conn.RemoteAddr())
}

// serveHome 提供静态文件服务
func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("📄 访问: %s %s", r.Method, r.URL.Path)
	
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

// apiStatus 返回服务器状态API
func apiStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := map[string]interface{}{
		"status":       "running",
		"online_users": hub.GetOnlineCount(),
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
		"uptime":       time.Since(startTime).String(),
	}
	
	json.NewEncoder(w).Encode(status)
	log.Printf("📊 状态查询: 在线用户 %d 人", hub.GetOnlineCount())
}

var startTime time.Time

func main() {
	// 记录启动时间
	startTime = time.Now()

	// 打印启动信息
	printBanner()

	// 创建Hub实例
	hub = newHub()
	go hub.run()

	// 注册路由
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	http.HandleFunc("/api/status", apiStatus)

	// 静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 启动服务器
	addr := ":8080"
	log.Printf("🌐 服务器启动成功！")
	log.Printf("📱 访问地址: http://localhost%s", addr)
	log.Printf("🔗 WebSocket地址: ws://localhost%s/ws", addr)
	log.Printf("⏰ 启动时间: %s", startTime.Format("2006-01-02 15:04:05"))
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("💥 服务器启动失败: %v", err)
	}
}

// printBanner 打印启动横幅
func printBanner() {
	banner := `
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   🚀 WebSocket 实时聊天室 v1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   功能特性:
   ✅ 实时双向通信
   ✅ 多客户端广播
   ✅ 私聊功能
   ✅ 在线用户列表
   ✅ 心跳保持连接
   ✅ 自动重连
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`
	fmt.Println(banner)
}
