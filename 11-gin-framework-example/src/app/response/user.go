package response

import "time"

// UserResponse defines the fields returned for a user.
// It omits sensitive information like the password.
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Mobile    string    `json:"mobile"`
	Level     uint8     `json:"level"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
