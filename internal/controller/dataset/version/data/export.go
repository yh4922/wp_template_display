package ctxDatasetVersionData

import (
	"fmt"
	"os"
	"path/filepath"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"
	"wp_template_display/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionDataExport(c *fiber.Ctx) error {
	// 获取ID
	ID, err := c.ParamsInt("id", 0)
	if err != nil {
		return ctx.CtxError(c, 5001, "参数异常", err.Error())
	}

	if ID == 0 {
		return ctx.CtxError(c, 5002, "参数异常", "ID不能为空")
	}

	// 获取版本
	version, err := m.GetDatasetVerById(uint32(ID))
	if err != nil {
		return ctx.CtxError(c, 5003, "版本不存在", err.Error())
	}

	// 导出数据
	taskPath, taskData, err := m.ExportDatasetVerForTrain(version)
	if err != nil {
		return ctx.CtxError(c, 5004, "导出数据失败", err.Error())
	}
	// 删除任务文件夹
	defer os.RemoveAll(taskPath)

	// 导出压缩包
	zipPath := filepath.Join(taskPath, "data.zip")
	utils.CompressZip(taskPath, zipPath)

	// 设置下载名称
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", taskData.UUID))
	return c.Status(fiber.StatusOK).SendFile(zipPath)
}
