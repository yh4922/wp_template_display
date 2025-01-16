package ctxDatasetVersionData

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionDataDelete(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/data/delete")

	// 获取ID
	ID, err := c.ParamsInt("id", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "参数异常", err.Error())
	}

	// 不能为空
	if ID == 0 {
		return ctx.CtxError(c, 5002, "参数异常", "ID不能为空")
	}

	// 获取数据集版本图片
	verImage, err := m.GetDatasetVerImageById(uint32(ID))
	if err != nil {
		if err.Error() == "record not found" {
			return ctx.CtxError(c, 5003, "数据集版本图片不存在", "数据集版本图片不存在")
		}
		return ctx.CtxError(c, 5003, "获取数据集版本图片失败", err.Error())
	}

	// 获取版本
	version, err := m.GetDatasetVerById(verImage.VersionID)
	if err != nil {
		return ctx.CtxError(c, 5004, "获取数据集版本失败", err.Error())
	}
	if version.Id == 0 {
		return ctx.CtxError(c, 5005, "数据集版本不存在", "数据集版本不存在")
	}
	if version.Status == 1 {
		return ctx.CtxError(c, 5006, "数据集版本已发布", "数据集版本已发布")
	}

	// 删除数据集版本图片
	UserId := c.Locals("UserId").(uint32)
	err = m.DeleteDatasetVerImage(verImage, UserId)
	if err != nil {
		return ctx.CtxError(c, 5007, "删除数据集版本图片失败", err.Error())
	}

	return ctx.CtxSuccess(c, nil)
}
