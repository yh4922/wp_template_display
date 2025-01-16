package utils

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	pathUtil "path"
	"path/filepath"
	"strings"
	"time"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"
	"wp_template_display/pkg"

	"github.com/dromara/carbon/v2"
	"github.com/google/uuid"
	"github.com/gookit/config/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MarkupZipTask struct {
	TaskId       string
	ZipPath      string
	DstPath      string
	Version      *m.SysDatasetVersion
	UploadBy     uint32
	IsError      bool
	IsDone       bool
	Status       string
	SuccessCount uint32
	FailCount    uint32
	JumpCount    uint32
	TotalCount   uint32
}

// 图片信息
type MarkupZipDataJsonImage struct {
	ID       uint64 `json:"id"`
	FileName string `json:"file_name"` // 图片名称
	Md5      string `json:"md5"`       // 图片MD5值
	Width    uint16 `json:"width"`     // 图片宽度
	Height   uint16 `json:"height"`    // 图片高度
	Valid    bool   `json:"valid"`     // 是否有效
	Rotate   uint   `json:"rotate"`    // 旋转角度
}

// 标注框信息
type MarkupZipDataJsonAnnotation struct {
	ID           uint64      `json:"id"`
	DBID         uint32      `json:"dbid"`
	ImageID      uint64      `json:"image_id"`
	IsCrowd      uint8       `json:"iscrowd"`
	Segmentation [][]float64 `json:"segmentation"`
	Area         float64     `json:"area"`
	Bbox         []float64   `json:"bbox"`
	CategoryID   uint16      `json:"category_id"`
	Order        uint8       `json:"order"`
}

// 系统图片数据
type SysImageData struct {
	m.SysImage
	ImageID  uint64                      `json:"image_id"`  // JSON里面的图片ID
	VerImage m.SysDatasetVersionImage    `json:"ver_image"` // 版本图片
	Jump     bool                        `json:"jump"`      // 是否跳过
	FileName string                      `json:"file_name"` // 图片名称
	Valid    bool                        `json:"valid"`     // 是否有效
	Rotate   uint                        `json:"rotate"`    // 旋转角度
	Markups  []m.SysDatasetVersionMarkup `json:"markups"`   // 标注信息
}

// 标签类别
type MarkupZipDataJsonCategorie struct {
	ID            uint16 `json:"id"`
	Name          string `json:"name"`
	SuperCategory string `json:"supercategory"`
}

// 系统标签类别
type SysCategorieData struct {
	m.SysDatasetVersionCategory
	CategorieID   uint16 `json:"categorie_id"`
	SuperCategory string `json:"supercategory"`
}

// 标注文件
type MarkupZipDataJson struct {
	Images      []MarkupZipDataJsonImage      `json:"images"`
	Annotations []MarkupZipDataJsonAnnotation `json:"annotations"`
	Categories  []MarkupZipDataJsonCategorie  `json:"categories"`
}

var HandleMarkupZipQueue = map[string]*MarkupZipTask{}

func init() {
	HandleMarkupZipQueue = make(map[string]*MarkupZipTask)
}

