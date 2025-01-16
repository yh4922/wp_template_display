package ctxDatasetVersionData

import (
	"time"
	ctx "wp_template_display/internal/controller"
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

type DatasetVersionDataListReq struct {
	Page        int     `query:"page"`                                     // 页码
	Size        int     `query:"size"`                                     // 每页数量
	DatasetId   uint32  `query:"dataset_id" validate:"required"`           // 数据集ID
	VersionId   uint32  `query:"version_id" validate:"required"`           // 数据集版本ID
	Type        int     `query:"type" validate:"required,oneof=1 2 3"`     // 类型 1未标注  2已标注  3待确认
	Name        string  `query:"name"`                                     // 名称（模糊查询）
	UpStart     int64   `query:"up_start"`                                 // 上传开始时间 shi
	UpEnd       int64   `query:"up_end"`                                   // 上传结束时间
	MarkerStart int64   `query:"marker_start"`                             // 标注开始时间
	MarkerEnd   int64   `query:"marker_end"`                               // 标注结束时间
	MinWidth    int     `query:"min_width"`                                // 最小宽度
	MaxWidth    int     `query:"max_width"`                                // 最大宽度
	MinRatio    float64 `query:"min_ratio"`                                // 最小宽高比
	MaxRatio    float64 `query:"max_ratio"`                                // 最大宽高比
	MarkerBy    int     `query:"marker_by"`                                // 标注人员
	Categorys   string  `query:"categorys"`                                // 标签类别
	Order       string  `query:"order" validate:"required,oneof=DESC ASC"` // 排序 DESC降序  ASC升序
}

func DatasetVersionDataList(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/dataset/version/data/list")

	req := new(DatasetVersionDataListReq)
	if err := c.QueryParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 默认数据
	req.Page = c.QueryInt("page", 1)
	req.Size = c.QueryInt("size", 15)

	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 验证数据集是否存在
	_, err = m.GetDatasetById(req.DatasetId)
	if err != nil {
		return ctx.CtxError(c, 5003, "数据集不存在", err.Error())
	}

	// 验证数据集版本是否存在
	_, err = m.GetDatasetVerById(req.VersionId)
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集版本不存在", err.Error())
	}

	var dataList []m.SysDatasetVersionImage
	var total int64

	// 构建查询
	query := g.DB.Model(&m.SysDatasetVersionImage{})

	// 关联数据集版本ID
	query = query.Where("version_id = ?", req.VersionId)

	// Name图片名称模糊查询
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 图片上传时间
	if req.UpStart != 0 && req.UpEnd != 0 {
		query = query.Where(
			"created_at BETWEEN ? AND ?",
			time.UnixMilli(int64(req.UpStart)),
			time.UnixMilli(int64(req.UpEnd)),
		)
	}

	// 图片宽度
	if req.MinWidth != 0 || req.MaxWidth != 0 {
		query = query.Where("image_width BETWEEN ? AND ?", req.MinWidth, req.MaxWidth)
	}

	// 图片宽高比
	if req.MinRatio != 0 || req.MaxRatio != 0 {
		query = query.Where("image_ratio BETWEEN ? AND ?", req.MinRatio, req.MaxRatio)
	}

	if req.Type == 1 {
		// 1未标注
		query = query.Where("is_marker = 0")
	} else if req.Type == 2 {
		//2.已标注
		query = query.Where("is_marker != 0 AND is_wait_confirm = 0")

		// 标注时间
		if req.MarkerStart != 0 && req.MarkerEnd != 0 {
			query = query.Where(
				"marker_at BETWEEN ? AND ?",
				time.UnixMilli(int64(req.MarkerStart)),
				time.UnixMilli(int64(req.MarkerEnd)),
			)
		}

		// 标注人员
		if req.MarkerBy != 0 {
			query = query.Where("marker_by = ?", req.MarkerBy)
		}

		// 标注标签
		if req.Categorys != "" {
			query = query.Where("marker_categorys LIKE ?", "%"+req.Categorys+"%")
		}

	} else if req.Type == 3 {
		//3.待确认
		query = query.Where("is_wait_confirm = 1")
	}

	// CreatedAt 排序
	if req.Order == "DESC" {
		query = query.Order("created_at DESC")
	} else {
		query = query.Order("created_at ASC")
	}

	// 总数
	if err := query.Count(&total).Error; err != nil {
		return ctx.CtxError(c, 5005, "获取数据集版本失败", err.Error())
	}

	// 获取数据
	if err := query.Offset((req.Page - 1) * req.Size).Limit(req.Size).Find(&dataList).Error; err != nil {
		return ctx.CtxError(c, 5006, "获取数据集版本失败", err.Error())
	}

	// 隐藏某些字段

	// 返回数据
	data := fiber.Map{
		"list": dataList,
		"page": map[string]int{
			"page":  req.Page,
			"size":  req.Size,
			"count": int(total),
		},
	}

	// TODO: 后面需要验证查询是否正确，是否需要格式化数据
	return ctx.CtxSuccess(c, data)
}
