package main

import (
	"log"

	"github.com/nsqio/go-nsq"
)

type Handler struct{}

func (h *Handler) HandleMessage(m *nsq.Message) error {
	log.Println("receive:", string(m.Body))
	return nil
}

func main() {
	cfg := nsq.NewConfig()
	consumer, _ := nsq.NewConsumer("test_topic", "ch1", cfg)
	consumer.AddHandler(&Handler{})

	err := consumer.ConnectToNSQLookupd("127.0.0.1:4161")
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
