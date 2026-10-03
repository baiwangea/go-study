package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
)

const (
	topic       = "test_topic"
	channel     = "ch1"
	lookupdAddr = "127.0.0.1:4161"
)

// Handler 实现 nsq.Handler 接口，处理消费到的消息
type Handler struct{}

func (h *Handler) HandleMessage(m *nsq.Message) error {
	log.Println("receive:", string(m.Body))
	return nil
}

func main() {
	cfg := nsq.NewConfig()

	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		log.Fatalf("创建 NSQ Consumer 失败: %v", err)
	}
	consumer.AddHandler(&Handler{})

	if err := consumer.ConnectToNSQLookupd(lookupdAddr); err != nil {
		log.Fatalf("连接 NSQ Lookupd %s 失败: %v", lookupdAddr, err)
	}
	log.Printf("✅ NSQ Consumer 已启动，订阅 %s/%s ...", topic, channel)

	// 等待终止信号，收到后优雅停止消费（在途消息处理完再退出）
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("🛑 收到信号 %v，正在关闭 NSQ Consumer...", sig)
	consumer.Stop()
	<-consumer.StopChan
	log.Println("👋 NSQ Consumer 已退出")
}
