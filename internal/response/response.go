package response

import "github.com/gin-gonic/gin"

const (
	CodeSuccess             = 0
	CodeParamError          = 10001
	CodeUnauthorized        = 10002
	CodeForbidden           = 10003
	CodeNotFound            = 10004
	CodeInvalidState        = 10005
	CodeInvalidRegistration = 10006
	CodeInvalidCredentials  = 10007
	CodeDuplicateRequest    = 10008
	CodeServerError         = 20001
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
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
		Data: nil,
	})
}
