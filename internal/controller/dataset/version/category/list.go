package ctxDatasetVersionCategory

import (
	ctx "wp_template_display/internal/controller"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionCategoryList(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/category/list")

	// 获取参数
	datasetId, err := c.ParamsInt("dataset", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "参数异常", err.Error())
	}
	versionId, err := c.ParamsInt("version", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "参数异常", err.Error())
	}

	if datasetId == 0 || versionId == 0 {
		return ctx.CtxError(c, 5002, "参数异常", "数据集ID或版本ID不能为空")
	}

	// 验证数据集是否存在
	_, err = m.GetDatasetById(uint32(datasetId))
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集不存在", err.Error())
	}

	// 验证数据集版本是否存在
	version, err := m.GetDatasetVerById(uint32(versionId))
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集版本不存在", err.Error())
	}

	if version.DatasetID != uint32(datasetId) {
		return ctx.CtxError(c, 5004, "版本ID不匹配", "版本关联的数据集ID与请求参数不匹配")
	}

	// 获取标签类别
	var categories []m.SysDatasetVersionCategory
	err = g.DB.Where("version_id = ?", versionId).Find(&categories).Error
	if err != nil {
		return ctx.CtxError(c, 5004, "查找标签类别失败", err.Error())
	}

	// 设置默认的颜色
	for _, category := range categories {
		if category.Color == "" {
			category.Color = "#FF0000"
		}
	}

	return ctx.CtxSuccess(c, categories)
}
