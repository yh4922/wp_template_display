package models

import (
	"errors"
	g "wp_template_display/internal/global"
)

type SysDataset struct {
	Model
	Name       string `json:"name" gorm:"size:20;not null;comment: 数据集名称"`
	DataType   string `json:"data_type" gorm:"size:10;not null;comment: 数据集类型"` // train 训练集   test 测试集
	Remark     string `json:"remark" gorm:"size:150;comment: 备注信息"`
	VersionNum uint32 `json:"version_num" gorm:"comment: 版本数量"`
	AlgType    string `json:"alg_type" gorm:"size:10;comment: 对应算法类型"` // 存放当前最新版本的算法类型和版本号
	Version    uint32 `json:"version" gorm:"comment: 最新版本号"`
	ModelTime
	ControlBy // 公共模字段
}

// 创建数据集
func CreateDataset(dataset *SysDataset) error {
	// 开启事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查是否存在同名数据集
	var count int64
	if err := tx.Model(&SysDataset{}).Where("name = ?", dataset.Name).Count(&count).Error; err != nil {
		tx.Rollback()
		return err
	}
	if count > 0 {
		tx.Rollback()
		return errors.New("数据集名称已存在")
	}

	// 2. 创建数据集
	if err := tx.Create(dataset).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 3. 创建数据集初始版本
	datasetVersion := &SysDatasetVersion{
		DatasetID: dataset.Id,
		Version:   dataset.VersionNum,
		AlgType:   dataset.AlgType,
		Status:    0, // 初始状态
		ControlBy: ControlBy{
			CreatedBy: dataset.CreatedBy,
		},
	}

	// 执行创建
	if err := tx.Create(datasetVersion).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// 获取数据集通过ID
func GetDatasetById(id uint32) (*SysDataset, error) {
	var dataset SysDataset
	if err := g.DB.Where("id = ?", id).First(&dataset).Error; err != nil {
		return nil, err
	}
	return &dataset, nil
}

// 删除数据集
// @description 同步删除关联的版本、图片、标注、标签类别
func DeleteDataset(id uint32) error {
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除标注
	if err := tx.Where("dataset_id = ?", id).Delete(&SysDatasetVersionMarkup{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 删除标签类别
	if err := tx.Where("dataset_id = ?", id).Delete(&SysDatasetVersionCategory{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 3. 删除图片
	if err := tx.Where("dataset_id = ?", id).Delete(&SysDatasetVersionImage{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4. 删除版本
	if err := tx.Where("dataset_id = ?", id).Delete(&SysDatasetVersion{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 5. 删除数据集本体
	if err := tx.Delete(&SysDataset{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// 获取数据集列表
func GetDatasetList(page int, size int, dataType string, name string, algType string) ([]SysDataset, int64, error) {
	var datasets []SysDataset
	var total int64

	// 构建查询
	query := g.DB.Model(&SysDataset{}).Where("deleted_at IS NULL")

	// 添加查询条件
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if dataType != "" {
		query = query.Where("data_type = ?", dataType)
	}
	if algType != "" {
		query = query.Where("alg_type = ?", algType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&datasets).Error; err != nil {
		return nil, 0, err
	}

	return datasets, total, nil
}
