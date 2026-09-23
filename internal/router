package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lost-found-server/internal/handler"
	"lost-found-server/internal/middleware"
	"lost-found-server/internal/pkg/response"
)

// InitRouter 初始化路由，完全按照 API.md 规范
func InitRouter() *gin.Engine {
	r := gin.New()

	// 1. 全局中间件
	r.Use(middleware.Cors())      // 跨域
	r.Use(middleware.Recovery())  // 统一异常捕获 (防panic)
	r.Use(middleware.Logger())    // 日志记录

	// 健康检查 (API 文档 第 9 节：接口实现顺序建议放第一位)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success) // 返回统一成功格式
	})

	// 全局 API 版本前缀
	v1 := r.Group("/api/v1")

	// ==========================================
	// 2. 用户与认证 (无需登录)
	// ==========================================
	auth := v1.Group("/auth")
	{
		auth.POST("/register", handler.Register) // 2.1 用户注册
		auth.POST("/login", handler.Login)       // 2.2 用户登录
	}

	// 无需登录的公开接口
	v1.GET("/items", handler.GetItemList)        // 3.1 获取信息列表 (无需登录)
	v1.GET("/items/:id", handler.GetItemDetail)  // 3.2 获取信息详情 (无需登录)
	v1.GET("/announcements", handler.GetAnnouncementList) // 7.1 获取公告列表

	// ==========================================
	// 需要登录的接口 (使用 JWT 鉴权)
	// ==========================================
	authRequired := v1.Group("")
	authRequired.Use(middleware.JWTAuth()) // JWT 鉴权中间件，解析 token 并将 uid 和 role 存入 Context
	{
		// 2.3 获取当前用户
		authRequired.GET("/auth/me", handler.GetCurrentUser)

		// 3. 失物和拾物信息
		authRequired.POST("/items", handler.CreateItem)         // 3.3 发布失物或拾物信息
		authRequired.PUT("/items/:id", handler.UpdateItem)      // 3.4 修改自己的信息
		authRequired.DELETE("/items/:id", handler.DeleteItem)   // 3.5 删除自己的信息
		authRequired.GET("/me/items", handler.GetMyItems)       // 3.6 获取我发布的信息

		// 4. 图片上传
		authRequired.POST("/upload", handler.UploadImage)       // 4.1 上传图片

		// 5. 认领申请 (普通用户行为)
		authRequired.POST("/items/:id/claims", handler.SubmitClaim) // 5.1 提交认领申请
		authRequired.GET("/me/claims", handler.GetMyClaims)         // 5.2 获取我提交的申请
	}

	// ==========================================
	// 管理员接口 (需要登录 + 特定角色)
	// ==========================================
	
	// 失物招领管理员 (item_admin) 或系统管理员 (system_admin) 均可访问
	itemAdminAuth := v1.Group("/admin")
	itemAdminAuth.Use(middleware.JWTAuth())
	itemAdminAuth.Use(middleware.AdminAuth("item_admin", "system_admin"))
	{
		// 5.3 & 5.4 认领申请管理
		itemAdminAuth.GET("/claims", handler.GetAdminClaims)          // 管理员获取待处理申请
		itemAdminAuth.PUT("/claims/:id", handler.ReviewClaim)         // 管理员审核认领申请

		// 6.1 审核失物或拾物信息
		itemAdminAuth.GET("/items", handler.GetAdminItems)
		itemAdminAuth.PUT("/items/:id", handler.ReviewItem)
	}

	// 系统管理员 (system_admin) 专属接口
	systemAdminAuth := v1.Group("/admin")
	systemAdminAuth.Use(middleware.JWTAuth())
	systemAdminAuth.Use(middleware.AdminAuth("system_admin"))
	{
		// 6.2 & 6.3 用户管理
		systemAdminAuth.GET("/users", handler.GetUserList)
		systemAdminAuth.PUT("/users/:id", handler.UpdateUserRoleOrStatus)

		// 7.2 管理公告
		systemAdminAuth.POST("/announcements", handler.CreateAnnouncement)
		systemAdminAuth.PUT("/announcements/:id", handler.UpdateAnnouncement)
		systemAdminAuth.DELETE("/announcements/:id", handler.DeleteAnnouncement)

		// 7.3 获取统计数据
		systemAdminAuth.GET("/statistics", handler.GetStatistics)
	}

	return r
}
