package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"lost-found-server/internal/response"
)

const maxUploadSize int64 = 5 * 1024 * 1024

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "请选择图片文件")
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxUploadSize {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "图片大小不能超过 5MB")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "图片文件无效")
		return
	}
	contentType := http.DetectContentType(buffer[:n])
	extension, ok := allowedImageExtension(contentType)
	if !ok {
		response.Error(c, http.StatusBadRequest, response.CodeParamError, "只允许 jpg、png、webp 图片")
		return
	}

	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
		return
	}
	name, err := randomFilename(extension)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
		return
	}
	target := filepath.Join(h.uploadDir, name)
	if err := c.SaveUploadedFile(fileHeader, target); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeServerError, "服务器内部错误")
		return
	}
	response.Success(c, gin.H{"url": "/uploads/" + name})
}

func allowedImageExtension(contentType string) (string, bool) {
	switch strings.ToLower(contentType) {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func randomFilename(extension string) (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer) + extension, nil
}
