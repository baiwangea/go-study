package handler

import (
	"gin-framework-example/src/app/model"
	"gin-framework-example/src/app/response"
	"gin-framework-example/src/app/service"
	"gin-framework-example/src/pkg/e"
	"strconv"

	"github.com/gin-gonic/gin"
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

	createdUser, err := service.CreateUser(&user)
	if err != nil {
		response.Result(e.ERROR, err.Error(), nil, c)
		return
	}

	response.SuccessWithData(createdUser, c)
}

func GetUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := service.GetUser(uint(id))
	if err != nil {
		response.Result(e.ERROR, err.Error(), nil, c)
		return
	}
	response.SuccessWithData(user, c)
}

func UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var updatedUser model.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		response.Result(e.INVALID_PARAMS, err.Error(), nil, c)
		return
	}

	user, err := service.UpdateUser(uint(id), &updatedUser)
	if err != nil {
		response.Result(e.ERROR, err.Error(), nil, c)
		return
	}

	response.SuccessWithData(user, c)
}

func DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := service.DeleteUser(uint(id))
	if err != nil {
		response.Result(e.ERROR, err.Error(), nil, c)
		return
	}

	response.Success(c)
}

func Login(c *gin.Context) {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		response.Result(e.INVALID_PARAMS, err.Error(), nil, c)
		return
	}

	user, token, err := service.Login(loginReq.Username, loginReq.Password)
	if err != nil {
		response.Result(e.ERROR, err.Error(), nil, c)
		return
	}

	response.SuccessWithData(gin.H{"user": user, "token": token}, c)
}

func Jtel(c *gin.Context) {
	var jtelReq struct {
		SPID       string `json:"spid"`
		Password   string `json:"password"`
		Mobiles    string `json:"mobiles"`
		Valistamps string `json:"valistamps"`
	}
	if err := c.ShouldBindJSON(&jtelReq); err != nil {
		response.Result(e.INVALID_PARAMS, err.Error(), nil, c)
		return
	}
	response.SuccessWithData(gin.H{"status": 200}, c)
}
