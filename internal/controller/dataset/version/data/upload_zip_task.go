package ctxDatasetVersionData

import (
	ctx "wp_template_display/internal/controller"
	"wp_template_display/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func DatasetVersionDataUploadZipTask(c *fiber.Ctx) error {
	taskId := c.Params("task_id", "")
	if taskId == "" {
		return ctx.CtxError(c, 5001, "任务ID不能为空", "")
	}

	task, ok := utils.HandleMarkupZipQueue[taskId]
	if !ok {
		return ctx.CtxError(c, 5002, "任务不存在", "")
	}

	if task.IsError {
		return ctx.CtxError(c, 5003, "上传失败", task.Status)
	}

	// 返回任务进程
	return ctx.CtxSuccess(c, fiber.Map{
		"status":  task.Status,
		"success": task.SuccessCount,
		"fail":    task.FailCount,
		"total":   task.TotalCount,
		"jump":    task.JumpCount,
		"done":    task.IsDone,
	})
}
