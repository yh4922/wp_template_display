package models

import (
	"fmt"
	"time"
	g "wp_template_display/internal/global"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// INFO: 为了兼容分表，暂时取消外键约束
// Dataset         SysDataset                `json:"dataset" gorm:"foreignKey:DatasetID;references:Id"`
// Version         SysDatasetVersion         `json:"version" gorm:"foreignKey:VersionID;references:Id"`
// VersionImage    SysDatasetVersionImage    `json:"version_image" gorm:"foreignKey:VersionImageID;references:Id"`
// TagCategory     SysDatasetVersionCategory `json:"tag_category" gorm:"foreignKey:TagCategoryId;references:Id"`

const TableCount = 5

type SysDatasetVersionMarkup struct {
	Id              string    `json:"id" gorm:"size:40;primaryKey;unique;comment:主键编码"`
	DatasetID       uint32    `json:"dataset_id" gorm:"not null;comment:数据集ID"`
	VersionID       uint32    `json:"version_id" gorm:"not null;comment:数据集版本ID"`
	VersionImageID  uint32    `json:"version_image_id" gorm:"not null;comment:图片ID"`
	Serial          uint16    `json:"serial" gorm:"not null;comment:序号"`
	MarkupType      string    `json:"markup_type" gorm:"size:20;not null;comment:标记类型"`
	MarkupPoints    string    `json:"markup_points" gorm:"type:text;not null;comment:点位信息"`
	TagCategoryId   uint32    `json:"tag_category_id" gorm:"not null;comment:类别标识"`
	TagCategoryName string    `json:"tag_category_name" gorm:"size:10;not null;comment:类别标识"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:最后更新时间"`
	ControlBy
}

type SysDatasetVersionMarkupJson struct {
	Id            string    `json:"id" validate:"required"`
	Serial        uint      `json:"serial" validate:"required"`
	MarkupType    string    `json:"markup_type" validate:"required"`
	MarkupPoints  []float64 `json:"markup_points" validate:"required"`
	TagCategoryId uint32    `json:"tag_category_id" validate:"required"`
}

func (m *SysDatasetVersionMarkup) TableName() string {
	tableIndex := m.VersionID % TableCount
	tableName := fmt.Sprintf("sys_dataset_version_markups_%d", tableIndex)
	return tableName
}

// 通过版本ID获取分表
func MarkupTableOfVer(versionID uint32) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tableIndex := versionID % TableCount
		tableName := fmt.Sprintf("sys_dataset_version_markups_%d", tableIndex)
		return db.Table(tableName)
	}
}

// 生成标记ID
func GenMarkupId() string {
	newId, err := uuid.NewV7()
	if err != nil {
		return ""
	}
	return newId.String()
}

// 初始化分表
func InitDatasetVerMarkupTables() {
	// 创建分表
	for i := 0; i < TableCount; i++ {
		g.DB.Table(fmt.Sprintf("sys_dataset_version_markups_%d", i)).AutoMigrate(&SysDatasetVersionMarkup{})
	}

	// println("测试", 1231231)
	// m := SysDatasetVersionMarkup{
	// 	Id:              "1212",
	// 	DatasetID:       1,
	// 	VersionID:       200,
	// 	VersionImageID:  1,
	// 	Serial:          12,
	// 	MarkupType:      "test",
	// 	MarkupPoints:    "test",
	// 	TagCategoryId:   1,
	// 	TagCategoryName: "test",
	// }
}
