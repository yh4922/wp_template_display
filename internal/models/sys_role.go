package models

import (
	g "wp_template_display/internal/global"
)

type SysRole struct {
	Model
	Alias  string `json:"alias" gorm:"size:20;not null;comment:别名"`
	Label  string `json:"label" gorm:"size:64;not null;comment:名称标签"`
	Sort   int    `json:"sort" gorm:"comment:排序"`
	Status uint8  `json:"status" gorm:"default:1;comment:状态"` // ENABLE 1   DISABLE 0
	ModelTime
	ControlBy // 公共模字段
}

// / 当前系统采用固定的角色  为了方便后续扩展还是写入数据库 初始化时写入默认角色
// / 初始化角色
func InitRole() *SysRole {
	var admin_role = SysRole{
		Model:  Model{Id: 1},
		Alias:  "admin",
		Label:  "管理员",
		Sort:   0,
		Status: 1,
	}

	// 不存在责创建
	initCreate(&admin_role)

	var mark_role = SysRole{
		Model:  Model{Id: 2},
		Alias:  "mark",
		Label:  "标记员",
		Sort:   1,
		Status: 1,
	}

	initCreate(&mark_role)

	// 返回
	return &admin_role
}

func initCreate(newRole *SysRole) {
	// 不存在责创建
	var role SysRole
	g.DB.Where("id = ?", newRole.Id).First(&role)
	if role.Id == 0 {
		g.DB.Create(newRole)
	}
}
