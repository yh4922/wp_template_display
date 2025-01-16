package ctxDatasetVersionCategory

import (
	ctx "wp_template_display/internal/controller"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

// 创建数据集版本类别
type DatasetVersionCategoryEditReq struct {
	ID          uint32 `json:"id" validate:"required"`    // 类别ID
	Label       string `json:"label" validate:"required"` // 类别名称
	Color       string `json:"color" validate:"required"` // 类别颜色
	ShortcutKey string `json:"shortcut_key"`              // 快捷键
	Remark      string `json:"remark"`                    // 备注
}

func DatasetVersionCategoryEdit(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/category/edit")

	// 验证字段
	req := new(DatasetVersionCategoryEditReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 查找标签类别
	category, err := m.GetCategoryById(req.ID)
	if err != nil {
		return ctx.CtxError(c, 5003, "查找标签类别失败", err.Error())
	}

	// 验证版本状态
	version, err := m.GetDatasetVerById(category.VersionID)
	if err != nil {
		return ctx.CtxError(c, 5004, "查找版本失败", err.Error())
	}
	if version.Status == 1 {
		return ctx.CtxError(c, 5005, "版本已发布", "版本已发布，不允许修改")
	}

	// 更新标签类别
	category.Label = req.Label
	category.Color = req.Color
	category.ShortcutKey = req.ShortcutKey
	category.Remark = req.Remark

	// 更新数据库
	tx := g.DB.Save(category)
	if tx.Error != nil {
		return ctx.CtxError(c, 5004, "更新失败", tx.Error.Error())
	}

	return ctx.CtxSuccess(c, nil)
}
