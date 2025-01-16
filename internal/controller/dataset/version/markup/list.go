package ctxDatasetVersionMarkup

import (
	ctx "wp_template_display/internal/controller"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionMarkupList(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/markup/list")

	// 获取图片ID
	imageId, err := c.ParamsInt("id", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "图片ID错误", err.Error())
	}
	if imageId == 0 {
		return ctx.CtxError(c, 5002, "图片ID不能为空", "")
	}

	// 查询图片
	image, err := m.GetDatasetVerImageById(uint32(imageId))
	if err != nil {
		return ctx.CtxError(c, 5003, "图片不存在", err.Error())
	}

	// 查询图片
	var markups []m.SysDatasetVersionMarkup
	err = g.DB.Scopes(m.MarkupTableOfVer(image.VersionID)).Where("version_image_id = ?", imageId).Order("serial ASC").Find(&markups).Error
	if err != nil {
		return ctx.CtxError(c, 5004, "获取标注失败", err.Error())
	}

	return ctx.CtxSuccess(c, markups)
}
