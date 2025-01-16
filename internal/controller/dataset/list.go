package ctxDataset

import (
	"wp_template_display/internal/consts"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetListReq struct {
	Page     int    `query:"page"`                                  // 页码
	Size     int    `query:"size"`                                  // 每页数量
	DataType string `query:"data_type" validate:"oneof=train test"` // 数据集类型
	Name     string `query:"name"`                                  // 名称（模糊查询）
	AlgType  string `query:"alg_type"`                              // 算法类型
}

func DatasetList(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/list")

	req := new(DatasetListReq)
	if err := c.QueryParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 默认数据
	req.Page = c.QueryInt("page", 1)
	req.Size = c.QueryInt("size", 15)
	req.DataType = c.Query("data_type", "train")

	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 判断 req.AlgType 是否在 consts.AlgorithmTypeList 中
	if req.AlgType != "" && !consts.CheckAlgorithmTypeValidity(req.AlgType) {
		return ctx.CtxError(c, 5003, "算法类型不存在", nil)
	}

	// 查询数据集列表
	datasets, total, err := m.GetDatasetList(req.Page, req.Size, req.DataType, req.Name, req.AlgType)
	if err != nil {
		return ctx.CtxError(c, 5004, "查询失败", err.Error())
	}

	// 返回数据
	data := fiber.Map{
		"list": datasets,
		"page": map[string]int{
			"page":  req.Page,
			"size":  req.Size,
			"count": int(total),
		},
	}

	return ctx.CtxSuccess(c, data)
}
