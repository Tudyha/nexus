package dao

import (
	"context"

	"github.com/Tudyha/nexus/internal/model"
	"gorm.io/gorm"
)

type taskDao struct {
	db *gorm.DB
}

type taskExecutionDao struct {
	db *gorm.DB
}

func newTaskDao(db *gorm.DB) TaskDao {
	return &taskDao{db: db}
}

func newTaskExecutionDao(db *gorm.DB) TaskExecutionDao {
	return &taskExecutionDao{db: db}
}

func (d *taskDao) Create(ctx context.Context, task *model.Task) error {
	return d.db.WithContext(ctx).Create(task).Error
}

func (d *taskDao) GetByID(ctx context.Context, id uint64) (*model.Task, error) {
	var task model.Task
	return &task, d.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
}

func (d *taskExecutionDao) Create(ctx context.Context, execs []*model.TaskExecution) error {
	return d.db.WithContext(ctx).Create(execs).Error
}

func (d *taskExecutionDao) Update(ctx context.Context, exec *model.TaskExecution) error {
	return d.db.WithContext(ctx).Updates(exec).Error
}

func (d *taskExecutionDao) GetByID(ctx context.Context, id uint64) (*model.TaskExecution, error) {
	var exec model.TaskExecution
	return &exec, d.db.WithContext(ctx).Where("id = ?", id).First(&exec).Error
}

func (d *taskExecutionDao) GetLatestByClientID(ctx context.Context, clientID uint64) (*model.TaskExecution, error) {
	var exec model.TaskExecution
	return &exec, d.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("id desc").
		First(&exec).Error
}

func (d *taskExecutionDao) GetLatestByClientIDs(ctx context.Context, clientIDs []uint64) (map[uint64]*model.TaskExecution, error) {
	if len(clientIDs) == 0 {
		return nil, nil
	}

	// 子查询：每个 client_id 的最大 id
	subQuery := d.db.WithContext(ctx).
		Model(&model.TaskExecution{}).
		Select("MAX(id) as max_id").
		Where("client_id IN ?", clientIDs).
		Group("client_id")

	var executions []*model.TaskExecution
	if err := d.db.WithContext(ctx).
		Model(&model.TaskExecution{}).
		Where("id IN (?)", subQuery).
		Find(&executions).Error; err != nil {
		return nil, err
	}

	result := make(map[uint64]*model.TaskExecution, len(executions))
	for _, exec := range executions {
		result[exec.ClientID] = exec
	}
	return result, nil
}
