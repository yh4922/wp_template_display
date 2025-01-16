package ctxDatasetVersionData

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionDataPreview(c *fiber.Ctx) error {
	// 获取ID
	ID, err := c.ParamsInt("id", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "参数异常", err.Error())
	}

	if ID == 0 {
		return ctx.CtxError(c, 5002, "参数异常", "ID不能为空")
	}

	// 获取图片
	image, err := m.GetImageById(uint64(ID))
	if err != nil {
		return ctx.CtxError(c, 5003, "图片不存在", err.Error())
	}
	// println(image.FilePath)

	return c.Status(fiber.StatusOK).SendFile(image.FilePath)
}
