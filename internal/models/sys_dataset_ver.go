package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wp_template_display/internal/consts"
	g "wp_template_display/internal/global"
	"wp_template_display/internal/rules"

	"github.com/google/uuid"
	"github.com/gookit/config/v2"
	"gorm.io/gorm"
)

type SysDatasetVersion struct {
	Model
	// Dataset         SysDataset `json:"dataset" gorm:"foreignKey:DatasetID;references:Id"`
	DatasetID       uint32     `json:"dataset_id" gorm:"not null;comment:数据集ID"`
	AlgType         string     `json:"alg_type" gorm:"size:10;not null;comment: 对应算法类型"`
	Version         uint32     `json:"version" gorm:"size:5;not null;comment:版本"`    // 数字版本号递增
	Inherit         uint32     `json:"inherit" gorm:"size:5;default:0;comment:继承版本"` // 0表示是全新创建版本  其他则表示继承的版本号
	InheritRelation string     `json:"inherit_relation" gorm:"comment:继承关系"`         // 按照继承顺序 [[1,创建,发布]]
	ImageNum        uint32     `json:"image_num" gorm:"default:0;comment:图片数量"`
	MarkupImageNum  uint32     `json:"markup_image_num" gorm:"default:0;comment:标注图片数量"`
	MarkupNum       uint32     `json:"markup_num" gorm:"default:0;comment:标注数量"`
	Status          uint8      `json:"status" gorm:"size:1;default:0;comment:状态"` // 0 未发布  1 已发布
	Remark          string     `json:"remark" gorm:"size:150;comment:备注"`
	PublishAt       *time.Time `json:"publish_at" gorm:"comment:发布时间"` // 发布时间
	ModelTime
	ControlBy // 公共模字段
}

type CategoryItem struct {
	Id     uint32 `json:"id"`      // 标签ID
	Name   string `json:"name"`    // 标签值
	Label  string `json:"label"`   // 标签名称
	Count  uint32 `json:"count"`   // 标注数量
	ImgNum uint32 `json:"img_num"` // 图片数量
	Pass   bool   `json:"pass"`    // 当前标签是否满足训练条件
}

type InheritItem struct {
	// Status    uint8  `json:"status"`     // 状态
	Version   uint32 `json:"version"`    // 版本
	CreatedAt int64  `json:"created_at"` // 创建时间
	PublishAt int64  `json:"publish_at"` // 发布时间
}

type SysDatasetVersionDetail struct {
	SysDatasetVersion
	Dataset             SysDataset        `json:"dataset"`           // 数据集
	TrainStatusIsOk     bool              `json:"train_status_isok"` // 是否满足训练标准
	TrainRules          []rules.RulesItem `json:"train_rules"`
	UnmarkupImageNum    uint              `json:"unmarkup_image_num"`     // 未标注图片数量
	WaitConfirmImageNum uint              `json:"wait_confirm_image_num"` // 待确认图片数量
	NotMarkerImageNum   int64             `json:"not_marker_image_num"`   // 无目标图片数量
	CategoryList        []CategoryItem    `json:"category_list"`          // 标签列表
	InheritInfo         []InheritItem     `json:"inherit_info"`           // 继承信息
}

