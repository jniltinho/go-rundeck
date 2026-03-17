package repository

import (
	"go-rundeck/internal/model"

	"gorm.io/gorm"
)

// ProjectRepository handles persistence for Project entities.
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new ProjectRepository.
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create inserts a new project.
func (r *ProjectRepository) Create(p *model.Project) error {
	return r.db.Create(p).Error
}

// GetByID retrieves a project by its primary key.
func (r *ProjectRepository) GetByID(id uint) (*model.Project, error) {
	var p model.Project
	err := r.db.First(&p, id).Error
	return &p, err
}

// List returns all active projects, newest first.
func (r *ProjectRepository) List() ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Where("active = ?", true).Order("created_at desc").Find(&projects).Error
	return projects, err
}

// Update saves changes to an existing project.
func (r *ProjectRepository) Update(p *model.Project) error {
	return r.db.Save(p).Error
}

// Delete soft-deletes a project.
func (r *ProjectRepository) Delete(id uint) error {
	return r.db.Delete(&model.Project{}, id).Error
}

// Count returns the total number of active projects.
func (r *ProjectRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Project{}).Where("active = ?", true).Count(&count).Error
	return count, err
}
