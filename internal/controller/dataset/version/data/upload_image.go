package ctxDatasetVersionData

import (
	"fmt"
	"math"
	ctx "wp_template_display/internal/controller"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"
	"wp_template_display/pkg"

	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gookit/config/v2"
)

type DatasetVersionDataUploadImageReq struct {
	DatasetID   uint32 `form:"dataset_id" validate:"required"` // 数据集ID
	VersionId   uint32 `form:"version_id" validate:"required"` // 数据版本ID
	UploadToken string `form:"token" validate:"required"`      // 文件名称
}

// 上传图片
func DatasetVersionDataUploadImage(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/data/upload-image")
	c.Locals("Content", "上传图片")
	c.Locals("ContentBody", map[string]interface{}{
		"dataset_id": c.Query("dataset_id"),
		"version_id": c.Query("version_id"),
		"token":      c.Query("token"),
	})

	file, err := c.FormFile("image")
	if err != nil {
		return ctx.CtxError(c, 5001, "文件异常", err.Error())
	}

	req := DatasetVersionDataUploadImageReq{}
	if err := c.BodyParser(&req); err != nil {
		return ctx.CtxError(c, 5001, "请求参数错误", err.Error())
	}

	// 验证字段
	err = ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 验证数据集是否存在
	_, err = m.GetDatasetById(req.DatasetID)
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集不存在", err.Error())
	}

	// 验证数据集版本是否存在
	version, err := m.GetDatasetVerById(req.VersionId)
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集版本不存在", err.Error())
	}

	if version.Status == 1 {
		return ctx.CtxError(c, 5001, "数据集版本已发布", "数据集版本已发布")
	}

	// 获取文件MD5
	fileMd5, err := pkg.GetFileMd5ByForm(file)
	if err != nil {
		return ctx.CtxError(c, 5001, "文件异常", err.Error())
	}

	// 解析图片
	imageWidth, imageHeight, err := pkg.GetImageSize(file)
	if err != nil {
		return ctx.CtxError(c, 5001, "图片异常", err.Error())
	}

	// 限制宽度 图片分辨率限制：64px*64px ≤ 图片分辨率 ≤ 4096px*4096px
	if imageWidth < 64 || imageWidth > 4096 || imageHeight < 64 || imageHeight > 4096 {
		return ctx.CtxError(c, 5003, "图片分辨率超出限制", "图片分辨率限制：64px*64px ≤ 图片分辨率 ≤ 4096px*4096px")
	}

	// 分割字符串 解析分割token 内容
	singStr := pkg.AesDecrypt(req.UploadToken)
	parts := strings.Split(singStr, "-")
	if len(parts) != 4 {
		return ctx.CtxError(c, 5004, "凭证无效", "凭证格式错误")
	}

	// 解析过期时间
	expireTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return ctx.CtxError(c, 5004, "凭证无效", "过期时间解析失败")
	}

	// 验证是否过期
	if time.Now().Unix() > expireTime {
		return ctx.CtxError(c, 5004, "凭证无效", "凭证已过期")
	}

	// 验证图片宽高
	width, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil || int32(width) != int32(imageWidth) {
		return ctx.CtxError(c, 5004, "凭证无效", "图片宽度不匹配")
	}

	height, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil || int32(height) != int32(imageHeight) {
		return ctx.CtxError(c, 5004, "凭证无效", "图片高度不匹配")
	}

	// 获取文件MD5
	tokenMd5 := parts[3]
	if tokenMd5 != fileMd5 {
		return ctx.CtxError(c, 5004, "凭证无效", "文件MD5不匹配")
	}

	// 数据库文件已存在
	dbImage, _ := m.GetImageByMd5(fileMd5)
	if dbImage.Id == 0 {
		// 图片扩展名
		fileExt := strings.ToLower(filepath.Ext(file.Filename)) // 获取文件后缀
		// 生成图片保存路径
		runtimePath := config.String("RuntimePath", "runtime")
		now := carbon.Now()               // 当前时间
		uploadTime := now.Format("His")   // 上传时间
		uploadDate := now.Format("Y-m-d") // 上传日期
		fileName := fmt.Sprintf("%s-%s%s", fileMd5, uploadTime, fileExt)
		filePath := filepath.Join(runtimePath, "uploads", uploadDate, fileName)
		filePath, err = filepath.Abs(filePath)
		if err != nil {
			return ctx.CtxError(c, 5001, "创建文件失败", err.Error())
		}

		// 创建图片文件夹
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return ctx.CtxError(c, 5001, "创建文件夹失败", err.Error())
		}

		// 保存文件
		if err := c.SaveFile(file, filePath); err != nil {
			return ctx.CtxError(c, 5001, "文件保存失败", err.Error())
		}

		// 保存图片到数据库
		imageId, err := m.GenImageId()
		if err != nil {
			return ctx.CtxError(c, 5001, "创建图片ID失败", err.Error())
		}

		dbImage.Id = imageId
		dbImage.FilePath = filePath
		dbImage.Md5 = fileMd5
		dbImage.Width = uint16(imageWidth)
		dbImage.Height = uint16(imageHeight)
		dbImage.Ratio = math.Round(float64(imageWidth)/float64(imageHeight)*100) / 100
		dbImage.Size = uint(file.Size)

		tx := g.DB.Create(&dbImage)
		if tx.Error != nil {
			return ctx.CtxError(c, 5001, "保存图片到数据库失败", tx.Error.Error())
		}
	}

	// 查询已存在
	existVerImage, _ := m.GetDatasetVerImageByImageId(req.VersionId, dbImage.Id)
	if existVerImage.Id != 0 {
		return ctx.CtxSuccess(c, fiber.Map{
			"id":       existVerImage.Id,
			"image_id": existVerImage.ImageID,
		}, "图片已存在，无需重新上传")
	}

	// 保存到数据集
	UserId := c.Locals("UserId").(uint32)
	verImage, err := m.AppendDatasetVersionImage(req.VersionId, &dbImage, file.Filename, UserId)
	if err != nil {
		return ctx.CtxError(c, 5006, "保存到数据集失败", err.Error())
	}

	// 计入日志数据
	c.Locals("ContentBody", map[string]interface{}{
		"dataset_id":   req.DatasetID,
		"version_id":   req.VersionId,
		"image_id":     verImage.ImageID,
		"image_name":   file.Filename,
		"image_size":   file.Size,
		"image_width":  width,
		"image_height": height,
	})

	// 返回图片ID 和 版本数据ID
	return ctx.CtxSuccess(c, fiber.Map{
		"id":       verImage.Id,
		"image_id": fmt.Sprint(verImage.ImageID),
	})
}
