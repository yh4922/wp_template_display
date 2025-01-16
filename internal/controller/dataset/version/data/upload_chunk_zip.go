package ctxDatasetVersionData

import (
	"fmt"
	"os"
	pathUtil "path"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"
	"wp_template_display/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gookit/config/v2"
)

// ChunkUploadTasks :=

type DatasetVersionDataUploadChunkZipReq struct {
	DatasetID   uint32 `form:"dataset_id" validate:"required"`  // 数据集ID
	VersionId   uint32 `form:"version_id" validate:"required"`  // 数据版本ID
	Deduplicate bool   `form:"deduplicate" validate:"required"` // 是否去重 true 删除重复图片数据导入新的  false 保留现有不导入标注
	ChunkTaskId string `form:"chunk_task_id"`                   // 分片任务ID
	ChunkIndex  uint   `form:"chunk_index" validate:"required"` // 分片下标
	ChunkCount  uint   `form:"chunk_count" validate:"required"` // 分片总数
	// ChunkData  uint   `form:"chunk_data" validate:"required"` // 分片总数
}

func DatasetVersionDataUploadChunkZip(c *fiber.Ctx) error {
	// 上传文件
	ChunkFile, err := c.FormFile("chunk")
	if err != nil {
		return ctx.CtxError(c, 5001, "文件异常", err.Error())
	}

	// 解析请求参数
	req := DatasetVersionDataUploadChunkZipReq{}
	if err := c.BodyParser(&req); err != nil {
		return ctx.CtxError(c, 5001, "请求参数错误", err.Error())
	}

	// 验证字段
	err = ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// 为0 生成任务ID
	if req.ChunkIndex == 0 {
		req.ChunkTaskId = m.GenMarkupId()
	}

	// 任务ID不能为空
	if req.ChunkTaskId == "" {
		return ctx.CtxError(c, 5001, "任务ID不能为空", "")
	}

	// 验证数据集是否存在
	_, err = m.GetDatasetById(req.DatasetID)
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集不存在", err.Error())
	}

	// 验证数据集版本是否存在
	version, err := m.GetDatasetVerById(req.VersionId)
	if err != nil {
		return ctx.CtxError(c, 5004, "数据集版本不存在", err.Error())
	}

	// 数据集已经发布 则不允许修改
	if version.Status == 1 {
		return ctx.CtxError(c, 5001, "数据集版本已发布", "数据集版本已发布")
	}

	// 创建临时目录
	runtimePath := config.String("RuntimePath", "runtime")
	chunkDir := pathUtil.Join(runtimePath, "temp", req.ChunkTaskId)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return ctx.CtxError(c, 5001, "创建临时目录失败", err.Error())
	}

	// 保存文件
	chunkPath := pathUtil.Join(chunkDir, fmt.Sprintf("%d", req.ChunkIndex))
	if err := c.SaveFile(ChunkFile, chunkPath); err != nil {
		return ctx.CtxError(c, 5001, "保存文件失败", err.Error())
	}

	if req.ChunkIndex == req.ChunkCount-1 {
		// 读取全部分片文件 合并
		zipData := []byte{}
		for i := 0; i < int(req.ChunkCount); i++ {
			chunkPath := pathUtil.Join(chunkDir, fmt.Sprintf("%d", i))
			chunkData, err := os.ReadFile(chunkPath)
			if err != nil {
				return ctx.CtxError(c, 5001, "读取文件失败", err.Error())
			}
			zipData = append(zipData, chunkData...)
		}

		// 删除文件夹 chunkDir
		_ = os.RemoveAll(chunkDir)

		zipPath := pathUtil.Join(runtimePath, "temp", fmt.Sprintf("%s.zip", req.ChunkTaskId))
		if err := os.WriteFile(zipPath, zipData, 0644); err != nil {
			return ctx.CtxError(c, 5001, "保存文件失败", err.Error())
		}

		// 创建处理任务
		UserId := c.Locals("UserId").(uint32)
		taskId, err := utils.HandleMarkupZipFile(zipPath, version, UserId, req.Deduplicate)
		if err != nil {
			return ctx.CtxError(c, 5001, "文件处理失败", err.Error())
		}

		// 获取文件MD5
		return ctx.CtxSuccess(c, fiber.Map{
			"status":  1,
			"task_id": taskId,
		}, "ZIP上传成功，正在处理")
	} else {
		// 获取文件MD5
		return ctx.CtxSuccess(c, fiber.Map{
			"status":  0,
			"task_id": req.ChunkTaskId,
		}, "分片上传中")
	}
}
