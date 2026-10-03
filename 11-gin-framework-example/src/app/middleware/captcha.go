package middleware

import (
	"gin-framework-example/src/app/response"
	"gin-framework-example/src/pkg/e"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

func GenerateCaptcha(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(80, 240, 6, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		response.Result(e.ERROR, "Failed to generate captcha", nil, c)
		return
	}
	response.SuccessWithData(gin.H{
		"captchaId": id,
		"picPath":   b64s,
	}, c)
}

func VerifyCaptcha() gin.HandlerFunc {
	return func(c *gin.Context) {
		captchaId := c.Query("captchaId")
		captchaValue := c.Query("captchaValue")

		if captchaId == "" || captchaValue == "" {
			response.Result(e.INVALID_PARAMS, "Captcha ID and value are required", nil, c)
			c.Abort()
			return
		}

		if !store.Verify(captchaId, captchaValue, true) {
			response.Result(e.ERROR, "Invalid captcha", nil, c)
			c.Abort()
			return
		}

		c.Next()
	}
}
