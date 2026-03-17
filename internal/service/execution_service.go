package service

import (
	"errors"
	"sync"
	"time"

	"go-rundeck/internal/model"
	"go-rundeck/internal/repository"
)

// LogEvent is pushed over SSE channels.
type LogEvent struct {
	ExecutionID uint
	Log         model.ExecutionLog
}

// ExecutionService manages execution lifecycle and SSE streaming.
type ExecutionService struct {
	repo     *repository.ExecutionRepository
	mu       sync.RWMutex
	channels map[uint][]chan LogEvent // executionID -> list of subscriber channels
}

// NewExecutionService creates a new ExecutionService.
func NewExecutionService(repo *repository.ExecutionRepository) *ExecutionService {
	return &ExecutionService{
		repo:     repo,
		channels: make(map[uint][]chan LogEvent),
	}
}

// Create persists a new execution record.
func (s *ExecutionService) Create(jobID, projectID uint, triggerType model.TriggerType, triggeredBy *uint) (*model.Execution, error) {
	now := time.Now()
	e := &model.Execution{
		JobID:       jobID,
		ProjectID:   projectID,
		Status:      model.ExecutionStatusRunning,
		TriggeredBy: triggeredBy,
		TriggerType: triggerType,
		StartedAt:   now,
		CreatedAt:   now,
	}
	if err := s.repo.Create(e); err != nil {
		return nil, err
	}
	return e, nil
}

// GetByID retrieves an execution.
func (s *ExecutionService) GetByID(id uint) (*model.Execution, error) {
	return s.repo.GetByID(id)
}

// List returns paginated executions for a project.
func (s *ExecutionService) List(projectID uint, limit, offset int) ([]model.Execution, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListByProject(projectID, limit, offset)
}

// ListByJob returns recent executions for a job.
func (s *ExecutionService) ListByJob(jobID uint, limit int) ([]model.Execution, error) {
	return s.repo.ListByJob(jobID, limit)
}

// GetLogs returns all log entries for an execution.
func (s *ExecutionService) GetLogs(executionID uint) ([]model.ExecutionLog, error) {
	return s.repo.GetLogs(executionID)
}

// AddLog persists a log entry and fans it out to any SSE subscribers.
func (s *ExecutionService) AddLog(executionID uint, nodeName string, stepOrder int, level model.LogLevel, message string) error {
	log := &model.ExecutionLog{
		ExecutionID: executionID,
		NodeName:    nodeName,
		StepOrder:   stepOrder,
		LogLevel:    level,
		Message:     message,
		LoggedAt:    time.Now(),
	}
	if err := s.repo.AddLog(log); err != nil {
		return err
	}

	s.mu.RLock()
	subs := s.channels[executionID]
	s.mu.RUnlock()

	event := LogEvent{ExecutionID: executionID, Log: *log}
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
	return nil
}

// UpdateStatus changes execution status and records end time.
func (s *ExecutionService) UpdateStatus(id uint, status model.ExecutionStatus) error {
	e, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	e.Status = status
	if status != model.ExecutionStatusRunning {
		now := time.Now()
		e.EndedAt = &now
		dur := now.Sub(e.StartedAt).Seconds()
		e.DurationSec = &dur
	}
	return s.repo.Update(e)
}

// Abort marks an execution as aborted.
func (s *ExecutionService) Abort(id uint) error {
	e, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if e.Status != model.ExecutionStatusRunning {
		return errors.New("execution is not running")
	}
	return s.UpdateStatus(id, model.ExecutionStatusAborted)
}

// Subscribe returns a channel that receives log events for an execution.
// The caller must call Unsubscribe when done.
func (s *ExecutionService) Subscribe(executionID uint) chan LogEvent {
	ch := make(chan LogEvent, 100)
	s.mu.Lock()
	s.channels[executionID] = append(s.channels[executionID], ch)
	s.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (s *ExecutionService) Unsubscribe(executionID uint, ch chan LogEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs := s.channels[executionID]
	for i, c := range subs {
		if c == ch {
			s.channels[executionID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(s.channels[executionID]) == 0 {
		delete(s.channels, executionID)
	}
	close(ch)
}

// RecentActivity returns the most recent executions.
func (s *ExecutionService) RecentActivity(limit int) ([]model.Execution, error) {
	return s.repo.RecentActivity(limit)
}

// CountRunning returns the number of running executions.
func (s *ExecutionService) CountRunning() (int64, error) {
	return s.repo.CountRunning()
}
