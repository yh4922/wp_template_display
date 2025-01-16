package consts

import "math"

type ExportVersionImage struct {
	ID       uint64 `json:"id"`        // 图片ID
	FileName string `json:"file_name"` // 图片文件名
	Width    uint16 `json:"width"`     // 图片宽度
	Height   uint16 `json:"height"`    // 图片高度
	Valid    bool   `json:"valid"`     // 图片是否有效
	Rotate   uint16 `json:"rotate"`    // 图片旋转角度
}

type ExportVersionAnnotation struct {
	ID           uint        `json:"id"`           // 标注ID  直接使用数组下标
	ImageID      uint32      `json:"image_id"`     // 对应的图片ID
	IsCrowd      uint        `json:"iscrowd"`      // 默认 0
	Segmentation [][]float64 `json:"segmentation"` // 多边形点位
	Area         float64     `json:"area"`         // 外接矩形面积
	Bbox         []float64   `json:"bbox"`         // 外接矩形点位
	CategoryID   uint16      `json:"category_id"`  // 标注类别ID
	Order        uint8       `json:"order"`        // 默认 1
}

type ExportVersionCategorie struct {
	ID            uint   `json:"id"`            // 标签类别ID
	Name          string `json:"name"`          // 标签类别名称
	SuperCategory string `json:"supercategory"` // 默认为 ""
}

type ExportVersionData struct {
	UUID        string                    `json:"-"`           // 任务UUID
	Images      []ExportVersionImage      `json:"images"`      // 图片列表
	Annotations []ExportVersionAnnotation `json:"annotations"` // 标注列表
	Categories  []ExportVersionCategorie  `json:"categories"`  // 标注类别列表
}

// 多边形转外接矩形
func GetPolygonBoundingRectangle(polygon []float64) ([]float64, float64) {
	// 初始化最小值和最大值
	minX, maxX := polygon[0], polygon[0]
	minY, maxY := polygon[1], polygon[1]

	// 面积
	area := float64(0)
	// 点位长度
	pointLen := len(polygon)

	// 遍历所有点，points数组格式为 [x1,y1,x2,y2,...]
	for i := 0; i < len(polygon); i += 2 {
		x := polygon[i]
		y := polygon[i+1]
		x2 := polygon[(i+2)%pointLen]
		y2 := polygon[(i+3)%pointLen]

		area += x*y2 - x2*y

		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	// 外接矩形
	rect := []float64{
		minX,
		minY,
		maxX - minX,
		maxY - minY,
	}

	// 面积
	area = math.Abs(area / 2)
	return rect, area
}