// 创建全选数据集版本
func CreateDatasetVer(ver *SysDatasetVersion) error {
	// 验证数据集存在
	dataset := &SysDataset{}
	if err := g.DB.Model(&SysDataset{}).Where("id = ?", ver.DatasetID).First(dataset).Error; err != nil {
		return errors.New("数据集不存在")
	}

	// 查询当前数据集下是否存在未发布的版本
	var unPubVer SysDatasetVersion
	err := g.DB.Model(&SysDatasetVersion{}).Where("dataset_id = ?", ver.DatasetID).Where("status = ?", 0).First(&unPubVer).Error
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return err
	}
	if unPubVer.Id != 0 {
		return errors.New("数据集下存在未发布的版本，无法创建新版本")
	}

	// 开启事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	ver.Version = uint32(dataset.Version + 1)

	// 添加新版本
	if err := tx.Model(&SysDatasetVersion{}).Create(ver).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2.同步数据集版本数量和最新版本
	if err := tx.Model(&SysDataset{}).Where("id = ?", ver.DatasetID).Updates(map[string]interface{}{
		"version_num": gorm.Expr("version_num + 1"),
		"version":     ver.Version,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// 创建继承数据集版本
func CreateDatasetVerForInherit(ver *SysDatasetVersion) error {
	// 验证数据集存在
	dataset := &SysDataset{}
	if err := g.DB.Model(&SysDataset{}).Where("id = ?", ver.DatasetID).First(dataset).Error; err != nil {
		return errors.New("数据集不存在")
	}

	// 验证继承的数据集版本存在
	inherit := &SysDatasetVersion{}
	if err := g.DB.Model(&SysDatasetVersion{}).Where("dataset_id = ?", ver.DatasetID).Where("version = ?", ver.Inherit).First(inherit).Error; err != nil {
		return err
	}

	// 继承的数据集已经发布
	if inherit.Status != 1 {
		return errors.New("继承的数据集未发布")
	}

	// 查询当前数据集下是否存在未发布的版本
	var unPubVer SysDatasetVersion
	err := g.DB.Model(&SysDatasetVersion{}).Where("dataset_id = ?", ver.DatasetID).Where("status = ?", 0).First(&unPubVer).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if unPubVer.Id != 0 {
		return errors.New("数据集下存在未发布的版本，无法创建新版本")
	}

	println("验证完成开启事务")

	// 开启事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 补全版本信息
	ver.AlgType = inherit.AlgType
	ver.Version = uint32(dataset.Version + 1)
	ver.InheritRelation = inherit.InheritRelation // 继承关系
	if ver.InheritRelation != "" {
		ver.InheritRelation += ","
	}

	ver.InheritRelation += fmt.Sprintf(
		"[%d,%d,%d]",
		inherit.Version,
		inherit.CreatedAt.Unix(),
		func() int64 {
			if inherit.PublishAt == nil {
				return 0
			}
			return inherit.PublishAt.Unix()
		}(),
	)

	// 添加新版本
	if err := tx.Model(&SysDatasetVersion{}).Create(ver).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2.同步数据集版本数量和最新版本
	if err := tx.Model(&SysDataset{}).Where("id = ?", ver.DatasetID).Updates(map[string]interface{}{
		"version_num": gorm.Expr("version_num + 1"),
		"version":     ver.Version,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 接收数据
	var categories []SysDatasetVersionCategory
	var imageList = []SysDatasetVersionImage{}
	var markupList = []SysDatasetVersionMarkup{}

	// 3.查询 SysDatasetVersionCategory 中 VersionID 等于 inherit.Id 的数据 复制一份修改 VersionID 为 ver.Id
	if err := tx.Model(&SysDatasetVersionCategory{}).Where("version_id = ?", inherit.Id).Find(&categories).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 循环创建
	for _, category := range categories {
		category.Id = 0
		category.VersionID = ver.Id
		if err := tx.Model(&SysDatasetVersionCategory{}).Create(&category).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 查询全部图片
	if err := tx.Model(&SysDatasetVersionImage{}).Where("version_id = ?", inherit.Id).Find(&imageList).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 查询全部标注
	if err := tx.Model(&SysDatasetVersionMarkup{}).Where("version_id = ?", inherit.Id).Find(&markupList).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 循环创建copy图片
	// var copyImageList = []SysDatasetVersionImage{}
	var oldImageMap = make(map[uint32]uint32)
	for _, image := range imageList {
		newImage := SysDatasetVersionImage{
			DatasetID:   ver.DatasetID,
			VersionID:   ver.Id,
			ImageID:     image.ImageID,
			Name:        image.Name,
			IsMarker:    image.IsMarker,
			MarkerAt:    image.MarkerAt,
			MarkerCount: image.MarkerCount,
			ModelTime:   ModelTime{CreatedAt: image.ModelTime.CreatedAt, UpdatedAt: image.ModelTime.UpdatedAt},
			ControlBy:   ControlBy{CreatedBy: ver.CreatedBy, UpdatedBy: ver.CreatedBy},
		}
		// 添加
		if err := tx.Model(&SysDatasetVersionImage{}).Create(&newImage).Error; err != nil {
			tx.Rollback()
			return err
		}

		if newImage.Id == 0 {
			tx.Rollback()
			return errors.New("复制图片失败")
		}

		oldImageMap[image.Id] = newImage.Id
		// copyImageList = append(copyImageList, newImage)
	}

	// 循环复制标注
	for _, markup := range markupList {
		// 生成新的ID
		newId, err := uuid.NewV7()
		if err != nil {
			tx.Rollback()
			return err
		}
		markup.Id = newId.String()
		markup.VersionID = ver.Id
		markup.VersionImageID = oldImageMap[markup.VersionImageID]
		if err := tx.Model(&SysDatasetVersionMarkup{}).Create(&markup).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	return tx.Commit().Error
}

// 通过ID获取数据集版本
func GetDatasetVerById(id uint32) (*SysDatasetVersion, error) {
	ver := &SysDatasetVersion{}
	if err := g.DB.Model(&SysDatasetVersion{}).Where("id = ?", id).First(ver).Error; err != nil {
		return nil, err
	}
	return ver, nil
}

// 发布数据集版本
func PublishDatasetVer(ver *SysDatasetVersion) error {
	ver.Status = 1
	PublishAt := time.Now()
	ver.PublishAt = &PublishAt
	return g.DB.Model(&SysDatasetVersion{}).Where("id = ?", ver.Id).Updates(ver).Error
}

// 删除数据集版本
func DeleteDatasetVer(ver *SysDatasetVersion) error {
	// 开启事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除
	if err := tx.Model(&SysDatasetVersion{}).Where("id = ?", ver.Id).Delete(ver).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2.同步数据集版本数量和最新版本
	if err := tx.Model(&SysDataset{}).Where("id = ?", ver.DatasetID).Updates(map[string]interface{}{
		"version_num": gorm.Expr("version_num - 1"),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// TODO 同步删除关联的图片和标注

	// 提交事务
	return tx.Commit().Error
}

// 通过数据集ID获取数据集版本
func GetDatasetVerByDatasetID(page int, size int, datasetID uint32) ([]SysDatasetVersion, int64, error) {
	var verList []SysDatasetVersion
	var total int64

	// 构建查询
	query := g.DB.Model(&SysDatasetVersion{})

	// 关联数据集ID
	query = query.Where("dataset_id = ?", datasetID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&verList).Error; err != nil {
		return nil, 0, err
	}

	return verList, total, nil
}

// 导出数据集版本用于训练的数据
func ExportDatasetVerForTrain(ver *SysDatasetVersion) (string, *consts.ExportVersionData, error) {
	// 全部图片 外键关联Image
	var verImageList []SysDatasetVersionImage
	if err := g.DB.Model(&SysDatasetVersionImage{}).Preload("Image").Where("version_id = ?", ver.Id).Find(&verImageList).Error; err != nil {
		return "", nil, err
	}

	// 读取全部标签
	var verMarkupList []SysDatasetVersionMarkup
	err := g.DB.Scopes(MarkupTableOfVer(ver.Id)).Where("version_id = ?", ver.Id).Find(&verMarkupList).Error
	if err != nil {
		return "", nil, err
	}

	// 读取全部标签类别
	var verCategoryList []SysDatasetVersionCategory
	err = g.DB.Model(&SysDatasetVersionCategory{}).Where("version_id = ?", ver.Id).Find(&verCategoryList).Error
	if err != nil {
		return "", nil, err
	}

	// 创建任务路径
	runtimePath := config.String("RuntimePath", "runtime")
	taskUUID := GenMarkupId()
	taskDir, err := filepath.Abs(filepath.Join(runtimePath, "tasks", taskUUID))
	if err != nil {
		return "", nil, err
	}

	// 创建任务文件夹
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return "", nil, err
	}

	// 初始化任务数据
	taskData := consts.ExportVersionData{
		UUID:        taskUUID,
		Images:      []consts.ExportVersionImage{},
		Annotations: []consts.ExportVersionAnnotation{},
		Categories:  []consts.ExportVersionCategorie{},
	}

	// 图片文件夹
	imageDir := filepath.Join(taskDir, "images")
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		os.RemoveAll(taskDir)
		return "", nil, err
	}

	// 复制图片到图片文件夹
	for _, image := range verImageList {
		imageName := filepath.Base(image.Image.FilePath)
		imagePath := filepath.Join(imageDir, imageName)

		// 复制文件
		srcFile, err := os.Open(image.Image.FilePath)
		if err != nil {
			continue
		}
		defer srcFile.Close()

		// 创建目标文件
		dstFile, err := os.Create(imagePath)
		if err != nil {
			continue
		}
		defer dstFile.Close()

		// 复制文件内容
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			continue
		}

		// 添加图片信息
		taskData.Images = append(taskData.Images, consts.ExportVersionImage{
			ID:       uint64(image.Id),
			FileName: imageName,
			Width:    uint16(image.Image.Width),
			Height:   uint16(image.Image.Height),
			Valid:    true,
			Rotate:   0,
		})
	}

	// 循环标签 并创建一个 map 对应表 方便后续使用
	var categoryMap = make(map[string]int)
	for index, category := range verCategoryList {
		categoryMap[category.Name] = index + 1
		taskData.Categories = append(taskData.Categories, consts.ExportVersionCategorie{
			ID:            uint(index + 1),
			Name:          category.Name,
			SuperCategory: "",
		})
	}

	// 循环标注
	for index, markup := range verMarkupList {
		// 解析标注点位
		var points []float64
		if err := json.Unmarshal([]byte(markup.MarkupPoints), &points); err != nil {
			continue
		}

		// 外接矩形 和 面积
		rect, area := consts.GetPolygonBoundingRectangle(points)

		// 添加标注
		taskData.Annotations = append(taskData.Annotations, consts.ExportVersionAnnotation{
			ID:           uint(index + 1),
			ImageID:      markup.VersionImageID,
			IsCrowd:      0,
			CategoryID:   uint16(categoryMap[markup.TagCategoryName]),
			Segmentation: [][]float64{points},
			Area:         area,
			Bbox:         rect,
			Order:        1,
		})
	}

	// 转为JSON 写入到
	jsonFile := filepath.Join(taskDir, "data.json")
	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return "", nil, err
	}

	// 格式化数据
	return taskDir, &taskData, nil
}

// 获取数据集版本详情
func GetDatasetVerDetail(versionID uint32) (*SysDatasetVersionDetail, error) {
	// 获取数据集版本
	ver, err := GetDatasetVerById(versionID)
	if err != nil {
		return nil, err
	}

	dataset, err := GetDatasetById(ver.DatasetID)
	if err != nil {
		return nil, err
	}

	// 版本详情
	detail := SysDatasetVersionDetail{
		SysDatasetVersion: *ver,
		Dataset:           *dataset,
		UnmarkupImageNum:  uint(ver.ImageNum) - uint(ver.MarkupImageNum),
	}

	// 获取图片
	var imageList []SysDatasetVersionImage
	err = g.DB.Model(&SysDatasetVersionImage{}).Where("version_id = ?", ver.Id).Find(&imageList).Error
	if err != nil {
		return nil, err
	}

	// 统计图片数量
	detail.ImageNum = uint32(len(imageList))

	// 查询全部标注到内存
	var markupList []SysDatasetVersionMarkup
	err = g.DB.Scopes(MarkupTableOfVer(versionID)).Where("version_id = ?", ver.Id).Find(&markupList).Error
	if err != nil {
		return nil, err
	}

	// 标注数量
	markupNum := len(markupList)
	detail.MarkupNum = uint32(markupNum)

	// 标注图片数量
	markupImageNum := 0
	for _, image := range imageList {
		if image.IsMarker == 1 {
			markupImageNum++
		}
	}
	detail.MarkupImageNum = uint32(markupImageNum)

	// 无目标图片数量
	notMarkerImageNum := 0
	for _, image := range imageList {
		if image.IsWaitConfirm == 1 {
			notMarkerImageNum++
		}
	}
	detail.NotMarkerImageNum = int64(notMarkerImageNum)

	// 是否满足训练条件
	detail.TrainStatusIsOk = true // 默认满足
	rule1 := rules.RulesItem{Rule: "至少有1个检测标签下的目标数≥10个", Pass: false}
	rule2 := rules.RulesItem{Rule: "目标数≥10个的所有检测标签下的图片总数≥10张", Pass: true}

	// 标签列表
	var categoryList []SysDatasetVersionCategory
	err = g.DB.Model(&SysDatasetVersionCategory{}).Where("version_id = ?", ver.Id).Find(&categoryList).Error
	if err != nil {
		return nil, err
	}
	// 标签列表
	detail.CategoryList = []CategoryItem{}
	for _, category := range categoryList {
		// 标注数量
		count := 0
		for _, markup := range markupList {
			if markup.TagCategoryName == category.Name {
				count++
			}
		}

		// 标注图片数量
		imageNum := 0
		for _, image := range imageList {
			if strings.Contains(image.MarkerCategorys, category.Name) {
				imageNum++
			}
		}

		// 标签详情
		detail.CategoryList = append(detail.CategoryList, CategoryItem{
			Id:     category.Id,
			Name:   category.Name,
			Label:  category.Label,
			Count:  uint32(count),
			ImgNum: uint32(imageNum),
			Pass:   count >= 10,
		})

		// 至少有1个检测标签下的目标数≥10个
		if count >= 10 {
			rule1.Pass = true
		}
	}

	// 目标数≥10个的所有检测标签下的图片总数≥10张
	number := 0
	for _, image := range imageList {
		if image.MarkerCount >= 10 {
			number++
		}
		if number >= 10 {
			rule2.Pass = true
			break
		}
	}

	// 训练标准
	detail.TrainRules = []rules.RulesItem{rule1, rule2}

	// 继承信息
	detail.InheritInfo = []InheritItem{}
	inheritInfoText := fmt.Sprintf("[%s]", ver.InheritRelation) // [[1,2,3],[1,2,3]]
	var inheritInfo [][]int64
	if err := json.Unmarshal([]byte(inheritInfoText), &inheritInfo); err != nil {
		return nil, err
	}

	// 循环继承信息
	for _, info := range inheritInfo {
		detail.InheritInfo = append(detail.InheritInfo, InheritItem{
			Version:   uint32(info[0]),
			CreatedAt: info[2],
			PublishAt: info[3],
		})
	}

	return &detail, nil
}
