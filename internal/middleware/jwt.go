package mid

import (
	m "wp_template_display/internal/models"
	"wp_template_display/pkg"

	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func JwtLoginMiddleware(c *fiber.Ctx) error {
	// 获取token
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": 401,
			"msg":  "未登录",
		})
	}

	// 截取token
	token = strings.TrimPrefix(token, "Bearer ")

	// 解析token
	username, err := pkg.JwtCheckToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": 401,
			"msg":  "token过期",
		})
	}

	// 挂载username
	c.Locals("Username", username)
	user, err := m.FindUserByUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": 401,
				"msg":  "用户不存在",
				"data": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code": 401,
			"msg":  "查询失败",
			"data": err.Error(),
		})
	}

	// FindUserByUsername 已限制用户状态 这里不需要判断
	// if user.Status == 0 {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
	// 		"code": 403,
	// 		"msg":  "用户被禁用",
	// 	})
	// }

	c.Locals("User", user)
	c.Locals("UserId", user.Id)

	// 刷新token
	newToken, err := pkg.JwtGenerateToken(username)
	if err == nil {
		c.Set("Access-Control-Expose-Headers", "Authorization")
		c.Set("Authorization", newToken)
	}

	return c.Next()
}

// 根据url获取token
func JwtLoginMiddlewareByUrl(c *fiber.Ctx) error {
	// 获取token
	token := c.Query("_")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": 401,
			"msg":  "未登录",
		})
	}

	// 截取token
	token = strings.TrimPrefix(token, "Bearer ")

	// 解析token
	username, err := pkg.JwtCheckToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": 401,
			"msg":  "token过期",
		})
	}

	// 挂载username
	c.Locals("Username", username)
	user, err := m.FindUserByUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": 404,
				"msg":  "token无效",
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code": 404,
			"msg":  err.Error(),
		})
	}

	if user.Status == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code": 403,
			"msg":  "用户被禁用",
		})
	}

	c.Locals("User", user)
	c.Locals("UserId", user.Id)

	// 刷新token
	newToken, err := pkg.JwtGenerateToken(username)
	if err == nil {
		c.Set("Authorization", "Bearer "+newToken)
	}

	return c.Next()
}
