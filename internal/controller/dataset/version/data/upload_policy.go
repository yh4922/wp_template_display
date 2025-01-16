package ctxDatasetVersionData

import (
	"fmt"
	"path"
	"strings"
	"time"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"
	"wp_template_display/pkg"

	"github.com/gofiber/fiber/v2"
)

type DatasetVersionDataUploadPolicyReq struct {
	DatasetID   uint32 `json:"dataset_id" validate:"required"`   // 数据集ID
	VersionId   uint32 `json:"version_id" validate:"required"`   // 数据版本ID
	FileName    string `json:"file_name" validate:"required"`    // 文件名称
	FileMd5     string `json:"file_md5" validate:"required"`     // 文件md5
	ImageWidth  int32  `json:"image_width" validate:"required"`  // 图片宽度
	ImageHeight int32  `json:"image_height" validate:"required"` // 图片高度
}

func DatasetVersionDataUploadPolicy(c *fiber.Ctx) error {
	// // 模型名称
	// c.Locals("Action", "/api/v1/dataset/version/data/upload-policy")

	req := DatasetVersionDataUploadPolicyReq{}
	if err := c.BodyParser(&req); err != nil {
		return ctx.CtxError(c, 5001, "请求参数错误", err.Error())
	}

	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 限制宽度 图片分辨率限制：64px*64px ≤ 图片分辨率 ≤ 4096px*4096px
	if req.ImageWidth < 64 || req.ImageWidth > 4096 || req.ImageHeight < 64 || req.ImageHeight > 4096 {
		return ctx.CtxError(c, 5003, "图片分辨率超出限制", "图片分辨率限制：64px*64px ≤ 图片分辨率 ≤ 4096px*4096px")
	}

	// 支持图片格式：jpg、jpeg、png
	supportFormat := []string{".jpg", ".jpeg", ".png"}
	fileExt := strings.ToLower(path.Ext(req.FileName)) // 获取文件后缀
	println("fileExt", fileExt)
	isSupport := false // 判断是否支持的格式
	for _, ext := range supportFormat {
		if ext == fileExt {
			isSupport = true
			break
		}
	}
	if !isSupport {
		return ctx.CtxError(c, 5003, "不支持的图片格式", "支持图片格式：jpg、jpeg、png")
	}

	// 支持图片名称长度：1-100字符
	if len(req.FileName) < 1 || len(req.FileName) > 100 {
		return ctx.CtxError(c, 5003, "图片名称长度超出限制", "图片名称长度限制：1-100字符")
	}

	// 图片名称不可以包含特殊符号 比如 $、%、&
	specialSymbol := []string{"$", "%", "&"}
	for _, symbol := range specialSymbol {
		if strings.Contains(req.FileName, symbol) {
			return ctx.CtxError(c, 5003, "图片名称包含特殊符号", "图片名称不可以包含特殊符号 比如 $、%、&")
		}
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

	// 查找图片
	image, err := m.GetImageByMd5(req.FileMd5)
	if err != nil && err.Error() != "record not found" {
		return ctx.CtxError(c, 5005, "查询失败", err.Error())
	}

	// 图片存在
	if image.Id != 0 {
		UserId := c.Locals("UserId").(uint32)

		// 保存到数据集
		verImage, err := m.AppendDatasetVersionImage(req.VersionId, &image, req.FileName, UserId)
		if err != nil {
			return ctx.CtxError(c, 5006, "保存到数据集失败", err.Error())
		}

		// 返回图片ID 和 版本数据ID
		return ctx.CtxSuccess(c, fiber.Map{
			"id":       verImage.Id,
			"image_id": fmt.Sprint(verImage.ImageID),
		}, "已存在重复图片，无需重新上传")
	}

	// 生成过期时间 需要在30s内上传
	expireTime := time.Now().Add(time.Second * 30).Unix()

	// 创建凭证
	singStr := fmt.Sprintf("%d-%d-%d-%s", expireTime, req.ImageWidth, req.ImageHeight, req.FileMd5)
	singStr = pkg.AesEncrypt(singStr)

	// 返回凭证
	return ctx.CtxSuccess(c, fiber.Map{"token": singStr})
}
