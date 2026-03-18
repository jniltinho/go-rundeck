package repository

import (
	"go-rundeck/internal/model"

	"gorm.io/gorm"
)

// ExecutionRepository handles persistence for Execution and ExecutionLog entities.
type ExecutionRepository struct {
	db *gorm.DB
}

// NewExecutionRepository creates a new ExecutionRepository.
func NewExecutionRepository(db *gorm.DB) *ExecutionRepository {
	return &ExecutionRepository{db: db}
}

// Create inserts a new execution record.
func (r *ExecutionRepository) Create(e *model.Execution) error {
	return r.db.Create(e).Error
}

// GetByID retrieves an execution by its primary key.
func (r *ExecutionRepository) GetByID(id uint) (*model.Execution, error) {
	var e model.Execution
	err := r.db.Preload("Job").Preload("Options").First(&e, id).Error
	return &e, err
}

// ListByProject returns paginated executions for a project, newest first.
func (r *ExecutionRepository) ListByProject(projectID uint, limit, offset int) ([]model.Execution, error) {
	var execs []model.Execution
	err := r.db.Where("project_id = ?", projectID).
		Preload("Job").
		Order("created_at desc").
		Limit(limit).Offset(offset).
		Find(&execs).Error
	return execs, err
}

// ListByJob returns executions for a specific job.
func (r *ExecutionRepository) ListByJob(jobID uint, limit int) ([]model.Execution, error) {
	var execs []model.Execution
	err := r.db.Where("job_id = ?", jobID).
		Order("created_at desc").
		Limit(limit).Find(&execs).Error
	return execs, err
}

// Update saves status changes to an execution.
func (r *ExecutionRepository) Update(e *model.Execution) error {
	return r.db.Save(e).Error
}

// AddLog appends a log entry for an execution.
func (r *ExecutionRepository) AddLog(l *model.ExecutionLog) error {
	return r.db.Create(l).Error
}

// GetLogs retrieves all log entries for an execution, ordered by time.
func (r *ExecutionRepository) GetLogs(executionID uint) ([]model.ExecutionLog, error) {
	var logs []model.ExecutionLog
	err := r.db.Where("execution_id = ?", executionID).
		Order("logged_at asc, id asc").
		Find(&logs).Error
	return logs, err
}

// CountByProject returns the total number of executions for a project.
func (r *ExecutionRepository) CountByProject(projectID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Execution{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

// CountRunning returns the number of currently running executions.
func (r *ExecutionRepository) CountRunning() (int64, error) {
	var count int64
	err := r.db.Model(&model.Execution{}).
		Where("status = ?", model.ExecutionStatusRunning).Count(&count).Error
	return count, err
}

// CountLastDay returns the total executions in the last 24 hours.
func (r *ExecutionRepository) CountLastDay() (int64, error) {
	var count int64
	err := r.db.Model(&model.Execution{}).
		Where("created_at >= NOW() - INTERVAL 1 DAY").Count(&count).Error
	return count, err
}

// CountFailedLastDay returns the failed executions in the last 24 hours.
func (r *ExecutionRepository) CountFailedLastDay() (int64, error) {
	var count int64
	err := r.db.Model(&model.Execution{}).
		Where("created_at >= NOW() - INTERVAL 1 DAY AND status = ?", model.ExecutionStatusFailed).Count(&count).Error
	return count, err
}

// Delete removes an execution and its associated logs and options (cascade).
func (r *ExecutionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Execution{}, id).Error
}

// RecentActivity returns the most recent executions across all projects.
func (r *ExecutionRepository) RecentActivity(limit int) ([]model.Execution, error) {
	var execs []model.Execution
	err := r.db.Preload("Job").
		Order("created_at desc").
		Limit(limit).Find(&execs).Error
	return execs, err
}
