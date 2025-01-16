package models

import (
	"fmt"
	"strings"
	g "wp_template_display/internal/global"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 颜色列表
var CategoryColorList = []string{
	"#4883EF",
	"#FF7BA1",
	"#FFD15A",
	"#00C2B7",
	"#53A5FF",
	"#9A98FF",
	"#FFC866",
	"#69BB2F",
	"#53C6FF",
	"#CE7AE2",
	"#FF9888",
	"#1BC572",
	"#65C7D9",
	"#A0A4BC",
	"#FFFFFF",
}

// 标签类别
type SysDatasetVersionCategory struct {
	Model
	Dataset     SysDataset        `json:"-" gorm:"foreignKey:DatasetID;references:Id"`
	DatasetID   uint32            `json:"dataset_id" gorm:"not null;comment:数据集ID"`
	Version     SysDatasetVersion `json:"-" gorm:"foreignKey:VersionID;references:Id"`
	VersionID   uint32            `json:"version_id" gorm:"not null;comment:数据集版本ID"`
	Name        string            `json:"name" gorm:"size:20;not null;comment:类别值"`
	Label       string            `json:"label" gorm:"size:20;not null;comment:类别名称"`
	Color       string            `json:"color" gorm:"size:20;comment:类别名颜色"`
	ShortcutKey string            `json:"shortcut_key" gorm:"size:20;comment:快捷键"`
	Remark      string            `json:"remark" gorm:"size:150;comment:备注信息"`
	ModelTime
	ControlBy // 公共模字段
}

// 根据版本ID和名称获取标签类别
func GetCategoryByName(versionID uint32, name string) (*SysDatasetVersionCategory, error) {
	var sysCategorie SysDatasetVersionCategory
	err := g.DB.Where("version_id = ? AND name = ?", versionID, name).First(&sysCategorie).Error
	return &sysCategorie, err
}

// 根据ID获取标签类别
func GetCategoryById(id uint32) (*SysDatasetVersionCategory, error) {
	var sysCategorie SysDatasetVersionCategory
	err := g.DB.Where("id = ?", id).First(&sysCategorie).Error
	return &sysCategorie, err
}

// 通过事务锁读取类别
func GetCategoryByTx(id uint32) (*SysDatasetVersionCategory, error) {
	// 开启事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var categorie SysDatasetVersionCategory
	// 直接使用 FOR UPDATE 锁定记录
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&categorie).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &categorie, nil
}

// 删除标注类别
func DeleteCategory(category *SysDatasetVersionCategory) (uint32, uint32, error) {
	// 开启事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			// println("defer 触发回滚")
			tx.Rollback()
		}
	}()

	// 使用 Clauses 方式加锁，确保锁定强度
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", category.VersionID).First(&SysDatasetVersion{}).Error; err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	// TODO 删除流程
	// 1.删除标注类别
	if err := tx.Delete(&SysDatasetVersionCategory{}, category.Id).Error; err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	// 2.删除此类别标注对象
	var markups []SysDatasetVersionMarkup
	if err := tx.Scopes(MarkupTableOfVer(category.VersionID)).Where("tag_category_id = ?", category.Id).Find(&markups).Delete(&markups).Error; err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	// 存在影响的标注 则需要同步版本和图片的数量信息
	clearMarkerImageNum := 0
	if len(markups) > 0 {
		// 图片ID
		var imageIds []uint32
		imageMarkerCategorys := make(map[uint32]uint16)
		for _, markup := range markups {
			// 检查imageIds中是否已包含当前markup的VersionImageID
			found := false
			for _, id := range imageIds {
				if id == markup.VersionImageID {
					found = true
					break
				}
			}
			// 如果不存在则添加
			if !found {
				imageIds = append(imageIds, markup.VersionImageID)
			}

			// 添加数量
			imageMarkerCategorys[markup.VersionImageID] += 1
		}
		var images []SysDatasetVersionImage
		if err := tx.Model(&SysDatasetVersionImage{}).Where("id IN (?)", imageIds).Find(&images).Error; err != nil {
			tx.Rollback()
			return 0, 0, err
		}

		// 3.同步图片的标注数量
		// is_marker
		// marker_count
		// marker_categorys

		for _, image := range images {
			oldMarkerCount := image.MarkerCount
			delCount := imageMarkerCategorys[image.Id]
			// 相等 说明所有标注都被删除了
			if oldMarkerCount == delCount {
				clearMarkerImageNum += 1
				// 2.同步数据集版本数量和最新版本
				if err := tx.Model(&SysDatasetVersionImage{}).Where("id = ?", image.Id).Updates(map[string]interface{}{
					"is_marker":        0,
					"marker_at":        nil,
					"marker_by":        0,
					"marker_count":     0,
					"marker_categorys": "",
				}).Error; err != nil {
					tx.Rollback()
					return 0, 0, err
				}
			} else {
				categorys := strings.Split(image.MarkerCategorys, ",")
				//  删除 category.Name
				for i, cat := range categorys {
					if cat == category.Name {
						categorys = append(categorys[:i], categorys[i+1:]...)
						break
					}
				}
				categorysStr := strings.Join(categorys, ",")

				markerCount := gorm.Expr(fmt.Sprintf("marker_count - %d", delCount))
				// 2.同步数据集版本数量和最新版本
				if err := tx.Model(&SysDatasetVersionImage{}).Where("id = ?", image.Id).Updates(map[string]interface{}{
					"marker_count":     markerCount,
					"marker_categorys": categorysStr,
				}).Error; err != nil {
					tx.Rollback()
					return 0, 0, err
				}
			}
		}

		// 4.同步版本的标注数量
		if err := tx.Model(&SysDatasetVersion{}).Where("id = ?", category.VersionID).Updates(map[string]interface{}{
			"markup_num":       gorm.Expr(fmt.Sprintf("markup_num - %d", len(markups))),
			"markup_image_num": gorm.Expr(fmt.Sprintf("markup_image_num - %d", clearMarkerImageNum)),
		}).Error; err != nil {
			tx.Rollback()
			return 0, 0, err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, 0, err
	}

	// 受影响的标注  和 图片数量
	return uint32(len(markups)), uint32(clearMarkerImageNum), nil
}
