package ctxDatasetVersion

import (
	"wp_template_display/internal/consts"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetVersionCreateReq struct {
	DatasetID uint32 `json:"dataset_id" validate:"required"`         // 数据集ID
	Type      uint   `json:"type" validate:"required,oneof=1 2"`     // 1:继承版本 2:全新版本
	Inherit   uint32 `json:"inherit" validate:"required_if=Type 1"`  // 继承的版本 如果type=1 则必传
	AlgType   string `json:"alg_type" validate:"required_if=Type 2"` // 算法类型 如果type=2 则必传
	Remark    string `json:"remark"`                                 // 版本备注
}

// 创建
func DatasetVersionCreate(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/create")

	req := new(DatasetVersionCreateReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 判断算法类型是否存在
	if req.Type == 2 {
		algTypeValid := consts.CheckAlgorithmTypeValidity(req.AlgType)
		if !algTypeValid {
			return ctx.CtxError(c, 5003, "算法类型不存在", nil)
		}
	}

	UserId := c.Locals("UserId").(uint32)

	if req.Type == 1 { // 继承版本
		ver := &m.SysDatasetVersion{
			DatasetID: req.DatasetID,
			Inherit:   req.Inherit,
			Remark:    req.Remark,
			ControlBy: m.ControlBy{
				CreatedBy: UserId, // 创建者
			},
		}

		// 创建继承版本
		err := m.CreateDatasetVerForInherit(ver)
		if err != nil {
			return ctx.CtxError(c, 5004, "创建失败", err.Error())
		}
	} else if req.Type == 2 { // 全新创建
		ver := &m.SysDatasetVersion{
			DatasetID: req.DatasetID,
			AlgType:   req.AlgType,
			Remark:    req.Remark,
			ControlBy: m.ControlBy{
				CreatedBy: UserId, // 创建者
			},
		}

		// 创建全新版本
		err := m.CreateDatasetVer(ver)
		if err != nil {
			return ctx.CtxError(c, 5005, "创建失败", err.Error())
		}
	}

	return ctx.CtxSuccess(c, "创建成功")
}
