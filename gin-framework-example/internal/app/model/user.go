package model

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Email    string `json:"email"    binding:"required,email"`
}
