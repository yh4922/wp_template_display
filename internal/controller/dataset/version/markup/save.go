package ctxDatasetVersionMarkup

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetVersionMarkupSaveReq struct {
	ImageId    uint32                          `json:"image_id" validate:"required"`
	MarkupList []m.SysDatasetVersionMarkupJson `json:"markup_list" validate:"required"`
}

func DatasetVersionMarkupSave(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/markup/save")

	req := new(DatasetVersionMarkupSaveReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}
	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 查找图片
	image, err := m.GetPreDatasetVerImageById(uint32(req.ImageId))
	if err != nil {
		return ctx.CtxError(c, 5003, "图片不存在", err.Error())
	}
	// 验证版本是否存在
	if image.Version.Id == 0 {
		return ctx.CtxError(c, 5004, "版本不存在", "数据集版本不存在")
	}
	// 验证版本是否已发布
	if image.Version.Status == 1 {
		return ctx.CtxError(c, 5005, "版本已发布", "数据集版本已发布，无法进行标记")
	}

	// 验证当前是否存在智能标注任务 存在则报错
	// task :=

	// TODO: 保存标记信息
	// 1.查找旧的标记
	// 2.删除旧标记
	// 3.保存新标记
	// 4.更新标记数量

	// 返回
	return ctx.CtxSuccess(c, image.Version)
}
