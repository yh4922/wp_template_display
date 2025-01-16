package models

import (
	g "wp_template_display/internal/global"
)

type SysUser struct {
	Model
	Username  string `json:"username" gorm:"not null;unique;size:20;comment:用户名"`
	Avatar    string `json:"avatar" gorm:"size:255;comment:头像"`
	Nickname  string `json:"nickname" gorm:"size:64;comment:昵称"`
	Password  string `json:"password" gorm:"size:255;not null;comment:密码"`
	Email     string `json:"email" gorm:"size:64;comment:邮箱"`
	Phone     string `json:"phone" gorm:"size:20;comment:手机号"`
	Status    uint8  `json:"status" gorm:"default:1;comment:状态"` // ENABLE 1   DISABLE 0
	ParentUid uint32 `json:"parent_uid" gorm:"comment:父级ID"`
	ModelTime
	ControlBy
}

// 通过 username 查找用户
func FindUserByUsername(username string) (*SysUser, error) {
	var user SysUser
	tx := g.DB.Where("username = ? AND status = 1", username).First(&user)
	if tx.Error != nil {
		return &user, tx.Error
	}
	return &user, nil
}

// 获取用户基础关联信息 通过 id
func GetUserBaseInfoById(id uint32) (*SysUser, []*SysRole, error) {
	var user SysUser
	tx := g.DB.Where("id = ? AND status = 1 AND deleted_at IS NULL", id).First(&user)
	if tx.Error != nil {
		return &user, []*SysRole{}, tx.Error
	}

	// 查找用户关联的角色
	var userRoles []SysUserRole
	tx = g.DB.Where("user_id = ?", id).Find(&userRoles)
	if tx.Error != nil {
		return &user, []*SysRole{}, tx.Error
	}

	var roleIds []uint32
	for _, v := range userRoles {
		roleIds = append(roleIds, v.RoleID)
	}

	// 查找角色
	var roles []*SysRole
	tx = g.DB.Where("id IN (?)", roleIds).Find(&roles)
	if tx.Error != nil {
		return &user, []*SysRole{}, tx.Error
	}

	return &user, roles, nil
}
