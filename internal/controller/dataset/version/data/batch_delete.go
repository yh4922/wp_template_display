package ctxDatasetVersionData

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetVersionDataBatchDeleteReq struct {
	Ids []uint32 `json:"ids" validate:"required"`
}

func DatasetVersionDataBatchDelete(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/data/batch-delete")

	req := new(DatasetVersionDataBatchDeleteReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}
	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 查询图片列表
	verImages, err := m.GetDatasetVerImageListByIds(req.Ids)
	if err != nil {
		return ctx.CtxError(c, 5003, "查询失败", err.Error())
	}

	// 有不存在的图片
	if len(verImages) == 0 || len(verImages) != len(req.Ids) {
		return ctx.CtxError(c, 5004, "图片不存在", "图片不存在")
	}

	// 版本不一致
	var versionId uint32 = 0
	for _, verImage := range verImages {
		if versionId == 0 {
			versionId = verImage.VersionID
		}
		if versionId != verImage.VersionID {
			return ctx.CtxError(c, 5005, "版本不一致", "只能操作同一个版本的数据")
		}
	}

	version, err := m.GetDatasetVerById(versionId)
	if err != nil {
		return ctx.CtxError(c, 5006, "查询版本失败", err.Error())
	}

	// 删除图片
	UserId := c.Locals("UserId").(uint32)
	err = m.BatchDeleteDatasetVerImage(version, verImages, UserId)
	if err != nil {
		return ctx.CtxError(c, 5006, "删除失败", err.Error())
	}

	return ctx.CtxSuccess(c, nil)
}
