package ctxDataset

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

func DatasetDelete(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/delete")
	c.Locals("Content", "删除数据集")

	// 获取参数
	datasetId, err := c.ParamsInt("id")
	if err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 查询数据集
	dataset, err := m.GetDatasetById(uint32(datasetId))
	if err != nil || dataset.Id == 0 {
		return ctx.CtxError(c, 5002, "数据集不存在", err.Error())
	}

	// TODO: 后期可能要验证数据集是否存在正在标注的算法任务

	// 删除数据集
	if err := m.DeleteDataset(uint32(datasetId)); err != nil {
		return ctx.CtxError(c, 5003, "删除数据集失败", err.Error())
	}

	return ctx.CtxSuccess(c, "删除成功")
}
