package response

import (
	"gin-framework-example/src/pkg/e"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Result(code int, msg string, data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

func Success(c *gin.Context) {
	Result(e.SUCCESS, e.GetMsg(e.SUCCESS), nil, c)
}

func SuccessWithData(data interface{}, c *gin.Context) {
	Result(e.SUCCESS, e.GetMsg(e.SUCCESS), data, c)
}

func Fail(c *gin.Context) {
	Result(e.ERROR, e.GetMsg(e.ERROR), nil, c)
}

func FailWithMessage(message string, c *gin.Context) {
	Result(e.ERROR, message, nil, c)
}

func FailWithCode(code int, c *gin.Context) {
	Result(code, e.GetMsg(code), nil, c)
}
