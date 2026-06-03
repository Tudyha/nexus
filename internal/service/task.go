package service

import (
	"context"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/proto"
)

type taskService struct {
	taskDao          dao.TaskDao
	taskExecutionDao dao.TaskExecutionDao
}

func newTaskService() TaskService {
	return &taskService{
		taskDao:          dao.GetTaskDao(),
		taskExecutionDao: dao.GetTaskExecutionDao(),
	}
}

func (s *taskService) CreateTask(ctx context.Context, taskType int32, clientIDs []uint64) ([]*model.TaskExecution, error) {
	task := &model.Task{
		TaskType: taskType,
	}
	if err := s.taskDao.Create(ctx, task); err != nil {
		return nil, err
	}
	var executions []*model.TaskExecution
	for _, clientID := range clientIDs {
		exec := &model.TaskExecution{
			TaskID:   task.ID,
			ClientID: clientID,
			TaskType: taskType,
			Status:   enum.TaskStatusPending,
		}
		executions = append(executions, exec)
	}
	if err := s.taskExecutionDao.Create(ctx, executions); err != nil {
		return nil, err
	}
	return executions, nil
}

func (s *taskService) UpdateExecution(ctx context.Context, exec *model.TaskExecution) error {
	return s.taskExecutionDao.Update(ctx, exec)
}

func (s *taskService) GetExecutionByID(ctx context.Context, id uint64) (*model.TaskExecution, error) {
	return s.taskExecutionDao.GetByID(ctx, id)
}

func (s *taskService) GetLatestByClientID(ctx context.Context, clientID uint64) (*model.TaskExecution, error) {
	return s.taskExecutionDao.GetLatestByClientID(ctx, clientID)
}

func (s *taskService) GetLatestByClientIDs(ctx context.Context, clientIDs []uint64) (map[uint64]*model.TaskExecution, error) {
	return s.taskExecutionDao.GetLatestByClientIDs(ctx, clientIDs)
}

// UpdateProgress 更新任务执行进度
func (s *taskService) UpdateProgress(ctx context.Context, execID uint64, p *proto.TaskProgress) error {
	update := &model.TaskExecution{
		Progress: p.Progress,
		Message:  p.Message,
	}
	update.ID = execID
	if p.Done {
		if p.Success {
			update.Status = enum.TaskStatusSuccess // done
		} else {
			update.Status = enum.TaskStatusFailed // failed
			update.Error = p.Error
		}
	}
	return s.taskExecutionDao.Update(ctx, update)
}
