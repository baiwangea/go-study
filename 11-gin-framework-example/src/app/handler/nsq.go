package handler

import (
	"fmt"
	"gin-framework-example/src/app/response"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nsqio/go-nsq"
)

func ProductNsq(c *gin.Context) {
	cfg := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", cfg)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= 10; i++ {
		msg := []byte("Hello NSQ " + fmt.Sprint(i))
		err := producer.Publish("test_topic", msg)
		if err != nil {
			log.Println("publish error:", err)
		}
	}

	producer.Stop()
	response.SuccessWithData(gin.H{"status": 200}, c)
}
