package handler

import (
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

var usernamePattern = regexp.MustCompile(`^[\p{Han}A-Za-z0-9]+$`)

type AuthHandler struct {
	db *gorm.DB
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
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

func validUsername(username string) bool {
	length := utf8.RuneCountInString(username)
	return length >= 3 && length <= 10 && usernamePattern.MatchString(username)
}

func validPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= 6 && length <= 20
}
