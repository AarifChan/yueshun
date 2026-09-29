package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"zhizhang-server/internal/pkg/response"
)

const maxUploadSize = 5 << 20

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// UploadHandler 文件上传处理器
type UploadHandler struct {
	dir string
}

func NewUploadHandler(dir string) *UploadHandler {
	return &UploadHandler{dir: dir}
}

func (h *UploadHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/upload", h.Upload)
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}
	if file.Size > maxUploadSize {
		response.BadRequest(c, "文件大小不能超过 5MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		response.BadRequest(c, "仅支持 jpg/jpeg/png/gif/webp 格式图片")
		return
	}

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		log.Error().Err(err).Msg("generate random filename failed")
		response.ServerError(c, "上传失败")
		return
	}
	filename := fmt.Sprintf("%d%s%s", time.Now().UnixNano(), hex.EncodeToString(buf), ext)

	if err := c.SaveUploadedFile(file, filepath.Join(h.dir, filename)); err != nil {
		log.Error().Err(err).Msg("save uploaded file failed")
		response.ServerError(c, "上传失败")
		return
	}
	response.Ok(c, gin.H{"url": "/uploads/" + filename})
}
