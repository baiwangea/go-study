package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeEmailDelivery = "email:deliver"

func NewEmailTask(userID int, templateID string) (*asynq.Task, error) {
	payload := map[string]interface{}{
		"user_id":     userID,
		"template_id": templateID,
	}
	data, _ := json.Marshal(payload)
	return asynq.NewTask(TypeEmailDelivery, data), nil
}
