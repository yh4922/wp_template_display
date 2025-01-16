package ctxDataset

import (
	ctx "wp_template_display/internal/controller"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetEditReq struct {
	Id     uint32 `json:"id" validate:"required"`
	Name   string `json:"name" validate:"required"`
	Remark string `json:"remark"`
}

func DatasetEdit(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/edit")
	c.Locals("Content", "编辑数据集")

	req := new(DatasetEditReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	dataset, err := m.GetDatasetById(req.Id)
	if err != nil {
		return ctx.CtxError(c, 5003, "数据不存在", err.Error())
	}

	// 修改数据
	dataset.Name = req.Name
	dataset.Remark = req.Remark
	dataset.ControlBy.UpdatedBy = c.Locals("UserId").(uint32)

	// 更新数据库
	tx := g.DB.Save(dataset)
	if tx.Error != nil {
		return ctx.CtxError(c, 5004, "更新失败", tx.Error.Error())
	}

	return ctx.CtxSuccess(c, nil)
}
