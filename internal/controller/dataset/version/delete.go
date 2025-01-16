package ctxDatasetVersion

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionDelete(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/delete")

	// 获取ID
	ID, err := c.ParamsInt("id", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "参数异常", err.Error())
	}

	if ID == 0 {
		return ctx.CtxError(c, 5002, "参数异常", "ID不能为空")
	}

	// 获取数据集版本
	ver, err := m.GetDatasetVerById(uint32(ID))
	if err != nil {
		return ctx.CtxError(c, 5003, "数据集版本不存在", err.Error())
	}

	// 删除数据集版本
	if err := m.DeleteDatasetVer(ver); err != nil {
		return ctx.CtxError(c, 5004, "数据集版本删除失败", err.Error())
	}

	return ctx.CtxSuccess(c, "数据集版本删除成功")
}
