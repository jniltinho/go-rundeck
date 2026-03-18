package repository

import (
	"go-rundeck/internal/model"

	"gorm.io/gorm"
)

// JobRepository handles persistence for Job and JobStep entities.
type JobRepository struct {
	db *gorm.DB
}

// NewJobRepository creates a new JobRepository.
func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Create inserts a new job with its steps.
func (r *JobRepository) Create(j *model.Job) error {
	return r.db.Create(j).Error
}

// GetByID retrieves a job by its primary key, preloading steps and options.
func (r *JobRepository) GetByID(id uint) (*model.Job, error) {
	var j model.Job
	err := r.db.Preload("Options").Preload("Schedules").Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order asc")
	}).Preload("Project").First(&j, id).Error
	return &j, err
}

// ListByProject returns all jobs for a project.
func (r *JobRepository) ListByProject(projectID uint) ([]model.Job, error) {
	var jobs []model.Job
	err := r.db.Where("project_id = ?", projectID).
		Order("name asc").Find(&jobs).Error
	return jobs, err
}

// Update saves changes to an existing job.
func (r *JobRepository) Update(j *model.Job) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(j).Error
}

// Delete soft-deletes a job.
func (r *JobRepository) Delete(id uint) error {
	return r.db.Delete(&model.Job{}, id).Error
}

// ReplaceSteps removes all existing steps and inserts new ones.
func (r *JobRepository) ReplaceSteps(jobID uint, steps []model.JobStep) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("job_id = ?", jobID).Delete(&model.JobStep{}).Error; err != nil {
			return err
		}
		if len(steps) > 0 {
			for i := range steps {
				steps[i].JobID = jobID
			}
			return tx.Create(&steps).Error
		}
		return nil
	})
}

// ReplaceOptions removes all existing options and inserts new ones.
func (r *JobRepository) ReplaceOptions(jobID uint, opts []model.JobOption) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("job_id = ?", jobID).Delete(&model.JobOption{}).Error; err != nil {
			return err
		}
		if len(opts) > 0 {
			for i := range opts {
				opts[i].JobID = jobID
			}
			return tx.Create(&opts).Error
		}
		return nil
	})
}

// CountByProject returns the total number of jobs for a project.
func (r *JobRepository) CountByProject(projectID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Job{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}
