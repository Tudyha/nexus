package v1

import (
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
)

type TaskController struct {
	taskService service.TaskService
}

func newTaskController() *TaskController {
	return &TaskController{
		taskService: service.GetTaskService(),
	}
}

func (h *TaskController) Create(ctx *gin.Context) {
	var req struct {
		TaskType  int32    `json:"task_type" binding:"required,oneof=1 2"`
		ClientIDs []uint64 `json:"client_ids" binding:"required,min=1"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, nil, "参数错误: "+err.Error())
		return
	}
	execs, err := h.taskService.CreateTask(ctx, req.TaskType, req.ClientIDs)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	var list []response.TaskExecutionResponse
	for _, e := range execs {
		list = append(list, response.TaskExecutionResponse{
			ID:        e.ID,
			TaskID:    e.TaskID,
			ClientID:  e.ClientID,
			TaskType:  e.TaskType,
			Status:    e.Status,
			Progress:  e.Progress,
			Message:   e.Message,
			Error:     e.Error,
			CreatedAt: e.CreatedAt,
			UpdatedAt: e.UpdatedAt,
		})
	}
	response.Success(ctx, list)
}

// GetExecution 获取任务执行记录
func (h *TaskController) GetExecution(ctx *gin.Context) {
	taskID := utils.StringToUint64(ctx.Param("taskId"))
	clientID := utils.StringToUint64(ctx.Param("clientId"))
	if taskID == 0 || clientID == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}

	// 通过 taskId + clientId 查询 execution
	// 先用 GetLatestByClientID 获取最新执行记录，验证匹配
	exec, err := h.taskService.GetLatestByClientID(ctx, clientID)
	if err != nil {
		response.Success(ctx, nil)
		return
	}
	if exec == nil || exec.TaskID != taskID {
		response.Success(ctx, nil)
		return
	}

	response.Success(ctx, response.TaskExecutionResponse{
		ID:        exec.ID,
		TaskID:    exec.TaskID,
		ClientID:  exec.ClientID,
		TaskType:  exec.TaskType,
		Status:    exec.Status,
		Progress:  exec.Progress,
		Message:   exec.Message,
		Error:     exec.Error,
		CreatedAt: exec.CreatedAt,
		UpdatedAt: exec.UpdatedAt,
	})
}
