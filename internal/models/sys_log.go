package models

import "time"

type SysLog struct {
	Model
	IP          string    `json:"ip" gorm:"size:20;comment:操作者IP"`
	Location    string    `json:"location" gorm:"size:255;comment:位置信息"`
	Username    string    `json:"username" gorm:"size:64;comment:操作人账号"`
	Content     string    `json:"content" gorm:"size:255;comment:操作描述"`
	Action      string    `json:"action" gorm:"size:64;comment:操作对应的action"`
	Header      string    `json:"header" gorm:"type:text;comment:请求头"`
	Params      string    `json:"params" gorm:"type:text;comment:请求参数"`
	Result      string    `json:"result" gorm:"type:text;comment:请求结果"`
	Status      string    `json:"status" gorm:"size:10;comment:请求状态"`
	Method      string    `json:"method" gorm:"size:10;comment:请求方式"`
	TriggerTime time.Time `json:"trigger_time" gorm:"comment:触发时间"`
	CompletTime time.Time `json:"complet_time" gorm:"comment:完成时间"`
	Duration    uint32    `json:"duration" gorm:"comment:请求耗时"`
}
