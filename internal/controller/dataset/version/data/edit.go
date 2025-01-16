package ctxDatasetVersionData

import (
	ctx "wp_template_display/internal/controller"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetVersionDataEditReq struct {
	Id   uint32 `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func DatasetVersionDataEdit(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/data/edit")

	req := new(DatasetVersionDataEditReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	verImage, err := m.GetDatasetVerImageById(req.Id)
	if err != nil {
		if err.Error() == "record not found" {
			return ctx.CtxError(c, 5003, "数据不存在", "图片不存在")
		}
		return ctx.CtxError(c, 5003, "查询失败", err.Error())
	}

	// 修改数据
	err = g.DB.Model(&verImage).Updates(map[string]interface{}{
		"name": req.Name,
	}).Error
	if err != nil {
		return ctx.CtxError(c, 5004, "修改失败", err.Error())
	}

	return ctx.CtxSuccess(c, nil)
}
