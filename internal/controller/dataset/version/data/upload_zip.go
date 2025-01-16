package ctxDatasetVersionData

import (
	"os"
	pathUtil "path"
	"strings"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"
	"wp_template_display/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/gookit/config/v2"
)

type DatasetVersionDataUploadZipReq struct {
	DatasetID   uint32 `form:"dataset_id" validate:"required"`  // 数据集ID
	VersionId   uint32 `form:"version_id" validate:"required"`  // 数据版本ID
	Deduplicate bool   `form:"deduplicate" validate:"required"` // 是否去重 true 删除重复图片数据导入新的  false 保留现有不导入标注
	// ZipFile   string `form:"file" validate:"required"`       // 压缩包文件
}

func DatasetVersionDataUploadZip(c *fiber.Ctx) error {
	// 上传文件
	ZipFile, err := c.FormFile("file")
	if err != nil {
		return ctx.CtxError(c, 5001, "文件异常", err.Error())
	}

	// 验证文件后缀名
	if !strings.HasSuffix(strings.ToLower(ZipFile.Filename), ".zip") {
		return ctx.CtxError(c, 5001, "文件格式错误", "仅支持zip格式文件")
	}

	// 解析请求参数
	req := DatasetVersionDataUploadZipReq{}
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

	// 数据集已经发布 则不允许修改
	if version.Status == 1 {
		return ctx.CtxError(c, 5001, "数据集版本已发布", "数据集版本已发布")
	}

	// 生成文件名
	var fileId string
	newId, err := uuid.NewV7()
	if err != nil {
		return ctx.CtxError(c, 5001, "任务ID生成失败", err.Error())
	}
	fileId = newId.String()

	// 临时目录
	runtimePath := config.String("RuntimePath", "runtime")
	tempPath := pathUtil.Join(runtimePath, "temp", fileId+".zip")
	tempDir := pathUtil.Dir(tempPath)

	// 创建临时目录
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return ctx.CtxError(c, 5001, "临时目录创建失败", err.Error())
	}

	// 保存文件
	if err := c.SaveFile(ZipFile, tempPath); err != nil {
		return ctx.CtxError(c, 5001, "文件保存失败", err.Error())
	}

	// 创建处理任务
	UserId := c.Locals("UserId").(uint32)
	taskId, err := utils.HandleMarkupZipFile(tempPath, version, UserId, req.Deduplicate)
	if err != nil {
		return ctx.CtxError(c, 5001, "文件处理失败", err.Error())
	}

	// 获取文件MD5
	return ctx.CtxSuccess(c, fiber.Map{
		"status":  1,
		"task_id": taskId,
	}, "ZIP上传成功，正在处理")
}
