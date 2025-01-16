package models

import (
	g "wp_template_display/internal/global"

	"github.com/sony/sonyflake"
)

var flake *sonyflake.Sonyflake

func init() {
	// 创建一个新的 sonyflake 实例
	flake = sonyflake.NewSonyflake(sonyflake.Settings{})
}

type SysImage struct {
	Id       uint64  `json:"id" gorm:"primaryKey;comment:主键编码"`
	FilePath string  `json:"file_path" gorm:"size:100;not null;comment:图片本地路径"`
	Md5      string  `json:"md5" gorm:"index;size:32;not null;unique;comment:图片MD5值"`
	Width    uint16  `json:"width" gorm:"not null;comment:图片宽度"`
	Height   uint16  `json:"height" gorm:"not null;comment:图片高度"`
	Ratio    float64 `json:"ratio" gorm:"not null;comment:图片宽高比"`
	Size     uint    `json:"size" gorm:"not null;comment:图片大小字节"`
}

// 生成图片ID
func GenImageId() (uint64, error) {
	return flake.NextID()
}

// 根据ID获取图片
func GetImageById(id uint64) (SysImage, error) {
	var image SysImage
	err := g.DB.Where("id = ?", id).First(&image).Error
	return image, err
}

// 根据md5获取图片
func GetImageByMd5(md5 string) (SysImage, error) {
	var image SysImage
	err := g.DB.Where("md5 = ?", md5).First(&image).Error
	return image, err
}
