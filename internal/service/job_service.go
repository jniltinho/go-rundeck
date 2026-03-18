package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
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
	keySvc   *KeyService
}

// NewJobService creates a new JobService.
func NewJobService(
	jobRepo *repository.JobRepository,
	nodeRepo *repository.NodeRepository,
	execSvc *ExecutionService,
	sshSvc *SSHService,
	keySvc *KeyService,
) *JobService {
	return &JobService{
		jobRepo:  jobRepo,
		nodeRepo: nodeRepo,
		execSvc:  execSvc,
		sshSvc:   sshSvc,
		keySvc:   keySvc,
	}
}

// Create validates and persists a new job and its initial steps/options.
func (s *JobService) Create(j *model.Job, steps []model.JobStep, opts []model.JobOption) (*model.Job, error) {
	if j.Name == "" {
		return nil, errors.New("job name is required")
	}
	for i, step := range steps {
		if strings.TrimSpace(step.Content) == "" {
			return nil, fmt.Errorf("step %d has no command/script content", i+1)
		}
	}
	if err := s.jobRepo.Create(j); err != nil {
		return nil, err
	}
	
	if len(steps) > 0 {
		_ = s.jobRepo.ReplaceSteps(j.ID, steps)
	}
	if len(opts) > 0 {
		_ = s.jobRepo.ReplaceOptions(j.ID, opts)
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

// Update saves changes to a job and its steps/options.
func (s *JobService) Update(j *model.Job, steps []model.JobStep, opts []model.JobOption) error {
	for i, step := range steps {
		if strings.TrimSpace(step.Content) == "" {
			return fmt.Errorf("step %d has no command/script content", i+1)
		}
	}
	if err := s.jobRepo.Update(j); err != nil {
		return err
	}
	if err := s.jobRepo.ReplaceSteps(j.ID, steps); err != nil {
		return err
	}
	return s.jobRepo.ReplaceOptions(j.ID, opts)
}

// Delete soft-deletes a job.
func (s *JobService) Delete(id uint) error {
	return s.jobRepo.Delete(id)
}

// Run creates an execution and dispatches it asynchronously.
func (s *JobService) Run(jobID uint, triggeredBy *uint, triggerType model.TriggerType, opts map[string]string) (*model.Execution, error) {
	job, err := s.jobRepo.GetByID(jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	finalOpts := make(map[string]string)
	for _, opt := range job.Options {
		val, provided := opts[opt.Name]
		if (!provided || val == "") && opt.DefaultVal != "" {
			val = opt.DefaultVal
		}
		if opt.Required && val == "" {
			return nil, fmt.Errorf("required option '%s' is missing", opt.Name)
		}
		finalOpts[opt.Name] = val
	}
	// Copy any extra options provided that weren't in job definition (optional mapping)
	for k, v := range opts {
		if _, exists := finalOpts[k]; !exists {
			finalOpts[k] = v
		}
	}

	exec, err := s.execSvc.Create(jobID, job.ProjectID, triggerType, triggeredBy, finalOpts)
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

	var hasFailed bool
	var mu sync.Mutex

	runOnNode := func(node model.Node) {
		nodeFailed := false
		for _, step := range job.Steps {
			label := step.Label
			if label == "" {
				label = fmt.Sprintf("Step %d", step.StepOrder)
			}
			msg := fmt.Sprintf("[%s] executing: %s", label, step.Content)
			_ = s.execSvc.AddLog(exec.ID, node.Name, step.StepOrder, model.LogLevelInfo, msg)

			result, err := s.runStep(node, step, exec.Options)
			if err != nil {
				_ = s.execSvc.AddLog(exec.ID, node.Name, step.StepOrder, model.LogLevelError,
					fmt.Sprintf("[%s] error: %s", label, err.Error()))
				nodeFailed = true
				if job.OnError == model.OnErrorStop {
					break
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
				nodeFailed = true
				if job.OnError == model.OnErrorStop {
					break
				}
			}
		}

		if nodeFailed {
			mu.Lock()
			hasFailed = true
			mu.Unlock()
		}
	}

	if job.ExecStrategy == model.ExecStrategyParallel {
		var wg sync.WaitGroup
		for _, node := range nodes {
			wg.Add(1)
			go func(n model.Node) {
				defer wg.Done()
				runOnNode(n)
			}(node)
		}
		wg.Wait()
	} else {
		for _, node := range nodes {
			runOnNode(node)
			if hasFailed && job.OnError == model.OnErrorStop {
				break
			}
		}
	}

	status := model.ExecutionStatusSucceeded
	if hasFailed {
		status = model.ExecutionStatusFailed
	}
	_ = s.execSvc.UpdateStatus(exec.ID, status)
}

func (s *JobService) runStep(node model.Node, step model.JobStep, execOpts []model.ExecutionOption) (*SSHResult, error) {
	cmd := step.Content

	// Interpolate variables
	cmd = strings.ReplaceAll(cmd, "${node.name}", node.Name)
	cmd = strings.ReplaceAll(cmd, "${node.hostname}", node.Hostname)
	cmd = strings.ReplaceAll(cmd, "${node.os_family}", node.OSFamily)
	cmd = strings.ReplaceAll(cmd, "${node.tags}", node.Tags)

	for _, opt := range execOpts {
		cmd = strings.ReplaceAll(cmd, fmt.Sprintf("${option.%s}", opt.OptionName), opt.Value)
	}

	if step.Type == model.StepTypeScript {
		interp := step.Interpreter
		if interp == "" {
			interp = "/bin/sh"
		}
		// Upload script inline via heredoc.
		cmd = fmt.Sprintf("%s -s << 'GORUNDECK_EOF'\n%s\nGORUNDECK_EOF", interp, cmd)
	}
	if step.Args != "" {
		args := step.Args
		for _, opt := range execOpts {
			args = strings.ReplaceAll(args, fmt.Sprintf("${option.%s}", opt.OptionName), opt.Value)
		}
		cmd = cmd + " " + args
	}

	timeout := 300 * time.Second

	_ = timeout // used for future implementation with context

	if node.AuthType == model.AuthTypePassword {
		password := ""
		if node.KeyID != nil {
			password, _ = s.keySvc.GetDecryptedContent(*node.KeyID)
		}
		return s.sshSvc.RunCommandWithPassword(node.Hostname, node.SSHPort, node.SSHUser, password, cmd)
	}
	
	// Key auth
	if node.KeyID != nil {
		pemKeyStr, err := s.keySvc.GetDecryptedContent(*node.KeyID)
		if err == nil && pemKeyStr != "" {
			return s.sshSvc.RunCommandWithKey(node.Hostname, node.SSHPort, node.SSHUser, []byte(pemKeyStr), cmd)
		}
	}
	
	return nil, fmt.Errorf("no valid credentials found or key could not be decrypted for node %s", node.Name)
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
