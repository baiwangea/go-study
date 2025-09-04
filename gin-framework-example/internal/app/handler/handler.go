package handler

import (
	"strconv"

	"gin-framework-example/internal/app/model"
	"gin-framework-example/internal/app/response"
	"gin-framework-example/internal/pkg/e"

	"github.com/gin-gonic/gin"
)

// In-memory database
var (
	users  = []model.User{}
	nextID = 1
)

func Ping(c *gin.Context) {
	response.SuccessWithData("pong", c)
}

func CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.Result(e.INVALID_PARAMS, err.Error(), nil, c)
		return
	}

	user.ID = nextID
	nextID++
	users = append(users, user)

	response.SuccessWithData(user, c)
}

func GetUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	for _, user := range users {
		if user.ID == id {
			response.SuccessWithData(user, c)
			return
		}
	}
	response.Result(e.ERROR, "User not found", nil, c)
}

func UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var updatedUser model.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		response.Result(e.INVALID_PARAMS, err.Error(), nil, c)
		return
	}

	for i, user := range users {
		if user.ID == id {
			users[i].Username = updatedUser.Username
			users[i].Email = updatedUser.Email
			// Password updates would typically be handled separately and securely
			response.SuccessWithData(users[i], c)
			return
		}
	}

	response.Result(e.ERROR, "User not found", nil, c)
}

func DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			response.Success(c)
			return
		}
	}

	response.Result(e.ERROR, "User not found", nil, c)
}