// 处理标注的压缩包数据
func HandleMarkupZipFile(zipPath string, version *m.SysDatasetVersion, uploadBy uint32, deduplicate bool) (taskId string, err error) {
	// 生成任务ID
	newId, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	taskId = newId.String()

	// ZIP处理任务
	HandleMarkupZipQueue[taskId] = &MarkupZipTask{
		TaskId:       taskId,   // 任务ID
		ZipPath:      zipPath,  // 压缩包路径
		Version:      version,  // 数据集版本
		UploadBy:     uploadBy, // 上传者ID
		Status:       "解压中",    // 状态信息
		SuccessCount: 0,        // 成功数量
		FailCount:    0,        // 失败数量
		JumpCount:    0,        // 跳过数量
		TotalCount:   0,        // 总数
	}

	go func() {
		// 0.生成临时解压目录
		runtimePath := config.String("RuntimePath", "runtime")
		now := carbon.Now()
		// 路径
		HandleMarkupZipQueue[taskId].DstPath = pathUtil.Join(runtimePath, "temp", taskId)
		// 1.在临时目录解压ZIP文件夹
		if err := unZip(HandleMarkupZipQueue[taskId].DstPath, zipPath); err != nil {
			HandleMarkupZipQueue[taskId].Status = "解压失败"
			HandleMarkupZipQueue[taskId].IsError = true
			go func() {
				time.Sleep(time.Minute * 10)
				DelHandleMarkupZipTask(taskId)
			}()
			return
		}
		HandleMarkupZipQueue[taskId].Status = "格式化中"
		// 2.判断文件格式
		dataJSON, err := checkZipFile(HandleMarkupZipQueue[taskId].DstPath)
		if err != nil {
			HandleMarkupZipQueue[taskId].Status = fmt.Sprintf("格式化失败: %s", err.Error())
			HandleMarkupZipQueue[taskId].IsError = true
			go func() {
				time.Sleep(time.Minute * 10)
				DelHandleMarkupZipTask(taskId)
			}()
			return
		}
		HandleMarkupZipQueue[taskId].TotalCount = uint32(len(dataJSON.Images)) // 总数
		// 3处理图片确认那些图片可以上传那些不可以上传
		// TODO: 后期压缩包的图片可能需要进行过滤
		var validImages []*SysImageData
		for _, image := range dataJSON.Images {
			// 图片路径
			imgPath := pathUtil.Join(HandleMarkupZipQueue[taskId].DstPath, "images", image.FileName)
			// 读取文件MD5
			md5, err := pkg.GetFileMD5(imgPath)
			if err != nil {
				HandleMarkupZipQueue[taskId].FailCount++
				continue
			}
			image.Md5 = md5
			// 查询图片是否存在 不存在则创建
			sysImage, err := m.GetImageByMd5(image.Md5)
			if err != nil {
				// 生成图片保存路径
				fileExt := pathUtil.Ext(image.FileName)
				uploadTime := now.Format("His")   // 上传时间
				uploadDate := now.Format("Y-m-d") // 上传日期
				fileName := fmt.Sprintf("%s-%s%s", md5, uploadTime, fileExt)
				filePath := pathUtil.Join(runtimePath, "uploads", uploadDate, fileName)
				filePath, err = filepath.Abs(filePath)
				if err != nil {
					HandleMarkupZipQueue[taskId].FailCount++
					continue
				}
				// 创建图片文件夹
				if err := os.MkdirAll(pathUtil.Dir(filePath), 0755); err != nil {
					HandleMarkupZipQueue[taskId].FailCount++
					continue
				}

				// 移动 imgPath 到 filePath
				if err := os.Rename(imgPath, filePath); err != nil {
					HandleMarkupZipQueue[taskId].FailCount++
					continue
				}

				// 读取文件信息
				fileInfo, err := os.Stat(filePath)
				if err != nil {
					HandleMarkupZipQueue[taskId].FailCount++
					continue
				}

				// 保存图片到数据库
				imageId, err := m.GenImageId()
				if err != nil {
					HandleMarkupZipQueue[taskId].FailCount++
					continue
				}

				sysImage = m.SysImage{
					Id:       imageId,
					FilePath: filePath,
					Md5:      md5,
					Width:    image.Width,
					Height:   image.Height,
					Ratio:    math.Round(float64(image.Width)/float64(image.Height)*100) / 100,
					Size:     uint(fileInfo.Size()),
				}
				tx := g.DB.Create(&sysImage)
				if tx.Error != nil {
					HandleMarkupZipQueue[taskId].FailCount++
					continue
				}
			}

			// 获取版本图片
			isJump := false
			verImage, err := m.GetDatasetVerImageByImageId(version.Id, sysImage.Id)
			if err != nil {
				// 不存在 创建
				verImage, err = m.AppendDatasetVersionImage(version.Id, &sysImage, image.FileName, uploadBy)
				if err != nil {
					HandleMarkupZipQueue[taskId].FailCount++
					continue
				}
			} else {
				isJump = !deduplicate
			}

			// 设置图片
			validImage := SysImageData{
				SysImage: sysImage,
				ImageID:  image.ID,
				VerImage: *verImage,
				FileName: image.FileName,
				Valid:    true,
				Rotate:   image.Rotate,
				Jump:     isJump, // 记录跳过状态
			}
			validImages = append(validImages, &validImage)
		}
		// 4.提取标签类别
		var labelCategories []SysCategorieData
		for _, categorie := range dataJSON.Categories {
			if categorie.Name != "" {
				// 查询数据库标注类别 不存在则创建
				sysSategorie, err := m.GetCategoryByName(version.Id, categorie.Name)
				if err != nil {
					sysSategorie = &m.SysDatasetVersionCategory{
						DatasetID: version.DatasetID,
						VersionID: version.Id,
						Name:      categorie.Name,
						Label:     categorie.Name,
						ControlBy: m.ControlBy{CreatedBy: uploadBy},
					}
					// 创建  报错则跳过
					tx := g.DB.Create(&sysSategorie)
					if tx.Error != nil {
						continue
					}
				}

				// 记录
				labelCategorie := SysCategorieData{
					SysDatasetVersionCategory: *sysSategorie,
					CategorieID:               categorie.ID,
					SuperCategory:             categorie.SuperCategory,
				}
				labelCategories = append(labelCategories, labelCategorie)
			}
		}
		HandleMarkupZipQueue[taskId].Status = "同步标注信息"
		// 5.循环格式化标注对象
		serial := 0

		// println("标注数量", len(dataJSON.Annotations))
		for _, annotation := range dataJSON.Annotations {
			// time.Sleep(time.Millisecond * 200)
			serial++
			// 查找对应的图片
			var validImage *SysImageData
			for _, img := range validImages {
				if img.ImageID == annotation.ImageID {
					validImage = img
					break
				}
			}
			if validImage.ImageID == 0 {
				// 未匹配到图像
				continue
			}

			// 获取标签类别
			var tagCategory SysCategorieData
			for _, categorie := range labelCategories {
				if categorie.CategorieID == annotation.CategoryID {
					tagCategory = categorie
					break
				}
			}

			// 点位数据转为JSON字符串
			pointJSON, err := json.Marshal(annotation.Segmentation[0])
			if err != nil {
				continue
			}

			// 通过判断 点位数据确认是多边形还是矩形
			markupType := "polygon"
			if checkPointInRectangle(annotation.Segmentation[0]) {
				markupType = "rect"
			}
			// 生成标注信息
			markup := m.SysDatasetVersionMarkup{
				Id:              m.GenMarkupId(),
				DatasetID:       version.DatasetID,
				VersionID:       version.Id,
				VersionImageID:  validImage.VerImage.Id,
				Serial:          uint16(serial),
				MarkupType:      markupType,
				MarkupPoints:    string(pointJSON),
				TagCategoryId:   tagCategory.Id,
				TagCategoryName: tagCategory.Name,
				ControlBy:       m.ControlBy{CreatedBy: uploadBy},
			}
			// 记录到图片上
			validImage.Markups = append(validImage.Markups, markup)
		}
		// 保存图片标注
		for _, image := range validImages {
			// 跳过
			if image.Jump {
				HandleMarkupZipQueue[taskId].JumpCount++
				HandleMarkupZipQueue[taskId].SuccessCount++
				continue
			}
			// 保存图片标注
			err := savImageMarkups(image, version, uploadBy)
			if err != nil {
				HandleMarkupZipQueue[taskId].FailCount++
			} else {
				HandleMarkupZipQueue[taskId].SuccessCount++
			}
		}

		// 标记完成
		HandleMarkupZipQueue[taskId].Status = "上传完成"
		HandleMarkupZipQueue[taskId].IsDone = true
		go func() {
			time.Sleep(time.Minute * 10)
			DelHandleMarkupZipTask(taskId)
		}()
	}()

	return taskId, nil
}

