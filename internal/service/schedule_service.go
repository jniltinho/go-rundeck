package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go-rundeck/internal/model"

	"gorm.io/gorm"
)

// ScheduleService checks enabled schedules and triggers job runs.
type ScheduleService struct {
	db      *gorm.DB
	jobSvc  *JobService
	enabled bool
	ticker  *time.Ticker
	done    chan struct{}
}

// NewScheduleService creates a new ScheduleService.
func NewScheduleService(db *gorm.DB, jobSvc *JobService, enabled bool, checkIntervalSec int) *ScheduleService {
	if checkIntervalSec <= 0 {
		checkIntervalSec = 30
	}
	return &ScheduleService{
		db:      db,
		jobSvc:  jobSvc,
		enabled: enabled,
		done:    make(chan struct{}),
	}
}

// Start begins the schedule checker loop.
func (s *ScheduleService) Start(checkIntervalSec int) {
	if !s.enabled {
		log.Println("[scheduler] disabled")
		return
	}
	if checkIntervalSec <= 0 {
		checkIntervalSec = 30
	}
	s.ticker = time.NewTicker(time.Duration(checkIntervalSec) * time.Second)
	go s.loop()
	log.Printf("[scheduler] started, check interval: %ds", checkIntervalSec)
}

// Stop shuts down the scheduler loop.
func (s *ScheduleService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.done)
}

func (s *ScheduleService) loop() {
	for {
		select {
		case <-s.done:
			return
		case t := <-s.ticker.C:
			s.checkSchedules(t)
		}
	}
}

func (s *ScheduleService) checkSchedules(now time.Time) {
	var schedules []model.Schedule
	if err := s.db.Where("enabled = ? AND next_run <= ?", true, now).Find(&schedules).Error; err != nil {
		log.Printf("[scheduler] query error: %v", err)
		return
	}

	for _, sched := range schedules {
		go func(sc model.Schedule) {
			if _, err := s.jobSvc.Run(sc.JobID, nil, model.TriggerTypeSchedule); err != nil {
				log.Printf("[scheduler] failed to run job %d: %v", sc.JobID, err)
			}

			next := nextCronRun(sc.CronExpr, now)
			sc.LastRun = &now
			sc.NextRun = next
			if err := s.db.Save(&sc).Error; err != nil {
				log.Printf("[scheduler] save schedule error: %v", err)
			}
		}(sched)
	}
}

// nextCronRun is a simplified cron parser that handles "*/N" minute intervals.
// For production use, replace with a proper cron library.
func nextCronRun(expr string, from time.Time) *time.Time {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		t := from.Add(5 * time.Minute)
		return &t
	}

	// Handle simple "*/N" minute expressions
	minPart := parts[0]
	if strings.HasPrefix(minPart, "*/") {
		var n int
		if _, err := fmt.Sscanf(minPart, "*/%d", &n); err == nil && n > 0 {
			t := from.Add(time.Duration(n) * time.Minute)
			return &t
		}
	}

	// Default: next minute boundary
	t := from.Truncate(time.Minute).Add(time.Minute)
	return &t
}
