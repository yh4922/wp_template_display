package ctxDataset

import (
	"wp_template_display/internal/consts"
	ctx "wp_template_display/internal/controller"

	"github.com/gofiber/fiber/v2"
)

func DatasetGetAlgType(c *fiber.Ctx) error {
	return ctx.CtxSuccess(c, consts.AlgorithmTypeList)
}
