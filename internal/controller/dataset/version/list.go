package ctxDatasetVersion

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetVersionListReq struct {
	Page      int    `query:"page"`                           // 页码
	Size      int    `query:"size"`                           // 每页数量
	DatasetId uint32 `query:"dataset_id" validate:"required"` // 数据集ID
}

func DatasetVersionList(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/list")

	req := new(DatasetVersionListReq)
	if err := c.QueryParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 默认数据
	req.Page = c.QueryInt("page", 1)
	req.Size = c.QueryInt("size", 15)

	if req.DatasetId == 0 {
		return ctx.CtxError(c, 5002, "参数错误", "数据集ID不能为空")
	}

	// 获取数据集
	_, err := m.GetDatasetById(req.DatasetId)
	if err != nil {
		return ctx.CtxError(c, 5003, "数据集不存在", err.Error())
	}

	// 获取数据集版本
	versionList, total, err := m.GetDatasetVerByDatasetID(req.Page, req.Size, req.DatasetId)
	if err != nil {
		return ctx.CtxError(c, 5004, "获取数据集版本失败", err.Error())
	}

	// 返回数据
	data := fiber.Map{
		"list": versionList,
		"page": map[string]int{
			"page":  req.Page,
			"size":  req.Size,
			"count": int(total),
		},
	}

	return ctx.CtxSuccess(c, data)
}
