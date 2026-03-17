package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go-rundeck/internal/model"
	"go-rundeck/internal/repository"
)

// JobService provides business logic for jobs.
type JobService struct {
	jobRepo  *repository.JobRepository
	nodeRepo *repository.NodeRepository
	execSvc  *ExecutionService
	sshSvc   *SSHService
}

// NewJobService creates a new JobService.
func NewJobService(
	jobRepo *repository.JobRepository,
	nodeRepo *repository.NodeRepository,
	execSvc *ExecutionService,
	sshSvc *SSHService,
) *JobService {
	return &JobService{
		jobRepo:  jobRepo,
		nodeRepo: nodeRepo,
		execSvc:  execSvc,
		sshSvc:   sshSvc,
	}
}

// Create validates and persists a new job.
func (s *JobService) Create(j *model.Job) (*model.Job, error) {
	if j.Name == "" {
		return nil, errors.New("job name is required")
	}
	if err := s.jobRepo.Create(j); err != nil {
		return nil, err
	}
	return j, nil
}

// GetByID retrieves a job with its steps.
func (s *JobService) GetByID(id uint) (*model.Job, error) {
	return s.jobRepo.GetByID(id)
}

// ListByProject returns all jobs for a project.
func (s *JobService) ListByProject(projectID uint) ([]model.Job, error) {
	return s.jobRepo.ListByProject(projectID)
}

// Update saves changes to a job.
func (s *JobService) Update(j *model.Job) error {
	return s.jobRepo.Update(j)
}

// Delete soft-deletes a job.
func (s *JobService) Delete(id uint) error {
	return s.jobRepo.Delete(id)
}

// Run creates an execution and dispatches it asynchronously.
func (s *JobService) Run(jobID uint, triggeredBy *uint, triggerType model.TriggerType) (*model.Execution, error) {
	job, err := s.jobRepo.GetByID(jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	exec, err := s.execSvc.Create(jobID, job.ProjectID, triggerType, triggeredBy)
	if err != nil {
		return nil, fmt.Errorf("create execution: %w", err)
	}

	go s.dispatch(job, exec)
	return exec, nil
}

// dispatch runs the job steps against matched nodes in a goroutine.
func (s *JobService) dispatch(job *model.Job, exec *model.Execution) {
	nodes, err := s.resolveNodes(job)
	if err != nil {
		_ = s.execSvc.AddLog(exec.ID, "scheduler", 0, model.LogLevelError, "node resolution failed: "+err.Error())
		_ = s.execSvc.UpdateStatus(exec.ID, model.ExecutionStatusFailed)
		return
	}

	if len(nodes) == 0 {
		_ = s.execSvc.AddLog(exec.ID, "scheduler", 0, model.LogLevelWarn, "no nodes matched filter: "+job.NodeFilter)
		_ = s.execSvc.UpdateStatus(exec.ID, model.ExecutionStatusFailed)
		return
	}

	failed := false
	for _, node := range nodes {
		for _, step := range job.Steps {
			label := step.Label
			if label == "" {
				label = fmt.Sprintf("Step %d", step.StepOrder)
			}
			_ = s.execSvc.AddLog(exec.ID, node.Name, step.StepOrder, model.LogLevelInfo,
				fmt.Sprintf("[%s] executing: %s", label, step.Content))

			result, err := s.runStep(node, step)
			if err != nil {
				_ = s.execSvc.AddLog(exec.ID, node.Name, step.StepOrder, model.LogLevelError,
					fmt.Sprintf("[%s] error: %s", label, err.Error()))
				failed = true
				if job.OnError == model.OnErrorStop {
					_ = s.execSvc.UpdateStatus(exec.ID, model.ExecutionStatusFailed)
					return
				}
				continue
			}

			if result.Stdout != "" {
				_ = s.execSvc.AddLog(exec.ID, node.Name, step.StepOrder, model.LogLevelInfo, result.Stdout)
			}
			if result.Stderr != "" {
				_ = s.execSvc.AddLog(exec.ID, node.Name, step.StepOrder, model.LogLevelWarn, result.Stderr)
			}
			if result.ExitCode != 0 {
				_ = s.execSvc.AddLog(exec.ID, node.Name, step.StepOrder, model.LogLevelError,
					fmt.Sprintf("[%s] exit code %d", label, result.ExitCode))
				failed = true
				if job.OnError == model.OnErrorStop {
					_ = s.execSvc.UpdateStatus(exec.ID, model.ExecutionStatusFailed)
					return
				}
			}
		}
	}

	status := model.ExecutionStatusSucceeded
	if failed {
		status = model.ExecutionStatusFailed
	}
	_ = s.execSvc.UpdateStatus(exec.ID, status)
}

func (s *JobService) runStep(node model.Node, step model.JobStep) (*SSHResult, error) {
	cmd := step.Content
	if step.Type == model.StepTypeScript {
		interp := step.Interpreter
		if interp == "" {
			interp = "/bin/sh"
		}
		// Upload script inline via heredoc.
		cmd = fmt.Sprintf("%s -s << 'GORUNDECK_EOF'\n%s\nGORUNDECK_EOF", interp, step.Content)
	}
	if step.Args != "" {
		cmd = cmd + " " + step.Args
	}

	timeout := 300 * time.Second

	_ = timeout // used for future implementation with context

	if node.AuthType == model.AuthTypePassword {
		return s.sshSvc.RunCommandWithPassword(node.Hostname, node.SSHPort, node.SSHUser, "", cmd)
	}
	// Key auth would decrypt the key from KeyStorage - simplified here
	return s.sshSvc.RunCommandWithPassword(node.Hostname, node.SSHPort, node.SSHUser, "", cmd)
}

// resolveNodes returns nodes matching the job's NodeFilter.
func (s *JobService) resolveNodes(job *model.Job) ([]model.Node, error) {
	filter := strings.TrimSpace(job.NodeFilter)
	if filter == "" || filter == "*" {
		return s.nodeRepo.ListByProject(job.ProjectID)
	}

	// Support "tag:value" or plain name filter
	if strings.HasPrefix(filter, "tag:") {
		tag := strings.TrimPrefix(filter, "tag:")
		return s.nodeRepo.FindByTags(job.ProjectID, []string{tag})
	}

	// Filter by name
	all, err := s.nodeRepo.ListByProject(job.ProjectID)
	if err != nil {
		return nil, err
	}
	var matched []model.Node
	for _, n := range all {
		if strings.Contains(strings.ToLower(n.Name), strings.ToLower(filter)) {
			matched = append(matched, n)
		}
	}
	return matched, nil
}
