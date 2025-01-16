package models

import "time"

// 上传
type SysSmartMarkTask struct {
	Model
	Dataset             SysDataset        `json:"-" gorm:"foreignKey:DatasetID;references:Id"`
	DatasetID           uint32            `json:"dataset_id" gorm:"not null;comment:数据集ID"`
	DataType            string            `json:"data_type" gorm:"size:10;not null;comment: 数据集类型"` // train 训练集   test 测试集
	Version             SysDatasetVersion `json:"-" gorm:"foreignKey:VersionID;references:Id"`
	VersionID           uint32            `json:"version_id" gorm:"not null;comment:数据集版本ID"`
	CategoryNum         uint16            `json:"category_num" gorm:"not null;comment:标注类别数量"`
	Status              uint8             `json:"status" gorm:"not null;comment:任务状态"` // 0初始化  1运行中  2待确认  3已完成   4已终止   5已完成
	ExpectedCompletedAt *time.Time        `json:"expected_completed_at" gorm:"comment:预计完成时间"`
	CompletedAt         *time.Time        `json:"completed_at" gorm:"comment:实际完成时间"`
	ModelTime
	ControlBy // 公共模字段
}
