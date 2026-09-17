package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 5 << 20 // 5MB

var allowedImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// Upload 图片上传：保存到 ./web/uploads/{tenant}/{yyyymm}/{rand}.{ext}
func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "未收到文件")
		return
	}
	if file.Size > maxUploadSize {
		response.BadRequest(c, "图片不能超过 5MB")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		response.BadRequest(c, "仅支持 jpg / png / gif / webp")
		return
	}

	tenantID := context.GetTenantID(c.Request.Context())
	sub := fmt.Sprintf("%d/%s", tenantID, time.Now().Format("200601"))
	dir := filepath.Join("web", "uploads", filepath.FromSlash(sub))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		response.ServerError(c, "创建目录失败")
		return
	}

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		response.ServerError(c, "生成文件名失败")
		return
	}
	name := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), hex.EncodeToString(buf), ext)
	dst := filepath.Join(dir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.ServerError(c, "保存文件失败")
		return
	}

	response.OK(c, gin.H{"url": "/uploads/" + sub + "/" + name})
}
