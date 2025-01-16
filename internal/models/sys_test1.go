package models

import (
	"fmt"
	g "wp_template_display/internal/global"

	"gorm.io/gorm"
)

const testTableCount = 5

// 测试数据库
type SysTest struct {
	Model
	A int `json:"a" gorm:"comment:a"`
	B int `json:"b" gorm:"comment:b"`
	C int `json:"c" gorm:"comment:c"`
	D int `json:"d" gorm:"comment:d"`
	E int `json:"e" gorm:"comment:e"`
	F int `json:"f" gorm:"comment:f"`
	G int `json:"g" gorm:"comment:g"`
	H int `json:"h" gorm:"comment:h"`
	I int `json:"i" gorm:"comment:i"`
	J int `json:"j" gorm:"comment:j"`
	ModelTime
	ControlBy // 公共模字段
}

func TestTableOfA(a int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tableIndex := a % testTableCount
		tableName := fmt.Sprintf("sys_tests_%d", tableIndex)
		return db.Table(tableName)
	}
}

func InitTestTables() {
	// 创建分表
	for i := 0; i < testTableCount; i++ {
		g.DB.Table(fmt.Sprintf("sys_tests_%d", i)).AutoMigrate(&SysTest{})
	}

	// // 插入数据
	// g.DB.Scopes(TestTableOfA(2)).Create(&SysTest{A: 2, B: 1, C: 1, D: 1, E: 1, F: 1, G: 1, H: 1, I: 1, J: 1})
	// g.DB.Scopes(TestTableOfA(2)).Create(&SysTest{A: 2, B: 2, C: 2, D: 2, E: 2, F: 2, G: 2, H: 2, I: 2, J: 2})
	// g.DB.Scopes(TestTableOfA(2)).Create(&SysTest{A: 2, B: 3, C: 3, D: 3, E: 3, F: 3, G: 3, H: 3, I: 3, J: 3})
	// g.DB.Scopes(TestTableOfA(2)).Create(&SysTest{A: 2, B: 4, C: 4, D: 4, E: 4, F: 4, G: 4, H: 4, I: 4, J: 4})
	// g.DB.Scopes(TestTableOfA(2)).Create(&SysTest{A: 2, B: 5, C: 5, D: 5, E: 5, F: 5, G: 5, H: 5, I: 5, J: 5})

	// g.DB.Scopes(TestTableOfA(3)).Create(&SysTest{A: 3, B: 1, C: 1, D: 1, E: 1, F: 1, G: 1, H: 1, I: 1, J: 1})

	// // 查询数据
	// var count int64
	// g.DB.Scopes(TestTableOfA(2)).Where("b = ? AND deleted_at IS NULL", 3).Count(&count)
	// println("count", count)

	// var count1 int64
	// g.DB.Model(&SysConfig{}).Where("id = 1").Count(&count1)
	// // g.DB.Scopes(func(db *gorm.DB) *gorm.DB {
	// // 	return db.Table("sys_configs")
	// // }).Where("id = 1").Count(&count1)
	// println("count1", count1)
}
