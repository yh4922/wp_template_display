package models

import (
	"fmt"

	"time"
	g "wp_template_display/internal/global"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SysDatasetVersionImage struct {
	Model
	Dataset         SysDataset        `json:"-" gorm:"foreignKey:DatasetID;references:Id"`
	DatasetID       uint32            `json:"dataset_id" gorm:"not null;comment:数据集ID"`
	Version         SysDatasetVersion `json:"-" gorm:"foreignKey:VersionID;references:Id"`
	VersionID       uint32            `json:"version_id" gorm:"not null;comment:数据集版本ID"`
	Image           SysImage          `json:"-" gorm:"foreignKey:ImageID;references:Id"`
	ImageID         uint64            `json:"image_id" gorm:"not null;comment:图片ID"`
	ImageLongID     string            `json:"image_long_id" gorm:"size:18;comment:图片长ID"`
	ImageWidth      uint16            `json:"image_width" gorm:"not null;comment:图片宽度"`
	ImageHeight     uint16            `json:"image_height" gorm:"not null;comment:图片高度"`
	ImageRatio      float64           `json:"image_ratio" gorm:"not null;comment:图片宽高比"`
	Name            string            `json:"name" gorm:"size:100;not null;comment:图片名称"`
	IsMarker        uint8             `json:"is_marker" gorm:"default:0;comment:是否已标记"`
	MarkerAt        *time.Time        `json:"marker_at" gorm:"comment:标记时间"`
	MarkerBy        uint32            `json:"marker_by" gorm:"index;comment:标注人员"` // 空表示未标注 或者机器标注
	MarkerCount     uint16            `json:"marker_count" gorm:"default:0;comment:标记数量"`
	NotMarker       uint8             `json:"not_marker" gorm:"default:0;comment:是否无目标"`      // 1无目标  0有目标或者未标注
	MarkerCategorys string            `json:"marker_categorys" gorm:"type:text;comment:标记类别"` // 字符串存储 "pig,none,egg"
	IsWaitConfirm   uint8             `json:"is_wait_confirm" gorm:"default:0;comment:是否等待确认"`
	UploadBy        uint32            `json:"upload_by" gorm:"default:0;comment:创建者"`
	ModelTime
	ControlBy // 公共模字段
}

// 获取数据集版本图片
func GetDatasetVerImageById(ID uint32) (*SysDatasetVersionImage, error) {
	var verImage SysDatasetVersionImage
	err := g.DB.Where("id = ?", ID).First(&verImage).Error
	return &verImage, err
}

// 获取数据集版本图片
func GetPreDatasetVerImageById(ID uint32) (*SysDatasetVersionImage, error) {
	var verImage SysDatasetVersionImage
	err := g.DB.Preload("Version").Where("id = ?", ID).First(&verImage).Error
	return &verImage, err
}

// 获取数据集版本图片列表
func GetDatasetVerImageListByIds(ids []uint32) ([]*SysDatasetVersionImage, error) {
	var verImages []*SysDatasetVersionImage
	err := g.DB.Where("id IN (?)", ids).Find(&verImages).Error
	return verImages, err
}

func GetDatasetVerImageByImageId(versionID uint32, imageId uint64) (*SysDatasetVersionImage, error) {
	var verImage SysDatasetVersionImage
	err := g.DB.Where("version_id = ? AND image_id = ?", versionID, imageId).First(&verImage).Error
	return &verImage, err
}

// 添加数据集版本图片
func AppendDatasetVersionImage(versionID uint32, image *SysImage, name string, uploadBy uint32) (*SysDatasetVersionImage, error) {
	// 查找数据集版本
	version, err := GetDatasetVerById(versionID)
	if err != nil {
		return nil, err
	}

	// 创建数据集版本图片
	verImage := &SysDatasetVersionImage{
		DatasetID:   version.DatasetID,
		VersionID:   versionID,
		ImageID:     image.Id,
		ImageLongID: fmt.Sprint(image.Id),
		ImageWidth:  image.Width,
		ImageHeight: image.Height,
		ImageRatio:  image.Ratio,
		Name:        name,
		UploadBy:    uploadBy,
	}

	// 创建事务
	tx := g.DB.Begin()

	// 添加锁，确保版本更新的顺序性
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&SysDatasetVersion{}, versionID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建数据集版本图片
	if err := tx.Create(verImage).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 2.同步数据集版本图片数量 - 使用一致的更新顺序
	if err := tx.Model(&SysDatasetVersion{}).
		Where("id = ?", versionID).
		Updates(map[string]interface{}{
			"image_num":  gorm.Expr("image_num + 1"),
			"updated_at": time.Now(), // 添加更新时间
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return verImage, nil
}

// 删除数据集版本图片
func DeleteDatasetVerImage(verImage *SysDatasetVersionImage, uploadBy uint32) error {
	// 创建事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除数据集版本图片
	if err := tx.Delete(&SysDatasetVersionImage{}, verImage.Id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除数据集版本标注
	markupDB := MarkupTableOfVer(verImage.VersionID)
	var markupCount int64
	if err := tx.Scopes(markupDB).Where("version_image_id = ?", verImage.Id).Count(&markupCount).Delete(&SysDatasetVersionMarkup{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var markupImageNum clause.Expr
	if verImage.IsMarker == 0 {
		markupImageNum = gorm.Expr("markup_image_num + ?", 0)
	} else {
		markupImageNum = gorm.Expr("markup_image_num - ?", 1)
	}

	// 更新标注版本上的数据信息
	if err := tx.Model(&SysDatasetVersion{}).Where("id = ?", verImage.VersionID).Updates(map[string]interface{}{
		"markup_image_num": markupImageNum,                           // 标注图片数量
		"markup_num":       gorm.Expr("markup_num - ?", markupCount), // 标注数量
		"image_num":        gorm.Expr("image_num - ?", 1),            // 图片数量
		"updated_by":       uploadBy,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// 批量删除数据集版本图片
func BatchDeleteDatasetVerImage(version *SysDatasetVersion, verImageList []*SysDatasetVersionImage, uploadBy uint32) error {
	// 创建事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	markupDB := MarkupTableOfVer(version.Id)

	for _, verImage := range verImageList {
		// 删除数据集版本图片
		if err := tx.Delete(&SysDatasetVersionImage{}, verImage.Id).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 删除数据集版本标注
		var markupCount int64
		if err := tx.Scopes(markupDB).Where("version_image_id = ?", verImage.Id).Count(&markupCount).Delete(&SysDatasetVersionMarkup{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		var markupImageNum clause.Expr
		if verImage.IsMarker == 0 {
			markupImageNum = gorm.Expr("markup_image_num + ?", 0)
		} else {
			markupImageNum = gorm.Expr("markup_image_num - ?", 1)
		}

		// 更新标注版本上的数据信息
		if err := tx.Model(&SysDatasetVersion{}).Where("id = ?", verImage.VersionID).Updates(map[string]interface{}{
			"markup_image_num": markupImageNum,                           // 标注图片数量
			"markup_num":       gorm.Expr("markup_num - ?", markupCount), // 标注数量
			"image_num":        gorm.Expr("image_num - ?", 1),            // 图片数量
			"updated_by":       uploadBy,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// 批量设置图片为无目标
func BatchSetVersionImageToNone(version *SysDatasetVersion, verImageList []*SysDatasetVersionImage, uploadBy uint32) error {
	// 创建事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	markupDB := MarkupTableOfVer(version.Id)
	for _, verImage := range verImageList {
		// 修改图标为无目标
		if err := tx.Model(verImage).Where("id = ?", verImage.Id).Updates(map[string]interface{}{
			"is_marker":        1,
			"marker_at":        time.Now(),
			"marker_by":        uploadBy,
			"marker_count":     0,
			"not_marker":       1,
			"marker_categorys": "",
			"is_wait_confirm":  0,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 删除数据集版本标注
		var markupCount int64
		if err := tx.Scopes(markupDB).Where("version_image_id = ?", verImage.Id).Count(&markupCount).Delete(&SysDatasetVersionMarkup{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 标注图片数量更新
		var markupImageNum clause.Expr
		if verImage.IsMarker == 0 {
			markupImageNum = gorm.Expr("markup_image_num + ?", 1)
		} else {
			markupImageNum = gorm.Expr("markup_image_num + ?", 0)
		}

		// 更新数据集版本信息
		if err := tx.Model(&SysDatasetVersion{}).Where("id = ?", verImage.VersionID).Updates(map[string]interface{}{
			"markup_image_num": markupImageNum,                           // 标注图片数量
			"markup_num":       gorm.Expr("markup_num - ?", markupCount), // 标注数量
			"updated_by":       uploadBy,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// 批量设置图片为未标注
func BatchSetVersionImageToEmpty(version *SysDatasetVersion, verImageList []*SysDatasetVersionImage, uploadBy uint32) error {
	// 创建事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	markupDB := MarkupTableOfVer(version.Id)
	for _, verImage := range verImageList {
		// 修改图标为无目标
		if err := tx.Model(verImage).Where("id = ?", verImage.Id).Updates(map[string]interface{}{
			"is_marker":        0,
			"marker_by":        0,
			"marker_count":     0,
			"not_marker":       0,
			"marker_categorys": "",
			"is_wait_confirm":  0,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 删除数据集版本标注
		var markupCount int64
		if err := tx.Scopes(markupDB).Where("version_image_id = ?", verImage.Id).Count(&markupCount).Delete(&SysDatasetVersionMarkup{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 标注图片数量更新
		var markupImageNum clause.Expr
		if verImage.IsMarker == 1 {
			markupImageNum = gorm.Expr("markup_image_num - ?", 1)
		} else {
			markupImageNum = gorm.Expr("markup_image_num + ?", 0)
		}

		// 更新数据集版本信息
		if err := tx.Model(&SysDatasetVersion{}).Where("id = ?", verImage.VersionID).Updates(map[string]interface{}{
			"markup_image_num": markupImageNum,                           // 标注图片数量
			"markup_num":       gorm.Expr("markup_num - ?", markupCount), // 标注数量
			"updated_by":       uploadBy,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// 保存图片的标注信息
func SaveVersionImageMarkup(image *SysDatasetVersionImage, list SysDatasetVersionMarkupJson) error {
	// 查找并删除旧的标注信息
	// 添加新的标注信息
	// 更新图片上标注数量
	// 更新版本上标注数量

	return nil
}