// 删除任务
func DelHandleMarkupZipTask(taskId string) {
	// 0.查看任务是否存在
	if _, ok := HandleMarkupZipQueue[taskId]; !ok {
		return
	}
	// 1.删除临时目录
	os.RemoveAll(HandleMarkupZipQueue[taskId].DstPath)
	// 2.删除压缩包
	os.Remove(HandleMarkupZipQueue[taskId].ZipPath)
	// 3.删除任务
	delete(HandleMarkupZipQueue, taskId)
}

// 复制压缩包中的文件
func copyZipFile(file *zip.File, path string) error {
	// DEBUG 测试添加1秒延时
	time.Sleep(time.Millisecond * 100)

	// 获取到 Reader
	fr, err := file.Open()
	if err != nil {
		return err
	}
	defer fr.Close()

	// 创建要写出的文件对应的 Write
	fw, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer fw.Close()

	_, err = io.Copy(fw, fr)
	if err != nil {
		return err
	}

	// fmt.Printf("成功解压 %s ，共写入了 %d 个字符的数据\n", path, n)

	return nil
}

// 保存图片标注
func savImageMarkups(image *SysImageData, version *m.SysDatasetVersion, uploadBy uint32) error {
	// println("保存图片标注", image.ImageID, len(image.Markups))
	// 创建事务
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 标注的分表
	markupDB := m.MarkupTableOfVer(version.Id)
	verImageID := image.VerImage.Id

	// 查找旧的标注 获取数量 并删除
	var oldNum int64
	if err := tx.Scopes(markupDB).Where("version_image_id = ?", verImageID).Count(&oldNum).Delete(&m.SysDatasetVersionMarkup{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果有标注则插入新的标注对象数据
	if len(image.Markups) > 0 {
		if err := tx.Scopes(markupDB).Create(&image.Markups).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 是否无标注
	notMarker := 0
	if len(image.Markups) == 0 {
		notMarker = 1
	}

	// 标签类别
	markerCategorys := ""
	for _, markup := range image.Markups {
		// 如果标签类别不存在则追加
		if markerCategorys == "" {
			markerCategorys = markup.TagCategoryName
		} else if !strings.Contains(markerCategorys, markup.TagCategoryName) {
			markerCategorys = markerCategorys + "," + markup.TagCategoryName
		}
	}

	// 更新版本图片数量
	if err := tx.Model(&m.SysDatasetVersionImage{}).Where("id = ?", verImageID).Updates(map[string]interface{}{
		"is_marker":        1,
		"marker_at":        time.Now(),
		"marker_by":        0,
		"marker_count":     uint16(len(image.Markups)),
		"not_marker":       notMarker,
		"marker_categorys": markerCategorys,
		"is_wait_confirm":  0,
		"updated_by":       uploadBy,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// gorm.Expr("marker_count + ?", len(image.Markups))
	var markupImageNum clause.Expr
	if image.VerImage.IsMarker == 0 {
		markupImageNum = gorm.Expr("markup_image_num + ?", 1)
	} else {
		markupImageNum = gorm.Expr("markup_image_num + ?", 0)
	}

	var markupNum clause.Expr
	newNum := int64(len(image.Markups))
	if oldNum > newNum {
		markupNum = gorm.Expr("markup_num - ?", oldNum-newNum)
	} else {
		markupNum = gorm.Expr("markup_num + ?", newNum-oldNum)
	}

	// 更新标注版本上的数据信息
	if err := tx.Model(&m.SysDatasetVersion{}).Where("id = ?", version.Id).Updates(map[string]interface{}{
		"markup_image_num": markupImageNum, // 标注图片数量
		"markup_num":       markupNum,      // 标注数量
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

// 解压压缩包
func unZip(dst, src string) (err error) {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return
	}
	defer zr.Close()

	// 如果解压后不是放在当前目录就按照保存目录去创建目录
	if dst != "" {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return err
		}
	}

	// 遍历 zr ，将文件写入到磁盘
	for _, file := range zr.File {
		path := filepath.Join(dst, file.Name)

		// 如果是目录，就创建目录
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return err
			}
			// 因为是目录，跳过当前循环，因为后面都是文件的处理
			continue
		}

		if err := copyZipFile(file, path); err != nil {
			return err
		}
	}
	return nil
}

// 验证压缩包格式
func checkZipFile(dstPath string) (*MarkupZipDataJson, error) {
	// data.json文件是否存在
	dataJsonPath := pathUtil.Join(dstPath, "data.json")
	if _, err := os.Stat(dataJsonPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("data.json文件不存在")
	}

	// images目录是否存在
	imagesDir := pathUtil.Join(dstPath, "images")
	// 判断图片目录是否存在
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("images目录不存在")
	}

	// 读取data.json文件
	dataJsonFile, err := os.Open(dataJsonPath)
	if err != nil {
		return nil, fmt.Errorf("data.json文件读取失败")
	}
	defer dataJsonFile.Close()

	// 读取data.json文件
	dataJsonData, err := io.ReadAll(dataJsonFile)
	if err != nil {
		return nil, fmt.Errorf("data.json文件读取失败")
	}

	// 解析data.json文件
	var dataJson MarkupZipDataJson
	if err := json.Unmarshal(dataJsonData, &dataJson); err != nil {
		return nil, fmt.Errorf("data.json文件解析失败")
	}

	return &dataJson, nil
}

// 判断点位是否是矩形
func checkPointInRectangle(points []float64) bool {
	// 100, 100, 400, 100, 400, 400, 100, 400
	if len(points) != 8 {
		return false
	}

	// x1, y1, x2, y1, x2, y2, x1, y2
	if points[0] == points[6] && points[1] == points[3] && points[2] == points[4] && points[5] == points[7] {
		return true
	}

	// 默认不是矩形
	return false
}
