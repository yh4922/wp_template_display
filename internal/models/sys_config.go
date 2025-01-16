package models

type SysConfig struct {
	Model
	Key    string `json:"key" gorm:"size:20;not null;unique;comment:键"`
	Label  string `json:"label" gorm:"size:20;not null;comment:标签"`
	Value  string `json:"value" gorm:"size:20;comment:值"`
	Remark string `json:"remark" gorm:"size:150;comment:备注"`
	Status uint8  `json:"status" gorm:"size:1;default:1;comment:状态"` // ENABLE 1   DISABLE 0
	ModelTime
	ControlBy // 公共模字段
}
