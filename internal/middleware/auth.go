package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lost-found-server/internal/auth"
	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

const (
	ContextUID  = "uid"
	ContextUser = "user"
	ContextRole = "role"
)

type AuthMiddleware struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthMiddleware(db *gorm.DB, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{db: db, jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			response.Error(c, 401, response.CodeUnauthorized, "未登录或令牌无效")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(m.jwtSecret, token)
		if err != nil {
			response.Error(c, 401, response.CodeUnauthorized, "未登录或令牌无效")
			c.Abort()
			return
		}

		var user model.User
		if err := m.db.First(&user, "uid = ?", claims.UID).Error; err != nil || user.Status != model.StatusActive {
			response.Error(c, 401, response.CodeUnauthorized, "未登录或令牌无效")
			c.Abort()
			return
		}

		c.Set(ContextUID, user.UID)
		c.Set(ContextUser, user)
		c.Set(ContextRole, user.Role)
		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.Next()
			return
		}

		claims, err := auth.ParseToken(m.jwtSecret, token)
		if err != nil {
			c.Next()
			return
		}

		var user model.User
		if err := m.db.First(&user, "uid = ?", claims.UID).Error; err != nil || user.Status != model.StatusActive {
			c.Next()
			return
		}

		c.Set(ContextUID, user.UID)
		c.Set(ContextUser, user)
		c.Set(ContextRole, user.Role)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextRole)
		if !exists {
			response.Error(c, 401, response.CodeUnauthorized, "未登录或令牌无效")
			c.Abort()
			return
		}
		if _, ok := allowed[role.(string)]; !ok {
			response.Error(c, 403, response.CodeForbidden, "没有权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentUID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(ContextUID)
	if !exists {
		return 0, false
	}
	uid, ok := value.(uint64)
	return uid, ok
}

func CurrentUser(c *gin.Context) (model.User, bool) {
	value, exists := c.Get(ContextUser)
	if !exists {
		return model.User{}, false
	}
	user, ok := value.(model.User)
	return user, ok
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}
