package ctxDatasetVersion

import (
	"time"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"
	"wp_template_display/internal/rules"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionPublish(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/publish")

	// 获取ID
	ID, err := c.ParamsInt("id", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "参数异常", err.Error())
	}

	if ID == 0 {
		return ctx.CtxError(c, 5002, "参数异常", "ID不能为空")
	}

	// 获取数据集版本
	ver, err := m.GetDatasetVerById(uint32(ID))
	if err != nil {
		return ctx.CtxError(c, 5003, "数据集版本不存在", err.Error())
	}

	// 验证数据集是否已发布
	if ver.Status == 1 {
		return ctx.CtxError(c, 5004, "数据集版本已发布", nil)
	}

	// TODO: 通过预设规则，判断数据集是否满足发布条件
	if ver.ImageNum < 10 {
		return ctx.CtxError(c, 5005, "发布失败", []rules.RulesItem{
			{
				Rule: "图片数量≥10",
				Pass: false,
			},
		})
	}

	// // 发布数据集版本
	// // 发布数据集版本
	ver.Status = 1
	PublishAt := time.Now()
	ver.PublishAt = &PublishAt

	// 保存数据集版本
	if err := m.PublishDatasetVer(ver); err != nil {
		return ctx.CtxError(c, 5005, "数据集版本发布失败", err.Error())
	}

	return ctx.CtxSuccess(c, "发布成功")
}
