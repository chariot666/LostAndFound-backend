package handler

import (
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"lost-found-server/internal/auth"
	"lost-found-server/internal/middleware"
	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

var usernamePattern = regexp.MustCompile(`^[\p{Han}A-Za-z0-9]+$`)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	UID      uint64 `json:"uid"`
	Password string `json:"password"`
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UID == 0 || req.Password == "" {
		response.Error(c, http.StatusUnauthorized, response.CodeInvalidCredentials, "UID或密码错误")
		return
	}

	var user model.User
	if err := h.db.First(&user, "uid = ?", req.UID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusUnauthorized, response.CodeInvalidCredentials, "UID或密码错误")
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
		return
	}

	if user.Status != model.StatusActive {
		response.Error(c, http.StatusForbidden, response.CodeForbidden, "账号已被禁用")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, response.CodeInvalidCredentials, "UID或密码错误")
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.UID, user.Username, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"uid":      user.UID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeInvalidRegistration, "注册信息不符合规范")
		return
	}

	if !validUsername(req.Username) || !validPassword(req.Password) {
		response.Error(c, http.StatusBadRequest, response.CodeInvalidRegistration, "注册信息不符合规范")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
		return
	}

	user := model.User{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Role:         model.RoleUser,
		Status:       model.StatusActive,
	}
	if err := h.db.Create(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
		return
	}

	response.Success(c, gin.H{
		"uid":      user.UID,
		"username": user.Username,
		"role":     user.Role,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录或令牌无效")
		return
	}

	response.Success(c, gin.H{
		"uid":        user.UID,
		"username":   user.Username,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	})
}

func validUsername(username string) bool {
	length := utf8.RuneCountInString(username)
	return length >= 3 && length <= 10 && usernamePattern.MatchString(username)
}

func validPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= 6 && length <= 20
}
