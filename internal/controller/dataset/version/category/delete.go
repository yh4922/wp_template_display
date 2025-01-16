package ctxDatasetVersionCategory

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionCategoryDelete(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/category/delete")

	// 类别ID
	categoryId, err := c.ParamsInt("id", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "参数异常", err.Error())
	}
	if categoryId == 0 {
		return ctx.CtxError(c, 5002, "参数异常", "类别ID不能为空")
	}

	// 查找类别
	category, err := m.GetCategoryById(uint32(categoryId))
	if err != nil {
		return ctx.CtxError(c, 5003, "查找类别失败", err.Error())
	}

	// 查找数据版本
	version, err := m.GetDatasetVerById(category.VersionID)
	if err != nil {
		return ctx.CtxError(c, 5004, "查找版本失败", err.Error())
	}

	// 验证版本状态
	if version.Status == 1 {
		return ctx.CtxError(c, 5005, "版本已发布", "版本已发布，不允许删除")
	}

	// 调用数据库删除函数
	markupNum, imageNum, err := m.DeleteCategory(category)
	if err != nil {
		return ctx.CtxError(c, 5006, "删除类别失败", err.Error())
	}

	return ctx.CtxSuccess(c, map[string]interface{}{
		"markup_num":       markupNum,
		"markup_image_num": imageNum,
	})
}
