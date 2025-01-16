package ctxDataset

import (
	"wp_template_display/internal/consts"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetCreateReq struct {
	Name     string `json:"name" validate:"required"`
	AlgType  string `json:"alg_type" validate:"required"`
	DataType string `json:"data_type" validate:"required,oneof=train test"`
	Remark   string `json:"remark"`
}

func DatasetCreate(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/create")

	req := new(DatasetCreateReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 判断 req.AlgType 是否在 consts.AlgorithmTypeList 中
	algTypeValid := consts.CheckAlgorithmTypeValidity(req.AlgType)
	if !algTypeValid {
		return ctx.CtxError(c, 5003, "算法类型不存在", nil)
	}

	UserId := c.Locals("UserId").(uint32)

	dataset := &m.SysDataset{
		Name:       req.Name,
		DataType:   req.DataType,
		Remark:     req.Remark,
		AlgType:    req.AlgType,
		VersionNum: 1,
		Version:    1,
		ControlBy: m.ControlBy{
			CreatedBy: UserId, // 创建者
		},
	}

	err = m.CreateDataset(dataset)
	if err != nil {
		return ctx.CtxError(c, 5004, err.Error(), nil)
	}

	return ctx.CtxSuccess(c, dataset)
}
