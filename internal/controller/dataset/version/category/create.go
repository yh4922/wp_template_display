package ctxDatasetVersionCategory

import (
	ctx "wp_template_display/internal/controller"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

// 创建数据集版本类别
type DatasetVersionCategoryCreateReq struct {
	DatasetID   uint32 `json:"dataset_id" validate:"required"` // 数据集ID
	VersionID   uint32 `json:"version_id" validate:"required"` // 版本ID
	Name        string `json:"name" validate:"required"`       // 类别值
	Label       string `json:"label" validate:"required"`      // 类别名称
	Color       string `json:"color" validate:"required"`      // 类别颜色
	ShortcutKey string `json:"shortcut_key"`                   // 快捷键
	Remark      string `json:"remark"`                         // 备注
}

func DatasetVersionCategoryCreate(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/category/create")

	// 验证字段
	req := new(DatasetVersionCategoryCreateReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 验证数据集是否存在
	_, err = m.GetDatasetById(req.DatasetID)
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集不存在", err.Error())
	}

	// 验证数据集版本是否存在
	version, err := m.GetDatasetVerById(req.VersionID)
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集版本不存在", err.Error())
	}

	// 数据集已经发布 则不允许修改
	if version.Status == 1 {
		return ctx.CtxError(c, 5001, "数据集版本已发布", "数据集版本已发布")
	}

	// 验证类别是否存在
	category, err := m.GetCategoryByName(req.VersionID, req.Name)
	if err != nil && err.Error() != "record not found" {
		return ctx.CtxError(c, 5003, "查询失败", err.Error())
	}
	if category.Id != 0 {
		return ctx.CtxError(c, 5003, "类别已存在", "类别已存在")
	}

	// 创建新标签类别
	newCategory := m.SysDatasetVersionCategory{
		DatasetID:   req.DatasetID,
		VersionID:   req.VersionID,
		Name:        req.Name,
		Label:       req.Label,
		Color:       req.Color,
		ShortcutKey: req.ShortcutKey,
		Remark:      req.Remark,
	}
	err = g.DB.Create(&newCategory).Error
	if err != nil {
		return ctx.CtxError(c, 5003, "创建失败", err.Error())
	}

	// 创建成功
	return ctx.CtxSuccess(c, nil)
}
