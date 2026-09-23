package response

import "net/http"

// 统一响应格式
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

var (
	Success          = Response{Code: 0, Msg: "success"}
	ErrParam         = Response{Code: 10001, Msg: "参数错误"}
	ErrUnauthorized  = Response{Code: 10002, Msg: "未登录或令牌无效"}
	ErrForbidden     = Response{Code: 10003, Msg: "没有权限"}
	ErrNotFound      = Response{Code: 10004, Msg: "资源不存在"}
	ErrStatus        = Response{Code: 10005, Msg: "资源状态不允许当前操作"}
	ErrRegister      = Response{Code: 10006, Msg: "注册信息不符合规范"}
	ErrLogin         = Response{Code: 10007, Msg: "uid或密码错误"}
	ErrDuplicate     = Response{Code: 10008, Msg: "重复提交"}
	ErrServer        = Response{Code: 20001, Msg: "服务器内部错误"}
)

// 自定义错误返回格式
func Fail(c *gin.Context, r Response) {
	c.JSON(http.StatusOK, r) // 注意：HTTP 状态码通常返回 200，业务错误码在 JSON 中体现
	c.Abort()
}


func AdminAuth(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Fail(c, response.ErrUnauthorized)
			return
		}

		userRole := role.(string)
		allowed := false
		for _, r := range roles {
			if userRole == r {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Fail(c, response.ErrForbidden)
			return
		}
		c.Next()
	}
}


if pageSize > 100 { pageSize = 100 }
if pageSize <= 0 { pageSize = 10 }
