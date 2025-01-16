package mid

import (
	"slices"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

// 用户权限中间件
func UserAuthMiddleware(roles []string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := c.Locals("User").(*m.SysUser)

		// 查找用户关联的角色
		var userRoles []m.SysUserRole
		tx := g.DB.Preload("Role").Where("user_id = ?", user.Id).Find(&userRoles)
		if tx.Error != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"msg":  "权限不足",
			})
		}

		isAuth := false

		for _, role := range userRoles {
			if slices.Contains(roles, role.Role.Alias) {
				isAuth = true
				break
			}
		}

		if !isAuth {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code": 403,
				"msg":  "权限不足",
			})
		}

		return c.Next()
	}
}
