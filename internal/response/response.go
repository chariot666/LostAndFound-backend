package response

import "github.com/gin-gonic/gin"

const (
	CodeSuccess      = 0
	CodeParamError   = 10001
	CodeUnauthorized = 10002
	CodeForbidden    = 10003
	CodeNotFound     = 10004
	CodeServerError  = 20001
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Body{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	})
}

func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Body{
		Code: code,
		Msg:  msg,
	})
}
