package ctxUser

import (
	ctx "wp_template_display/internal/controller"

	"github.com/gofiber/fiber/v2"
)

func UserLoginPwdKey(c *fiber.Ctx) error {
	rsaKey, publicKey, err := ctx.GetRsaKey()
	if err != nil {
		return ctx.CtxError(c, 500, err.Error(), nil)
	}

	return ctx.CtxSuccess(c, fiber.Map{
		"key":   rsaKey,
		"value": publicKey,
	})
}
