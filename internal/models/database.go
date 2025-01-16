package models

import (
	"time"
	g "wp_template_display/internal/global"
	"wp_template_display/pkg"

	"gorm.io/gorm"
)

type Model struct {
	Id uint32 `json:"id" gorm:"primaryKey;autoIncrement;comment:主键编码"`
}

type ControlBy struct {
	CreatedBy uint32 `json:"create_by" gorm:"index;comment:创建者"`
	UpdatedBy uint32 `json:"update_by" gorm:"index;comment:更新者"`
}

type ModelTime struct {
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:最后更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:删除时间"`
}

// 初始化数据库
func InitModel() {
	// 初始化数据表
	{
		// 日志
		g.DB.AutoMigrate(&SysLog{})
		// 图片文件
		g.DB.AutoMigrate(&SysImage{})
		// 配置
		g.DB.AutoMigrate(&SysConfig{})
		// 用户角色
		g.DB.AutoMigrate(&SysUser{})
		g.DB.AutoMigrate(&SysRole{})
		g.DB.AutoMigrate(&SysUserRole{})
		// 数据集相关
		g.DB.AutoMigrate(&SysDataset{})
		g.DB.AutoMigrate(&SysDatasetVersion{})
		g.DB.AutoMigrate(&SysDatasetVersionCategory{})
		g.DB.AutoMigrate(&SysDatasetVersionImage{})
		InitDatasetVerMarkupTables() // 标注表
		// 任务相关
		g.DB.AutoMigrate(&SysSmartMarkTask{}) // 智能标注任务表
		// 模型训练任务表
		// 模型测试任务表

		// 测试表
		InitTestTables() // 标注表
	}

	/// 初始化角色
	role := InitRole()

	// 查找SysUser 是否存在 Username = admin 的用户
	var user SysUser
	g.DB.Where("username = ?", "admin").First(&user)
	if user.Id == 0 {
		user = SysUser{
			Username: "admin",
			Nickname: "管理员",
			Password: pkg.PwdEncode("123456"),
			Email:    "superadmin@example.com",
			Phone:    "1234567890",
			Status:   1,
		}
		g.DB.Create(&user)
	}

	// 查找sys_user_role 是否存在 UserID = user.Id RoleID = role.Id
	var userRole SysUserRole
	g.DB.Where("user_id = ? AND role_id = ?", user.Id, role.Id).First(&userRole)
	if userRole.Id == 0 {
		userRole = SysUserRole{UserID: user.Id, RoleID: role.Id}
		g.DB.Create(&userRole)
	}
}
