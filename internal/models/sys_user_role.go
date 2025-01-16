package models

type SysUserRole struct {
	Model
	User   SysUser `json:"user" gorm:"foreignKey:UserID;references:Id"`
	UserID uint32  `json:"user_id" gorm:"not null;comment:用户ID"`
	Role   SysRole `json:"role" gorm:"foreignKey:RoleID;references:Id"`
	RoleID uint32  `json:"role_id" gorm:"not null;comment:角色ID"`
}

// 查询外键需要使用 Preload 预加载需要关联的数据
// var test SysUserRole
// g.DB.Preload("User").Preload("Role").First(&test)
